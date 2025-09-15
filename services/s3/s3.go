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
	return nil
}

func (s *S3Service) CreateBucketUser(ctx context.Context, bucketName string, accessKey string, secretKey string) error {
	err := s.adminClient.AddUser(ctx, accessKey, secretKey)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create user", "error", err)
		return errors.New("failed to create user")
	}
	return nil
}
