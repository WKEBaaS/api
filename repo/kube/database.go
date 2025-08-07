// Package kube
//
// kubernetes related repository for project management
package kube

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

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

func (r *kubeProjectRepository) ReadDatabasePassword(ctx context.Context, namespace string, ref string) (*string, error) {
	secretName := fmt.Sprintf("%s-app", ref)
	secret, err := r.clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		slog.Error("Failed to read database secret", "error", err)
		return nil, ErrFailedToReadDatabaseSecret
	}

	passwordBytes, ok := secret.Data["password"]
	if !ok {
		slog.Error("Password not found in database secret", "secretName", secretName)
		return nil, ErrFailedToReadDatabaseSecret
	}

	password := string(passwordBytes)
	return &password, nil
}

func (r *kubeProjectRepository) ResetDatabasePassword(ctx context.Context, namespace, ref, password string) error {
	secretName := fmt.Sprintf("%s-app", ref)
	secret, err := r.clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		slog.Error("Failed to read database secret", "error", err)
		return ErrFailedToReadDatabaseSecret
	}

	// Update the secret with the new password
	secret.Data["password"] = []byte(password)
	_, err = r.clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		slog.Error("Failed to update database secret with new password", "error", err)
		return ErrFailedToResetDatabasePassword
	}

	return nil
}
