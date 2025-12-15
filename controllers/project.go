package controllers

import (
	"context"
	"net/http"
	"time"

	"baas-api/config"
	"baas-api/controllers/middlewares"
	"baas-api/dto"
	"baas-api/services/kubeproject"
	"baas-api/services/project"
	"baas-api/services/usersdb"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
)

type ProjectControllerInterface interface {
	RegisterProjectAPIs(api huma.API)
	getProjectByRef(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetProjectByRefOutput, error)
	patchProjectSettings(ctx context.Context, in *dto.UpdateProjectInput) (*struct{}, error)
	getProjectStatus(ctx context.Context, in *dto.GetProjectByRefInput, send sse.Sender)
	getProjectSettings(ctx context.Context, in *dto.GetProjectSettingsInput) (*dto.GetProjectSettingsOutput, error)
	createProject(ctx context.Context, in *dto.CreateProjectInput) (*dto.CreateProjectOutput, error)
	deleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByIDInput) (*dto.DeleteProjectByIDOutput, error)
	resetDatabasePassword(ctx context.Context, in *dto.ResetDatabasePasswordInput) (*dto.ResetDatabasePasswordOutput, error)
	getUsersProjects(ctx context.Context, in *dto.GetUsersProjectsInput) (*dto.GetUsersProjectsOutput, error)
	getUsersRootClasses(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetUsersFirstLevelClassesOutput, error)
	getUsersChildClasses(ctx context.Context, in *dto.GetUsersChildClassesInput) (*dto.GetUsersChildClassesOutput, error)
	getUsersClassByID(ctx context.Context, in *dto.GetUsersClassByIDInput) (*dto.GetUsersClassByIDOutput, error)
	getUsersClassPermissions(ctx context.Context, in *dto.GetUsersClassPermissionsInput) (*dto.GetUsersClassPermissionsOutput, error)
	updateUsersClassPermissions(ctx context.Context, in *dto.UpdateUsersClassPermissionsInput) (*struct{}, error)
}

type ProjectController struct {
	config         *config.Config                          `di.inject:"config"`
	kubeProject    kubeproject.KubeProjectServiceInterface `di.inject:"kubeProjectService"`
	projectService project.ProjectServiceInterface         `di.inject:"projectService"`
	usersdb        usersdb.UsersDBServiceInterface         `di.inject:"usersdbService"`
}

func NewProjectController() ProjectControllerInterface {
	return &ProjectController{}
}

