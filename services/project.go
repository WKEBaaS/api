package services

import (
	"baas-api/config"
	"baas-api/dto"
	"baas-api/models"
	"baas-api/repo"
	"baas-api/services/kube_project"
	"context"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/lo"
)

type ProjectServiceInterface interface {
	CreateProject(ctx context.Context, in *dto.CreateProjectInput, userID *string) (*dto.CreateProjectOutput, error)
	DeleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput, userID string) (*dto.DeleteProjectByRefOutput, error)
	PatchProjectSettings(ctx context.Context, in *dto.PatchProjectSettingInput, userID string) error
	GetUsersProjects(ctx context.Context, userID string) ([]*models.ProjectView, error)
	GetUserProjectByRef(ctx context.Context, ref, userID string) (*models.ProjectView, error)
	GetUserProjectStatusByRef(ctx context.Context, c chan any, ref, userID string) error
	GetProjectSettings(ctx context.Context, in *dto.GetProjectSettingsInput, userID string) (*dto.GetProjectSettingsOutput, error)
	ResetDatabasePassword(ctx context.Context, in *dto.ResetDatabasePasswordInput, userID string) (*dto.ResetDatabasePasswordOutput, error)
}

type ProjectService struct {
	config             *config.Config                             `di.inject:"config"`
	kube               kube_project.KubeProjectServiceInterface   `di.inject:"kubeProjectService"`
	entity             repo.EntityRepositoryInterface             `di.inject:"entityRepository"`
	project            repo.ProjectRepositoryInterface            `di.inject:"projectRepository"`
	projectAuthSetting repo.ProjectAuthSettingRepositoryInterface `di.inject:"projectAuthSettingRepository"`
}

func NewProjectService() ProjectServiceInterface {
	service := &ProjectService{}
	return service
}

func (s *ProjectService) CreateProject(ctx context.Context, in *dto.CreateProjectInput, userID *string) (*dto.CreateProjectOutput, error) {
	projectEntity, err := s.entity.GetByChineseName(ctx, "專案")
	if err != nil {
		return nil, err
	}

	// Cleanup if any error occurs
	var cleanupFuncs []func()
	var success bool
	defer func() {
		if success {
			cleanupFuncs = nil // Clear cleanup functions if successful
			return
		}

		for i := len(cleanupFuncs) - 1; i >= 0; i-- {
			cleanupFuncs[i]()
		}
	}()

	id, ref, err := s.project.Create(ctx, in.Body.Name, in.Body.Description, projectEntity.ID, userID)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.project.DeleteByID(ctx, *id)
	})

	projectAuthSetting := &models.ProjectSettings{
		ProjectID: *id,
	}
	err = s.projectAuthSetting.Create(ctx, projectAuthSetting)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.projectAuthSetting.DeleteByProjectID(ctx, *id)
	})

	err = s.kube.CreateCluster(ctx, *ref, in.Body.StorageSize)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.kube.DeleteCluster(ctx, *ref)
	})

	err = s.kube.CreateDatabase(ctx, s.config.Kube.Project.Namespace, *ref)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.kube.DeleteDatabase(ctx, s.config.Kube.Project.Namespace, *ref)
	})

	// Create default email/password provider
	emailPasswordProvider := &models.ProjectAuthProvider{
		Name:         "email",
		ProjectID:    *id,
		Enabled:      true,
		ClientID:     "",
		ClientSecret: "",
	}
	_, err = s.projectAuthSetting.CreateOAuthProvider(ctx, emailPasswordProvider)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.projectAuthSetting.DeleteByProjectID(ctx, *id) // This will cascade delete OAuth providers
	})

	err = s.kube.CreateAuthAPIDeployment(ctx,
		*ref,
		kube_project.NewAPIDeploymentOption().
			WithBetterAuthSecret(projectAuthSetting.Secret).
			WithEmailAndPasswordAuth(true),
	)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.kube.DeleteAuthAPIDeployment(ctx, *ref)
	})

	err = s.kube.CreateAuthAPIService(ctx, *ref)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.kube.DeleteAuthAPIService(ctx, *ref)
	})

	err = s.kube.CreateIngressRoute(ctx, *ref)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.kube.DeleteIngressRoute(ctx, *ref)
	})

	err = s.kube.CreateIngressRouteTCP(ctx, *ref)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.kube.DeleteIngressRouteTCP(ctx, *ref)
	})

	success = true // Mark as successful if we reach here without errors
	out := &dto.CreateProjectOutput{}
	out.Body.ID = *id
	out.Body.Reference = *ref

	return out, nil
}

