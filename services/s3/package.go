package s3

import (
	"baas-api/config"

	"github.com/minio/madmin-go/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/samber/do/v2"
)

func NewMinIOClient(i do.Injector) (*minio.Client, error) {
	cfg := do.MustInvoke[*config.Config](i)
	minioClient, err := minio.New(cfg.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, ""),
		Secure: cfg.S3.UseSSL,
	})
	return minioClient, err
}

func NewMinIOAdminClient(i do.Injector) (*madmin.AdminClient, error) {
	cfg := do.MustInvoke[*config.Config](i)
	minioAdminClient, err := madmin.NewWithOptions(cfg.S3.Endpoint, &madmin.Options{
		Creds:  credentials.NewStaticV4(cfg.S3.AccessKeyID, cfg.S3.SecretAccessKey, ""),
		Secure: cfg.S3.UseSSL,
	})
	return minioAdminClient, err
}

var Package = do.Package(
	do.Lazy(NewMinIOClient),
	do.Lazy(NewMinIOAdminClient),
	do.Lazy(NewS3Service),
)
