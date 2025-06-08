package services

import (
	"baas-api/internal/configs"
	"baas-api/internal/dto"
	"baas-api/internal/models"
	"baas-api/internal/repo"
	"context"

	"github.com/danielgtaylor/huma/v2"
)

type ProjectService interface {
	CreateProject(ctx context.Context, in *dto.CreateProjectInput, userID *string) (*dto.CreateProjectOutput, error)
	DeleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput, userID string) (*dto.DeleteProjectByRefOutput, error)
	GetUsersProjects(ctx context.Context, userID string) ([]*models.ProjectView, error)
	GetUserProjectByRef(ctx context.Context, ref, userID string) (*models.ProjectView, error)
}

type projectService struct {
	entityRepo      repo.EntityRepository
	projectRepo     repo.ProjectRepository
	kubeProjectRepo repo.KubeProjectRepository
	namespace       string
}

func NewProjectService(config *configs.Config, ep repo.EntityRepository, pp repo.ProjectRepository, kp repo.KubeProjectRepository) ProjectService {
	return &projectService{
		entityRepo:      ep,
		projectRepo:     pp,
		kubeProjectRepo: kp,
		namespace:       config.Kube.ProjectsNamespace,
	}
}

func (s *projectService) CreateProject(ctx context.Context, in *dto.CreateProjectInput, userID *string) (*dto.CreateProjectOutput, error) {
	projectEntity, err := s.entityRepo.GetByChineseName(ctx, "專案")
	if err != nil {
		return nil, err
	}

	id, ref, err := s.projectRepo.Create(ctx, in.Body.Name, in.Body.Description, projectEntity.ID, userID)
	if err != nil {
		return nil, err
	}

	err = s.kubeProjectRepo.CreateCluster(ctx, s.namespace, *ref, in.Body.StorageSize)
	if err != nil {
		// Tey to clean up the cluster if creation fails
		_ = s.projectRepo.DeleteByIDPermanently(ctx, *id)
		return nil, err
	}

	err = s.kubeProjectRepo.CreateDatabase(ctx, s.namespace, *ref)
	if err != nil {
		// Tey to clean up the cluster if database creation fails
		_ = s.kubeProjectRepo.DeleteCluster(ctx, s.namespace, *ref)
		_ = s.projectRepo.DeleteByIDPermanently(ctx, *id)
		return nil, err
	}

	err = s.kubeProjectRepo.CreateIngressRouteTCP(ctx, s.namespace, *ref)
	if err != nil {
		// Tey to clean up the cluster and database if ingress route creation fails
		_ = s.kubeProjectRepo.DeleteDatabase(ctx, s.namespace, *ref)
		_ = s.kubeProjectRepo.DeleteCluster(ctx, s.namespace, *ref)
		_ = s.projectRepo.DeleteByIDPermanently(ctx, *id)
		return nil, err
	}

	out := &dto.CreateProjectOutput{}
	out.Body.ID = *id
	out.Body.Ref = *ref

	return out, nil
}

func (s *projectService) DeleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput, userID string) (*dto.DeleteProjectByRefOutput, error) {
	project, err := s.projectRepo.FindByRef(ctx, in.Ref)
	if err != nil {
		return nil, err
	}
	if project.OwnerID != userID {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	err = s.kubeProjectRepo.DeleteCluster(ctx, s.namespace, in.Ref)
	if err != nil {
		return nil, err
	}

	err = s.kubeProjectRepo.DeleteDatabase(ctx, s.namespace, in.Ref)
	if err != nil {
		// If the database deletion fails, we should not proceed with deleting the project entry
		return nil, err
	}

	err = s.projectRepo.DeleteByRefPermanently(ctx, in.Ref)
	if err != nil {
		return nil, err
	}

	err = s.kubeProjectRepo.DeleteIngressRouteTCP(ctx, s.namespace, in.Ref)
	if err != nil {
		return nil, err
	}

	out := &dto.DeleteProjectByRefOutput{}
	out.Body.Success = true

	return out, nil
}

func (s *projectService) GetUsersProjects(ctx context.Context, userID string) ([]*models.ProjectView, error) {
	projects, err := s.projectRepo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *projectService) GetUserProjectByRef(ctx context.Context, ref, userID string) (*models.ProjectView, error) {
	project, err := s.projectRepo.FindByRef(ctx, ref)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	return project, nil
}
