package services

import (
	"baas-api/internal/configs"
	"baas-api/internal/dto"
	"baas-api/internal/repo"
	"context"
)

type ProjectService interface {
	CreateProject(ctx context.Context, in *dto.CreateProjectInput) (*dto.CreateProjectOutput, error)
	DeleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput) (*dto.DeleteProjectByRefOutput, error)
}

type projectService struct {
	projectRepo     repo.ProjectRepository
	kubeProjectRepo repo.KubeProjectRepository
	namespace       string
}

func NewProjectService(config *configs.Config, p repo.ProjectRepository, kp repo.KubeProjectRepository) ProjectService {
	return &projectService{
		projectRepo:     p,
		kubeProjectRepo: kp,
		namespace:       config.Kube.ProjectsNamespace,
	}
}

func (s *projectService) CreateProject(ctx context.Context, in *dto.CreateProjectInput) (*dto.CreateProjectOutput, error) {
	id, ref, err := s.projectRepo.CreateProject(ctx, in.Body.Name)
	if err != nil {
		return nil, err
	}

	err = s.kubeProjectRepo.CreateCluster(ctx, s.namespace, *ref, in.Body.StorageSize)
	if err != nil {
		// Tey to clean up the cluster if creation fails
		_ = s.projectRepo.DeleteProjectByIDPermanently(ctx, *id)
		return nil, err
	}

	err = s.kubeProjectRepo.CreateDatabase(ctx, s.namespace, *ref)
	if err != nil {
		// Tey to clean up the cluster if database creation fails
		_ = s.kubeProjectRepo.DeleteCluster(ctx, s.namespace, *ref)
		_ = s.projectRepo.DeleteProjectByIDPermanently(ctx, *id)
		return nil, err
	}

	err = s.kubeProjectRepo.CreateIngressRouteTCP(ctx, s.namespace, *ref)
	if err != nil {
		// Tey to clean up the cluster and database if ingress route creation fails
		_ = s.kubeProjectRepo.DeleteDatabase(ctx, s.namespace, *ref)
		_ = s.kubeProjectRepo.DeleteCluster(ctx, s.namespace, *ref)
		_ = s.projectRepo.DeleteProjectByIDPermanently(ctx, *id)
		return nil, err
	}

	out := &dto.CreateProjectOutput{}
	out.Body.ID = *id
	out.Body.Ref = *ref

	return out, nil
}

func (s *projectService) DeleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput) (*dto.DeleteProjectByRefOutput, error) {
	err := s.kubeProjectRepo.DeleteCluster(ctx, s.namespace, in.Ref)
	if err != nil {
		return nil, err
	}

	err = s.kubeProjectRepo.DeleteDatabase(ctx, s.namespace, in.Ref)
	if err != nil {
		// If the database deletion fails, we should not proceed with deleting the project entry
		return nil, err
	}

	err = s.projectRepo.DeleteProjectByRefPermanently(ctx, in.Ref)
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
