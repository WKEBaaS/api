// Package usersdb
package usersdb

import (
	"context"

	"baas-api/internal/dto"
	"baas-api/internal/middlewares"
	"baas-api/internal/utils"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/do/v2"
)

type Controller interface {
	RegisterGetRootClasses(api huma.API)
	RegisterGetClassChildren(api huma.API)
	RegisterGetClassByID(api huma.API)
	RegisterGetClassPermissions(api huma.API)
	RegisterGetClassesChildBatched(api huma.API)
	RegisterUpdateUsersClassPermissions(api huma.API)
	// RegisterCreateClassAPI(api huma.API)
}

type controller struct {
	authMiddleware middlewares.AuthMiddleware
	usersdb        Service
}

var _ Controller = (*controller)(nil)

func NewController(i do.Injector) (Controller, error) {
	return &controller{
		authMiddleware: do.MustInvoke[middlewares.AuthMiddleware](i),
		usersdb:        do.MustInvokeAs[Service](i),
	}, nil
}

func (c *controller) RegisterGetRootClasses(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-root-classes",
		Method:      "GET",
		Path:        "/project/root-classes",
		Summary:     "Get User's Root Classes",
		Description: "Retrieve the root classes for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetUsersFirstLevelClassesOutput, error) {
		session, err := utils.GetSessionFromContext(ctx)
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
	})
}

func (c *controller) RegisterGetClassChildren(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-class-children",
		Method:      "GET",
		Path:        "/project/class-children",
		Summary:     "Get User's Class Children",
		Description: "Retrieve the child classes for a given parent class ID (PCID) for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetUsersChildClassesInput) (*dto.GetUsersChildClassesOutput, error) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			return nil, err
		}

		db, err := c.usersdb.GetDB(ctx, in.Ref, session.UserID, "superuser")
		if err != nil {
			return nil, err
		}
		classes, err := c.usersdb.GetChildClasses(ctx, db, in.PCID)
		if err != nil {
			return nil, err
		}

		out := &dto.GetUsersChildClassesOutput{}
		out.Body.Classes = classes

		return out, nil
	})
}

func (c *controller) RegisterGetClassByID(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-class-by-id",
		Method:      "GET",
		Path:        "/project/class-by-id",
		Summary:     "Get User's Class by ID",
		Description: "Retrieve a specific class by its ID for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetUsersClassByIDInput) (*dto.GetUsersClassByIDOutput, error) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			return nil, err
		}

		db, err := c.usersdb.GetDB(ctx, in.Ref, session.UserID, "superuser")
		if err != nil {
			return nil, err
		}

		class, err := c.usersdb.GetClassByID(ctx, db, in.ClassID)
		if err != nil {
			return nil, err
		}

		out := &dto.GetUsersClassByIDOutput{}
		out.Body.Class = *class

		return out, nil
	})
}

func (c *controller) RegisterGetClassPermissions(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-class-permissions",
		Method:      "GET",
		Path:        "/project/class-permissions",
		Summary:     "Get User's Class Permissions",
		Description: "Retrieve permissions for a specific class ID for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetUsersClassPermissionsInput) (*dto.GetUsersClassPermissionsOutput, error) {
		session, err := utils.GetSessionFromContext(ctx)
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
	})
}

func (c *controller) RegisterGetClassesChildBatched(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-classes-child-batched",
		Method:      "GET",
		Path:        "/project/child-classes-batched",
		Summary:     "Get User's Child Classes (Batched)",
		Description: "Retrieve the child classes for a given parent class ID (PCID) for the authenticated user in the specified project, batched.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetUsersClassesChildInput) (*dto.GetUsersClassesChildOutput, error) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			return nil, err
		}

		db, err := c.usersdb.GetDB(ctx, in.Ref, session.UserID, "superuser")
		if err != nil {
			return nil, err
		}

		classes, err := c.usersdb.GetClassesChild(ctx, db, in.ClassIDs)
		if err != nil {
			return nil, err
		}

		out := &dto.GetUsersClassesChildOutput{}
		out.Body.Classes = classes

		return out, nil
	})
}

func (c *controller) RegisterUpdateUsersClassPermissions(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "update-users-class-permissions",
		Method:      "PUT",
		Path:        "/project/class-permissions",
		Summary:     "Update User's Class Permissions",
		Description: "Update permissions for a specific class ID for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.UpdateUsersClassPermissionsInput) (*struct{}, error) {
		session, err := utils.GetSessionFromContext(ctx)
		if err != nil {
			return nil, err
		}

		db, err := c.usersdb.GetDB(ctx, in.Body.Ref, session.UserID, "superuser")
		if err != nil {
			return nil, err
		}

		err = c.usersdb.UpdateClassPermissions(ctx, db, in.Body.ClassID, in.Body.Permissions)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
}
