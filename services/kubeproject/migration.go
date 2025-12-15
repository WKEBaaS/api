package kubeproject

import (
	"context"
	"errors"
	"log/slog"

	"github.com/samber/lo"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const MigrationFinishedTTLSeconds = 300

func (s *KubeProjectService) CreateMigrationJob(ctx context.Context, ref string) error {
	migJobName := s.GetMigrationJobName(ref)
	roleAppSecretName := s.GetDatabaseRoleSecretName(ref, RoleApp)
	jwkConfigMapName := s.GetJWKSConfigMapName(ref)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      migJobName,
			Namespace: s.namespace,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: lo.ToPtr(int32(300)),
			BackoffLimit:            lo.ToPtr(int32(4)),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  migJobName,
							Image: "ghcr.io/amacneil/dbmate:2",
							Args:  []string{"--wait", "up"},
							Env: []corev1.EnvVar{
								{
									Name: "DATABASE_URL", ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											Key:                  "uri",
											LocalObjectReference: corev1.LocalObjectReference{Name: roleAppSecretName},
										},
									},
								},
								{Name: "DBMATE_MIGRATIONS_DIR", Value: "/migrations"},
							},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "migrations", MountPath: "/migrations", ReadOnly: true},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "migrations",
							VolumeSource: corev1.VolumeSource{
								Projected: &corev1.ProjectedVolumeSource{
									Sources: []corev1.VolumeProjection{
										{ConfigMap: &corev1.ConfigMapProjection{LocalObjectReference: corev1.LocalObjectReference{Name: "migrations"}}},
										{ConfigMap: &corev1.ConfigMapProjection{LocalObjectReference: corev1.LocalObjectReference{Name: jwkConfigMapName}}},
									},
								},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	_, err := s.clientset.BatchV1().Jobs(s.namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create migration job", "error", err, "jobName", migJobName)
		return errors.New("failed to create migration job")
	}

	return nil
}
