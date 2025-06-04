package repo

import (
	"baas-api/internal/configs"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Errors for kubeProjectRepository
var (
	// cluster errors
	ErrFailedToOpenPostgresClusterYAML        = errors.New("failed to open Postgres cluster YAML file")
	ErrFailedToDecodePostgresClusterYAML      = errors.New("failed to decode Postgres cluster YAML")
	ErrFailedToGetSpecFromPostgresClusterYAML = errors.New("failed to get spec from Postgres cluster YAML")
	ErrSpecNotFoundInPostgresClusterYAML      = errors.New("spec not found in Postgres cluster YAML")
	ErrFailedToSetSpecStorageSize             = errors.New("failed to set storage size in Postgres cluster spec")
	ErrFailedToCreatePostgresCluster          = errors.New("failed to create Postgres cluster")
	ErrFailedToDeletePostgresCluster          = errors.New("failed to delete Postgres cluster")
	// database errors
	ErrFailedToOpenPostgresDatabaseYAML   = errors.New("failed to open Postgres database YAML file")
	ErrFailedToDecodePostgresDatabaseYAML = errors.New("failed to decode Postgres database YAML")
	ErrFailedToSetSpecClusterName         = errors.New("failed to set cluster name in Postgres database spec")
	ErrFeiledToCreatePostgresDatabase     = errors.New("failed to create Postgres database")
	ErrFailedToDeletePostgresDatabase     = errors.New("failed to delete Postgres database")
	// ingress route TCP errors
	ErrFailedToOpenIngressRouteTCPYAML    = errors.New("failed to open IngressRouteTCP YAML file")
	ErrFailedToDecodeIngressRouteTCPYAML  = errors.New("failed to decode IngressRouteTCP YAML")
	ErrFailedToSetSpecRouteTCPName        = errors.New("failed to set route TCP name in IngressRouteTCP spec")
	ErrFailedToSetSpecRouteTCPServiceName = errors.New("failed to set service name in IngressRouteTCP spec")
	ErrFailedToSetSpecTLSSecretName       = errors.New("failed to set TLS secret name in IngressRouteTCP spec")
	ErrFailedToCreateIngressRouteTCP      = errors.New("failed to create IngressRouteTCP")
	ErrFailedToDeleteIngressRouteTCP      = errors.New("failed to delete IngressRouteTCP")
)

type KubeProjectRepository interface {
	CreateCluster(ctx context.Context, namespace string, ref string, storageSize string) error
	DeleteCluster(ctx context.Context, namespace string, ref string) error

	CreateDatabase(ctx context.Context, namespace string, ref string) error
	DeleteDatabase(ctx context.Context, namespace string, ref string) error

	CreateIngressRouteTCP(ctx context.Context, namespace string, ref string) error
	DeleteIngressRouteTCP(ctx context.Context, namespace string, ref string) error
}

type kubeProjectRepository struct {
	kubeConfig                    *rest.Config
	clientset                     *kubernetes.Clientset
	dynamicClient                 *dynamic.DynamicClient
	projectsHost                  string
	projectsWildcardTLSSecretName string
}

func NewKubeProjectRepository(config *configs.Config) KubeProjectRepository {
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
	repo.clientset = clientset
	repo.dynamicClient = dynamicClient
	repo.projectsHost = config.PROJECTS_HOST
	repo.projectsWildcardTLSSecretName = config.Kube.ProjectsWildcardTLSSecretName

	return repo
}

// CloudNativePG Cluster GVR
var clusterGVR = schema.GroupVersionResource{
	Group:    "postgresql.cnpg.io",
	Version:  "v1",
	Resource: "clusters", // Resource Name: 通常是CRD定義中的 `spec.names.plural`
}

var databaseGVR = schema.GroupVersionResource{
	Group:    "postgresql.cnpg.io",
	Version:  "v1",
	Resource: "databases",
}

var ingressRouteTCPGVR = schema.GroupVersionResource{
	Group:    "traefik.io",
	Version:  "v1alpha1",
	Resource: "ingressroutetcps",
}

func (r *kubeProjectRepository) CreateCluster(ctx context.Context, namespace string, ref string, storageSize string) error {
	pgClusterYAML, err := os.Open("kube-files/project-cnpg-cluster.yaml")
	if err != nil {
		slog.Error("Failed to open Postgres cluster YAML file", "error", err)
		return ErrFailedToOpenPostgresClusterYAML
	}
	defer pgClusterYAML.Close()

	pgClusterUnstructured := &unstructured.Unstructured{}
	decoder := yaml.NewYAMLOrJSONDecoder(pgClusterYAML, 1024)
	if err := decoder.Decode(pgClusterUnstructured); err != nil {
		slog.Error("Failed to decode Postgres cluster YAML", "error", err)
		return ErrFailedToDecodePostgresClusterYAML
	}

	// set metadata
	pgClusterUnstructured.SetName(ref)
	pgClusterUnstructured.SetNamespace(namespace)

	// set spec.storage.size
	if err := unstructured.SetNestedField(pgClusterUnstructured.Object, storageSize, "spec", "storage", "size"); err != nil {
		slog.Error("Failed to set storage size in Postgres cluster spec", "error", err)
		return ErrFailedToSetSpecStorageSize
	}

	// 使用 dynamicClient 創建資源
	_, err = r.dynamicClient.Resource(clusterGVR).
		Namespace(namespace).
		Create(ctx, pgClusterUnstructured, metav1.CreateOptions{})
	if err != nil {
		slog.Error("Failed to create Postgres cluster", "error", err)
		return ErrFailedToCreatePostgresCluster
	}

	return nil
}

func (r *kubeProjectRepository) DeleteCluster(ctx context.Context, namespace string, ref string) error {
	// 使用 dynamicClient 刪除資源
	err := r.dynamicClient.Resource(clusterGVR).
		Namespace(namespace).
		Delete(ctx, ref, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete Postgres cluster", "error", err)
		return ErrFailedToDeletePostgresCluster
	}

	return nil
}

func (r *kubeProjectRepository) CreateDatabase(ctx context.Context, namespace string, ref string) error {
	pgDatabaseYAML, err := os.Open("kube-files/project-cnpg-database.yaml")
	if err != nil {
		slog.Error("Failed to open Postgres database YAML file", "error", err)
		return ErrFailedToOpenPostgresDatabaseYAML
	}
	defer pgDatabaseYAML.Close()

	pgDatabaseUnstructured := &unstructured.Unstructured{}
	decoder := yaml.NewYAMLOrJSONDecoder(pgDatabaseYAML, 1024)
	if err := decoder.Decode(pgDatabaseUnstructured); err != nil {
		slog.Error("Failed to decode Postgres database YAML", "error", err)
		return ErrFailedToDecodePostgresDatabaseYAML
	}

	// set metadata
	pgDatabaseUnstructured.SetName(ref)
	pgDatabaseUnstructured.SetNamespace(namespace)

	// set spec.cluster.name
	if err := unstructured.SetNestedField(pgDatabaseUnstructured.Object, ref, "spec", "cluster", "name"); err != nil {
		slog.Error("Failed to set name in Postgres database spec", "error", err)
		return ErrFailedToSetSpecClusterName
	}

	// 使用 dynamicClient 創建資源
	_, err = r.dynamicClient.Resource(databaseGVR).
		Namespace(namespace).
		Create(ctx, pgDatabaseUnstructured, metav1.CreateOptions{})
	if err != nil {
		slog.Error("Failed to create Postgres database", "error", err)
		return ErrFeiledToCreatePostgresDatabase
	}

	return nil
}

func (r *kubeProjectRepository) DeleteDatabase(ctx context.Context, namespace string, ref string) error {
	// 使用 dynamicClient 刪除資源
	err := r.dynamicClient.Resource(databaseGVR).
		Namespace(namespace).
		Delete(ctx, ref, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete Postgres database", "error", err)
		return ErrFailedToDeletePostgresDatabase
	}

	return nil
}

func (r *kubeProjectRepository) CreateIngressRouteTCP(ctx context.Context, namespace string, ref string) error {
	ingressRouteTCPYAML, err := os.Open("kube-files/project-ingressroutetcp.yaml")
	if err != nil {
		slog.Error("Failed to open IngressRouteTCP YAML file", "error", err)
		return ErrFailedToOpenIngressRouteTCPYAML
	}
	defer ingressRouteTCPYAML.Close()

	ingressRouteTCPUnstructured := &unstructured.Unstructured{}
	decoder := yaml.NewYAMLOrJSONDecoder(ingressRouteTCPYAML, 1024)
	if err := decoder.Decode(ingressRouteTCPUnstructured); err != nil {
		slog.Error("Failed to decode IngressRouteTCP YAML", "error", err)
		return ErrFailedToDecodeIngressRouteTCPYAML
	}

	// set metadata
	ingressRouteTCPUnstructured.SetName(fmt.Sprintf("%s-ingressroutetcp", ref))
	ingressRouteTCPUnstructured.SetNamespace(namespace)

	// set spec.routes[0]
	projectHostSNI := fmt.Sprintf("HostSNI(`%s.%s`)", ref, r.projectsHost)
	serviceName := fmt.Sprintf("%s-rw", ref)
	unstructured.SetNestedSlice(ingressRouteTCPUnstructured.Object, []any{
		map[string]any{
			"match": projectHostSNI,
			"services": []any{
				map[string]any{
					"name": serviceName,
					"port": int64(5432),
				},
			},
		},
	}, "spec", "routes")
	// set spec.tls.secretName
	if err := unstructured.SetNestedField(ingressRouteTCPUnstructured.Object, r.projectsWildcardTLSSecretName, "spec", "tls", "secretName"); err != nil {
		slog.Error("Failed to set TLS secret name in IngressRouteTCP spec", "error", err)
		return ErrFailedToSetSpecTLSSecretName
	}

	// 使用 dynamicClient 創建資源
	_, err = r.dynamicClient.Resource(ingressRouteTCPGVR).
		Namespace(namespace).
		Create(ctx, ingressRouteTCPUnstructured, metav1.CreateOptions{})
	if err != nil {
		slog.Error("Failed to create IngressRouteTCP", "error", err)
		return ErrFailedToCreateIngressRouteTCP
	}

	return nil
}

func (r *kubeProjectRepository) DeleteIngressRouteTCP(ctx context.Context, namespace string, ref string) error {
	// 使用 dynamicClient 刪除資源
	err := r.dynamicClient.Resource(ingressRouteTCPGVR).
		Namespace(namespace).
		Delete(ctx, fmt.Sprintf("%s-ingressroutetcp", ref), metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete IngressRouteTCP", "error", err)
		return ErrFailedToDeleteIngressRouteTCP
	}

	return nil
}
