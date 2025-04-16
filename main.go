package main

import (
	"i3s-service/internal/configs"
	"i3s-service/internal/repo"
	"i3s-service/internal/router"
	"i3s-service/internal/services"
)

func main() {
	config := configs.LoadConfig()

	//////////// Init Repo, Service //////////
	repo := repo.InitRepository(config)
	service := services.InitService(config)

	cli := router.InitAPI(config, repo, service)

	cli.Run()
}
