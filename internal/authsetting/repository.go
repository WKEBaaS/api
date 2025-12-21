// Package authsetting implements the repository for project authentication settings.
package authsetting

import (
	"context"
	"errors"
	"log/slog"

	"baas-api/internal/models"

	"github.com/samber/do/v2"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrProjectAuthSettingCreateFailed   = errors.New("ProjectAuthSettingCreateFailed")
	ErrProjectAuthSettingNotFound       = errors.New("ProjectAuthSettingNotFound")
	ErrProjectOAuthProviderCreateFailed = errors.New("ProjectOAuthProviderCreateFailed")
	ErrDatabaseError                    = errors.New("DatabaseError")
)

type Repository interface {
	Create(ctx context.Context, s *models.ProjectAuthSettings) error
	FindByProjectID(ctx context.Context, projectID string) (*models.ProjectAuthSettings, error)
	Update(ctx context.Context, s *models.ProjectAuthSettings) error
	DeleteByProjectID(ctx context.Context, projectID string) error
	CreateOAuthProvider(ctx context.Context, provider *models.ProjectAuthProvider) (*string, error)
	UpsertOAuthProviders(ctx context.Context, providers []*models.ProjectAuthProvider) error
	FindAllOAuthProviders(ctx context.Context, projectID string) ([]*models.ProjectAuthProvider, error)
	UpdateOrInsertOAuthProvider(ctx context.Context, provider *models.ProjectAuthProvider) error
}

var _ Repository = (*repository)(nil)

type repository struct {
	db *gorm.DB
}

func NewRepository(i do.Injector) (*repository, error) {
	db := do.MustInvoke[*gorm.DB](i)
	return &repository{
		db: db,
	}, nil
}

func (r *repository) Create(ctx context.Context, s *models.ProjectAuthSettings) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create project auth setting", "error", err)
		return ErrProjectAuthSettingCreateFailed
	}
	return nil
}

func (r *repository) FindByProjectID(ctx context.Context, projectID string) (*models.ProjectAuthSettings, error) {
	var setting models.ProjectAuthSettings
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "Project auth setting not found", "projectID", projectID)
			return nil, ErrProjectAuthSettingNotFound
		}
		slog.ErrorContext(ctx, "Failed to find project auth setting", "error", err)
		return nil, ErrDatabaseError
	}

	return &setting, nil
}

func (r *repository) Update(ctx context.Context, s *models.ProjectAuthSettings) error {
	if err := r.db.WithContext(ctx).Model(&models.ProjectAuthSettings{}).Where("project_id = ?", s.ProjectID).Updates(s).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to update project auth setting", "error", err)
		return ErrDatabaseError
	}
	return nil
}

func (r *repository) DeleteByProjectID(ctx context.Context, projectID string) error {
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Delete(&models.ProjectAuthSettings{}).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to delete project auth setting", "error", err)
		return ErrDatabaseError
	}
	return nil
}

func (r *repository) CreateOAuthProvider(ctx context.Context, provider *models.ProjectAuthProvider) (*string, error) {
	if err := r.db.WithContext(ctx).Create(provider).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create project OAuth provider", "error", err)
		return nil, ErrProjectOAuthProviderCreateFailed
	}
	return lo.ToPtr(provider.ID), nil
}

func (r *repository) FindAllOAuthProviders(ctx context.Context, projectID string) ([]*models.ProjectAuthProvider, error) {
	var providers []*models.ProjectAuthProvider
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&providers).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to find OAuth providers", "error", err)
		return nil, ErrDatabaseError
	}
	return providers, nil
}

func (r *repository) UpdateOrInsertOAuthProvider(ctx context.Context, provider *models.ProjectAuthProvider) error {
	var existingProvider models.ProjectAuthProvider
	err := r.db.WithContext(ctx).Where("project_id = ? AND name = ?", provider.ProjectID, provider.Name).First(&existingProvider).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.ErrorContext(ctx, "Failed to find existing OAuth provider", "error", err)
		return ErrDatabaseError
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new provider
		if _, err := r.CreateOAuthProvider(ctx, provider); err != nil {
			return err
		}
	} else {
		// Update existing provider
		provider.ID = existingProvider.ID // Ensure we use the existing ID
		if err := r.db.WithContext(ctx).Save(provider).Error; err != nil {
			slog.ErrorContext(ctx, "Failed to update OAuth provider", "error", err)
			return ErrDatabaseError
		}
	}

	return nil
}

func (r *repository) UpsertOAuthProviders(ctx context.Context, providers []*models.ProjectAuthProvider) error {
	// FIX: add debug here, since gorm will raise a weird error that insert with  all value null/empty
	err := r.db.Debug().WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "project_id"},
			{Name: "name"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"enabled",
			"client_id",
			"client_secret",
			"extra_config",
			"updated_at",
		}),
	}).Create(providers).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to upsert OAuth providers", "error", err)
		return ErrDatabaseError
	}
	return nil
}
