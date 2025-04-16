package controllers

import (
	"context"
	"i3s-service/internal/controllers/inputs"
	"i3s-service/internal/controllers/outputs"
	"i3s-service/internal/repo"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type AuthController struct {
	repo *repo.Repository
}

func InitAuthController(repo *repo.Repository) *AuthController {
	controller := &AuthController{}
	controller.repo = repo

	return controller
}

func (c *AuthController) RegisterAuthAPIs(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "auth-callback",
		Method:      http.MethodGet,
		Path:        "/auth/callback",
		Summary:     "Callback",
		Description: "Callback",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, input *inputs.AuthCallbackInput) (*outputs.AuthCallbackOutput, error) {
		resp := &outputs.AuthCallbackOutput{}

		userID, err := c.repo.GetUserIDByIdentity(ctx, "keycloak", "3bb24321-baa4-41f8-9bf5-c1eb16a3f336")
		if err != nil {
			slog.Error(err.Error())
			resp.Status = http.StatusNotFound
			resp.Body.Message = "user not found"
			return resp, nil
		}

		resp.Status = http.StatusOK
		resp.Body.Message = *userID
		return resp, nil
	})
}
