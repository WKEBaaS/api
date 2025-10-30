package kubeproject

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const InsertJwksSQLFilename = "001001_insert_jwks.sql"

type CreateJWKSConfigMapOption struct {
	Ref        string
	KID        string
	PublicKey  string
	PrivateKey string
}

func (s *KubeProjectService) CreateJWKSConfigMap(ctx context.Context, opt CreateJWKSConfigMapOption) error {
	configMapName := s.GetJWKSConfigMapName(opt.Ref)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: s.namespace,
		},
		Data: map[string]string{
			InsertJwksSQLFilename: fmt.Sprintf(`-- migrate:up
INSERT INTO auth.jwks (id, public_key, private_key) VALUES ('%s', '%s', '%s');
-- migrate:down
`, opt.KID, opt.PublicKey, opt.PrivateKey),
		},
	}
	_, err := s.clientset.CoreV1().ConfigMaps(s.namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create JWKS ConfigMap", "error", err, "configMapName", configMapName)
		return errors.New("failed to create JWKS ConfigMap")
	}
	return nil
}

func (s *KubeProjectService) DeleteJWKSConfigMap(ctx context.Context, ref string) error {
	configMapName := s.GetJWKSConfigMapName(ref)
	err := s.clientset.CoreV1().ConfigMaps(s.namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete JWKS ConfigMap", "error", err, "configMapName", configMapName)
		return errors.New("failed to delete JWKS ConfigMap")
	}
	return nil
}
