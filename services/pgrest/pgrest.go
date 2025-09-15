// Package pgrest implements a RESTful API server for PostgreSQL databases.
package pgrest

import (
	"baas-api/config"
	"context"
)

type PgRestServiceInterface interface {
	// Project Management
	CreateProject(ctx context.Context, jwt string, name string, description string) (*CreateProjectOutput, error)
	// DeleteProject deletes a project by its ID.
	//
	// Returns the `reference` of the deleted project.
	DeleteProject(ctx context.Context, jwt string, id string) (*string, error)
}

type PgRestService struct {
	config *config.Config `di.inject:"config"`
}

func NewPgRestService() PgRestServiceInterface {
	return &PgRestService{}
}
