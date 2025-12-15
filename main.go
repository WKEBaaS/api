package main

import (
	"context"
	"log"
	"reflect"
	"time"

	"baas-api/config"
	"baas-api/controllers"
	"baas-api/repo"
	"baas-api/router"
	"baas-api/services/kubeproject"
	"baas-api/services/pgrest"
	"baas-api/services/project"
	"baas-api/services/s3"
	"baas-api/services/usersdb"

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
	_, _ = di.RegisterBean("entityRepository", reflect.TypeFor[*repo.EntityRepository]())
	_, _ = di.RegisterBean("projectRepository", reflect.TypeFor[*repo.ProjectRepository]())
	_, _ = di.RegisterBean("projectAuthSettingRepository", reflect.TypeFor[*repo.ProjectAuthSettingRepository]())

	//////////// Services //////////
	_, _ = di.RegisterBean("pgrestService", reflect.TypeFor[*pgrest.PgRestService]())
	_, _ = di.RegisterBean("s3Service", reflect.TypeFor[*s3.S3Service]())
	_, _ = di.RegisterBean("kubeProjectService", reflect.TypeFor[*kubeproject.KubeProjectService]())
	_, _ = di.RegisterBean("projectService", reflect.TypeFor[*project.ProjectService]())
	_, _ = di.RegisterBean("usersdbService", reflect.TypeFor[*usersdb.UsersDBService]())

	//////////// Controllers //////////
	_, _ = di.RegisterBean("projectController", reflect.TypeFor[*controllers.ProjectController]())

	/////////// Initialize Container //////////
	_ = di.InitializeContainer()
}

func main() {
	config := di.GetInstance("config").(*config.Config)
	projectController := di.GetInstance("projectController").(controllers.ProjectControllerInterface)
	cli := router.NewAPICli(config, projectController)

	cli.Run()
}
