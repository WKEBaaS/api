package repo

import (
	"baas-api/internal/configs"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Repository struct {
	DB *gorm.DB
}

func InitRepository(config *configs.Config) *Repository {
	repo := &Repository{}

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

	repo.DB = db
	return repo
}
