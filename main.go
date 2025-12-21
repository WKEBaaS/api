package main

import (
	"baas-api/internal/authsetting"
	"baas-api/internal/cache"
	"baas-api/internal/config"
	"baas-api/internal/database"
	"baas-api/internal/kubeproject"
	"baas-api/internal/middlewares"
	"baas-api/internal/minio"
	"baas-api/internal/pgrest"
	"baas-api/internal/project"
	"baas-api/internal/router"
	"baas-api/internal/usersdb"

	"github.com/samber/do/v2"
)

func main() {
	i := do.New()

	// Config
	config.Package(i)
	cache.Package(i)
	database.Package(i)

	// Services
	minio.Package(i)
	pgrest.Package(i)
	kubeproject.Package(i)

	// Middlewares
	middlewares.Package(i)

	// Domains
	project.Package(i)
	authsetting.Package(i)
	usersdb.Package(i)

	// Router
	router.Package(i)

	router := do.MustInvoke[*router.BaaSRouter](i)
	router.RegisterControllers()
	router.Start()
}
