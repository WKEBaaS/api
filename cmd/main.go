package main

import (
	"baas-api/internal/configs"
	"baas-api/internal/controllers"
	"baas-api/internal/i3s"
	"baas-api/internal/repo"
	"baas-api/internal/router"
	"baas-api/internal/services"
	"log"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func main() {
	config := configs.LoadConfig()

	//////////// Init Gorm Database //////////
	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
			NoLowerCase:   false,
		},
	})
	if err != nil {
		panic(err)
	}

	//////////// Migrate I3S Schema //////////
	i3s := i3s.NewI3S(config)
	if err := i3s.Migrate(); err != nil {
		log.Fatalf("failed to migrate database: %v\n", err)
	}

	//////////// Init Cache //////////
	cache := cache.New(15*time.Minute, 20*time.Minute)

	//////////// Init Repo, Service //////////
	// Repositories
	projectRepo := repo.NewProjectRepository(db)
	kubeProjectRepo := repo.NewKubeProjectRepository(config)
	entityRepo := repo.NewEntityRepository(db, cache)
	userRepo := repo.NewUserRepository(db, cache)
	// Services
	projectService := services.NewProjectService(config, projectRepo, kubeProjectRepo)
	authService := services.NewAuthService(config, entityRepo, userRepo)

	//////////// Init Controllers //////////
	authController := controllers.NewAuthController(config, authService)
	projectController := controllers.NewProjectController(config, projectService)

	cli := router.NewApiCli(config, authController, projectController)

	cli.Run()
}
