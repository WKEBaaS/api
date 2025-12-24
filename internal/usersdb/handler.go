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
	RegisterGetRootClass(api huma.API)
	RegisterGetRootClasses(api huma.API)
	RegisterGetClassChildren(api huma.API)
	RegisterGetClassByID(api huma.API)
	RegisterGetClassPermissions(api huma.API)
	RegisterGetClassesChildBatched(api huma.API)
	RegisterUpdateUsersClassPermissions(api huma.API)
	RegisterCreateClass(api huma.API)
	RegisterDeleteUsersClass(api huma.API)
}

type controller struct {
	authMiddleware middlewares.AuthMiddleware
	usersdb        Service
}

var _ Controller = (*controller)(nil)

func NewController(i do.Injector) (*controller, error) {
	return &controller{
		authMiddleware: do.MustInvoke[middlewares.AuthMiddleware](i),
		usersdb:        do.MustInvokeAs[Service](i),
	}, nil
}

func (c *controller) RegisterGetRootClass(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-users-root-class",
		Method:      "GET",
		Path:        "/project/root-class",
		Summary:     "Get User's Root Class",
		Description: "Retrieve the root class for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetUsersRootClassOutput, error) {
		// 1. 取得 JWT
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		// 2. 直接呼叫 Service，不再需要先 GetDB，role 也不用傳 (Service 內定為 superuser)
		class, err := c.usersdb.GetRootClass(ctx, jwt, in.Ref)
		if err != nil {
			return nil, err
		}

		out := &dto.GetUsersRootClassOutput{}
		out.Body.Class = *class
		return out, nil
	})
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
	}, func(ctx context.Context, in *dto.GetProjectByRefInput) (*dto.GetUsersRootClassesOutput, error) {
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		classes, err := c.usersdb.GetRootClasses(ctx, jwt, in.Ref)
		if err != nil {
			return nil, err
		}

		out := &dto.GetUsersRootClassesOutput{}
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
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		classes, err := c.usersdb.GetChildClasses(ctx, jwt, in.Ref, in.PCID)
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
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		class, err := c.usersdb.GetClassByID(ctx, jwt, in.Ref, in.ClassID)
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
		// 改用 GetJWTFromContext，因為 Service 層需要 JWT 字串來連線 DB
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		permissions, err := c.usersdb.GetClassPermissions(ctx, jwt, in.Ref, in.ClassID)
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
		// 改用 GetJWTFromContext
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		classes, err := c.usersdb.GetClassesChild(ctx, jwt, in.Ref, in.ClassIDs)
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
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		// 參數從 Body 中獲取
		err = c.usersdb.UpdateClassPermissions(ctx, jwt, in.Body.Ref, in.Body.ClassID, in.Body.Permissions)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
}

func (c *controller) RegisterCreateClass(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "create-users-class",
		Method:      "POST",
		Path:        "/project/class",
		Summary:     "Create User's Class",
		Description: "Create a new class under a specified parent class ID (PCID) for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.CreateClassInput) (*dto.CreateClassOutput, error) {
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		class, err := c.usersdb.CreateClass(ctx, jwt, in)
		if err != nil {
			return nil, err
		}

		out := &dto.CreateClassOutput{}
		out.Body.Class = *class
		return out, nil
	})
}

func (c *controller) RegisterDeleteUsersClass(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "delete-users-class",
		Method:      "DELETE",
		Path:        "/project/class",
		Summary:     "Delete User's Class",
		Description: "Delete a class by its ID for the authenticated user in the specified project.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.DeleteClassInput) (*struct{}, error) {
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		err = c.usersdb.DeleteClass(ctx, jwt, in)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
}
