package controllers

import (
	"baas-api/config"
	"baas-api/controllers/middlewares"
	"baas-api/dto"
	"baas-api/services"
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
)

type ProjectControllerInterface interface {
	RegisterProjectAPIs(api huma.API)
	getProjectByRef(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetProjectByRefOutput, error)
	patchProjectSettings(ctx context.Context, in *dto.PatchProjectSettingInput) (*struct{}, error)
	getProjectStatus(ctx context.Context, in *dto.GetProjectByRefInput, send sse.Sender)
	getProjectSettings(ctx context.Context, in *dto.GetProjectSettingsInput) (*dto.GetProjectSettingsOutput, error)
	createProject(ctx context.Context, in *dto.CreateProjectInput) (*dto.CreateProjectOutput, error)
	deleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput) (*dto.DeleteProjectByRefOutput, error)
	getUsersProjects(ctx context.Context, in *dto.GetUsersProjectsInput) (*dto.GetUsersProjectsOutput, error)
	resetDatabasePassword(ctx context.Context, in *dto.ResetDatabasePasswordInput) (*dto.ResetDatabasePasswordOutput, error)
}

type ProjectController struct {
	config         *config.Config                   `di.inject:"config"`
	projectService services.ProjectServiceInterface `di.inject:"projectService"`
}

func NewProjectController() ProjectControllerInterface {
	return &ProjectController{}
}

func (c *ProjectController) RegisterProjectAPIs(api huma.API) {
	authMiddleware := middlewares.NewAuthMiddleware(api, c.config)

	huma.Register(api, huma.Operation{
		OperationID: "get-project-by-ref",
		Method:      "GET",
		Path:        "/project/by-ref",
		Summary:     "Get Project by Reference",
		Description: "Retrieve a project by its reference. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.getProjectByRef)

	huma.Register(api, huma.Operation{
		OperationID: "create-project",
		Method:      "POST",
		Path:        "/project",
		Summary:     "Create Project",
		Description: "Create a new project with the specified name and storage size.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.createProject)

	huma.Register(api, huma.Operation{
		OperationID: "patch-project-settings",
		Method:      http.MethodPatch,
		Path:        "/project/settings",
		Summary:     "Patch Project Settings",
		Description: "Update a project settings by its reference. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.patchProjectSettings)

	huma.Register(api, huma.Operation{
		OperationID: "get-project-settings",
		Method:      "GET",
		Path:        "/project/settings/by-ref",
		Summary:     "Get Project Settings by Reference",
		Description: "Retrieve project settings including authentication and OAuth provider settings by reference.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.getProjectSettings)

	sse.Register(api, huma.Operation{
		OperationID: "get-project-status",
		Method:      http.MethodGet,
		Path:        "/project/status",
		Summary:     "Get Project Status (SSE)",
		Description: "Get the status of a project by its reference. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	},
		map[string]any{
			"project-status": dto.ProjectStatusEvent{},
			"error":          dto.ErrorEvent{},
		}, c.getProjectStatus)

	huma.Register(api, huma.Operation{
		OperationID: "delete-project-by-ref",
		Method:      "DELETE",
		Path:        "/project/by-ref",
		Summary:     "Delete Project by Reference",
		Description: "Delete a project by its reference. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.deleteProjectByRef)

	huma.Register(api, huma.Operation{
		OperationID: "get-users-projects",
		Method:      "GET",
		Path:        "/project/users",
		Summary:     "Get User's Projects",
		Description: "Retrieve all projects associated with the authenticated user.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.getUsersProjects)

	huma.Register(api, huma.Operation{
		OperationID: "reset-database-password",
		Method:      "POST",
		Path:        "/project/reset-database-password",
		Summary:     "Reset Database Password",
		Description: "Reset the database password for a project. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.resetDatabasePassword)
}

func (c *ProjectController) getProjectByRef(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetProjectByRefOutput, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.projectService.GetUserProjectByRef(ctx, in.Ref, session.UserID)
	if err != nil {
		return nil, err
	}
	return &dto.GetProjectByRefOutput{Body: *out}, nil
}

func (c *ProjectController) patchProjectSettings(ctx context.Context, in *dto.PatchProjectSettingInput) (*struct{}, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	err = c.projectService.PatchProjectSettings(ctx, in, session.UserID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (c *ProjectController) getProjectSettings(ctx context.Context, in *dto.GetProjectSettingsInput) (*dto.GetProjectSettingsOutput, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.projectService.GetProjectSettings(ctx, in, session.UserID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ProjectController) getProjectStatus(ctx context.Context, in *dto.GetProjectByRefInput, send sse.Sender) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		send.Data(dto.ErrorEvent{Message: "Unauthorized access"})
		return
	}

	dataChan := make(chan any, 1)

	go func() {
		defer close(dataChan)

		err := c.projectService.GetUserProjectStatusByRef(ctx, dataChan, in.Ref, session.UserID)
		if err != nil {
			dataChan <- dto.ErrorEvent{Message: err.Error()}
			return
		}
	}()

	for {
		select {
		case data, ok := <-dataChan:
			if !ok {
				// Channel was closed, so we are done.
				return
			}
			if err := send.Data(data); err != nil {
				return
			}
		case <-ctx.Done():
			// Context was canceled, so we are done.
			return
		}
	}
}

func (c *ProjectController) createProject(ctx context.Context, in *dto.CreateProjectInput) (*dto.CreateProjectOutput, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	out, err := c.projectService.CreateProject(ctx, in, &session.UserID)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ProjectController) deleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByRefInput) (*dto.DeleteProjectByRefOutput, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	out, err := c.projectService.DeleteProjectByRef(ctx, in, session.UserID)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *ProjectController) getUsersProjects(ctx context.Context, in *dto.GetUsersProjectsInput) (*dto.GetUsersProjectsOutput, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	projects, err := c.projectService.GetUsersProjects(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	out := &dto.GetUsersProjectsOutput{}
	out.Body.Projects = projects

	return out, nil
}

func (c *ProjectController) resetDatabasePassword(ctx context.Context, in *dto.ResetDatabasePasswordInput) (*dto.ResetDatabasePasswordOutput, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	out, err := c.projectService.ResetDatabasePassword(ctx, in, session.UserID)
	if err != nil {
		return nil, err
	}

	return out, nil
}
