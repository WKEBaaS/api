package kubeproject

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const INSERT_JWKS_SQL_FILENAME = "001001_insert_jwks.sql"

func (s *KubeProjectService) CreateJWKSConfigMap(ctx context.Context, ref string, publicKey string, privateKey string) error {
	configMapName := s.GetJWKSConfigMapName(ref)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: s.namespace,
		},
		Data: map[string]string{
			INSERT_JWKS_SQL_FILENAME: fmt.Sprintf(`-- migrate:up
INSERT INTO auth.jwks (public_key, private_key) VALUES ('%s', '%s');
-- migrate:down
`, publicKey, privateKey),
		},
	}
	_, err := s.clientset.CoreV1().ConfigMaps(s.namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create JWKS ConfigMap", "error", err, "configMapName", configMapName)
		return errors.New("Failed to create JWKS ConfigMap")
	}
	return nil
}

func (s *KubeProjectService) DeleteJWKSConfigMap(ctx context.Context, ref string) error {
	configMapName := s.GetJWKSConfigMapName(ref)
	err := s.clientset.CoreV1().ConfigMaps(s.namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete JWKS ConfigMap", "error", err, "configMapName", configMapName)
		return errors.New("Failed to delete JWKS ConfigMap")
	}
	return nil
}
