package services

import (
	"baas-api/internal/configs"
	"baas-api/internal/repo"
	"baas-api/internal/services/hasura"
)

type Service struct {
	config *configs.Config
	Hasura *hasura.HasuraService
	Repo   *repo.Repository
}

func InitService(config *configs.Config, repo *repo.Repository) *Service {
	service := &Service{}
	service.config = config
	service.Hasura = hasura.InitHasuraService(config)
	service.Repo = repo

	repo.CreateProject("Test Project", "1234567890")
	return service
}
