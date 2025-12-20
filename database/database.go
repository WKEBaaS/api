// Package database provides database connection using GORM and PostgreSQL.
package database

import (
	"baas-api/config"

	"github.com/samber/do/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func NewGormDB(i do.Injector) (*gorm.DB, error) {
	cfg := do.MustInvoke[*config.Config](i)
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
			NoLowerCase:   false,
		},
	})
	if err != nil {
		panic(err)
	}
	return db, nil
}

var Package = do.Package(
	do.Lazy(NewGormDB),
)
