package kubeproject

import (
	"context"
	"errors"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (s *KubeProjectService) CreateAuthAPIService(ctx context.Context, ref string) error {
	serviceName := s.GetAuthAPIServiceName(ref)
	deploymentName := s.GetAuthAPIDeploymentName(ref)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: s.namespace,
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

	_, err := s.clientset.CoreV1().Services(s.namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create Auth API service", "error", err)
		return errors.New("failed to create Auth API service")
	}

	return nil
}

func (s *KubeProjectService) DeleteAuthAPIService(ctx context.Context, ref string) error {
	serviceName := s.GetAuthAPIServiceName(ref)

	err := s.clientset.CoreV1().Services(s.namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete Auth API service", "error", err)
		return errors.New("failed to delete Auth API service")
	}

	return nil
}
