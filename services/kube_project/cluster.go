// Package kube_project
//
// kubernetes related repository for project management
package kube_project

import (
	"context"
	"errors"
	"log/slog"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (r *KubeProjectService) CreateCluster(ctx context.Context, namespace string, ref string, storageSize string) error {
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

func (r *KubeProjectService) DeleteCluster(ctx context.Context, namespace string, ref string) error {
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

func (r *KubeProjectService) FindClusterStatus(ctx context.Context, namespace string, ref string) (*string, error) {
	deployment, err := r.dynamicClient.Resource(clusterGVR).
		Namespace(namespace).
		Get(ctx, ref, metav1.GetOptions{})
	if err != nil {
		slog.Error("Failed to get postgres cluster", "error", err)
		return nil, errors.New("failed to get postgres cluster")
	}

	status := deployment.Object["status"].(map[string]any)
	phase, ok := status["phase"].(string)
	if !ok {
		phase = "Initializing Postgres cluster"
		return &phase, nil
	}

	return &phase, nil
}
