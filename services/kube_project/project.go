// Package kube
//
// kubernetes related repository for project management
package kube_project

import (
	"baas-api/config"
	"context"
	"log"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeProjectServiceInterface interface {
	// === 基礎設施層 ===
	// Prepare Cluster Required Resources
	CreateJWKSConfigMap(ctx context.Context, ref string, publicKey string, privateKey string) error
	DeleteJWKSConfigMap(ctx context.Context, ref string) error

	// CNPG Cluster 管理
	CreateCluster(ctx context.Context, ref string, storageSize string) error
	DeleteCluster(ctx context.Context, ref string) error
	FindClusterStatus(ctx context.Context, ref string) (*string, error)
	WaitClusterHealthy(ctx context.Context, ref string) error

	// Database Management
	CreateDatabase(ctx context.Context, ref string) error
	DeleteDatabase(ctx context.Context, ref string) error
	CreateMigrationJob(ctx context.Context, ref string) error

	// Database Role Management
	FindDatabaseRolePassword(ctx context.Context, ref, role string) (*string, error)
	CreateDatabaseRoleSecret(ctx context.Context, ref, role, password string) error
	UpdateDatabaseRoleSecret(ctx context.Context, ref, role, password string) error
	DeleteDatabaseRoleSecret(ctx context.Context, ref string, role string) error

	// === 應用層 ===
	// Auth API
	CreateAuthAPIDeployment(ctx context.Context, ref string, opt *APIDeploymentOption) error
	DeleteAuthAPIDeployment(ctx context.Context, ref string) error
	PatchAuthAPIDeployment(ctx context.Context, ref string, opt *APIDeploymentOption) error
	CreateAuthAPIService(ctx context.Context, ref string) error
	DeleteAuthAPIService(ctx context.Context, ref string) error

	// REST API (PostgREST)
	CreateRESTAPIDeployment(ctx context.Context, ref string, jwks string) error
	DeleteRESTAPIDeployment(ctx context.Context, ref string) error

	// === 網路層 ===
	// Ingress for REST API (PostgREST) and Auth API
	CreateIngressRoute(ctx context.Context, ref string) error
	DeleteIngressRoute(ctx context.Context, ref string) error
	CreateIngressRouteTCP(ctx context.Context, ref string) error
	DeleteIngressRouteTCP(ctx context.Context, ref string) error
}

type KubeProjectService struct {
	kubeConfig    *rest.Config           `di.inject:"kubeConfig"`
	clientset     *kubernetes.Clientset  `di.inject:"kubeClientset"`
	dynamicClient *dynamic.DynamicClient `di.inject:"kubeDynamicClient"`
	config        *config.Config         `di.inject:"config"`
	namespace     string
}

func (s *KubeProjectService) PostConstruct() error {
	s.namespace = s.config.Kube.Project.Namespace
	return nil
}

func NewKubeProjectService(config *config.Config) KubeProjectServiceInterface {
	repo := &KubeProjectService{}

	kc, err := clientcmd.BuildConfigFromFlags("", config.Kube.ConfigPath)
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
	repo.kubeConfig = kc
	repo.kubeConfig.WarningHandler = rest.NoWarnings{}
	repo.clientset = clientset
	repo.dynamicClient = dynamicClient
	repo.config = config
	repo.namespace = config.Kube.Project.Namespace

	return repo
}
