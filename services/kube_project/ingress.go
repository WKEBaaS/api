// Package kube
//
// kubernetes related repository for project management
package kube_project

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (r *KubeProjectService) CreateIngressRoute(ctx context.Context, namespace string, ref string) error {
	ingressRouteTCPYAML, err := os.Open("kube-files/project-ingressroute.yaml")
	if err != nil {
		slog.Error("Failed to open IngressRoute YAML file", "error", err)
		return errors.New("failed to open IngressRoute YAML file")
	}
	defer ingressRouteTCPYAML.Close()

	ingressRouteUnstructured := &unstructured.Unstructured{}
	decoder := yaml.NewYAMLOrJSONDecoder(ingressRouteTCPYAML, 1024)
	if err := decoder.Decode(ingressRouteUnstructured); err != nil {
		slog.Error("Failed to decode IngressRoute YAML", "error", err)
		return errors.New("failed to decode IngressRoute YAML")
	}

	// set metadata
	ingressRouteName := r.GetAPIIngressRouteName(ref)
	ingressRouteUnstructured.SetName(ingressRouteName)
	ingressRouteUnstructured.SetNamespace(namespace)

	// set spec.routes[0]
	projectHost := r.GenProjectHost(ref)
	serviceName := r.GetAuthAPIServiceName(ref)
	unstructured.SetNestedSlice(ingressRouteUnstructured.Object, []any{
		map[string]any{
			"match": fmt.Sprintf("Host(`%s`) && PathPrefix(`/api/auth`)", projectHost),
			"services": []any{
				map[string]any{
					"name": serviceName,
					"port": int64(3000),
				},
			},
		},
	}, "spec", "routes")

	// set spec.tls.secretName
	if err := unstructured.SetNestedField(ingressRouteUnstructured.Object, r.config.Kube.Project.TLSSecretName, "spec", "tls", "secretName"); err != nil {
		slog.Error("Failed to set TLS secret name in IngressRoute spec", "error", err)
		return errors.New("failed to set TLS secret name in IngressRoute spec")
	}

	// 使用 dynamicClient 創建資源
	_, err = r.dynamicClient.Resource(ingressRouteGVR).
		Namespace(namespace).
		Create(ctx, ingressRouteUnstructured, metav1.CreateOptions{})
	if err != nil {
		slog.Error("Failed to create IngressRoute", "error", err)
		return errors.New("failed to create IngressRoute")
	}

	return nil
}

func (r *KubeProjectService) DeleteIngressRoute(ctx context.Context, namespace string, ref string) error {
	// 使用 dynamicClient 刪除資源
	target := r.GetAPIIngressRouteName(ref)
	err := r.dynamicClient.Resource(ingressRouteGVR).
		Namespace(namespace).
		Delete(ctx, target, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete IngressRoute", "error", err)
		return errors.New("failed to delete IngressRoute")
	}

	return nil
}

func (r *KubeProjectService) CreateIngressRouteTCP(ctx context.Context, namespace string, ref string) error {
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
	dbIngressRouteTCPName := r.GetDBIngressRouteTCPName(ref)
	ingressRouteTCPUnstructured.SetName(dbIngressRouteTCPName)
	ingressRouteTCPUnstructured.SetNamespace(namespace)

	// set spec.routes[0]
	projectHost := r.GenProjectHost(ref)
	serviceName := fmt.Sprintf("%s-rw", ref)
	unstructured.SetNestedSlice(ingressRouteTCPUnstructured.Object, []any{
		map[string]any{
			"match": fmt.Sprintf("Host(`%s`))", projectHost),
			"services": []any{
				map[string]any{
					"name": serviceName,
					"port": int64(5432),
				},
			},
		},
	}, "spec", "routes")
	// set spec.tls.secretName
	if err := unstructured.SetNestedField(ingressRouteTCPUnstructured.Object, r.config.Kube.Project.TLSSecretName, "spec", "tls", "secretName"); err != nil {
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

func (r *KubeProjectService) DeleteIngressRouteTCP(ctx context.Context, namespace string, ref string) error {
	// 使用 dynamicClient 刪除資源
	target := r.GetDBIngressRouteTCPName(ref)
	err := r.dynamicClient.Resource(ingressRouteTCPGVR).
		Namespace(namespace).
		Delete(ctx, target, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete IngressRouteTCP", "error", err)
		return ErrFailedToDeleteIngressRouteTCP
	}

	return nil
}
