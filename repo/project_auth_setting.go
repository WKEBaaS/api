// Package repo
package repo

import (
	"baas-api/models"
	"context"
	"errors"
	"log/slog"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

var (
	ErrProjectAuthSettingCreateFailed   = errors.New("ProjectAuthSettingCreateFailed")
	ErrProjectAuthSettingNotFound       = errors.New("ProjectAuthSettingNotFound")
	ErrProjectOAuthProviderCreateFailed = errors.New("ProjectOAuthProviderCreateFailed")
)

type ProjectAuthSettingRepository interface {
	Create(ctx context.Context, s *models.ProjectAuthSettings) error
	FindByProjectID(ctx context.Context, projectID string) (*models.ProjectAuthSettings, error)
	DeleteByProjectID(ctx context.Context, projectID string) error
	CreateOAuthProvider(ctx context.Context, provider *models.ProjectOAuthProvider) (*string, error)
	FindAllOAuthProviders(ctx context.Context, projectID string) ([]*models.ProjectOAuthProvider, error)
}

type projectAuthSettingRepository struct {
	db *gorm.DB
}

func NewProjectAuthSettingRepository(db *gorm.DB) ProjectAuthSettingRepository {
	return &projectAuthSettingRepository{
		db: db,
	}
}

func (r *projectAuthSettingRepository) Create(ctx context.Context, s *models.ProjectAuthSettings) error {
	if err := r.db.WithContext(ctx).Create(s).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create project auth setting", "error", err)
		return ErrProjectAuthSettingCreateFailed
	}
	return nil
}

func (r *projectAuthSettingRepository) FindByProjectID(ctx context.Context, projectID string) (*models.ProjectAuthSettings, error) {
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

func (r *projectAuthSettingRepository) DeleteByProjectID(ctx context.Context, projectID string) error {
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Delete(&models.ProjectAuthSettings{}).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to delete project auth setting", "error", err)
		return ErrDatabaseError
	}
	return nil
}

func (r *projectAuthSettingRepository) CreateOAuthProvider(ctx context.Context, provider *models.ProjectOAuthProvider) (*string, error) {
	if err := r.db.WithContext(ctx).Create(provider).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to create project OAuth provider", "error", err)
		return nil, ErrProjectOAuthProviderCreateFailed
	}
	return lo.ToPtr(provider.ID), nil
}

func (r *projectAuthSettingRepository) FindAllOAuthProviders(ctx context.Context, projectID string) ([]*models.ProjectOAuthProvider, error) {
	var providers []*models.ProjectOAuthProvider
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&providers).Error; err != nil {
		slog.ErrorContext(ctx, "Failed to find OAuth providers", "error", err)
		return nil, ErrDatabaseError
	}
	return providers, nil
}
