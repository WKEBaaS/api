package models

import (
	"time"

	"gorm.io/gorm"
)

type Object struct {
	ID                 string         `gorm:"type:varchar(21);primaryKey;default:nanoid();unique"`
	EntityID           *string        `gorm:"type:varchar(21)"` // 可為 NULL，使用指標
	ChineseName        *string        `gorm:"type:varchar(512)"`
	ChineseDescription *string        `gorm:"type:varchar(4000)"`
	EnglishName        *string        `gorm:"type:varchar(512)"`
	EnglishDescription *string        `gorm:"type:varchar(4000)"`
	CreatedAt          time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt          time.Time      `gorm:"type:timestamptz;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt          gorm.DeletedAt `gorm:"index;type:timestamptz"` // GORM 的軟刪除
	OwnerID            *string        `gorm:"type:varchar(21)"`
	ClickCount         int            `gorm:"type:int;not null;default:0"`
	OutlinkCount       *int           `gorm:"type:int"`
	InlinkCount        *int           `gorm:"type:int"`
	IsHidden           bool           `gorm:"type:boolean;not null;default:false"`
}

func (Object) TableName() string {
	return "dbo.objects"
}
