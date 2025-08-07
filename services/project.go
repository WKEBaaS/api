package services

import (
	"baas-api/config"
	"baas-api/dto"
	"baas-api/models"
	"baas-api/repo"
	"baas-api/repo/kube"
	"context"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type ProjectService interface {
	CreateProject(ctx context.Context, in *dto.CreateProjectInput, userID *string) (*dto.CreateProjectOutput, error)
	DeleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput, userID string) (*dto.DeleteProjectByRefOutput, error)
	GetUsersProjects(ctx context.Context, userID string) ([]*models.ProjectView, error)
	GetUserProjectByRef(ctx context.Context, ref, userID string) (*models.ProjectView, error)
	GetUserProjectStatusByRef(ctx context.Context, c chan any, ref, userID string) error
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

	err = s.repo.kubeProject.CreateIngressRouteTCP(ctx, s.config.Kube.Project.Namespace, *ref)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.repo.kubeProject.DeleteIngressRouteTCP(ctx, s.config.Kube.Project.Namespace, *ref)
	})
	err = s.repo.kubeProject.CreateAPIDeployment(ctx,
		kube.NewAPIDeploymentOption().
			WithNamespace(s.config.Kube.Project.Namespace).
			WithRef(*ref).
			WithBetterAuthSecret(projectAuthSetting.Secret).
			WithEmailAndPasswordAuth(projectAuthSetting.EmailAndPasswordEnabled),
	)
	if err != nil {
		return nil, err
	}
	cleanupFuncs = append(cleanupFuncs, func() {
		_ = s.repo.kubeProject.DeleteAPIDeployment(ctx, s.config.Kube.Project.Namespace, *ref)
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

	err = s.repo.kubeProject.DeleteCluster(ctx, s.config.Kube.Project.Namespace, in.Reference)
	if err != nil {
		return nil, err
	}

	err = s.repo.kubeProject.DeleteDatabase(ctx, s.config.Kube.Project.Namespace, in.Reference)
	if err != nil {
		// If the database deletion fails, we should not proceed with deleting the project entry
		return nil, err
	}

	err = s.repo.project.DeleteByRef(ctx, in.Reference)
	if err != nil {
		return nil, err
	}

	err = s.repo.kubeProject.DeleteIngressRouteTCP(ctx, s.config.Kube.Project.Namespace, in.Reference)
	if err != nil {
		return nil, err
	}

	err = s.repo.kubeProject.DeleteAPIDeployment(ctx, s.config.Kube.Project.Namespace, in.Reference)
	if err != nil {
		return nil, err
	}

	out := &dto.DeleteProjectByRefOutput{}
	out.Body.Success = true

	return out, nil
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

	for {
		status, err := s.repo.kubeProject.FindClusterStatus(ctx, s.config.Kube.Project.Namespace, ref)
		if err != nil {
			return err
		}
		if status == nil {
			c <- dto.MessageEvent{Message: "Postgres cluster is not ready yet."}
			continue
		}
		if *status == "Cluster in healthy state" {
			c <- dto.MessageEvent{Message: "Postgres cluster is ready."}
			break
		}
		c <- dto.MessageEvent{Message: *status}
		time.Sleep(time.Second * 1)
	}
	return nil
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

func (s *projectService) ResetDatabasePassword(ctx context.Context, in *dto.ResetDatabasePasswordInput, userID string) (*dto.ResetDatabasePasswordOutput, error) {
	err := s.repo.kubeProject.ResetDatabasePassword(ctx, s.config.Kube.Project.Namespace, in.Body.Reference, in.Body.Password)
	if err != nil {
		return nil, err
	}
	out := &dto.ResetDatabasePasswordOutput{}
	out.Body.Success = true
	return out, nil
}
