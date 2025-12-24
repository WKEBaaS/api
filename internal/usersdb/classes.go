// Package usersdb
//
// Service methods for managing user database classes and permissions.
package usersdb

import (
	"context"

	"baas-api/internal/dto"
	"baas-api/internal/models"

	"gorm.io/gorm"
)

// GetRootClass 取得根節點
func (s *service) GetRootClass(ctx context.Context, jwt, ref string) (*models.Class, error) {
	// role 固定使用 "superuser"
	db, err := s.GetDB(ctx, jwt, ref, "superuser")
	if err != nil {
		return nil, err
	}

	var class models.Class

	err = db.WithContext(ctx).
		Model(&models.Class{}).
		Where("name_path = '/'").
		First(&class).Error
	if err != nil {
		return nil, err
	}

	return &class, nil
}

// GetRootClasses 取得根節點 (Level 0) 與第一層子節點 (Level 1)
func (s *service) GetRootClasses(ctx context.Context, jwt, ref string) ([]models.Class, error) {
	db, err := s.GetDB(ctx, jwt, ref, "superuser")
	if err != nil {
		return nil, err
	}

	var classes []models.Class

	// 查詢邏輯：
	// 1. hierarchy_level <= 1 (涵蓋 0 和 1)
	// 2. 依照層級排序 (確保 Root 在最上面)
	// 3. 依照 class_rank 排序 (確保同層級的顯示順序)
	err = db.WithContext(ctx).
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
func (s *service) GetChildClasses(ctx context.Context, jwt, ref string, pcid string) ([]models.Class, error) {
	db, err := s.GetDB(ctx, jwt, ref, "superuser")
	if err != nil {
		return nil, err
	}

	var classes []models.Class

	// 查詢邏輯：
	// 1. 從 inheritances 表開始找，篩選 pcid (父類別 ID)
	// 2. Join classes 表，將對應的 ccid (子類別 ID) 的詳細資料撈出來
	// 3. 依照 inheritance.rank 排序 (確保子選單順序正確)
	err = db.WithContext(ctx).
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

func (s *service) GetClassesChild(ctx context.Context, jwt, ref string, classIDs []string) ([]models.ClassWithPCID, error) {
	db, err := s.GetDB(ctx, jwt, ref, "superuser")
	if err != nil {
		return nil, err
	}

	var classes []models.ClassWithPCID

	err = db.WithContext(ctx).
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

func (s *service) GetClassByID(ctx context.Context, jwt, ref string, classID string) (*models.Class, error) {
	db, err := s.GetDB(ctx, jwt, ref, "superuser")
	if err != nil {
		return nil, err
	}

	var class models.Class

	err = db.WithContext(ctx).
		Where("id = ?", classID).
		First(&class).Error
	if err != nil {
		return nil, err
	}

	return &class, nil
}

func (s *service) GetClassPermissions(ctx context.Context, jwt, ref string, classID string) ([]models.PermissionWithRoleName, error) {
	db, err := s.GetDB(ctx, jwt, ref, "superuser")
	if err != nil {
		return nil, err
	}

	var permissions []models.PermissionWithRoleName

	selectClause := `
        p.*,
        CASE p.role_type
            WHEN 'USER' THEN u.name
            WHEN 'GROUP' THEN g.name
            ELSE ''
        END AS role_name
    `

	err = db.WithContext(ctx).
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
func (s *service) UpdateClassPermissions(ctx context.Context, jwt, ref string, classID string, permissions []models.Permission) error {
	db, err := s.GetDB(ctx, jwt, ref, "superuser")
	if err != nil {
		return err
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 先刪除舊的權限
		if err := tx.Where("class_id = ?", classID).Delete(&models.Permission{}).Error; err != nil {
			return err
		}

		// 2. 如果傳入的權限列表為空，代表只是想清空權限，直接返回
		if len(permissions) == 0 {
			return nil
		}

		// 3. 強制覆蓋 classID，防止呼叫端傳入錯誤的 ID，確保資料一致性
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

func (s *service) CreateClass(ctx context.Context, jwt string, in *dto.CreateClassInput) (*models.Class, error) {
	db, err := s.GetDB(ctx, jwt, in.Body.ProjectRef, "superuser")
	if err != nil {
		return nil, err
	}

	class := &models.Class{}

	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create the class
		query := `SELECT * FROM dbo.fn_insert_class(?, ?, ?, ?, ?, ?, ?)`
		tx.Raw(query,
			in.Body.ParentClassID,
			in.Body.EntityID,
			in.Body.ChineseName,
			in.Body.ChineseDesc,
			in.Body.EnglishName,
			in.Body.EnglishDesc,
			nil).
			Scan(class)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return class, nil
}

func (s *service) DeleteClass(ctx context.Context, jwt string, in *dto.DeleteClassInput) error {
	db, err := s.GetDB(ctx, jwt, in.Body.ProjectRef, "superuser")
	if err != nil {
		return err
	}

	// fn_delete_class(p_class_id character varying, p_recursive boolean DEFAULT false)
	query := `SELECT dbo.fn_delete_class(?, ?)`
	err = db.WithContext(ctx).
		Exec(query, in.Body.ClassID, in.Body.Recursive).
		Error

	return err
}
