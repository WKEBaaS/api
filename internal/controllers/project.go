package controllers

import (
	"baas-api/internal/configs"
	"baas-api/internal/controllers/middlewares"
	"baas-api/internal/dto"
	"baas-api/internal/services"
	"context"

	"github.com/danielgtaylor/huma/v2"
)

type ProjectController interface {
	RegisterProjectAPIs(api huma.API)
	createProject(ctx context.Context, in *dto.CreateProjectInput) (*dto.CreateProjectOutput, error)
	deleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput) (*dto.DeleteProjectByRefOutput, error)
}

type projectController struct {
	config         *configs.Config
	projectService services.ProjectService
}

func NewProjectController(config *configs.Config, projectService services.ProjectService) ProjectController {
	return &projectController{
		config:         config,
		projectService: projectService,
	}
}

func (c *projectController) RegisterProjectAPIs(api huma.API) {
	authMiddleware := middlewares.NewAuthMiddleWare(api, c.config, "baasAuth")

	huma.Register(api, huma.Operation{
		OperationID: "test",
		Method:      "GET",
		Path:        "/test",
		Summary:     "Test Endpoint",
		Description: "A simple test endpoint to verify the API is working.",
		Tags:        []string{"Test"},
		Security: []map[string][]string{
			{"baasAuth": {"project:create"}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, func(ctx context.Context, in *struct{}) (*struct{}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "create-project",
		Method:      "POST",
		Path:        "/project",
		Summary:     "Create Project",
		Description: "Create a new project with the specified name and storage size.",
		Tags:        []string{"Project"},
		Security: []map[string][]string{
			{"baasAuth": {"project:create"}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.createProject)

	huma.Register(api, huma.Operation{
		OperationID: "delete-project-by-ref",
		Method:      "DELETE",
		Path:        "/project/by-ref",
		Summary:     "Delete Project by Reference",
		Description: "Delete a project by its reference. The reference is a 20-character string.",
		Tags:        []string{"Project"},
	}, c.deleteProjectByRef)
}

func (c *projectController) createProject(ctx context.Context, in *dto.CreateProjectInput) (*dto.CreateProjectOutput, error) {
	out, err := c.projectService.CreateProject(ctx, in)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *projectController) deleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput) (*dto.DeleteProjectByRefOutput, error) {
	out, err := c.projectService.DeleteProjectByRef(ctx, in)
	if err != nil {
		return nil, err
	}

	return out, nil
}