func (s *ProjectService) DeleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput, userID string) (*dto.DeleteProjectByRefOutput, error) {
	project, err := s.project.FindByRef(ctx, in.Reference)
	if err != nil {
		return nil, err
	}
	if project.OwnerID != userID {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	var errors []error

	err = s.kube.DeleteCluster(ctx, in.Reference)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.kube.DeleteDatabase(ctx, s.config.Kube.Project.Namespace, in.Reference)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.project.DeleteByRef(ctx, in.Reference)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.kube.DeleteIngressRouteTCP(ctx, in.Reference)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.kube.DeleteAuthAPIDeployment(ctx, in.Reference)
	if err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return nil, huma.Error500InternalServerError("Failed to delete project resources", errors...)
	}

	out := &dto.DeleteProjectByRefOutput{}
	out.Body.Success = true

	return out, nil
}

func (s *ProjectService) PatchProjectSettings(ctx context.Context, in *dto.PatchProjectSettingInput, userID string) error {
	project, err := s.project.FindByRef(ctx, in.Body.Ref)
	if err != nil {
		return err
	}
	if project.OwnerID != userID {
		return huma.Error401Unauthorized("Unauthorized")
	}

	objectPayload := &models.Object{}
	projectPayload := &models.Project{}
	authProviders := []*models.ProjectAuthProvider{}
	needPatchDeployment := false

	if in.Body.Name != nil || in.Body.Description != nil {
		objectPayload.ChineseName = in.Body.Name
		objectPayload.ChineseDescription = in.Body.Description
		objectPayload.UpdatedAt = time.Now()
	}

	opt := kube_project.NewAPIDeploymentOption()
	if in.Body.TrustedOrigins != nil {
		needPatchDeployment = true
		opt.WithTrustedOrigins(in.Body.TrustedOrigins)
	}

	if in.Body.Auth != nil {
		needPatchDeployment = true
		if in.Body.Auth.Email != nil {
			opt.WithEmailAndPasswordAuth(in.Body.Auth.Email.Enabled)
			authProviders = append(authProviders, &models.ProjectAuthProvider{
				Name:      "email",
				ProjectID: project.ID,
				Enabled:   in.Body.Auth.Email.Enabled,
				UpdatedAt: time.Now(),
			})
		}
		if in.Body.Auth.Google != nil {
			opt.WithGoogle(in.Body.Auth.Google.ClientID, in.Body.Auth.Google.ClientSecret)
			authProviders = append(authProviders, &models.ProjectAuthProvider{
				Name:         "google",
				ProjectID:    project.ID,
				Enabled:      in.Body.Auth.Google.Enabled,
				ClientID:     in.Body.Auth.Google.ClientID,
				ClientSecret: in.Body.Auth.Google.ClientSecret,
				UpdatedAt:    time.Now(),
			})
		}
		if in.Body.Auth.GitHub != nil {
			opt.WithGitHub(in.Body.Auth.GitHub.ClientID, in.Body.Auth.GitHub.ClientSecret)
			authProviders = append(authProviders, &models.ProjectAuthProvider{
				Name:         "github",
				ProjectID:    project.ID,
				Enabled:      in.Body.Auth.GitHub.Enabled,
				ClientID:     in.Body.Auth.GitHub.ClientID,
				ClientSecret: in.Body.Auth.GitHub.ClientSecret,
				UpdatedAt:    time.Now(),
			})
		}
		if in.Body.Auth.Discord != nil {
			opt.WithDiscord(in.Body.Auth.Discord.ClientID, in.Body.Auth.Discord.ClientSecret)
			authProviders = append(authProviders, &models.ProjectAuthProvider{
				Name:         "discord",
				ProjectID:    project.ID,
				Enabled:      in.Body.Auth.Discord.Enabled,
				ClientID:     in.Body.Auth.Discord.ClientID,
				ClientSecret: in.Body.Auth.Discord.ClientSecret,
				UpdatedAt:    time.Now(),
			})
		}
	}

	if needPatchDeployment {
		err = s.kube.PatchAuthAPIDeployment(ctx, in.Body.Ref, opt)
		if err != nil {
			return err
		}
	}

	if len(authProviders) > 0 {
		err = s.projectAuthSetting.UpsertOAuthProviders(ctx, authProviders)
		if err != nil {
			return err
		}
	}

	err = s.project.UpdateByRef(ctx, in.Body.Ref, projectPayload, objectPayload)
	if err != nil {
		return err
	}

	return nil
}

