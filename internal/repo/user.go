package repo

import (
	"baas-api/internal/models"
	"context"
	"errors"
	"log/slog"

	"github.com/patrickmn/go-cache"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrCreateUserFailed                  = errors.New("failed to create user")
	ErrFailedToCheckUserExistsByProvider = errors.New("failed to check if user exists by provider and ID")
)

type UserRepository interface {
	// CreateUserFromIdentity creates a user from an identity provider.
	//
	// Returns the user ID if successful, or an error if the operation fails.
	CreateUserFromIdentity(ctx context.Context, email *string, userEntityID, username, provider, providerID string, identityData datatypes.JSON) (*string, error)
	// Check if the user exists by provider and provider ID.
	CheckUserExistsByProviderAndID(ctx context.Context, provider, providerID string) (bool, error)
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

func (r *userRepository) CreateUserFromIdentity(ctx context.Context, email *string, userEntityID, username, provider, providerID string, identityData datatypes.JSON) (*string, error) {
	object := &models.Object{
		EntityID: &userEntityID,
	}
	user := &models.User{
		Username: username,
		Email:    email,
	}

	identity := &models.Identity{
		Provider:     provider,
		ProviderID:   providerID,
		IdentityData: identityData,
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

	return &user.ID, err
}

func (r *userRepository) CheckUserExistsByProviderAndID(ctx context.Context, provider, providerID string) (bool, error) {
	// Check cache first
	var exists bool
	if cachedExists, found := r.cache.Get("user" + provider + providerID); found {
		exists = cachedExists.(bool)
		return exists, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Identity{}).
		Where("provider = ? AND provider_id = ?", provider, providerID).
		Count(&count).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to check user existence by provider and ID", "error", err)
		return false, ErrFailedToCheckUserExistsByProvider
	}

	// Store the result in cache
	r.cache.Set("user"+provider+providerID, count > 0, cache.DefaultExpiration)

	return count > 0, nil
}
