package kube

import (
	"baas-api/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type APIDeploymentOption struct {
	Namespace        string
	Ref              string
	BetterAuthSecret string
	TrustedOrigins   []string

	Auth AuthConfig
}

type AuthConfig struct {
	EmailAndPasswordEnabled bool

	Google  *OAuthProvider
	GitHub  *OAuthProvider
	Discord *OAuthProvider
}

type OAuthProvider struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
}

func NewAPIDeploymentOption() *APIDeploymentOption {
	return &APIDeploymentOption{}
}

func (o *APIDeploymentOption) WithNamespace(namespace string) *APIDeploymentOption {
	o.Namespace = namespace
	return o
}

func (o *APIDeploymentOption) WithRef(ref string) *APIDeploymentOption {
	o.Ref = ref
	return o
}

func (o *APIDeploymentOption) WithTrustedOrigins(origins []string) *APIDeploymentOption {
	o.TrustedOrigins = origins
	return o
}

func (o *APIDeploymentOption) WithEmailAndPasswordAuth(enabled bool) *APIDeploymentOption {
	o.Auth.EmailAndPasswordEnabled = enabled
	return o
}

func (o *APIDeploymentOption) WithBetterAuthSecret(secret string) *APIDeploymentOption {
	o.BetterAuthSecret = secret
	return o
}

func (o *APIDeploymentOption) WithGoogle(clientID, clientSecret string) *APIDeploymentOption {
	o.Auth.Google = &OAuthProvider{
		Enabled:      true,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	return o
}

func (o *APIDeploymentOption) WithGitHub(clientID, clientSecret string) *APIDeploymentOption {
	o.Auth.GitHub = &OAuthProvider{
		Enabled:      true,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	return o
}

func (o *APIDeploymentOption) WithDiscord(clientID, clientSecret string) *APIDeploymentOption {
	o.Auth.Discord = &OAuthProvider{
		Enabled:      true,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	return o
}

func (r *kubeProjectRepository) CreateAPIDeployment(ctx context.Context, opt *APIDeploymentOption) error {
	// Build environment variables dynamically
	envVars := []corev1.EnvVar{
		{
			Name:  "BETTER_AUTH_URL",
			Value: fmt.Sprintf("https://%s.%s", opt.Ref, r.config.App.ExternalDomain),
		},
		{
			Name:  "BETTER_AUTH_SECRET",
			Value: opt.BetterAuthSecret,
		},
		{
			Name:  "TRUSTED_ORIGINS",
			Value: strings.Join(opt.TrustedOrigins, ","),
		},
		{
			Name:  "EMAIL_AND_PASSWORD_ENABLED",
			Value: utils.BoolToString(opt.Auth.EmailAndPasswordEnabled),
		},
		{
			Name: "DATABASE_URL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s-app", opt.Ref),
					},
					Key: "uri",
				},
			},
		},
	}

	// Add OAuth provider environment variables
	if opt.Auth.Google != nil {
		envVars = append(envVars,
			corev1.EnvVar{Name: "GOOGLE_ENABLED", Value: utils.BoolToString(opt.Auth.Google.Enabled)},
			corev1.EnvVar{Name: "GOOGLE_CLIENT_ID", Value: opt.Auth.Google.ClientID},
			corev1.EnvVar{Name: "GOOGLE_CLIENT_SECRET", Value: opt.Auth.Google.ClientSecret},
		)
	}

	if opt.Auth.GitHub != nil {
		envVars = append(envVars,
			corev1.EnvVar{Name: "GITHUB_ENABLED", Value: utils.BoolToString(opt.Auth.GitHub.Enabled)},
			corev1.EnvVar{Name: "GITHUB_CLIENT_ID", Value: opt.Auth.GitHub.ClientID},
			corev1.EnvVar{Name: "GITHUB_CLIENT_SECRET", Value: opt.Auth.GitHub.ClientSecret},
		)
	}

	if opt.Auth.Discord != nil {
		envVars = append(envVars,
			corev1.EnvVar{Name: "DISCORD_ENABLED", Value: utils.BoolToString(opt.Auth.Discord.Enabled)},
			corev1.EnvVar{Name: "DISCORD_CLIENT_ID", Value: opt.Auth.Discord.ClientID},
			corev1.EnvVar{Name: "DISCORD_CLIENT_SECRET", Value: opt.Auth.Discord.ClientSecret},
		)
	}

	deploymentName := r.GenAPIDeploymentName(opt.Ref)
	authContainerName := fmt.Sprintf("%s-auth", opt.Ref)
	// Create the deployment object
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: opt.Namespace,
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
							Image: "ghcr.io/wkebaas/project-auth:v0.0.12",
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
	_, err := r.clientset.AppsV1().Deployments(opt.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create API deployment", "error", err)
		return errors.New("failed to create API deployment")
	}

	return nil
}

