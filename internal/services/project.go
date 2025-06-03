package services

import (
	"baas-api/internal/configs"
	"baas-api/internal/dto"
	"baas-api/internal/repo"
	"context"
)

type ProjectService interface {
	CreateProject(ctx context.Context, in *dto.CreateProjectInput) (*dto.CreateProjectOutput, error)
}

type projectService struct {
	projectRepo     repo.ProjectRepository
	kubeProjectRepo repo.KubeProjectRepository
	namespace       string
}

func NewProjectService(config configs.Config, p repo.ProjectRepository, kp repo.KubeProjectRepository) ProjectService {
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
		// If the cluster creation fails, we should clean up the project entry
		if delErr := s.projectRepo.DeleteProjectByIDPermanently(ctx, *id); delErr != nil {
			return nil, delErr
		}
		return nil, err
	}

	out := &dto.CreateProjectOutput{}
	out.Body.ID = *id
	out.Body.Ref = *ref

	return out, nil
}
