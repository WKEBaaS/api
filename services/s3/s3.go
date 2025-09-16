// Package s3 MinIO Related Code
package s3

import (
	"baas-api/config"
	"context"
	"errors"
	"log/slog"

	"github.com/minio/madmin-go/v4"
	"github.com/minio/minio-go/v7"
)

type S3ServiceInterface interface {
	CreateBucket(ctx context.Context, bucketName string) error
	DeleteBucket(ctx context.Context, bucketName string) error
	CreateBucketUser(ctx context.Context, accessKeyID string, secretAccessKey string) error
	DeleteBucketUser(ctx context.Context, accessKeyID string) error
	CreateBucketPolicy(ctx context.Context, ref string, bucketName string) error
	DeleteBucketPolicy(ctx context.Context, bucketname string) error
}

type S3Service struct {
	config      *config.Config      `di.inject:"config"`
	client      *minio.Client       `di.inject:"minioClient"`
	adminClient *madmin.AdminClient `di.inject:"minioAdminClient"`
}

func NewS3Service() S3ServiceInterface {
	return &S3Service{}
}

func (s *S3Service) CreateBucket(ctx context.Context, bucketName string) error {
	err := s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: s.config.S3.Region})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create bucket", "error", err)
		return errors.New("failed to create bucket")
	}

	err = s.adminClient.SetBucketQuota(ctx, bucketName, &madmin.BucketQuota{
		Type: madmin.HardQuota,
		Size: 1 << 30, // 1 GiB
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to set bucket quota", "error", err)
		return errors.New("failed to set bucket quota")
	}
	return nil
}

func (s *S3Service) DeleteBucket(ctx context.Context, bucketName string) error {
	err := s.client.RemoveBucket(ctx, bucketName)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete bucket", "error", err)
		return errors.New("failed to delete bucket")
	}
	return nil
}

func (s *S3Service) CreateBucketUser(ctx context.Context, accessKeyID string, secretAccessKey string) error {
	err := s.adminClient.AddUser(ctx, accessKeyID, secretAccessKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create user", "error", err)
		return errors.New("failed to create user")
	}

	return nil
}

func (s *S3Service) DeleteBucketUser(ctx context.Context, ref string) error {
	err := s.adminClient.RemoveUser(ctx, ref)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete user", "error", err)
		return errors.New("failed to delete user")
	}
	return nil
}

func (s *S3Service) CreateBucketPolicy(ctx context.Context, ref string, bucketName string) error {
	policy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"s3:GetBucketLocation",
					"s3:ListBucket"
				],
				"Resource": [
					"arn:aws:s3:::` + bucketName + `"
				]
			},
			{
				"Effect": "Allow",
				"Action": [
					"s3:PutObject",
					"s3:GetObject",
					"s3:DeleteObject"
				],
				"Resource": [
					"arn:aws:s3:::` + bucketName + `/*"
				]
			}
		]
	}`

	err := s.adminClient.AddCannedPolicy(ctx, bucketName, []byte(policy))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to set user policy", "error", err)
		return errors.New("failed to set user policy")
	}
	_, err = s.adminClient.AttachPolicy(ctx, madmin.PolicyAssociationReq{
		User:     ref,
		Policies: []string{bucketName},
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to attach user policy", "error", err)
		return errors.New("failed to attach user policy")
	}
	return nil
}

func (s *S3Service) DeleteBucketPolicy(ctx context.Context, bucketName string) error {
	err := s.adminClient.RemoveCannedPolicy(ctx, bucketName)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete user policy", "error", err)
		return errors.New("failed to delete user policy")
	}
	return nil
}
