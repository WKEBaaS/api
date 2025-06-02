package repo

import (
	"baas-api/internal/models"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

func (r *Repository) CreateProject(name string, ref string) (string, error) {
	object := &models.Object{
		ChineseName: lo.ToPtr(name),
	}

	project := &models.Project{
		Reference: ref,
	}

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(object).Error; err != nil {
			return err
		}

		project.ID = object.ID // Ensure project has the same ID as object
		if err := tx.Create(project).Error; err != nil {
			return err
		}
		return nil
	})

	return project.ID, err
}
