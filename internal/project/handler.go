// Package project implements the ProjectController for managing projects.
package project

import (
	"context"
	"net/http"
	"time"

	"baas-api/internal/config"
	"baas-api/internal/dto"
	"baas-api/internal/kubeproject"
	"baas-api/internal/middlewares"
	"baas-api/internal/utils"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/samber/do/v2"
)

type Controller interface {
	RegisterTestAny(api huma.API)
	RegisterGetProjectByRef(api huma.API)
	RegisterCreateProject(api huma.API)
	RegisterPatchProjectSettings(api huma.API)
	RegisterGetProjectSettings(api huma.API)
	RegisterGetProjectStatus(api huma.API)
	RegisterDeleteProjectByRef(api huma.API)
	RegisterGetUsersProjects(api huma.API)
	RegisterResetDatabasePassword(api huma.API)
}

type controller struct {
	config         *config.Config             `do:""`
	authMiddleware middlewares.AuthMiddleware `do:""`
	kube           kubeproject.Service        `do:""`
	project        Service                    `do:""`
}

var _ Controller = (*controller)(nil)

func NewController(i do.Injector) (*controller, error) {
	return &controller{
		authMiddleware: do.MustInvoke[middlewares.AuthMiddleware](i),
		config:         do.MustInvoke[*config.Config](i),
		kube:           do.MustInvokeAs[kubeproject.Service](i),
		project:        do.MustInvokeAs[Service](i),
	}, nil
}

func (c *controller) RegisterTestAny(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "project-test-any",
		Method:      "GET",
		Path:        "/project/test-any",
		Summary:     "Test Project REST Endpoint",
		Tags:        []string{"Debug"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *struct{}) (*struct{}, error) {
		return nil, nil
	})
}

func (c *controller) RegisterGetProjectByRef(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-project-by-ref",
		Method:      "GET",
		Path:        "/project/by-ref",
		Summary:     "Get Project by Reference",
		Description: "Retrieve a project by its reference. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetProjectByRefOutput, error) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := c.project.GetUserProjectByRef(ctx, in.Ref, session.UserID)
		if err != nil {
			return nil, err
		}
		return &dto.GetProjectByRefOutput{Body: *out}, nil
	})
}

func (c *controller) RegisterCreateProject(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "create-project",
		Method:      "POST",
		Path:        "/project",
		Summary:     "Create Project",
		Description: "Create a new project with the specified name and storage size.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.CreateProjectInput) (*dto.CreateProjectOutput, error) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		out, internalOut, err := c.project.CreateProject(ctx, in, jwt, &session.UserID)
		if err != nil {
			return nil, err
		}

		go func() {
			postCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			err := c.kube.WaitClusterHealthy(postCtx, out.Body.Reference)
			if err != nil {
				return
			}
			_ = c.project.CreateProjectPostInstall(postCtx, out.Body.Reference, internalOut.AuthSecret, internalOut.JWKSPublicKey)
		}()

		return out, nil
	})
}

func (c *controller) RegisterPatchProjectSettings(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "patch-project-settings",
		Method:      http.MethodPatch,
		Path:        "/project/settings",
		Summary:     "Patch Project Settings",
		Description: "Update a project settings by its reference. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.UpdateProjectInput) (*struct{}, error) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		err = c.project.PatchProjectSettings(ctx, jwt, in, session.UserID)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
}

func (c *controller) RegisterGetProjectSettings(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-project-settings",
		Method:      "GET",
		Path:        "/project/settings/by-ref",
		Summary:     "Get Project Settings by Reference",
		Description: "Retrieve project settings including authentication and OAuth provider settings by reference.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetProjectSettingsInput) (*dto.GetProjectSettingsOutput, error) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		out, err := c.project.GetProjectSettings(ctx, in, session.UserID)
		if err != nil {
			return nil, err
		}
		return out, nil
	})
}

func (c *controller) RegisterGetProjectStatus(api huma.API) {
	sse.Register(api, huma.Operation{
		OperationID: "get-project-status",
		Method:      http.MethodGet,
		Path:        "/project/status",
		Summary:     "Get Project Status (SSE)",
		Description: "Get the status of a project by its reference. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, map[string]any{
		"project-status": dto.ProjectStatusEvent{},
		"error":          dto.ErrorEvent{},
	}, func(ctx context.Context, in *dto.GetProjectByRefInput, send sse.Sender) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			send.Data(dto.ErrorEvent{Message: "Unauthorized access"})
			return
		}

		dataChan := make(chan any, 1)
		go func() {
			defer close(dataChan)
			err := c.project.GetUserProjectStatusByRef(ctx, dataChan, in.Ref, session.UserID)
			if err != nil {
				dataChan <- dto.ErrorEvent{Message: err.Error()}
				return
			}
		}()

		for {
			select {
			case data, ok := <-dataChan:
				if !ok {
					return
				}
				if err := send.Data(data); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	})
}

func (c *controller) RegisterDeleteProjectByRef(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "delete-project-by-ref",
		Method:      "DELETE",
		Path:        "/project",
		Summary:     "Delete Project by ID",
		Description: "Delete a project by its ID.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.DeleteProjectByIDInput) (*dto.DeleteProjectByIDOutput, error) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		out, err := c.project.DeleteProjectByID(ctx, jwt, in, session.UserID)
		if err != nil {
			return nil, err
		}

		return out, nil
	})
}

func (c *controller) RegisterGetUsersProjects(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-projects",
		Method:      "GET",
		Path:        "/project/users",
		Summary:     "Get User's Projects",
		Description: "Retrieve all projects associated with the authenticated user.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetUsersProjectsInput) (*dto.GetUsersProjectsOutput, error) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			return nil, err
		}

		projects, err := c.project.GetUsersProjects(ctx, session.UserID)
		if err != nil {
			return nil, err
		}

		out := &dto.GetUsersProjectsOutput{}
		out.Body.Projects = projects

		return out, nil
	})
}

func (c *controller) RegisterResetDatabasePassword(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "reset-database-password",
		Method:      "POST",
		Path:        "/project/reset-database-password",
		Summary:     "Reset Database Password",
		Description: "Reset the database password for a project. The reference is a 20-character string.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.ResetDatabasePasswordInput) (*dto.ResetDatabasePasswordOutput, error) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			return nil, err
		}

		out, err := c.project.ResetDatabasePassword(ctx, in, session.UserID)
		if err != nil {
			return nil, err
		}

		return out, nil
	})
}
