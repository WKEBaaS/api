package models

import (
	"time"

	"gorm.io/datatypes"
)

type User struct {
	ID                 string         `gorm:"type:varchar(21);primaryKey;not null"` // UNIQUE 已由 primaryKey 隱含
	Username           *string        `gorm:"type:varchar(255);unique"`
	Email              *string        `gorm:"type:varchar(255);unique"`
	EncryptedPassword  *string        `gorm:"type:varchar(255)"`
	ConfirmedAt        *time.Time     `gorm:"type:timestamptz"`
	InvitedAt          *time.Time     `gorm:"type:timestamptz"`
	ConfirmationToken  *string        `gorm:"type:varchar(255);index"`
	ConfirmationSentAt *time.Time     `gorm:"type:timestamptz"`
	RecoveryToken      *string        `gorm:"type:varchar(255);index"`
	RecoverySentAt     *time.Time     `gorm:"type:timestamptz"`
	EmailChangeToken   *string        `gorm:"type:varchar(255);index"`
	EmailChange        *string        `gorm:"type:varchar(255)"`
	EmailChangeSentAt  *time.Time     `gorm:"type:timestamptz"`
	LastSignInAt       *time.Time     `gorm:"type:timestamptz"`
	RawAppMetaData     datatypes.JSON `gorm:"type:jsonb"`
	RawUserMetaData    datatypes.JSON `gorm:"type:jsonb"`

	Object Object `gorm:"foreignKey:ID;references:ID"`
}

func (User) TableName() string {
	return "auth.users"
}

type Identity struct {
	ID           string         `gorm:"type:varchar(21);primaryKey;default:nanoid();not null"`
	ProviderID   string         `gorm:"type:text;not null;uniqueIndex:uq_auth_idp_id"`
	UserID       string         `gorm:"type:varchar(21);not null;index"` // 外鍵，建議加上索引
	IdentityData datatypes.JSON `gorm:"type:jsonb;not null"`
	Provider     string         `gorm:"type:text;not null;uniqueIndex:uq_auth_idp_id"`
	LastSignInAt *time.Time     `gorm:"type:timestamptz"`
	CreatedAt    *time.Time     `gorm:"type:timestamptz"`
	UpdatedAt    *time.Time     `gorm:"type:timestamptz"`
	Email        *string        `gorm:"type:text;->"` // GENERATED column, GORM 設為唯讀

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName 指定 GORM 使用的資料表名稱 (包含 schema)
func (Identity) TableName() string {
	return "auth.identities"
}