func (r *kubeProjectRepository) PatchAPIDeployment(ctx context.Context, namespace string, ref string, opt *APIDeploymentOption) error {
	deploymentName := r.GenAPIDeploymentName(ref)
	envVars := []corev1.EnvVar{}

	if opt.BetterAuthSecret != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "BETTER_AUTH_SECRET",
			Value: opt.BetterAuthSecret,
		})
	}
	if len(opt.TrustedOrigins) > 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "TRUSTED_ORIGINS",
			Value: strings.Join(opt.TrustedOrigins, ","),
		})
	}
	if opt.Auth.EmailAndPasswordEnabled {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "EMAIL_AND_PASSWORD_ENABLED",
			Value: utils.BoolToString(opt.Auth.EmailAndPasswordEnabled),
		})
	}
	if opt.Auth.Google != nil {
		envVars = append(envVars,
			corev1.EnvVar{Name: "GOOGLE_ENABLED", Value: utils.BoolToString(opt.Auth.Google.Enabled)},
			corev1.EnvVar{Name: "GOOGLE_CLIENT_ID", Value: opt.Auth.Google.ClientID},
			corev1.EnvVar{Name: "GOOGLE_CLIENT_SECRET", Value: opt.Auth.Google.ClientSecret},
		)
	}
	if opt.Auth.GitHub != nil {
		envVars = append(envVars,
			corev1.EnvVar{Name: "GITHUB_ENABLED", Value: utils.BoolToString(opt.Auth.GitHub.Enabled)},
			corev1.EnvVar{Name: "GITHUB_CLIENT_ID", Value: opt.Auth.GitHub.ClientID},
			corev1.EnvVar{Name: "GITHUB_CLIENT_SECRET", Value: opt.Auth.GitHub.ClientSecret},
		)
	}
	if opt.Auth.Discord != nil {
		envVars = append(envVars,
			corev1.EnvVar{Name: "DISCORD_ENABLED", Value: utils.BoolToString(opt.Auth.Discord.Enabled)},
			corev1.EnvVar{Name: "DISCORD_CLIENT_ID", Value: opt.Auth.Discord.ClientID},
			corev1.EnvVar{Name: "DISCORD_CLIENT_SECRET", Value: opt.Auth.Discord.ClientSecret},
		)
	}

	// Prepare JSON merge patch payload instead of YAML to avoid "invalid character" errors
	payload := map[string]any{
		"spec": map[string]any{
			"template": map[string]any{
				"spec": map[string]any{
					"containers": []map[string]any{
						{
							"name": deploymentName,
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
	_, err = r.clientset.AppsV1().Deployments(namespace).Patch(
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

func (r *kubeProjectRepository) DeleteAPIDeployment(ctx context.Context, namespace string, ref string) error {
	deploymentName := r.GenAPIDeploymentName(ref)

	// Delete the deployment
	err := r.clientset.AppsV1().Deployments(namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete API deployment: %w", err)
	}

	return nil
}
