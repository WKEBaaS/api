package repo

import (
	"baas-api/internal/models"
	"context"
	"errors"
	"log/slog"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrCreateProjectFailed = errors.New("failed to create project")
	ErrDeleteProjectFailed = errors.New("failed to delete project")
)

type ProjectRepository interface {
	// Crete Project
	//
	// params:
	// 	@param name project name
	// 	@param ref  must contain exactly 20 alphabetic characters [a-zA-Z]
	//
	// @return id, ref of the created project
	CreateProject(ctx context.Context, name string) (*string, *string, error)
	// DeleteProjectByIDSoft 依 ID 軟刪除專案 (及其關聯的 Object)。
	DeleteProjectByIDSoft(ctx context.Context, id string) error
	// DeleteProjectByIDPermanently 依 ID 永久刪除專案 (及其關聯的 Object)。
	DeleteProjectByIDPermanently(ctx context.Context, id string) error
	// DeleteProjectByRefSoft 依 Reference 軟刪除專案 (及其關聯的 Object)。
	DeleteProjectByRefSoft(ctx context.Context, ref string) error
	// DeleteProjectByRefPermanently 依 Reference 永久刪除專案 (及其關聯的 Object)。
	DeleteProjectByRefPermanently(ctx context.Context, ref string) error
	// GetProjectByID 依 ID 取得專案詳細資訊 (包含關聯的 Object)。
	GetProjectByID(ctx context.Context, id string) (*models.Project, error)
	// GetProjectByRef 依 Reference 取得專案詳細資訊 (包含關聯的 Object)。
	GetProjectByRef(ctx context.Context, ref string) (*models.Project, error)
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

func (r *projectRepository) CreateProject(ctx context.Context, name string) (*string, *string, error) {
	ref := gonanoid.MustGenerate(string(lo.LowerCaseLettersCharset), 20)

	object := &models.Object{
		ChineseName: lo.ToPtr(name),
	}

	project := &models.Project{
		Reference: ref,
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(object).Error; err != nil {
			slog.ErrorContext(ctx, "Failed to create object in transaction for project", "projectName", name, "error", err)
			return ErrCreateProjectFailed
		}

		project.ID = object.ID // Ensure project has the same ID as object
		if err := tx.Create(project).Error; err != nil {
			slog.ErrorContext(ctx, "Failed to create project in transaction", "projectName", name, "objectID", object.ID, "error", err)
			return ErrCreateProjectFailed
		}
		return nil
	})

	return lo.ToPtr(project.ID), lo.ToPtr(ref), err
}

func (r *projectRepository) DeleteProjectByIDSoft(ctx context.Context, id string) error {
	txErr := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&models.Project{}).Error; err != nil {
			// 檢查是否因為找不到記錄而刪除失敗
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.WarnContext(ctx, "Project not found for soft deletion by ID, no action taken", "projectID", id)
				return ErrProjectNotFound
			}
			slog.ErrorContext(ctx, "Failed to soft delete project by ID in transaction", "projectID", id, "error", err)
			return errors.New("failed to delete project")
		}

		if err := tx.Where("id = ?", id).Delete(&models.Object{}).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.WarnContext(ctx, "Associated object not found for soft deletion by ID, project might have been deleted", "objectID", id)
				return ErrProjectNotFound
			}
			slog.ErrorContext(ctx, "Failed to soft delete associated object by ID in transaction", "objectID", id, "error", err)
			return errors.New("failed to delete associated object")
		}
		return nil
	})

	if txErr != nil {
		// 檢查是否是因為 ErrProjectNotFound 導致的交易失敗
		if errors.Is(txErr, ErrProjectNotFound) {
			return ErrProjectNotFound
		}
		slog.ErrorContext(ctx, "Transaction failed during soft delete project by ID", "projectID", id, "error", txErr)
		return ErrTransactionFailed
	}
	slog.InfoContext(ctx, "Project soft deleted successfully by ID", "projectID", id)
	return nil
}

func (r *projectRepository) DeleteProjectByIDPermanently(ctx context.Context, id string) error {
	txErr := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("id = ?", id).Delete(&models.Project{}).Error; err != nil {
			// 檢查是否因為找不到記錄而刪除失敗
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.WarnContext(ctx, "Project not found for soft deletion by ID, no action taken", "projectID", id)
				return ErrProjectNotFound
			}
			slog.ErrorContext(ctx, "Failed to permanently delete project by ID in transaction", "projectID", id, "error", err)
			return errors.New("failed to permanently delete project")
		}

		if err := tx.Unscoped().Where("id = ?", id).Delete(&models.Object{}).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.WarnContext(ctx, "Associated object not found for permanent deletion by ID", "objectID", id)
				return ErrProjectNotFound
			}
			slog.ErrorContext(ctx, "Failed to permanently delete associated object by ID in transaction", "objectID", id, "error", err)
			return errors.New("failed to permanently delete associated object")
		}
		return nil
	})

	if txErr != nil {
		// 檢查是否是因為 ErrProjectNotFound 導致的交易失敗
		if errors.Is(txErr, ErrProjectNotFound) {
			return ErrProjectNotFound
		}
		slog.ErrorContext(ctx, "Transaction failed during permanent delete project by ID", "projectID", id, "error", txErr)
		return ErrTransactionFailed
	}
	slog.InfoContext(ctx, "Project permanently deleted successfully by ID", "projectID", id)
	return nil
}

