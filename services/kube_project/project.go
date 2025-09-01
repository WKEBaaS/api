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
	// CNPG Cluster 管理
	CreateCluster(ctx context.Context, namespace, ref, storageSize string) error
	DeleteCluster(ctx context.Context, namespace, ref string) error
	FindClusterStatus(ctx context.Context, namespace, ref string) (*string, error)

	// Database Management
	CreateDatabase(ctx context.Context, namespace, ref string) error
	DeleteDatabase(ctx context.Context, namespace, ref string) error

	// Database Role Management
	FindDatabaseRolePassword(ctx context.Context, ref, role string) (*string, error)
	CreateDatabaseRoleSecret(ctx context.Context, ref, role, password string) error
	UpdateDatabaseRoleSecret(ctx context.Context, ref, role, password string) error

	// === 應用層 ===
	// Auth API
	CreateAuthAPIDeployment(ctx context.Context, ref string, opt *APIDeploymentOption) error
	DeleteAuthAPIDeployment(ctx context.Context, namespace, ref string) error
	PatchAuthAPIDeployment(ctx context.Context, namespace, ref string, opt *APIDeploymentOption) error
	CreateAuthAPIService(ctx context.Context, namespace, ref string) error
	DeleteAuthAPIService(ctx context.Context, namespace, ref string) error

	// REST API (PostgREST)
	CreateRESTAPIDeployment(ctx context.Context, ref string, jwks string) error
	DeleteRESTAPIDeployment(ctx context.Context, namespace, ref string) error

	// === 網路層 ===
	// Ingress for REST API (PostgREST) and Auth API
	CreateIngressRoute(ctx context.Context, namespace, ref string) error
	DeleteIngressRoute(ctx context.Context, namespace, ref string) error
	CreateIngressRouteTCP(ctx context.Context, namespace, ref string) error
	DeleteIngressRouteTCP(ctx context.Context, namespace, ref string) error
}

type KubeProjectService struct {
	kubeConfig    *rest.Config
	clientset     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	config        *config.Config
	namespace     string
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
