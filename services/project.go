package services

import (
	"baas-api/config"
	"baas-api/dto"
	"baas-api/models"
	"baas-api/repo"
	"baas-api/repo/kube"
	"context"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/lo"
)

type ProjectService interface {
	CreateProject(ctx context.Context, in *dto.CreateProjectInput, userID *string) (*dto.CreateProjectOutput, error)
	DeleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput, userID string) (*dto.DeleteProjectByRefOutput, error)
	PatchProjectSettings(ctx context.Context, in *dto.PatchProjectSettingInput, userID string) error
	GetUsersProjects(ctx context.Context, userID string) ([]*models.ProjectView, error)
	GetUserProjectByRef(ctx context.Context, ref, userID string) (*models.ProjectView, error)
	GetUserProjectStatusByRef(ctx context.Context, c chan any, ref, userID string) error
	GetProjectSettings(ctx context.Context, in *dto.GetProjectSettingsInput, userID string) (*dto.GetProjectSettingsOutput, error)
	ResetDatabasePassword(ctx context.Context, in *dto.ResetDatabasePasswordInput, userID string) (*dto.ResetDatabasePasswordOutput, error)
}

type projectService struct {
	config *config.Config
	repo   struct {
		entity             repo.EntityRepository
		project            repo.ProjectRepository
		projectAuthSetting repo.ProjectAuthSettingRepository
		kubeProject        kube.KubeProjectRepository
	}
}

func NewProjectService(config *config.Config, ep repo.EntityRepository, pp repo.ProjectRepository, ps repo.ProjectAuthSettingRepository, kp kube.KubeProjectRepository) ProjectService {
	service := &projectService{}
	service.config = config
	service.repo.entity = ep
	service.repo.project = pp
	service.repo.projectAuthSetting = ps
	service.repo.kubeProject = kp
	return service
}

func (s *projectService) CreateProject(ctx context.Context, in *dto.CreateProjectInput, userID *string) (*dto.CreateProjectOutput, error) {
	projectEntity, err := s.repo.entity.GetByChineseName(ctx, "專案")
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

	id, ref, err := s.repo.project.Create(ctx, in.Body.Name, in.Body.Description, projectEntity.ID, userID)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.repo.project.DeleteByID(ctx, *id)
	})

	projectAuthSetting := &models.ProjectAuthSettings{
		ProjectID: *id,
	}
	err = s.repo.projectAuthSetting.Create(ctx, projectAuthSetting)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.repo.projectAuthSetting.DeleteByProjectID(ctx, *id)
	})

	err = s.repo.kubeProject.CreateCluster(ctx, s.config.Kube.Project.Namespace, *ref, in.Body.StorageSize)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.repo.kubeProject.DeleteCluster(ctx, s.config.Kube.Project.Namespace, *ref)
	})

	err = s.repo.kubeProject.CreateDatabase(ctx, s.config.Kube.Project.Namespace, *ref)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.repo.kubeProject.DeleteDatabase(ctx, s.config.Kube.Project.Namespace, *ref)
	})

	err = s.repo.kubeProject.CreateAPIDeployment(ctx,
		*ref,
		kube.NewAPIDeploymentOption().
			WithNamespace(s.config.Kube.Project.Namespace).
			WithBetterAuthSecret(projectAuthSetting.Secret).
			WithEmailAndPasswordAuth(projectAuthSetting.EmailAndPasswordEnabled),
	)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.repo.kubeProject.DeleteAPIDeployment(ctx, s.config.Kube.Project.Namespace, *ref)
	})

	err = s.repo.kubeProject.CreateAPIService(ctx, s.config.Kube.Project.Namespace, *ref)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.repo.kubeProject.DeleteAPIService(ctx, s.config.Kube.Project.Namespace, *ref)
	})

	err = s.repo.kubeProject.CreateIngressRoute(ctx, s.config.Kube.Project.Namespace, *ref)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.repo.kubeProject.DeleteIngressRoute(ctx, s.config.Kube.Project.Namespace, *ref)
	})

	err = s.repo.kubeProject.CreateIngressRouteTCP(ctx, s.config.Kube.Project.Namespace, *ref)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.repo.kubeProject.DeleteIngressRouteTCP(ctx, s.config.Kube.Project.Namespace, *ref)
	})

	success = true // Mark as successful if we reach here without errors
	out := &dto.CreateProjectOutput{}
	out.Body.ID = *id
	out.Body.Reference = *ref

	return out, nil
}

