package main

import (
	"baas-api/internal/configs"
	"baas-api/internal/repo"
	"baas-api/internal/router"
	"baas-api/internal/services"
)

func main() {
	config := configs.LoadConfig()

	//////////// Init Repo, Service //////////
	repo := repo.InitRepository(config)
	service := services.InitService(config, repo)

	cli := router.InitAPI(config, service, repo)

	cli.Run()
}
