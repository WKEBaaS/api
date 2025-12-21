// Package classfunc provides utilities for handling generate class functions.
package classfunc

import (
	"context"

	"baas-api/internal/dto"
	"baas-api/internal/middlewares"
	"baas-api/internal/pggen"
	"baas-api/internal/utils"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/do/v2"
)

type Controller interface {
	RegisterCreateClassFunc(api huma.API)
}

type controller struct {
	authMiddleware middlewares.AuthMiddleware
	pggen          pggen.Service
}

var _ Controller = (*controller)(nil)

func NewController(i do.Injector) (*controller, error) {
	return &controller{
		authMiddleware: do.MustInvoke[middlewares.AuthMiddleware](i),
		pggen:          do.MustInvokeAs[pggen.Service](i),
	}, nil
}

func (c *controller) RegisterCreateClassFunc(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "create-class-function",
		Method:      "POST",
		Path:        "/project/create-classes-function",
		Summary:     "Create a New Class Function",
		Description: "Generate a SQL function to create a new class api in the users database.",
		Tags:        []string{"UsersDB"},
		Middlewares: huma.Middlewares{c.authMiddleware},
	}, func(ctx context.Context, in *dto.CreateClassFunctionInput) (*struct{}, error) {
		jwt, err := utils.GetJWTFromContext(ctx)
		if err != nil {
			return nil, err
		}

		err = c.pggen.GenerateCreateClassFunction(ctx, jwt, in)
		if err != nil {
			return nil, err
		}

		return nil, nil
	})
}
