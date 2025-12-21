// Package kubeproject
//
// kubernetes related repository for project management
package kubeproject

import (
	"context"

	"baas-api/internal/config"

	"github.com/samber/do/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Service interface {
	// === 基礎設施層 ===
	// Prepare Cluster Required Resources
	CreateJWKSConfigMap(ctx context.Context, opt CreateJWKSConfigMapOption) error
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
	FindDatabaseRoleSecret(ctx context.Context, ref string, role string) (*corev1.Secret, error)
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
	CreateRESTAPIService(ctx context.Context, ref string) error
	DeleteRESTAPIService(ctx context.Context, ref string) error

	// === 網路層 ===
	// Ingress for REST API (PostgREST) and Auth API
	CreateIngressRoute(ctx context.Context, ref string) error
	DeleteIngressRoute(ctx context.Context, ref string) error
	CreateIngressRouteTCP(ctx context.Context, ref string) error
	DeleteIngressRouteTCP(ctx context.Context, ref string) error
}

var _ Service = (*service)(nil)

type service struct {
	config        *config.Config `do:""`
	kubeConfig    *rest.Config
	clientset     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	namespace     string
}

func NewService(i do.Injector) (*service, error) {
	cfg := do.MustInvoke[*config.Config](i)
	svc := &service{}

	kc, err := clientcmd.BuildConfigFromFlags("", cfg.Kube.ConfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(kc)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(kc)
	if err != nil {
		return nil, err
	}
	svc.kubeConfig = kc
	svc.kubeConfig.WarningHandler = rest.NoWarnings{}
	svc.clientset = clientset
	svc.dynamicClient = dynamicClient
	svc.config = cfg
	svc.namespace = cfg.Kube.Project.Namespace

	return svc, nil
}
