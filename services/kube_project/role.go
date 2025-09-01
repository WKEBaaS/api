package kube_project

import (
	"context"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *KubeProjectService) buildDatabaseRoleSecret(ref string, role string, password string) *corev1.Secret {
	secretName := r.GetDatabaseRoleSecretName(ref, role)

	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: r.namespace,
			Labels: map[string]string{
				"cnpg.io/reload": "true",
			},
		},
		Type: "kubernetes.io/basic-auth",
		StringData: map[string]string{
			"username": role,
			"password": password,
			"uri":      fmt.Sprintf("postgresql://%s:%s@%s-db-rw:5432/app", role, password, ref),
		},
	}
}

func (r *KubeProjectService) CreateDatabaseRoleSecret(ctx context.Context, ref string, role string, password string) error {
	secret := r.buildDatabaseRoleSecret(ref, role, password)

	_, err := r.clientset.CoreV1().Secrets(r.namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "failed to create database role secret",
			"secret_name", secret.Name,
			"namespace", r.namespace,
			"error", err,
		)
		return fmt.Errorf("failed to create database role secret")
	}

	return nil
}

func (r *KubeProjectService) UpdateDatabaseRoleSecret(ctx context.Context, ref string, role string, password string) error {
	secret := r.buildDatabaseRoleSecret(ref, role, password)

	_, err := r.clientset.CoreV1().Secrets(r.namespace).Update(ctx, secret, metav1.UpdateOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "failed to update database role secret",
			"secret_name", secret.Name,
			"namespace", r.namespace,
			"error", err,
		)
		return fmt.Errorf("failed to update database role secret")
	}

	return nil
}

func (r *KubeProjectService) FindDatabaseRolePassword(ctx context.Context, ref string, role string) (*string, error) {
	secretName := r.GetDatabaseRoleSecretName(ref, role)
	secret, err := r.clientset.CoreV1().Secrets(r.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		slog.Error("Failed to read database secret", "error", err)
		return nil, ErrFailedToReadDatabaseSecret
	}

	passwordBytes, ok := secret.Data["password"]
	if !ok {
		slog.Error("Password not found in database secret", "secretName", secretName)
		return nil, ErrFailedToReadDatabaseSecret
	}

	password := string(passwordBytes)
	return &password, nil
}
