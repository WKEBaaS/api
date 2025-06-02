package models

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