func (s *projectService) DeleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput, userID string) (*dto.DeleteProjectByRefOutput, error) {
	project, err := s.repo.project.FindByRef(ctx, in.Reference)
	if err != nil {
		return nil, err
	}
	if project.OwnerID != userID {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	var errors []error

	err = s.repo.kubeProject.DeleteCluster(ctx, s.config.Kube.Project.Namespace, in.Reference)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.repo.kubeProject.DeleteDatabase(ctx, s.config.Kube.Project.Namespace, in.Reference)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.repo.project.DeleteByRef(ctx, in.Reference)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.repo.kubeProject.DeleteIngressRouteTCP(ctx, s.config.Kube.Project.Namespace, in.Reference)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.repo.kubeProject.DeleteAPIDeployment(ctx, s.config.Kube.Project.Namespace, in.Reference)
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

func (s *projectService) PatchProjectSettings(ctx context.Context, in *dto.PatchProjectSettingInput, userID string) error {
	project, err := s.repo.project.FindByRef(ctx, in.Body.Ref)
	if err != nil {
		return err
	}
	if project.OwnerID != userID {
		return huma.Error401Unauthorized("Unauthorized")
	}

	if in.Body.TrustedOrigins == nil && in.Body.Auth == nil {
		return nil // No changes to apply
	}

	oauthProviders := []*models.ProjectOAuthProvider{}
	objectPayload := &models.Object{}
	projectPayload := &models.Project{}
	needPatchDeployment := false

	if in.Body.Name != nil || in.Body.Description != nil {
		objectPayload.ChineseName = in.Body.Name
		objectPayload.EnglishName = in.Body.Name
		objectPayload.UpdatedAt = time.Now()
	}

	opt := kube.NewAPIDeploymentOption()
	if in.Body.TrustedOrigins != nil {
		needPatchDeployment = true
		opt.WithTrustedOrigins(in.Body.TrustedOrigins)
	}

	if in.Body.Auth != nil {
		needPatchDeployment = true
		if in.Body.Auth.EmailAndPasswordEnabled != nil {
			opt.WithEmailAndPasswordAuth(*in.Body.Auth.EmailAndPasswordEnabled)
		}
		if in.Body.Auth.Google != nil {
			opt.WithGoogle(in.Body.Auth.Google.ClientID, in.Body.Auth.Google.ClientSecret)
			oauthProviders = append(oauthProviders, &models.ProjectOAuthProvider{
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
			oauthProviders = append(oauthProviders, &models.ProjectOAuthProvider{
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
			oauthProviders = append(oauthProviders, &models.ProjectOAuthProvider{
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
		err = s.repo.kubeProject.PatchAPIDeployment(ctx, s.config.Kube.Project.Namespace, in.Body.Ref, opt)
		if err != nil {
			return err
		}
	}

	if len(oauthProviders) > 0 {
		err = s.repo.projectAuthSetting.UpsertOAuthProviders(ctx, oauthProviders)
		if err != nil {
			return err
		}
	}

	err = s.repo.project.UpdateByRef(ctx, in.Body.Ref, projectPayload, objectPayload)
	if err != nil {
		return err
	}

	return nil
}

func (s *projectService) GetUsersProjects(ctx context.Context, userID string) ([]*models.ProjectView, error) {
	projects, err := s.repo.project.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *projectService) GetUserProjectStatusByRef(ctx context.Context, c chan any, ref, userID string) error {
	project, err := s.repo.project.FindByRef(ctx, ref)
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
		status, err := s.repo.kubeProject.FindClusterStatus(ctx, s.config.Kube.Project.Namespace, ref)
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
			s.repo.project.UpdateByRef(ctx, ref, &models.Project{
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

func (s *projectService) GetUserProjectByRef(ctx context.Context, ref, userID string) (*models.ProjectView, error) {
	project, err := s.repo.project.FindByRef(ctx, ref)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	return project, nil
}

func (s *projectService) GetProjectSettings(ctx context.Context, in *dto.GetProjectSettingsInput, userID string) (*dto.GetProjectSettingsOutput, error) {
	project, err := s.repo.project.FindByRef(ctx, in.Ref)
	if err != nil {
		return nil, err
	}
	if project.OwnerID != userID {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	authSettings, err := s.repo.projectAuthSetting.FindByProjectID(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	oauthProviders, err := s.repo.projectAuthSetting.FindAllOAuthProviders(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	out := &dto.GetProjectSettingsOutput{}
	out.Body.ID = project.ID
	out.Body.TrustedOrigins = authSettings.TrustedOrigins
	out.Body.CreatedAt = project.CreatedAt.Format(time.RFC3339)
	out.Body.UpdatedAt = project.UpdatedAt.Format(time.RFC3339)

	out.Body.Auth.EmailAndPasswordEnabled = authSettings.EmailAndPasswordEnabled

	for _, provider := range oauthProviders {
		providerInfo := &dto.ProjectOAuthProviderInfo{
			Enabled:      provider.Enabled,
			ClientID:     provider.ClientID,
			ClientSecret: provider.ClientSecret,
		}

		switch provider.Name {
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

func (s *projectService) ResetDatabasePassword(ctx context.Context, in *dto.ResetDatabasePasswordInput, userID string) (*dto.ResetDatabasePasswordOutput, error) {
	err := s.repo.kubeProject.ResetDatabasePassword(ctx, s.config.Kube.Project.Namespace, in.Body.Reference, in.Body.Password)
	if err != nil {
		return nil, err
	}

	_ = s.repo.project.UpdateByRef(ctx, in.Body.Reference, map[string]any{"password_expired_at": nil}, models.Object{
		UpdatedAt: time.Now(),
	})

	out := &dto.ResetDatabasePasswordOutput{}
	out.Body.Success = true
	return out, nil
}
