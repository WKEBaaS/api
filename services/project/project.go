// Package project implements the project service for managing projects in the BaaS API.
package project

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"baas-api/config"
	"baas-api/dto"
	"baas-api/models"
	"baas-api/repo"
	"baas-api/services/kubeproject"
	"baas-api/services/pgrest"
	"baas-api/services/s3"
	"baas-api/utils"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/samber/do/v2"
	"github.com/samber/lo"
)

type CreateProjectInternalOutput struct {
	AuthSecret    string
	JWKSPublicKey string
}

type ProjectServiceInterface interface {
	// CreateProject
	//
	// Returns dto.CreateProjectOutput, project's auth-secret, error
	CreateProject(ctx context.Context, in *dto.CreateProjectInput, jwt string, userID *string) (*dto.CreateProjectOutput, *CreateProjectInternalOutput, error)
	// CreateProjectPostInstall performs post-installation steps after the project's cluster is read.
	CreateProjectPostInstall(ctx context.Context, ref string, authSecret string, jwks string) error
	GetProjectJWKS(ctx context.Context, ref string) (*string, error)
	DeleteProjectByID(ctx context.Context, jwt string, in *dto.DeleteProjectByIDInput, userID string) (*dto.DeleteProjectByIDOutput, error)
	PatchProjectSettings(ctx context.Context, jwt string, in *dto.UpdateProjectInput, userID string) error
	GetUsersProjects(ctx context.Context, userID string) ([]*models.ProjectView, error)
	GetUserProjectByRef(ctx context.Context, ref, userID string) (*models.ProjectView, error)
	GetUserProjectStatusByRef(ctx context.Context, c chan any, ref, userID string) error
	GetProjectSettings(ctx context.Context, in *dto.GetProjectSettingsInput, userID string) (*dto.GetProjectSettingsOutput, error)
	ResetDatabasePassword(ctx context.Context, in *dto.ResetDatabasePasswordInput, userID string) (*dto.ResetDatabasePasswordOutput, error)
}

type ProjectService struct {
	config *config.Config
	// Services
	kube   kubeproject.KubeProjectServiceInterface
	pgrest pgrest.PgRestServiceInterface
	s3     s3.S3ServiceInterface
	// Repositories
	// entity             repo.EntityRepositoryInterface             `do:""`
	project            repo.ProjectRepositoryInterface
	projectAuthSetting repo.ProjectAuthSettingRepositoryInterface
}

func NewProjectService(i do.Injector) (ProjectServiceInterface, error) {
	service := &ProjectService{
		config:             do.MustInvoke[*config.Config](i),
		kube:               do.MustInvoke[kubeproject.KubeProjectServiceInterface](i),
		pgrest:             do.MustInvoke[pgrest.PgRestServiceInterface](i),
		s3:                 do.MustInvoke[s3.S3ServiceInterface](i),
		project:            do.MustInvoke[repo.ProjectRepositoryInterface](i),
		projectAuthSetting: do.MustInvoke[repo.ProjectAuthSettingRepositoryInterface](i),
	}
	return service, nil
}

var Package = do.Package(
	do.Lazy(NewProjectService),
)

