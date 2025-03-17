package main

import (
	"i3s-service/internal/configs"
	"i3s-service/internal/router"
	"i3s-service/internal/services"
)

func main() {
	config := configs.LoadConfig()
	service := services.InitService(config)

	cli := router.InitAPI(config, service)

	cli.Run()
}
