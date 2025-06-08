package repo

import (
	"baas-api/internal/models"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrCreateUserFailed            = errors.New("failed to create user")
	ErrFailedToGetUserIDByProvider = errors.New("failed to get user ID by provider and ID")
)

type CreateUserFromIdentityInput struct {
	// Object fields
	UserEntityID string
	Name         string
	// User fields
	Email    *string
	Username string
	// Identity fields
	Provider     string
	ProviderID   string
	IdentityData datatypes.JSON
}

type UserRepository interface {
	// CreateFromIdentity creates a user from an identity provider.
	//
	// Returns the user ID if successful, or an error if the operation fails.
	CreateFromIdentity(ctx context.Context, in *CreateUserFromIdentityInput) (*string, error)
	// GetIDByProviderAndID 透過 provider 和 providerID 取得使用者 ID
	// 如果找到，則回傳使用者 ID 和 true
	// 如果找不到，則回傳 0 和 false (假設 UserID 是 uint)
	// 如果發生其他錯誤，則回傳 0 和錯誤
	GetIDByProviderAndID(ctx context.Context, provider, providerID string) (*string, bool, error)
}

type userRepository struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewUserRepository(db *gorm.DB, cache *cache.Cache) UserRepository {
	return &userRepository{
		db:    db,
		cache: cache,
	}
}

func (r *userRepository) CreateFromIdentity(ctx context.Context, in *CreateUserFromIdentityInput) (*string, error) {
	object := &models.Object{
		EntityID:    &in.UserEntityID,
		ChineseName: &in.Name,
	}
	user := &models.User{
		Username: &in.Username,
		Email:    in.Email,
	}

	identity := &models.Identity{
		Provider:     in.Provider,
		ProviderID:   in.ProviderID,
		IdentityData: in.IdentityData,
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(object).Error; err != nil {
			slog.ErrorContext(ctx, "Failed to create object in transaction", "error", err)
			return ErrCreateUserFailed
		}

		user.ID = object.ID
		if err := tx.Create(user).Error; err != nil {
			slog.ErrorContext(ctx, "Failed to create user in transaction", "error", err)
			return ErrCreateUserFailed
		}

		identity.UserID = user.ID
		if err := tx.Create(identity).Error; err != nil {
			slog.ErrorContext(ctx, "Failed to create identity in transaction", "error", err)
			return ErrCreateUserFailed
		}

		return nil
	})

	// update userid to cache
	r.cache.Set("userid"+in.Provider+in.ProviderID, user.ID, cache.DefaultExpiration)

	return &user.ID, err
}

func (r *userRepository) GetIDByProviderAndID(ctx context.Context, provider, providerID string) (*string, bool, error) {
	// 嘗試從快取中取得 UserID
	cacheKey := "userid:" + provider + ":" + providerID
	if cachedUserID, found := r.cache.Get(cacheKey); found {
		if userID, ok := cachedUserID.(string); ok {
			slog.InfoContext(ctx, "User ID retrieved from cache", "provider", provider, "provider_id", providerID, "user_id", userID)
			return &userID, userID != "", nil
		}
	}

	var identity models.Identity
	err := r.db.WithContext(ctx).
		Select("user_id"). // 只選擇 user_id 欄位
		Where("provider = ? AND provider_id = ?", provider, providerID).
		First(&identity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			r.cache.Set(cacheKey, "", 5*time.Minute) // 快取 0 表示找不到，例如 5 分鐘
			slog.InfoContext(ctx, "User ID not found in DB", "provider", provider, "provider_id", providerID)
			return nil, false, nil // 找不到使用者
		}
		slog.ErrorContext(ctx, "Failed to get user ID by provider and ID from DB", "error", err, "provider", provider, "provider_id", providerID)
		return nil, false, ErrFailedToGetUserIDByProvider // 其他資料庫錯誤
	}

	// Store the result in cache
	r.cache.Set(cacheKey, identity.UserID, cache.DefaultExpiration)

	return &identity.UserID, true, nil
}
