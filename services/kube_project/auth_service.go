package kube_project

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *KubeProjectService) CreateAuthAPIService(ctx context.Context, namespace string, ref string) error {
	serviceName := r.GetAuthAPIServiceName(ref)
	deploymentName := r.GetAuthAPIDeploymentName(ref)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				"app": deploymentName,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       3000,
					TargetPort: intstr.FromInt(3000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	_, err := r.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create API service", "error", err)
		return errors.New("failed to create API service")
	}

	return nil
}

func (r *KubeProjectService) DeleteAuthAPIService(ctx context.Context, namespace string, ref string) error {
	serviceName := fmt.Sprintf("%s-api", ref)

	err := r.clientset.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete API service", "error", err)
		return errors.New("failed to delete API service")
	}

	return nil
}
