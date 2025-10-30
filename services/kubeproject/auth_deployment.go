package kubeproject

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"baas-api/dto"
	"baas-api/utils"

	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type APIDeploymentOption struct {
	BetterAuthSecret *string
	TrustedOrigins   []string
	ProxyURL         *string
	AuthProviders    map[string]dto.AuthProvider
}

func (s *KubeProjectService) CreateAuthAPIDeployment(ctx context.Context, ref string, opt *APIDeploymentOption) error {
	if opt.BetterAuthSecret == nil {
		return errors.New("BetterAuthSecret is required when creating Auth API deployment")
	}

	// Build environment variables dynamically
	envVars := []corev1.EnvVar{
		{
			Name:  "BETTER_AUTH_URL",
			Value: fmt.Sprintf("https://%s.%s", ref, s.config.App.ExternalDomain),
		},
		{
			Name:  "BETTER_AUTH_SECRET",
			Value: *opt.BetterAuthSecret,
		},
		{
			Name:  "TRUSTED_ORIGINS",
			Value: strings.Join(opt.TrustedOrigins, ","),
		},
		{
			Name:  "EMAIL_AND_PASSWORD_ENABLED",
			Value: utils.BoolToString(opt.AuthProviders["email"].Enabled),
		},
		{
			Name: "DATABASE_URL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-app", ref),
					},
					Key: "uri",
				},
			},
		},
	}

	// Add OAuth provider environment variables dynamically from AuthProviders
	for providerName, provider := range opt.AuthProviders {
		upperProviderName := strings.ToUpper(providerName)
		envVars = append(envVars,
			corev1.EnvVar{Name: upperProviderName + "_ENABLED", Value: utils.BoolToString(provider.Enabled)},
		)
		if provider.ClientID != nil {
			envVars = append(envVars,
				corev1.EnvVar{Name: upperProviderName + "_CLIENT_ID", Value: *provider.ClientID},
			)
		}
		if provider.ClientSecret != nil {
			envVars = append(envVars,
				corev1.EnvVar{Name: upperProviderName + "_CLIENT_SECRET", Value: *provider.ClientSecret},
			)
		}
	}

	deploymentName := s.GetAuthAPIDeploymentName(ref)
	authContainerName := s.GetAuthAPIContainerName(ref)
	// Create the deployment object
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
							Name:  authContainerName,
							Image: "ghcr.io/wkebaas/project-auth:v0.0.22",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 3000,
								},
							},
							Env: envVars,
						},
					},
				},
			},
		},
	}

	// Create the deployment
	_, err := s.clientset.AppsV1().Deployments(s.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create API deployment", "error", err)
		return errors.New("failed to create API deployment")
	}

	return nil
}

func (s *KubeProjectService) PatchAuthAPIDeployment(ctx context.Context, ref string, opt *APIDeploymentOption) error {
	deploymentName := s.GetAuthAPIDeploymentName(ref)
	authContainerName := s.GetAuthAPIContainerName(ref)
	envVars := []corev1.EnvVar{}

	if opt.ProxyURL != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "BETTER_AUTH_URL",
			Value: *opt.ProxyURL,
		})
	}

	if opt.BetterAuthSecret != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "BETTER_AUTH_SECRET",
			Value: *opt.BetterAuthSecret,
		})
	}
	if len(opt.TrustedOrigins) > 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "TRUSTED_ORIGINS",
			Value: strings.Join(opt.TrustedOrigins, ","),
		})
	}
	if provider, ok := opt.AuthProviders["email"]; ok {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "EMAIL_AND_PASSWORD_ENABLED",
			Value: utils.BoolToString(provider.Enabled),
		})
	}

	// Add OAuth provider environment variables dynamically from AuthProviders
	for providerName, provider := range opt.AuthProviders {
		upperProviderName := strings.ToUpper(providerName)
		envVars = append(envVars,
			corev1.EnvVar{Name: upperProviderName + "_ENABLED", Value: utils.BoolToString(provider.Enabled)},
		)
		if provider.ClientID != nil {
			envVars = append(envVars,
				corev1.EnvVar{Name: upperProviderName + "_CLIENT_ID", Value: *provider.ClientID},
			)
		}
		if provider.ClientSecret != nil {
			envVars = append(envVars,
				corev1.EnvVar{Name: upperProviderName + "_CLIENT_SECRET", Value: *provider.ClientSecret},
			)
		}
	}

	// Prepare JSON merge patch payload instead of YAML to avoid "invalid character" errors
	payload := map[string]any{
		"spec": map[string]any{
			"template": map[string]any{
				"spec": map[string]any{
					"containers": []map[string]any{
						{
							"name": authContainerName,
							"env":  envVars,
						},
					},
				},
			},
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to marshal patch data", "error", err)
		return errors.New("failed to marshal patch data")
	}

	// Patch the deployment
	_, err = s.clientset.AppsV1().Deployments(s.namespace).Patch(
		ctx,
		deploymentName,
		types.StrategicMergePatchType,
		data,
		metav1.PatchOptions{},
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to patch API deployment", "error", err)
		return fmt.Errorf("failed to patch API deployment")
	}

	return nil
}

func (s *KubeProjectService) DeleteAuthAPIDeployment(ctx context.Context, ref string) error {
	deploymentName := s.GetAuthAPIDeploymentName(ref)

	// Delete the deployment
	err := s.clientset.AppsV1().Deployments(s.namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete API deployment", "error", err)
		return errors.New("failed to delete API deployment")
	}

	return nil
}
