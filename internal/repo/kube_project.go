package repo

import (
	"baas-api/internal/configs"
	"context"
	"errors"
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

type KubeProjectRepository interface {
	CreateCluster(ctx context.Context, namespace string, ref string, storageSize string) error
	DeleteCluster(ctx context.Context, namespace string, ref string) error
}

type kubeProjectRepository struct {
	kubeConfig    *rest.Config
	clientset     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	appConfig     *configs.Config
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
	repo.appConfig = config

	return repo
}

// Errors for kubeProjectRepository
var (
	ErrFailedToOpenPostgresClusterYAML             = errors.New("failed to open Postgres cluster YAML file")
	ErrFailedToDecodePostgresClusterYAML           = errors.New("failed to decode Postgres cluster YAML")
	ErrFailedToGetSpecFromPostgresClusterYAML      = errors.New("failed to get spec from Postgres cluster YAML")
	ErrSpecNotFoundInPostgresClusterYAML           = errors.New("spec not found in Postgres cluster YAML")
	ErrFailedToSetStorageSizeInPostgresClusterSpec = errors.New("failed to set storage size in Postgres cluster spec")
	ErrFailedToCreatePostgresCluster               = errors.New("failed to create Postgres cluster")
	ErrFailedToDeletePostgresCluster               = errors.New("failed to delete Postgres cluster")
)

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

func (r *kubeProjectRepository) CreateCluster(ctx context.Context, namespace string, ref string, storageSize string) error {
	pgClusterYAML, err := os.Open("kube-files/postgres-cluster.yaml")
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

	// set spec
	pgClusterSpec, found, err := unstructured.NestedMap(pgClusterUnstructured.Object, "spec")
	if err != nil {
		slog.Error("Failed to get spec from Postgres cluster YAML", "error", err)
		return ErrFailedToGetSpecFromPostgresClusterYAML
	}
	if !found {
		slog.Error("Spec not found in Postgres cluster YAML")
		return ErrSpecNotFoundInPostgresClusterYAML
	}
	// set storage size
	if err := unstructured.SetNestedField(pgClusterSpec, storageSize, "storage", "size"); err != nil {
		slog.Error("Failed to set storage size in Postgres cluster spec", "error", err)
		return ErrFailedToSetStorageSizeInPostgresClusterSpec
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
