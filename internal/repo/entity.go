package repo

import (
	"baas-api/internal/models"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

var ErrEntityNotFound = errors.New("entity not found")

type EntityRepository interface {
	GetByChineseName(ctx context.Context, cname string) (*models.Entity, error)
}

type entityRepository struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewEntityRepository(db *gorm.DB, cache *cache.Cache) EntityRepository {
	return &entityRepository{
		db:    db,
		cache: cache,
	}
}

func (r *entityRepository) GetByChineseName(ctx context.Context, name string) (*models.Entity, error) {
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