func (s *ProjectService) CreateProject(ctx context.Context, in *dto.CreateProjectInput, jwt string, userID *string) (*dto.CreateProjectOutput, *CreateProjectInternalOutput, error) {
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

	///// Create database records /////
	project, err := s.pgrest.CreateProject(ctx, jwt, in.Body.Name, *in.Body.Description)
	if err != nil {
		return nil, nil, err
	}

	////// Create S3 resources /////
	err = s.s3.CreateBucket(ctx, project.S3Bucket)
	if err != nil {
		return nil, nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.s3.DeleteBucket(ctx, project.S3Bucket)
	})

	err = s.s3.CreateBucketUser(ctx, project.S3AccessKeyID, project.S3SecretAccessKey)
	if err != nil {
		return nil, nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.s3.DeleteBucketUser(ctx, project.S3AccessKeyID)
	})

	err = s.s3.CreateBucketPolicy(ctx, project.Ref, project.S3Bucket)
	if err != nil {
		return nil, nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.s3.DeleteBucketPolicy(ctx, project.S3Bucket)
	})

	///// Create Kubernetes resources /////
	ref := project.Ref
	jwkID, err := uuid.NewV7()
	if err != nil {
		return nil, nil, errors.New("failed to generate JWK ID")
	}

	publicKey, privateKey, err := utils.NewEd25519JWKWithKIDStringified(ctx, jwkID.String())
	if err != nil {
		return nil, nil, err
	}
	err = s.kube.CreateJWKSConfigMap(ctx, kubeproject.CreateJWKSConfigMapOption{
		Ref:        ref,
		KID:        jwkID.String(),
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	})
	if err != nil {
		return nil, nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.kube.DeleteJWKSConfigMap(ctx, ref)
	})

	password := utils.GenerateNewPassword(16)
	err = s.kube.CreateDatabaseRoleSecret(ctx, ref, kubeproject.RoleAuthenticator, password)
	if err != nil {
		return nil, nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.kube.DeleteDatabaseRoleSecret(ctx, ref, kubeproject.RoleAuthenticator)
	})

	err = s.kube.CreateCluster(ctx, ref, in.Body.StorageSize)
	if err != nil {
		return nil, nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.kube.DeleteCluster(ctx, ref)
	})

	err = s.kube.CreateDatabase(ctx, ref)
	if err != nil {
		return nil, nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.kube.DeleteDatabase(ctx, ref)
	})

	success = true // Mark as successful if we reach here without errors
	out := &dto.CreateProjectOutput{}
	out.Body.ID = project.ID
	out.Body.Reference = project.Ref

	internalOut := &CreateProjectInternalOutput{}
	internalOut.AuthSecret = project.AuthSecret
	internalOut.JWKSPublicKey = publicKey

	return out, internalOut, nil
}

