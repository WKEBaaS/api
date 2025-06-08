package models

import (
	"errors"
	"regexp"
)

var ErrInvalidReference = errors.New("invalid reference format, must be exactly 20 lower alphabetic characters [a-z]")

var refRegex = regexp.MustCompile(`^[a-z]{20}$`)

// Project 對應 dbo.projects 資料表
type Project struct {
	ID        string `gorm:"type:varchar(21);primaryKey;unique"` // VARCHAR(21) NOT NULL UNIQUE, 同時是主鍵
	Reference string `gorm:"type:varchar(20);not null;unique"`   // VARCHAR(20) NOT NULL UNIQUE, 也是外鍵

	// gorm one-to-one
	Object Object `gorm:"foreignKey:ID;references:ID"`
}

func (Project) TableName() string {
	return "dbo.projects"
}

func IsValidReference(ref string) bool {
	return refRegex.MatchString(ref)
}

// Views
type ProjectView struct {
	ID          string  `gorm:"type:varchar(21);primaryKey;unique" json:"id"`
	Name        string  `gorm:"type:varchar(255);not null" json:"name"`
	Description *string `gorm:"type:varchar(4000)" json:"description"`
	OwnerID     string  `gorm:"type:varchar(21)" json:"ownerID"`                   // 外鍵，指向 dbo.users 資料表
	EntityID    string  `gorm:"type:varchar(21);not null" json:"entityID"`         // 外鍵，指向 dbo.entities 資料表
	Reference   string  `gorm:"type:varchar(20);not null;unique" json:"reference"` // 外鍵，指向 dbo.projects 資料表
	CreatedAt   string  `gorm:"type:datetime;not null" json:"createdAt"`
	UpdatedAt   string  `gorm:"type:datetime;not null" json:"updatedAt"`
}

func (ProjectView) TableName() string {
	return "dbo.vd_projects"
}
