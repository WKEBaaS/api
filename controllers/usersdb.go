package controllers

import (
	"context"

	"baas-api/controllers/middlewares"
	"baas-api/dto"
	"baas-api/services/usersdb"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/do/v2"
)

type UsersDBControllerInterface interface {
	RegisterGetRootClasses(api huma.API)
	RegisterGetClassChildren(api huma.API)
	RegisterGetClassByID(api huma.API)
	RegisterGetClassPermissions(api huma.API)
	RegisterGetClassesChildBatched(api huma.API)
}

type UsersDBController struct {
	authMiddleware middlewares.AuthMiddleware      `do:""`
	usersdb        usersdb.UsersDBServiceInterface `do:""`
}

func NewUsersDBController(i do.Injector) (UsersDBControllerInterface, error) {
	return &UsersDBController{
		authMiddleware: do.MustInvoke[middlewares.AuthMiddleware](i),
		usersdb:        do.MustInvoke[usersdb.UsersDBServiceInterface](i),
	}, nil
}

func (c *UsersDBController) RegisterGetRootClasses(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-root-classes",
		Method:      "GET",
		Path:        "/project/root-classes",
		Summary:     "Get User's Root Classes",
		Description: "Retrieve the root classes for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetUsersFirstLevelClassesOutput, error) {
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
	})
}

func (c *UsersDBController) RegisterGetClassChildren(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-class-children",
		Method:      "GET",
		Path:        "/project/class-children",
		Summary:     "Get User's Class Children",
		Description: "Retrieve the child classes for a given parent class ID (PCID) for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetUsersChildClassesInput) (*dto.GetUsersChildClassesOutput, error) {
		session, err := GetSessionFromContext(ctx)
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

func (c *UsersDBController) RegisterGetClassByID(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-class-by-id",
		Method:      "GET",
		Path:        "/project/class-by-id",
		Summary:     "Get User's Class by ID",
		Description: "Retrieve a specific class by its ID for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetUsersClassByIDInput) (*dto.GetUsersClassByIDOutput, error) {
		session, err := GetSessionFromContext(ctx)
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

func (c *UsersDBController) RegisterGetClassPermissions(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-class-permissions",
		Method:      "GET",
		Path:        "/project/class-permissions",
		Summary:     "Get User's Class Permissions",
		Description: "Retrieve permissions for a specific class ID for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetUsersClassPermissionsInput) (*dto.GetUsersClassPermissionsOutput, error) {
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
	})
}

func (c *UsersDBController) RegisterGetClassesChildBatched(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-classes-child-batched",
		Method:      "GET",
		Path:        "/project/child-classes-batched",
		Summary:     "Get User's Child Classes (Batched)",
		Description: "Retrieve the child classes for a given parent class ID (PCID) for the authenticated user in the specified project, batched.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetUsersClassesChildInput) (*dto.GetUsersClassesChildOutput, error) {
		session, err := GetSessionFromContext(ctx)
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
