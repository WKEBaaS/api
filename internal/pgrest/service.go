// Package pgrest implements a RESTful API server for PostgreSQL databases.
package pgrest

import (
	"context"
	"encoding/json"
	"log/slog"

	"baas-api/internal/config"
	"baas-api/internal/dto"

	"github.com/danielgtaylor/huma/v2"
	"github.com/samber/do/v2"
)

type Service interface {
	// Project Management
	CreateProject(ctx context.Context, jwt string, name string, description string) (*CreateProjectOutput, error)
	// DeleteProject deletes a project by its ID.
	//
	// Returns the `reference` of the deleted project.
	DeleteProject(ctx context.Context, jwt string, id string) (*DeleteProjectOutput, error)
	// UpdateProject
	UpdateProject(ctx context.Context, jwt string, payload UpdateProjectPayload) (*UpdateProjectOutput, error)
	// CreateOrUpdateAuthProvider
	CreateOrUpdateAuthProvider(ctx context.Context, jwt string, payload CreateOrUpdateAuthProviderPayload) error
	// CheckProjectPermission checks if the user has permission to access the project.
	CheckProjectPermission(ctx context.Context, jwt string, projectID string) error
	CheckProjectPermissionByRef(ctx context.Context, jwt string, projectRef string) error

	CreateClassFunction(ctx context.Context, jwt string, in *dto.CreateClassFunctionInput) error
}

type service struct {
	config *config.Config
}

var _ Service = (*service)(nil)

func NewService(i do.Injector) (*service, error) {
	return &service{
		config: do.MustInvoke[*config.Config](i),
	}, nil
}

type PgRestError struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Detail  *string `json:"details"`
	Hint    *string `json:"hint"`
}

func (s *service) UnmarshalPgRestError(data []byte) (*PgRestError, error) {
	var err PgRestError
	if err := json.Unmarshal(data, &err); err != nil {
		return nil, err
	}
	return &err, nil
}

func (err PgRestError) ToHumaError() error {
	switch err.Code {
	case "PT401":
		return huma.Error401Unauthorized(err.Message)
	case "PT403":
		return huma.Error403Forbidden(err.Message)
	case "PT404":
		return huma.Error404NotFound(err.Message)
	case "PT409":
		return huma.Error409Conflict(err.Message)
	default:
		slog.Error("Unhandled PgRestError", "code", err.Code, "message", err.Message, "detail", err.Detail, "hint", err.Hint)
		return huma.Error500InternalServerError("Internal server error")
	}
}
