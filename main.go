package main

import (
	"baas-api/config"
	"baas-api/controllers"
	"baas-api/repo"
	"baas-api/router"
	"baas-api/services/kube_project"
	"baas-api/services/pgrest"
	"baas-api/services/project"
	"context"
	"log"
	"reflect"
	"time"

	"github.com/goioc/di"
	"github.com/minio/madmin-go/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/patrickmn/go-cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

	//////////// Kubernetes //////////
	kc, err := clientcmd.BuildConfigFromFlags("", cfg.Kube.ConfigPath)
	if err != nil {
		log.Panicf("Failed to build kube config %+v", err)
	}
	clientset, err := kubernetes.NewForConfig(kc)
	if err != nil {
		log.Panicf("Failed to create kube client %+v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(kc)
	if err != nil {
		log.Panicf("Failed to create dynamic client %+v", err)
	}
	_, _ = di.RegisterBeanInstance("kubeConfig", kc)
	_, _ = di.RegisterBeanInstance("kubeClientset", clientset)
	_, _ = di.RegisterBeanInstance("kubeDynamicClient", dynamicClient)

	//////////// S3 Clients //////////
	minioClient, err := minio.New(cfg.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, ""),
		Secure: cfg.S3.UseSSL,
	})
	if err != nil {
		log.Panicf("Failed to create minio client %+v", err)
	}
	minioAdminClient, err := madmin.NewWithOptions(cfg.S3.Endpoint, &madmin.Options{
		Creds:  credentials.NewStaticV4(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, ""),
		Secure: cfg.S3.UseSSL,
	})
	if err != nil {
		log.Panicf("Failed to create minio admin client %+v", err)
	}
	_, _ = di.RegisterBeanInstance("minioClient", minioClient)
	_, _ = di.RegisterBeanInstance("minioAdminClient", minioAdminClient)

	//////////// Repositories //////////
	_, _ = di.RegisterBean("entityRepository", reflect.TypeOf((*repo.EntityRepository)(nil)))
	_, _ = di.RegisterBean("projectRepository", reflect.TypeOf((*repo.ProjectRepository)(nil)))
	_, _ = di.RegisterBean("projectAuthSettingRepository", reflect.TypeOf((*repo.ProjectAuthSettingRepository)(nil)))

	//////////// Services //////////
	_, _ = di.RegisterBean("pgrestService", reflect.TypeOf((*pgrest.PgRestService)(nil)))
	_, _ = di.RegisterBean("kubeProjectService", reflect.TypeOf((*kube_project.KubeProjectService)(nil)))
	_, _ = di.RegisterBean("projectService", reflect.TypeOf((*project.ProjectService)(nil)))

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
