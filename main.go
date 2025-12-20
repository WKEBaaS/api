package main

import (
	"baas-api/cache"
	"baas-api/config"
	"baas-api/controllers"
	"baas-api/database"
	"baas-api/repo"
	"baas-api/router"
	"baas-api/services/kubeproject"
	"baas-api/services/pgrest"
	"baas-api/services/project"
	"baas-api/services/s3"
	"baas-api/services/usersdb"

	"github.com/samber/do/v2"
)

func main() {
	i := do.New()

	// Config
	config.Package(i)
	cache.Package(i)
	database.Package(i)

	// Repositories
	repo.Package(i)

	// Services
	s3.Package(i)
	pgrest.Package(i)
	kubeproject.Package(i)
	usersdb.Package(i)
	project.Package(i)

	// Controllers
	controllers.Package(i)

	// Router
	router.Package(i)

	router := do.MustInvoke[*router.BaaSRouter](i)
	router.RegisterControllers()
	router.Start()
}
