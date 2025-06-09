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
	getProjectByRef(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetProjectByRefOutput, error)
	createProject(ctx context.Context, in *dto.CreateProjectInput) (*dto.CreateProjectOutput, error)
	deleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput) (*dto.DeleteProjectByRefOutput, error)
	getUsersProjects(ctx context.Context, in *dto.GetUsersProjectsInput) (*dto.GetUsersProjectsOutput, error)
	resetDatabasePassword(ctx context.Context, in *dto.ResetDatabasePasswordInput) (*dto.ResetDatabasePasswordOutput, error)
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
			{"baasAuth": {"project:manage"}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, func(ctx context.Context, in *struct{}) (*struct{}, error) {
		return nil, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "get-project-by-ref",
		Method:      "GET",
		Path:        "/project/by-ref",
		Summary:     "Get Project by Reference",
		Description: "Retrieve a project by its reference. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Security: []map[string][]string{
			{"baasAuth": {"project:manage", "project:read"}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.getProjectByRef)

	huma.Register(api, huma.Operation{
		OperationID: "create-project",
		Method:      "POST",
		Path:        "/project",
		Summary:     "Create Project",
		Description: "Create a new project with the specified name and storage size.",
		Tags:        []string{"Project"},
		Security: []map[string][]string{
			{"baasAuth": {"project:manage", "project:create"}},
		},
		Middlewares: huma.Middlewares{authMiddleware, middlewares.TLSMiddleware},
	}, c.createProject)

	huma.Register(api, huma.Operation{
		OperationID: "delete-project-by-ref",
		Method:      "DELETE",
		Path:        "/project/by-ref",
		Summary:     "Delete Project by Reference",
		Description: "Delete a project by its reference. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Security: []map[string][]string{
			{"baasAuth": {"project:manage", "project:delete"}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.deleteProjectByRef)

	huma.Register(api, huma.Operation{
		OperationID: "get-users-projects",
		Method:      "GET",
		Path:        "/project/users",
		Summary:     "Get User's Projects",
		Description: "Retrieve all projects associated with the authenticated user.",
		Tags:        []string{"Project"},
		Security: []map[string][]string{
			{"baasAuth": {"project:manage", "project:read"}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.getUsersProjects)

	huma.Register(api, huma.Operation{
		OperationID: "reset-database-password",
		Method:      "POST",
		Path:        "/project/reset-database-password",
		Summary:     "Reset Database Password",
		Description: "Reset the database password for a project. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Security: []map[string][]string{
			{"baasAuth": {"project:manage", "project:write"}},
		},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.resetDatabasePassword)
}

func (c *projectController) getProjectByRef(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetProjectByRefOutput, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.projectService.GetUserProjectByRef(ctx, in.Ref, *userID)
	if err != nil {
		return nil, err
	}
	return &dto.GetProjectByRefOutput{Body: *out}, nil
}

func (c *projectController) createProject(ctx context.Context, in *dto.CreateProjectInput) (*dto.CreateProjectOutput, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	out, err := c.projectService.CreateProject(ctx, in, userID)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *projectController) deleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput) (*dto.DeleteProjectByRefOutput, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	out, err := c.projectService.DeleteProjectByRef(ctx, in, *userID)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *projectController) getUsersProjects(ctx context.Context, in *dto.GetUsersProjectsInput) (*dto.GetUsersProjectsOutput, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projects, err := c.projectService.GetUsersProjects(ctx, *userID)
	if err != nil {
		return nil, err
	}

	out := &dto.GetUsersProjectsOutput{}
	out.Body.Projects = projects

	return out, nil
}

func (c *projectController) resetDatabasePassword(ctx context.Context, in *dto.ResetDatabasePasswordInput) (*dto.ResetDatabasePasswordOutput, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	out, err := c.projectService.ResetDatabasePassword(ctx, in, *userID)
	if err != nil {
		return nil, err
	}

	return out, nil
}
