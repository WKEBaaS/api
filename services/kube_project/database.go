// Package kube
//
// kubernetes related repository for project management
package kube_project

import (
	"context"
	"log/slog"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (r *KubeProjectService) CreateDatabase(ctx context.Context, namespace string, ref string) error {
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

func (r *KubeProjectService) DeleteDatabase(ctx context.Context, namespace string, ref string) error {
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
