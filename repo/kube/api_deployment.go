package kube

import (
	"baas-api/utils"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateAPIDeploymentOption struct {
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

func NewAPIDeploymentOption() *CreateAPIDeploymentOption {
	return &CreateAPIDeploymentOption{}
}

func (o *CreateAPIDeploymentOption) WithNamespace(namespace string) *CreateAPIDeploymentOption {
	o.Namespace = namespace
	return o
}

func (o *CreateAPIDeploymentOption) WithRef(ref string) *CreateAPIDeploymentOption {
	o.Ref = ref
	return o
}

func (o *CreateAPIDeploymentOption) WithTrustedOrigins(origins []string) *CreateAPIDeploymentOption {
	o.TrustedOrigins = origins
	return o
}

func (o *CreateAPIDeploymentOption) WithEmailAndPasswordAuth(enabled bool) *CreateAPIDeploymentOption {
	o.Auth.EmailAndPasswordEnabled = enabled
	return o
}

func (o *CreateAPIDeploymentOption) WithBetterAuthSecret(secret string) *CreateAPIDeploymentOption {
	o.BetterAuthSecret = secret
	return o
}

func (o *CreateAPIDeploymentOption) WithGoogle(clientID, clientSecret string) *CreateAPIDeploymentOption {
	o.Auth.Google = &OAuthProvider{
		Enabled:      true,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	return o
}

func (o *CreateAPIDeploymentOption) WithGitHub(clientID, clientSecret string) *CreateAPIDeploymentOption {
	o.Auth.GitHub = &OAuthProvider{
		Enabled:      true,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	return o
}

func (o *CreateAPIDeploymentOption) WithDiscord(clientID, clientSecret string) *CreateAPIDeploymentOption {
	o.Auth.Discord = &OAuthProvider{
		Enabled:      true,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
	return o
}

func (r *kubeProjectRepository) CreateAPIDeployment(ctx context.Context, opt *CreateAPIDeploymentOption) error {
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

	deploymentName := fmt.Sprintf("%s-api", opt.Ref)
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
							Name:  deploymentName,
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
		return fmt.Errorf("failed to create API deployment")
	}

	return nil
}

func (r *kubeProjectRepository) DeleteAPIDeployment(ctx context.Context, namespace string, ref string) error {
	deploymentName := fmt.Sprintf("%s-api", ref)

	// Delete the deployment
	err := r.clientset.AppsV1().Deployments(namespace).Delete(ctx, deploymentName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete API deployment: %w", err)
	}

	return nil
}
