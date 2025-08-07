// Package kube
//
// kubernetes related repository for project management
package kube

import (
	"baas-api/config"
	"context"
	"log"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeProjectRepository interface {
	CreateCluster(ctx context.Context, namespace string, ref string, storageSize string) error
	DeleteCluster(ctx context.Context, namespace string, ref string) error
	FindClusterStatus(ctx context.Context, namespace string, ref string) (*string, error)

	CreateDatabase(ctx context.Context, namespace string, ref string) error
	DeleteDatabase(ctx context.Context, namespace string, ref string) error
	ReadDatabasePassword(ctx context.Context, namespace string, ref string) (*string, error)
	ResetDatabasePassword(ctx context.Context, namespace string, ref string, password string) error

	CreateIngressRouteTCP(ctx context.Context, namespace string, ref string) error
	DeleteIngressRouteTCP(ctx context.Context, namespace string, ref string) error

	CreateAPIDeployment(ctx context.Context, opt *CreateAPIDeploymentOption) error
	DeleteAPIDeployment(ctx context.Context, namespace string, ref string) error
}

type kubeProjectRepository struct {
	kubeConfig    *rest.Config
	clientset     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	config        *config.Config
}

func NewKubeProjectRepository(config *config.Config) KubeProjectRepository {
	repo := &kubeProjectRepository{}

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

	return repo
}
