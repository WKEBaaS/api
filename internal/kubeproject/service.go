// Package kubeproject
//
// kubernetes related repository for project management
package kubeproject

import (
	"context"
	"fmt"
	"path/filepath"

	"baas-api/internal/config"

	"github.com/samber/do/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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

	var kc *rest.Config
	var err error

	// 1. 優先嘗試讀取 In-Cluster Config (適用於 Pod 內部 / 生產環境)
	kc, err = rest.InClusterConfig()
	// 2. 如果 In-Cluster 失敗 (代表可能在 Local 開發環境)，則嘗試讀取 kubeconfig
	if err != nil {
		kubeConfigPath := cfg.Kube.ConfigPath

		// 如果設定檔中沒有指定路徑，則使用預設路徑 (~/.kube/config)
		if kubeConfigPath == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeConfigPath = filepath.Join(home, ".kube", "config")
			}
		}

		// 讀取 kubeconfig
		kc, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			// 如果兩種方式都失敗，才回傳錯誤
			return nil, fmt.Errorf("無法初始化 K8s Config (InCluster 失敗且無法讀取 %s): %w", kubeConfigPath, err)
		}
	}

	// 3. 初始化 Clients
	clientset, err := kubernetes.NewForConfig(kc)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(kc)
	if err != nil {
		return nil, err
	}

	// 4. 設定 Service 屬性
	svc.kubeConfig = kc
	svc.kubeConfig.WarningHandler = rest.NoWarnings{} // 忽略 API 警告
	svc.clientset = clientset
	svc.dynamicClient = dynamicClient
	svc.config = cfg
	svc.namespace = cfg.Kube.Project.Namespace

	return svc, nil
}
