package models

import (
	"time"

	"gorm.io/gorm"
)

type Entity struct {
	ID           string    `gorm:"type:varchar(21);primaryKey;default:nanoid();unique;not null"`
	Rank         int       `gorm:"column:rank;<-:false"` // `<-:false` prevents GORM from writing to this column
	ChineseName  string    `gorm:"type:varchar(50);column:chinese_name"`
	EnglishName  string    `gorm:"type:varchar(50);column:english_name"`
	IsRelational bool      `gorm:"column:is_relational;default:false;not null"`
	IsHideable   bool      `gorm:"column:is_hideable;default:false;not null"`
	IsDeletable  bool      `gorm:"column:is_deletable;default:false;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null"`
}

func (Entity) TableName() string {
	return "dbo.entities"
}

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
