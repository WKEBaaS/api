// Package kubeproject
//
// kubernetes related repository for project management
package kubeproject

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (s *KubeProjectService) CreateIngressRoute(ctx context.Context, ref string) error {
	ingressYAML, err := os.ReadFile("kube-files/project-ingressroute.yaml")
	if err != nil {
		slog.Error("Failed to open IngressRoute YAML file", "error", err)
		return errors.New("failed to open IngressRoute YAML file")
	}

	ingressYAMLString := string(ingressYAML)
	ingressData := map[string]any{
		"ProjectHost":     s.GetProjectHost(ref),
		"AuthServiceName": s.GetAuthAPIServiceName(ref),
		"RESTServiceName": s.GetRESTAPIServiceName(ref),
		"TLSSecretName":   s.config.Kube.Project.TLSSecretName,
	}
	ingressTmpl, err := template.New("yaml").Parse(ingressYAMLString)
	if err != nil {
		slog.Error("Failed to parse IngressRoute YAML template", "error", err)
		return errors.New("failed to parse IngressRoute YAML template")
	}
	var ingressRendered strings.Builder
	if err := ingressTmpl.Execute(&ingressRendered, ingressData); err != nil {
		slog.Error("Failed to execute IngressRoute YAML template", "error", err)
		return errors.New("failed to execute IngressRoute YAML template")
	}

	ingressRoute := &unstructured.Unstructured{}
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(ingressRendered.String()), 1024)
	if err := decoder.Decode(ingressRoute); err != nil {
		slog.Error("Failed to decode IngressRoute YAML", "error", err)
		return errors.New("failed to decode IngressRoute YAML")
	}

	// set metadata
	ingressRouteName := s.GetAPIIngressRouteName(ref)
	ingressRoute.SetName(ingressRouteName)
	ingressRoute.SetNamespace(s.namespace)

	// 使用 dynamicClient 創建資源
	_, err = s.dynamicClient.Resource(ingressRouteGVR).
		Namespace(s.namespace).
		Create(ctx, ingressRoute, metav1.CreateOptions{})
	if err != nil {
		slog.Error("Failed to create IngressRoute", "error", err)
		return errors.New("failed to create IngressRoute")
	}

	return nil
}

func (s *KubeProjectService) DeleteIngressRoute(ctx context.Context, ref string) error {
	// 使用 dynamicClient 刪除資源
	target := s.GetAPIIngressRouteName(ref)
	err := s.dynamicClient.Resource(ingressRouteGVR).
		Namespace(s.namespace).
		Delete(ctx, target, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete IngressRoute", "error", err)
		return errors.New("failed to delete IngressRoute")
	}

	return nil
}

func (s *KubeProjectService) CreateIngressRouteTCP(ctx context.Context, ref string) error {
	ingressTCPYAML, err := os.ReadFile("kube-files/project-ingressroutetcp.yaml")
	if err != nil {
		slog.Error("Failed to open IngressRouteTCP YAML file", "error", err)
		return ErrFailedToOpenIngressRouteTCPYAML
	}

	ingressTCPYAMLString := string(ingressTCPYAML)
	ingressData := map[string]any{
		"ProjectHost":          s.GetProjectHost(ref),
		"ProjectDBServiceName": s.GetDatabaseRWServiceName(ref),
		"TLSSecretName":        s.config.Kube.Project.TLSSecretName,
	}
	ingressTmpl, err := template.New("yaml").Parse(ingressTCPYAMLString)
	if err != nil {
		slog.Error("Failed to parse IngressRouteTCP YAML template", "error", err)
		return errors.New("failed to parse IngressRouteTCP YAML template")
	}
	var ingressRendered strings.Builder
	if err := ingressTmpl.Execute(&ingressRendered, ingressData); err != nil {
		slog.Error("Failed to execute IngressRouteTCP YAML template", "error", err)
		return errors.New("failed to execute IngressRouteTCP YAML template")
	}

	ingressRouteTCPUnstructured := &unstructured.Unstructured{}
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(ingressRendered.String()), 1024)
	if err := decoder.Decode(ingressRouteTCPUnstructured); err != nil {
		slog.Error("Failed to decode IngressRouteTCP YAML", "error", err)
		return ErrFailedToDecodeIngressRouteTCPYAML
	}

	// set metadata
	dbIngressRouteTCPName := s.GetDBIngressRouteTCPName(ref)
	ingressRouteTCPUnstructured.SetName(dbIngressRouteTCPName)
	ingressRouteTCPUnstructured.SetNamespace(s.namespace)

	// 使用 dynamicClient 創建資源
	_, err = s.dynamicClient.Resource(ingressRouteTCPGVR).
		Namespace(s.namespace).
		Create(ctx, ingressRouteTCPUnstructured, metav1.CreateOptions{})
	if err != nil {
		slog.Error("Failed to create IngressRouteTCP", "error", err)
		return ErrFailedToCreateIngressRouteTCP
	}

	return nil
}

func (s *KubeProjectService) DeleteIngressRouteTCP(ctx context.Context, ref string) error {
	// 使用 dynamicClient 刪除資源
	target := s.GetDBIngressRouteTCPName(ref)
	err := s.dynamicClient.Resource(ingressRouteTCPGVR).
		Namespace(s.namespace).
		Delete(ctx, target, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete IngressRouteTCP", "error", err)
		return ErrFailedToDeleteIngressRouteTCP
	}

	return nil
}
