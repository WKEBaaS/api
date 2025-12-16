// Package usersdb provides functionalities to interact with the users database.
package usersdb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"baas-api/models"
	"baas-api/repo"
	"baas-api/services/kubeproject"

	"github.com/patrickmn/go-cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type UsersDBServiceInterface interface {
	// GetDB by baas-project ref
	GetDB(ctx context.Context, ref, userID, role string) (*gorm.DB, error)
	GetRootClasses(ctx context.Context, db *gorm.DB) ([]models.Class, error)
	GetClassesChild(ctx context.Context, db *gorm.DB, classIDs []string) ([]models.ClassWithPCID, error)
	GetChildClasses(ctx context.Context, db *gorm.DB, pcid string) ([]models.Class, error)
	GetClassByID(ctx context.Context, db *gorm.DB, classID string) (*models.Class, error)
	GetClassPermissions(ctx context.Context, db *gorm.DB, classID string) ([]models.PermissionWithRoleName, error)
	UpdateClassPermissions(ctx context.Context, db *gorm.DB, classID string, permissions []models.Permission) error
}

type UsersDBService struct {
	// config *config.Config                          `di.inject:"config"`
	kube    kubeproject.KubeProjectServiceInterface `di.inject:"kubeProjectService"`
	project repo.ProjectRepositoryInterface         `di.inject:"projectRepository"`
	cache   *cache.Cache                            `di.inject:"cache"`
}

func NewUsersDBService() UsersDBServiceInterface {
	return &UsersDBService{}
}

func (s *UsersDBService) GetDB(ctx context.Context, ref, userID, role string) (*gorm.DB, error) {
	// 1. 驗證權限
	isOwner, err := s.project.IsOwner(ctx, ref, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner && role == "owner" {
		return nil, errors.New("only project owner can use owner role to connect database")
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
