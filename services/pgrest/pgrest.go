// Package pgrest implements a RESTful API server for PostgreSQL databases.
package pgrest

import (
	"context"
	"encoding/json"

	"baas-api/config"

	"github.com/samber/do/v2"
)

type PgRestServiceInterface interface {
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
}

type PgRestService struct {
	config *config.Config
}

func NewPgRestService(i do.Injector) (PgRestServiceInterface, error) {
	return &PgRestService{
		config: do.MustInvoke[*config.Config](i),
	}, nil
}

var Package = do.Package(
	do.Lazy(NewPgRestService),
)

type PgRestError struct {
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Detail  *string `json:"details"`
	Hint    *string `json:"hint"`
}

func (s *PgRestService) UnmarshalPgRestError(data []byte) (*PgRestError, error) {
	var err PgRestError
	if err := json.Unmarshal(data, &err); err != nil {
		return nil, err
	}
	return &err, nil
}
