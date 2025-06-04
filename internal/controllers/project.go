package controllers

import (
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
	projectService services.ProjectService
}

func NewProjectController(projectService services.ProjectService) ProjectController {
	return &projectController{
		projectService: projectService,
	}
}

func (c *projectController) RegisterProjectAPIs(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "create-project",
		Method:      "POST",
		Path:        "/project",
		Summary:     "Create Project",
		Description: "Create a new project with the specified name and storage size.",
	}, c.createProject)

	huma.Register(api, huma.Operation{
		OperationID: "delete-project-by-ref",
		Method:      "DELETE",
		Path:        "/project/by-ref",
		Summary:     "Delete Project by Reference",
		Description: "Delete a project by its reference. The reference is a 20-character string.",
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
