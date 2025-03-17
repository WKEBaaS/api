package services

import (
	"context"
	"i3s-service/internal/configs"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	DB *pgxpool.Pool
}

func InitService(config *configs.Config) *Service {
	db, err := pgxpool.New(context.Background(), config.DatabaseURL)

	if err != nil {
		panic(err)
	}

	service := &Service{}
	service.DB = db

	return service
}
