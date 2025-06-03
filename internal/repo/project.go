package repo

import (
	"baas-api/internal/models"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	// Crete Project
	//
	// params:
	// 	@param name project name
	// 	@param ref  must contain exactly 20 alphabetic characters [a-zA-Z]
	//
	// @return id, ref of the created project
	CreateProject(name string) (*string, *string, error)
	DeleteProjectByIDSoft(id string) error
	DeleteProjectByIDPermanently(id string) error
	DeleteProjectByRefSoft(ref string) error
	DeleteProjectByRefPermanently(ref string) error
	GetProjectByID(id string) (*models.Project, error)
	GetProjectByRef(ref string) (*models.Project, error)
}

type projectRepository struct {
	DB *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	repo := &projectRepository{
		DB: db,
	}

	repo.DB = db
	return repo
}

func (r *projectRepository) CreateProject(name string) (*string, *string, error) {
	ref := gonanoid.MustGenerate(string(lo.LettersCharset), 20)

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

	return lo.ToPtr(project.ID), lo.ToPtr(ref), err
}

func (r *projectRepository) DeleteProjectByIDSoft(id string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&models.Project{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&models.Object{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *projectRepository) DeleteProjectByIDPermanently(id string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("id = ?", id).Delete(&models.Project{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("id = ?", id).Delete(&models.Object{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *projectRepository) DeleteProjectByRefSoft(ref string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var project models.Project
		if err := tx.Where("reference = ?", ref).First(&project).Error; err != nil {
			return err
		}

		if err := tx.Where("id = ?", project.ID).Delete(&models.Project{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", project.ID).Delete(&models.Object{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *projectRepository) DeleteProjectByRefPermanently(ref string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var project models.Project
		if err := tx.Where("reference = ?", ref).First(&project).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Where("id = ?", project.ID).Delete(&models.Project{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("id = ?", project.ID).Delete(&models.Object{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *projectRepository) GetProjectByID(id string) (*models.Project, error) {
	var project models.Project
	if err := r.DB.Preload("Object").First(&project, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) GetProjectByRef(ref string) (*models.Project, error) {
	var project models.Project
	if err := r.DB.Preload("Object").First(&project, "reference = ?", ref).Error; err != nil {
		return nil, err
	}
	return &project, nil
}