func (s *ProjectService) CreateProjectPostInstall(ctx context.Context, ref string, authSecret string, jwks string) error {
	var err error
	err = s.kube.CreateMigrationJob(ctx, ref)
	if err != nil {
		return err
	}

	err = s.kube.CreateAuthAPIDeployment(ctx, ref,
		&kubeproject.APIDeploymentOption{
			BetterAuthSecret: &authSecret,
			TrustedOrigins:   []string{"*"},
			AuthProviders: map[string]dto.AuthProvider{
				"email": {
					Enabled: true,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	err = s.kube.CreateRESTAPIDeployment(ctx, ref, jwks)
	if err != nil {
		return err
	}

	err = s.kube.CreateAuthAPIService(ctx, ref)
	if err != nil {
		return err
	}

	err = s.kube.CreateRESTAPIService(ctx, ref)
	if err != nil {
		return err
	}

	err = s.kube.CreateIngressRoute(ctx, ref)
	if err != nil {
		return err
	}

	err = s.kube.CreateIngressRouteTCP(ctx, ref)
	if err != nil {
		return err
	}

	return nil
}

func (s *ProjectService) GetProjectJWKS(ctx context.Context, ref string) (*string, error) {
	jwksURL := url.URL{
		Scheme: "https",
		Host:   ref + "." + s.config.App.ExternalDomain,
		Path:   "/api/auth/jwks",
	}

	req, err := http.NewRequest(http.MethodGet, jwksURL.String(), nil)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create request for JWKS", "error", err)
		return nil, huma.Error500InternalServerError("Failed to create request for JWKS")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get JWKS", "error", err)
		return nil, huma.Error500InternalServerError("Failed to get JWKS")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.ErrorContext(ctx, "Failed to get JWKS", "status", resp.StatusCode)
		return nil, huma.Error500InternalServerError("Failed to get JWKS")
	}

	body, _ := io.ReadAll(resp.Body)
	jwks := string(body)
	return &jwks, nil
}

func (s *ProjectService) DeleteProjectByID(ctx context.Context, jwt string, in *dto.DeleteProjectByIDInput, userID string) (*dto.DeleteProjectByIDOutput, error) {
	var errors []error

	///// Delete database records /////
	deleted, err := s.pgrest.DeleteProject(ctx, jwt, in.ID)
	if err != nil {
		return nil, err
	}

	///// Delete S3 resources /////
	err = s.s3.DeleteBucket(ctx, deleted.S3Bucket)
	if err != nil {
		errors = append(errors, err)
	}
	err = s.s3.DeleteBucketUser(ctx, deleted.S3AccessKeyID)
	if err != nil {
		errors = append(errors, err)
	}
	err = s.s3.DeleteBucketPolicy(ctx, deleted.S3Bucket)
	if err != nil {
		errors = append(errors, err)
	}

	///// Delete Kubernetes resources /////
	err = s.kube.DeleteCluster(ctx, deleted.Ref)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.kube.DeleteDatabase(ctx, deleted.Ref)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.kube.DeleteIngressRouteTCP(ctx, deleted.Ref)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.kube.DeleteIngressRoute(ctx, deleted.Ref)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.kube.DeleteAuthAPIDeployment(ctx, deleted.Ref)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.kube.DeleteRESTAPIDeployment(ctx, deleted.Ref)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.kube.DeleteAuthAPIService(ctx, deleted.Ref)
	if err != nil {
		errors = append(errors, err)
	}
	err = s.kube.DeleteRESTAPIService(ctx, deleted.Ref)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.kube.DeleteDatabaseRoleSecret(ctx, deleted.Ref, kubeproject.RoleAuthenticator)
	if err != nil {
		errors = append(errors, err)
	}

	err = s.kube.DeleteJWKSConfigMap(ctx, deleted.Ref)
	if err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return nil, huma.Error500InternalServerError("Failed to delete project resources", errors...)
	}

	out := &dto.DeleteProjectByIDOutput{}
	out.Body.Success = true

	return out, nil
}

func (s *ProjectService) PatchProjectSettings(ctx context.Context, jwt string, in *dto.UpdateProjectInput, userID string) error {
	updated, err := s.pgrest.UpdateProject(ctx, jwt, pgrest.UpdateProjectPayload{
		ID:             in.Body.ID,
		Name:           in.Body.Name,
		Description:    in.Body.Description,
		TrustedOrigins: in.Body.TrustedOrigins,
		ProxyURL:       in.Body.ProxyURL,
	})
	if err != nil {
		return err
	}

	needPatchDeployment := false
	opt := &kubeproject.APIDeploymentOption{}
	if in.Body.TrustedOrigins != nil {
		needPatchDeployment = true
		opt.TrustedOrigins = in.Body.TrustedOrigins
	}

	if in.Body.ProxyURL != nil {
		needPatchDeployment = true
		opt.ProxyURL = in.Body.ProxyURL
	}

	if in.Body.Auth != nil {
		needPatchDeployment = true
		opt.AuthProviders = in.Body.Auth
	}

	if needPatchDeployment {
		err = s.kube.PatchAuthAPIDeployment(ctx, updated.Ref, opt)
		if err != nil {
			return err
		}
	}

	// Update Database Auth Providers
	if in.Body.Auth != nil {
		err := s.pgrest.CreateOrUpdateAuthProvider(ctx, jwt, pgrest.CreateOrUpdateAuthProviderPayload{
			ProjectID: in.Body.ID,
			Providers: in.Body.Auth,
		})
		if err != nil {
			return err
		}
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

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c <- dto.ProjectStatusEvent{Message: "Operation cancelled by client.", Step: -1, TotalStep: 4}
			return ctx.Err()
		case <-ticker.C:
			totalStep := 4
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
		}
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
	out.Body.ProxyURL = authSettings.ProxyURL
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
