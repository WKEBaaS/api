package services

import (
	"context"
	"i3s-service/internal/configs"
	"i3s-service/internal/services/hasura"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	config *configs.Config
	DB     *pgxpool.Pool
	Hasura *hasura.HasuraService
}

func InitService(config *configs.Config) *Service {
	db, err := pgxpool.New(context.Background(), config.DatabaseURL)

	if err != nil {
		panic(err)
	}

	service := &Service{}
	service.config = config
	service.DB = db
	service.Hasura = hasura.InitHasuraService(config)

	return service
}
