// Package usersdb provides functionalities to interact with the users database.
package usersdb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"baas-api/internal/kubeproject"
	"baas-api/internal/models"
	"baas-api/internal/pgrest"

	"github.com/patrickmn/go-cache"
	"github.com/samber/do/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Service interface {
	// GetDB by baas-project ref
	GetDB(ctx context.Context, jwt, ref, role string) (*gorm.DB, error)
	GetRootClasses(ctx context.Context, db *gorm.DB) ([]models.Class, error)
	GetClassesChild(ctx context.Context, db *gorm.DB, classIDs []string) ([]models.ClassWithPCID, error)
	GetChildClasses(ctx context.Context, db *gorm.DB, pcid string) ([]models.Class, error)
	GetClassByID(ctx context.Context, db *gorm.DB, classID string) (*models.Class, error)
	GetClassPermissions(ctx context.Context, db *gorm.DB, classID string) ([]models.PermissionWithRoleName, error)
	UpdateClassPermissions(ctx context.Context, db *gorm.DB, classID string, permissions []models.Permission) error
}

type service struct {
	// config *config.Config
	kube   kubeproject.Service
	pgrest pgrest.Service
	cache  *cache.Cache
}

var _ Service = (*service)(nil)

func NewService(i do.Injector) (*service, error) {
	return &service{
		kube:   do.MustInvokeAs[kubeproject.Service](i),
		pgrest: do.MustInvokeAs[pgrest.Service](i),
		cache:  do.MustInvoke[*cache.Cache](i),
	}, nil
}

func (s *service) GetDB(ctx context.Context, jwt, ref, role string) (*gorm.DB, error) {
	// 1. 驗證權限
	err := s.pgrest.CheckProjectPermission(ctx, jwt, ref)
	if err != nil {
		return nil, err
	}

	// 2. 生成緩存鍵
	cacheKey := "usersdb:" + ref + ":" + role

	// 3. 嘗試從緩存獲取連接
	if cached, found := s.cache.Get(cacheKey); found {
		if db, ok := cached.(*gorm.DB); ok {
			return db, nil
		}
	}

	// 4. 緩存未命中或連接不健康,創建新連接
	secret, err := s.kube.FindDatabaseRoleSecret(ctx, ref, role)
	if err != nil {
		return nil, err
	}

	uri, ok := secret.Data["uri"]
	if !ok {
		return nil, errors.New("database uri not found in secret")
	}

	dsn := strings.Replace(string(uri), "*", "app", 1)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
			NoLowerCase:   false,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 設置緩存過期時間為 1 小時,過期時自動關閉連接
	s.cache.Set(cacheKey, db, time.Hour)

	return db, nil
}
