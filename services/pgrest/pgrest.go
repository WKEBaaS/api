// Package pgrest implements a RESTful API server for PostgreSQL databases.
package pgrest

import (
	"baas-api/config"
	"context"
	"encoding/json"
)

type PgRestServiceInterface interface {
	// Project Management
	CreateProject(ctx context.Context, jwt string, name string, description string) (*CreateProjectOutput, error)
	// DeleteProject deletes a project by its ID.
	//
	// Returns the `reference` of the deleted project.
	DeleteProject(ctx context.Context, jwt string, id string) (*DeleteProjectOutput, error)
}

type PgRestService struct {
	config *config.Config `di.inject:"config"`
}

func NewPgRestService() PgRestServiceInterface {
	return &PgRestService{}
}

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
