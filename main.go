package main

import (
	"baas-api/config"
	"baas-api/controllers"
	"baas-api/repo"
	"baas-api/router"
	"baas-api/services"
	"baas-api/services/kube_project"
	"context"
	"reflect"
	"time"

	"github.com/goioc/di"
	"github.com/patrickmn/go-cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func init() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	/////////// APP Config //////////
	_, _ = di.RegisterBeanInstance("config", cfg)

	//////////// Gorm Database //////////
	_, _ = di.RegisterBeanFactory("db", di.Singleton, func(ctx context.Context) (any, error) {
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
	})
	//////////// Cache //////////
	_, _ = di.RegisterBeanFactory("cache", di.Singleton, func(ctx context.Context) (any, error) {
		return cache.New(15*time.Minute, 20*time.Minute), nil
	})

	//////////// Repositories //////////
	_, _ = di.RegisterBean("entityRepository", reflect.TypeOf((*repo.EntityRepository)(nil)))
	_, _ = di.RegisterBean("projectRepository", reflect.TypeOf((*repo.ProjectRepository)(nil)))
	_, _ = di.RegisterBean("projectAuthSettingRepository", reflect.TypeOf((*repo.ProjectAuthSettingRepository)(nil)))

	//////////// Services //////////
	_, _ = di.RegisterBean("kubeProjectService", reflect.TypeOf((*kube_project.KubeProjectService)(nil)))
	_, _ = di.RegisterBean("projectService", reflect.TypeOf((*services.ProjectService)(nil)))

	//////////// Controllers //////////
	_, _ = di.RegisterBean("projectController", reflect.TypeOf((*controllers.ProjectController)(nil)))

	/////////// Initialize Container //////////
	_ = di.InitializeContainer()
}

func main() {
	config := di.GetInstance("config").(*config.Config)
	projectController := di.GetInstance("projectController").(controllers.ProjectControllerInterface)
	cli := router.NewAPICli(config, projectController)

	cli.Run()
}
