// Package usersdb
package usersdb

import (
	"context"

	"baas-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetRootClasses 取得根節點 (Level 0) 與第一層子節點 (Level 1)
func (s *UsersDBService) GetRootClasses(ctx context.Context, db *gorm.DB) ([]models.Class, error) {
	var classes []models.Class

	// 查詢邏輯：
	// 1. hierarchy_level <= 1 (涵蓋 0 和 1)
	// 2. 依照層級排序 (確保 Root 在最上面)
	// 3. 依照 class_rank 排序 (確保同層級的顯示順序)
	err := db.WithContext(ctx).
		Model(&models.Class{}).
		Select("dbo.classes.id, dbo.classes.chinese_name").
		Where("hierarchy_level BETWEEN ? AND ?", 0, 1).
		Order("hierarchy_level ASC").
		Order("class_rank ASC").
		Find(&classes).Error
	if err != nil {
		return nil, err
	}

	return classes, nil
}

// GetChildClasses 根據父類別 ID (pcid) 取得其直接子類別
func (s *UsersDBService) GetChildClasses(ctx context.Context, db *gorm.DB, pcid string) ([]models.Class, error) {
	var classes []models.Class

	// 查詢邏輯：
	// 1. 從 inheritances 表開始找，篩選 pcid (父類別 ID)
	// 2. Join classes 表，將對應的 ccid (子類別 ID) 的詳細資料撈出來
	// 3. 依照 inheritance.rank 排序 (確保子選單順序正確)
	err := db.WithContext(ctx).
		Table("dbo.classes").
		Select("dbo.classes.id, dbo.classes.chinese_name"). // 只選取 Class 的欄位，忽略 Inheritance 的欄位
		Joins("JOIN dbo.inheritances ON dbo.inheritances.ccid = dbo.classes.id").
		Where("dbo.inheritances.pcid = ?", pcid).
		Order("dbo.inheritances.rank ASC"). // 通常子節點順序由繼承關係中的 rank 決定
		Find(&classes).Error
	if err != nil {
		return nil, err
	}

	return classes, nil
}

func (s *UsersDBService) GetClassByID(ctx context.Context, db *gorm.DB, classID string) (*models.Class, error) {
	var class models.Class

	err := db.WithContext(ctx).
		Where("id = ?", classID).
		First(&class).Error
	if err != nil {
		return nil, err
	}

	return &class, nil
}

func (s *UsersDBService) GetClassPermissions(ctx context.Context, db *gorm.DB, classID string) ([]models.Permission, error) {
	var permissions []models.Permission

	selectClause := `
		p.*,
		CASE p.role_type
			WHEN 'USER' THEN u.name
			WHEN 'GROUP' THEN g.name
			ELSE ''
		END AS role_name
	`

	err := db.WithContext(ctx).
		// 1. 指定要查詢的欄位
		Select(selectClause).
		// 2. 從 dbo.permissions 表開始 (別名為 p)
		Table(models.Permission{}.TableName()+" AS p").
		Joins(`LEFT JOIN auth.users AS u ON p.role_type = 'USER' AND p.role_id = u.id`).
		Joins(`LEFT JOIN auth.groups AS g ON p.role_type = 'GROUP' AND p.role_id = g.id`).
		Where("p.class_id = ?", classID).
		Scan(&permissions).Error
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

// UpdateClassPermissions Insert or Update class permissions
func (s *UsersDBService) UpdateClassPermissions(ctx context.Context, db *gorm.DB, classID string, permissions []models.Permission) error {
	// Begin a transaction
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Upsert each permission
	for _, perm := range permissions {
		perm.ClassID = classID
		if err := tx.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&perm).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	return tx.Commit().Error
}
