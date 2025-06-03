package main

import (
	"baas-api/internal/configs"
	"baas-api/internal/repo"
	"baas-api/internal/router"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	config := configs.LoadConfig()

	//////////// Init Gorm Database //////////
	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{
		Logger: gormLogger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogger.Config{
			// SlowThreshold: time.Second,
			LogLevel: gormLogger.Info,
			Colorful: true,
		}),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
			NoLowerCase:   false,
		},
	})
	if err != nil {
		panic(err)
	}

	//////////// Init Repo, Service //////////
	projectRepo := repo.NewProjectRepository(db)
	// service := services.NewService(config, projectRepo)

	cli := router.NewAPI(config, projectRepo)

	cli.Run()
}
