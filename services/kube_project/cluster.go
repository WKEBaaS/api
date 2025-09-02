// Package kube_project
//
// kubernetes related repository for project management
package kube_project

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"text/template"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (r *KubeProjectService) CreateCluster(ctx context.Context, ref string, storageSize string) error {
	clusterYAML, err := os.ReadFile("kube-files/project-cnpg-cluster.yaml")
	if err != nil {
		slog.Error("Failed to open Postgres cluster YAML file", "error", err)
		return ErrFailedToOpenPostgresClusterYAML
	}

	pgClusterYAMLString := string(clusterYAML)
	clusterData := map[string]any{
		"RoleAuthenticatorSecretName": r.GetDatabaseRoleSecretName(ref, RoleAuthenticator),
	}
	clusterTmpl, err := template.New("yaml").Parse(pgClusterYAMLString)
	if err != nil {
		slog.Error("Failed to parse Postgres cluster YAML template", "error", err)
		return errors.New("failed to parse Postgres cluster YAML template")
	}
	var clusterRendered bytes.Buffer
	if err := clusterTmpl.Execute(&clusterRendered, clusterData); err != nil {
		slog.Error("Failed to execute Postgres cluster YAML template", "error", err)
		return errors.New("failed to execute Postgres cluster YAML template")
	}

	cluster := &unstructured.Unstructured{}
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(clusterRendered.String()), 1024)
	if err := decoder.Decode(cluster); err != nil {
		slog.Error("Failed to decode Postgres cluster YAML", "error", err)
		return ErrFailedToDecodePostgresClusterYAML
	}

	// set metadata
	cluster.SetName(ref)
	cluster.SetNamespace(r.namespace)

	// set spec.storage.size
	if err := unstructured.SetNestedField(cluster.Object, storageSize, "spec", "storage", "size"); err != nil {
		slog.Error("Failed to set storage size in Postgres cluster spec", "error", err)
		return ErrFailedToSetSpecStorageSize
	}

	// 使用 dynamicClient 創建資源
	_, err = r.dynamicClient.Resource(clusterGVR).
		Namespace(r.namespace).
		Create(ctx, cluster, metav1.CreateOptions{})
	if err != nil {
		slog.Error("Failed to create Postgres cluster", "error", err)
		return ErrFailedToCreatePostgresCluster
	}

	return nil
}

func (r *KubeProjectService) DeleteCluster(ctx context.Context, ref string) error {
	// 使用 dynamicClient 刪除資源
	err := r.dynamicClient.Resource(clusterGVR).
		Namespace(r.namespace).
		Delete(ctx, ref, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete Postgres cluster", "error", err)
		return ErrFailedToDeletePostgresCluster
	}

	return nil
}

func (r *KubeProjectService) FindClusterStatus(ctx context.Context, ref string) (*string, error) {
	deployment, err := r.dynamicClient.Resource(clusterGVR).
		Namespace(r.namespace).
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

func (s *KubeProjectService) WaitClusterHealthy(ctx context.Context, ref string) error {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			status, err := s.FindClusterStatus(ctx, ref)
			if err != nil {
				return err
			}
			if *status == "Cluster in healthy state" {
				return nil
			}
		}
	}
}