func (r *projectRepository) DeleteProjectByRefSoft(ctx context.Context, ref string) error {
	var project models.Project
	txErr := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
			Where("reference = ?", ref).
			Delete(&project).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.WarnContext(ctx, "Project not found for soft deletion by reference", "projectRef", ref)
				return ErrProjectNotFound
			}
			slog.ErrorContext(ctx, "Failed to soft delete project by reference in transaction", "projectRef", ref, "error", err)
			return errors.New("failed to soft delete project by reference")
		}

		err = tx.Where("id = ?", project.ID).Delete(&models.Object{}).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.WarnContext(ctx, "Associated object not found for soft deletion by reference", "objectID", project.ID, "projectRef", ref)
				return ErrProjectNotFound
			}
			slog.ErrorContext(ctx, "Failed to soft delete associated object by ID (project found by ref) in transaction", "objectID", project.ID, "projectRef", ref, "error", err)
			return errors.New("failed to soft delete associated object by ID (project found by ref)")
		}

		return nil
	})

	if txErr != nil {
		if errors.Is(txErr, ErrProjectNotFound) {
			return ErrProjectNotFound
		}
		slog.ErrorContext(ctx, "Transaction failed during soft delete project by reference", "projectRef", ref, "error", txErr)
		return ErrTransactionFailed
	}
	slog.InfoContext(ctx, "Project soft deleted successfully by reference", "projectRef", ref, "deletedProjectID", project.ID)
	return nil
}

func (r *projectRepository) DeleteProjectByRefPermanently(ctx context.Context, ref string) error {
	var project models.Project
	txErr := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
			Where("reference = ?", ref).
			Delete(&project).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.WarnContext(ctx, "Project not found for permanent deletion by reference", "projectRef", ref)
				return ErrProjectNotFound
			}
			slog.ErrorContext(ctx, "Failed to permanently delete project by reference in transaction", "projectRef", ref, "error", err)
			return errors.New("failed to permanently delete project by reference")
		}

		err = tx.Unscoped().Where("id = ?", project.ID).Delete(&models.Object{}).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.WarnContext(ctx, "Associated object not found for permanent deletion by reference", "objectID", project.ID, "projectRef", ref)
				return ErrProjectNotFound
			}
			slog.ErrorContext(ctx, "Failed to permanently delete associated object by ID (project found by ref) in transaction", "objectID", project.ID, "projectRef", ref, "error", err)
			return errors.New("failed to permanently delete associated object by ID (project found by ref)")
		}

		return nil
	})

	if txErr != nil {
		if errors.Is(txErr, ErrProjectNotFound) {
			return ErrProjectNotFound
		}
		slog.ErrorContext(ctx, "Transaction failed during permanent delete project by reference", "projectRef", ref, "error", txErr)
		return errors.New("transaction failed during permanent delete project by reference")
	}
	slog.InfoContext(ctx, "Project permanently deleted successfully by reference", "projectRef", ref, "deletedProjectID", project.ID)
	return nil
}

func (r *projectRepository) GetProjectByID(ctx context.Context, id string) (*models.Project, error) {
	var project models.Project
	if err := r.db.WithContext(ctx).Preload("Object").First(&project, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "Project not found by ID", "projectID", id)
			return nil, ErrProjectNotFound
		}
		slog.ErrorContext(ctx, "Failed to get project by ID", "projectID", id, "error", err)
		return nil, errors.New("failed to get project by ID")
	}
	slog.InfoContext(ctx, "Project retrieved successfully by ID", "projectID", id)
	return &project, nil
}

func (r *projectRepository) GetProjectByRef(ctx context.Context, ref string) (*models.Project, error) {
	var project models.Project
	if err := r.db.WithContext(ctx).Preload("Object").First(&project, "reference = ?", ref).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.WarnContext(ctx, "Project not found by reference", "projectRef", ref)
			return nil, ErrProjectNotFound
		}
		slog.ErrorContext(ctx, "Failed to get project by reference", "projectRef", ref, "error", err)
		return nil, errors.New("failed to get project by reference")
	}
	slog.InfoContext(ctx, "Project retrieved successfully by reference", "projectRef", ref, "projectID", project.ID)
	return &project, nil
}
