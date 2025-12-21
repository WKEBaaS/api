package kubeproject

import (
	"context"
	"errors"
	"log/slog"

	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *service) CreateRESTAPIDeployment(ctx context.Context, ref string, jwks string) error {
	deploymentName := s.GetRESTAPIDeploymentName(ref)
	pgrstContainerName := s.GetRESTAPIContainerName(ref, PGRSTComponent)
	openapiContainerName := s.GetRESTAPIContainerName(ref, OpenAPIComponent)
	authenticatorSecretName := s.GetDatabaseRoleSecretName(ref, RoleAuthenticator)
	restURL := s.GetRESTAPIURL(ref)
	scalarConfig := s.GenerateScalarAPIConfig(restURL)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: s.namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: lo.ToPtr(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": deploymentName,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": deploymentName,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  pgrstContainerName,
							Image: "postgrest/postgrest:v13.0.4",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 3000,
							}},
							Env: []corev1.EnvVar{
								{Name: "PGRST_SERVER_PORT", Value: "3000"},
								{Name: "PGRST_DB_SCHEMA", Value: "api"},
								{Name: "PGRST_DB_ANON_ROLE", Value: "anon"},
								{Name: "PGRST_OPENAPI_SECURITY_ACTIVE", Value: "true"},
								{Name: "PGRST_OPENAPI_SERVER_PROXY_URI", Value: restURL},
								{Name: "PGRST_JWT_SECRET", Value: jwks},
								{Name: "PGRST_DB_URI", ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										Key: "uri",
										LocalObjectReference: corev1.LocalObjectReference{
											Name: authenticatorSecretName,
										},
									},
								}},
							},
						},
						{
							Name:  openapiContainerName,
							Image: "scalarapi/api-reference:0.2.25",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 8080,
							}},
							Env: []corev1.EnvVar{
								{
									Name:  "API_REFERENCE_CONFIG",
									Value: scalarConfig,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := s.clientset.AppsV1().Deployments(s.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create REST API deployment", "error", err)
		return errors.New("failed to create REST API deployment")
	}

	return nil
}

func (s *service) DeleteRESTAPIDeployment(ctx context.Context, ref string) error {
	deploymentName := s.GetRESTAPIDeploymentName(ref)

	// Delete the deployment
	err := s.clientset.AppsV1().Deployments(s.namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete API deployment", "error", err)
		return errors.New("failed to delete API deployment")
	}

	return nil
}