func (c *ProjectController) RegisterProjectAPIs(api huma.API) {
	authMiddleware := middlewares.NewAuthMiddleware(api, c.config)

	huma.Register(api, huma.Operation{
		OperationID: "project-test-any",
		Method:      "GET",
		Path:        "/project/test-any",
		Summary:     "Test Project REST Endpoint",
		Description: "",
		Tags:        []string{"Debug"},
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

	huma.Register(api, huma.Operation{
		OperationID: "get-users-root-classes",
		Method:      "GET",
		Path:        "/project/root-classes",
		Summary:     "Get User's Root Classes",
		Description: "Retrieve the root classes for the authenticated user in the specified project.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.getUsersRootClasses)

	huma.Register(api, huma.Operation{
		OperationID: "get-users-child-classes",
		Method:      "GET",
		Path:        "/project/child-classes",
		Summary:     "Get User's Child Classes",
		Description: "Retrieve the child classes for a given parent class ID (PCID) for the authenticated user in the specified project.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.getUsersChildClasses)

	huma.Register(api, huma.Operation{
		OperationID: "get-users-class-by-id",
		Method:      "GET",
		Path:        "/project/class-by-id",
		Summary:     "Get User's Class by ID",
		Description: "Retrieve a specific class by its ID for the authenticated user in the specified project.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.getUsersClassByID)

	huma.Register(api, huma.Operation{
		OperationID: "get-users-class-permissions",
		Method:      "GET",
		Path:        "/project/class-permissions",
		Summary:     "Get User's Class Permissions",
		Description: "Retrieve permissions for a specific class ID for the authenticated user in the specified project.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.getUsersClassPermissions)

	huma.Register(api, huma.Operation{
		OperationID: "update-users-class-permissions",
		Method:      "PUT",
		Path:        "/project/class-permissions",
		Summary:     "Update User's Class Permissions",
		Description: "Update permissions for a specific class ID for the authenticated user in the specified project.",
		Tags:        []string{"Project"},
		Middlewares: huma.Middlewares{authMiddleware},
	}, c.updateUsersClassPermissions)
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

func (c *ProjectController) patchProjectSettings(ctx context.Context, in *dto.UpdateProjectInput) (*struct{}, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	jwt, err := GetJWTFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = c.projectService.PatchProjectSettings(ctx, jwt, in, session.UserID)
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
	jwt, err := GetJWTFromContext(ctx)
	if err != nil {
		return nil, err
	}

	out, internalOut, err := c.projectService.CreateProject(ctx, in, jwt, &session.UserID)
	if err != nil {
		return nil, err
	}

	go func() {
		postCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		err := c.kubeProject.WaitClusterHealthy(postCtx, out.Body.Reference)
		if err != nil {
			return
		}
		c.projectService.CreateProjectPostInstall(postCtx, out.Body.Reference, internalOut.AuthSecret, internalOut.JWKSPublicKey)
	}()

	return out, nil
}

func (c *ProjectController) deleteProjectByRef(ctx context.Context, in *dto.DeleteProjectByIDInput) (*dto.DeleteProjectByIDOutput, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	jwt, err := GetJWTFromContext(ctx)
	if err != nil {
		return nil, err
	}

	out, err := c.projectService.DeleteProjectByID(ctx, jwt, in, session.UserID)
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

func (c *ProjectController) getUsersRootClasses(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetUsersFirstLevelClassesOutput, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	db, err := c.usersdb.GetDB(ctx, in.Ref, session.UserID, "superuser")
	if err != nil {
		return nil, err
	}

	classes, err := c.usersdb.GetRootClasses(ctx, db)
	if err != nil {
		return nil, err
	}

	out := &dto.GetUsersFirstLevelClassesOutput{}
	out.Body.Classes = classes
	return out, nil
}

func (c *ProjectController) getUsersChildClasses(ctx context.Context, in *dto.GetUsersChildClassesInput) (*dto.GetUsersChildClassesOutput, error) {
	// 1. Get User Session
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Connect to the specific project/tenant DB
	// Uses in.Ref to find the DB, and session.UserID for logging/permissions
	// "superuser" is kept consistent with your example, but adjust if role logic differs here
	db, err := c.usersdb.GetDB(ctx, in.Ref, session.UserID, "superuser")
	if err != nil {
		return nil, err
	}

	// 3. Call DAO
	// Pass the active 'db' connection and the 'pcid' from input
	classes, err := c.usersdb.GetChildClasses(ctx, db, in.PCID)
	if err != nil {
		return nil, err
	}

	// 4. Construct Output
	out := &dto.GetUsersChildClassesOutput{}
	out.Body.Classes = classes

	return out, nil
}

func (c *ProjectController) getUsersClassByID(ctx context.Context, in *dto.GetUsersClassByIDInput) (*dto.GetUsersClassByIDOutput, error) {
	// 1. Get User Session
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Connect to the specific project/tenant DB
	db, err := c.usersdb.GetDB(ctx, in.Ref, session.UserID, "superuser")
	if err != nil {
		return nil, err
	}

	// 3. Call DAO
	class, err := c.usersdb.GetClassByID(ctx, db, in.ClassID)
	if err != nil {
		return nil, err
	}

	// 4. Construct Output
	out := &dto.GetUsersClassByIDOutput{}
	out.Body.Class = *class

	return out, nil
}

func (c *ProjectController) getUsersClassPermissions(ctx context.Context, in *dto.GetUsersClassPermissionsInput) (*dto.GetUsersClassPermissionsOutput, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	db, err := c.usersdb.GetDB(ctx, in.Ref, session.UserID, "superuser")
	if err != nil {
		return nil, err
	}

	permissions, err := c.usersdb.GetClassPermissions(ctx, db, in.ClassID)
	if err != nil {
		return nil, err
	}

	out := &dto.GetUsersClassPermissionsOutput{}
	out.Body.Permissions = permissions

	return out, nil
}

func (c *ProjectController) updateUsersClassPermissions(ctx context.Context, in *dto.UpdateUsersClassPermissionsInput) (*struct{}, error) {
	session, err := GetSessionFromContext(ctx)
	if err != nil {
		return nil, err
	}

	db, err := c.usersdb.GetDB(ctx, in.Ref, session.UserID, "superuser")
	if err != nil {
		return nil, err
	}

	err = c.usersdb.UpdateClassPermissions(ctx, db, in.ClassID, in.Body)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
