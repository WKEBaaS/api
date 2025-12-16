// Package usersdb
package usersdb

import (
	"context"

	"baas-api/models"

	"gorm.io/gorm"
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

func (s *UsersDBService) GetClassesChild(ctx context.Context, db *gorm.DB, classIDs []string) ([]models.ClassWithPCID, error) {
	var classes []models.ClassWithPCID

	err := db.WithContext(ctx).
		Debug().
		Table("dbo.classes AS c").
		Select("c.id, c.chinese_name, i.pcid").
		Joins("JOIN dbo.inheritances AS i ON i.ccid = c.id").
		Where("i.pcid in (?)", classIDs).
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

func (s *UsersDBService) GetClassPermissions(ctx context.Context, db *gorm.DB, classID string) ([]models.PermissionWithRoleName, error) {
	var permissions []models.PermissionWithRoleName

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
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 先刪除舊的權限
		// 注意：這裡使用 Unscoped() 是可選的，如果你使用了 Soft Delete 但想要真的刪除資料，請加上 Unscoped()
		if err := tx.Where("class_id = ?", classID).Delete(&models.Permission{}).Error; err != nil {
			return err
		}

		// 2. 如果傳入的權限列表為空，代表只是想清空權限，直接返回
		if len(permissions) == 0 {
			return nil
		}

		// 3. 強制覆蓋 classID，防止呼叫端傳入錯誤的 ID，確保資料一致性
		// 由於 permissions 是 slice，這裡修改會影響到底層 array，這通常是預期行為
		for i := range permissions {
			permissions[i].ClassID = classID
		}

		// 4. 直接批量創建 (Batch Create)
		if err := tx.Create(&permissions).Error; err != nil {
			return err
		}

		return nil
	})
}