func (s *ProjectService) GetUsersProjects(ctx context.Context, userID string) ([]*models.ProjectView, error) {
	projects, err := s.project.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *ProjectService) GetUserProjectStatusByRef(ctx context.Context, c chan any, ref, userID string) error {
	project, err := s.project.FindByRef(ctx, ref)
	if err != nil {
		return err
	}
	if project.OwnerID != userID {
		return huma.Error401Unauthorized("Unauthorized")
	}
	if project.InitializedAt != nil {
		return huma.Error400BadRequest("Project already initialized")
	}

	totalStep := 4
	for {
		status, err := s.kube.FindClusterStatus(ctx, ref)
		if err != nil {
			return err
		}
		if status == nil {
			c <- dto.ProjectStatusEvent{Message: "Postgres cluster is not ready yet.", Step: 0, TotalStep: totalStep}
			continue
		}
		switch *status {
		case "Initializing Postgres cluster":
			c <- dto.ProjectStatusEvent{Message: "Postgres cluster is initializing...", Step: 1, TotalStep: totalStep}
		case "Setting up primary":
			c <- dto.ProjectStatusEvent{Message: "Postgres cluster is setting up primary...", Step: 2, TotalStep: totalStep}
		case "Waiting for the instances to become active":
			c <- dto.ProjectStatusEvent{Message: "Postgres Waiting for the instances to become active", Step: 3, TotalStep: totalStep}
		case "Cluster in healthy state":
			s.project.UpdateByRef(ctx, ref, &models.Project{
				InitializedAt: lo.ToPtr(time.Now()),
			}, models.Object{
				UpdatedAt: time.Now(),
			})
			c <- dto.ProjectStatusEvent{Message: "Postgres cluster is ready.", Step: 4, TotalStep: totalStep}
			return nil
		default:
			slog.Warn("Unknown Postgres cluster status", "status", *status)
		}
		time.Sleep(time.Second * 1)
	}
}

func (s *ProjectService) GetUserProjectByRef(ctx context.Context, ref, userID string) (*models.ProjectView, error) {
	project, err := s.project.FindByRef(ctx, ref)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	return project, nil
}

func (s *ProjectService) GetProjectSettings(ctx context.Context, in *dto.GetProjectSettingsInput, userID string) (*dto.GetProjectSettingsOutput, error) {
	project, err := s.project.FindByRef(ctx, in.Ref)
	if err != nil {
		return nil, err
	}
	if project.OwnerID != userID {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	authSettings, err := s.projectAuthSetting.FindByProjectID(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	oauthProviders, err := s.projectAuthSetting.FindAllOAuthProviders(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	out := &dto.GetProjectSettingsOutput{}
	out.Body.ID = project.ID
	out.Body.TrustedOrigins = authSettings.TrustedOrigins
	out.Body.CreatedAt = project.CreatedAt.Format(time.RFC3339)
	out.Body.UpdatedAt = project.UpdatedAt.Format(time.RFC3339)

	for _, provider := range oauthProviders {
		providerInfo := &dto.ProjectAuthProviderInfo{
			Enabled:      provider.Enabled,
			ClientID:     provider.ClientID,
			ClientSecret: provider.ClientSecret,
		}

		switch provider.Name {
		case "email":
			out.Body.Auth.Email = providerInfo
		case "google":
			out.Body.Auth.Google = providerInfo
		case "github":
			out.Body.Auth.GitHub = providerInfo
		case "discord":
			out.Body.Auth.Discord = providerInfo
		}
	}

	return out, nil
}

func (s *ProjectService) ResetDatabasePassword(ctx context.Context, in *dto.ResetDatabasePasswordInput, userID string) (*dto.ResetDatabasePasswordOutput, error) {
	err := s.kube.UpdateDatabaseRoleSecret(ctx, in.Body.Reference, "app", in.Body.Password)
	if err != nil {
		return nil, err
	}

	_ = s.project.UpdateByRef(ctx, in.Body.Reference, map[string]any{"password_expired_at": nil}, models.Object{
		UpdatedAt: time.Now(),
	})

	out := &dto.ResetDatabasePasswordOutput{}
	out.Body.Success = true
	return out, nil
}
