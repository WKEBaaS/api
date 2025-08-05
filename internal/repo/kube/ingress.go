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
	projectHostSNI := fmt.Sprintf("HostSNI(`%s.%s`)", ref, r.config.App.Host)
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
