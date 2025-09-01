package repo

import (
	"baas-api/models"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

var ErrEntityNotFound = errors.New("entity not found")

type EntityRepositoryInterface interface {
	GetByChineseName(ctx context.Context, cname string) (*models.Entity, error)
}

type EntityRepository struct {
	db    *gorm.DB     `di.inject:"db"`
	cache *cache.Cache `di.inject:"cache"`
}

func NewEntityRepository(db *gorm.DB, cache *cache.Cache) EntityRepositoryInterface {
	return &EntityRepository{
		db:    db,
		cache: cache,
	}
}

func (r *EntityRepository) GetByChineseName(ctx context.Context, name string) (*models.Entity, error) {
	// Check cache first
	var entity models.Entity
	if cachedEntity, found := r.cache.Get("entity" + name); found {
		entity = cachedEntity.(models.Entity)
		return &entity, nil
	}

	// If not found in cache, query the database
	if err := r.db.WithContext(ctx).Where("chinese_name = ?", name).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "Entity not found in database", "chinese_name", name)
			return nil, ErrEntityNotFound
		}
		slog.ErrorContext(ctx, "Failed to query entity from database", "error", err)
		return nil, ErrTransactionFailed
	}

	// Store the entity in cache
	r.cache.Set("entity"+name, entity, time.Hour)

	return &entity, nil
}
