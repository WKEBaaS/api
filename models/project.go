// Package models
package models

import (
	"errors"
	"regexp"
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

var ErrInvalidReference = errors.New("invalid reference format, must be exactly 20 lower alphabetic characters [a-z]")

var refRegex = regexp.MustCompile(`^[a-z]{20}$`)

// Project 對應 dbo.projects 資料表
type Project struct {
	ID                string     `gorm:"type:varchar(21);primaryKey;unique"` // VARCHAR(21) NOT NULL UNIQUE, 同時是主鍵
	Reference         string     `gorm:"type:varchar(20);not null;unique"`   // VARCHAR(20) NOT NULL UNIQUE, 也是外鍵
	PasswordExpiredAt *time.Time `gorm:"type:timestamptz;default:now()" json:"password_expired_at"`
	InitializedAt     *time.Time `gorm:"type:timestamptz" json:"initialized_at"`

	// gorm one-to-one
	Object Object `gorm:"foreignKey:ID;references:ID"`
}

func (Project) TableName() string {
	return "dbo.projects"
}

type ProjectAuthSettings struct {
	ID                      string         `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	ProjectID               string         `gorm:"type:uuid;not null;unique" json:"project_id"`
	Secret                  string         `gorm:"type:text;not null;default:encode(gen_random_bytes(32), 'base64')" json:"secret"`
	TrustedOrigins          pq.StringArray `gorm:"type:text[];not null;default:'{}'" json:"trusted_origins"`
	EmailAndPasswordEnabled bool           `gorm:"not null;default:true" json:"email_and_password_enabled"`
	CreatedAt               time.Time      `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt               time.Time      `gorm:"not null;default:now()" json:"updated_at"`
}

// TableName specifies the table name for ProjectAuthSettings
func (ProjectAuthSettings) TableName() string {
	return "dbo.project_auth_settings"
}

// ProjectOAuthProvider represents the project_oauth_providers table
type ProjectOAuthProvider struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Enabled      bool           `gorm:"not null;default:false" json:"enabled"`
	Name         string         `gorm:"type:varchar(50);not null" json:"name"`
	ProjectID    string         `gorm:"column:project_id;type:uuid;not null" json:"project_id"` // 加上 column:project_id
	CreatedAt    time.Time      `gorm:"not null;default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null;default:current_timestamp" json:"updated_at"`
	ClientID     string         `gorm:"column:client_id;type:text;not null" json:"client_id"`         // 加上 column:client_id
	ClientSecret string         `gorm:"column:client_secret;type:text;not null" json:"client_secret"` // 加上 column:client_secret
	ExtraConfig  datatypes.JSON `gorm:"column:extra_config;type:jsonb" json:"extra_config"`           // 加上 column:extra_config
}

// TableName specifies the table name for ProjectOAuthProvider
func (ProjectOAuthProvider) TableName() string {
	return "dbo.project_oauth_providers"
}

func IsValidReference(ref string) bool {
	return refRegex.MatchString(ref)
}

// Views

type ProjectView struct {
	ID                string     `gorm:"type:varchar(21);primaryKey;unique" json:"id"`
	Name              string     `gorm:"type:varchar(255);not null" json:"name"`
	Description       *string    `gorm:"type:varchar(4000)" json:"description"`
	OwnerID           string     `gorm:"type:varchar(21)" json:"ownerID"`                   // 外鍵，指向 dbo.users 資料表
	EntityID          string     `gorm:"type:varchar(21);not null" json:"entityID"`         // 外鍵，指向 dbo.entities 資料表
	Reference         string     `gorm:"type:varchar(20);not null;unique" json:"reference"` // 外鍵，指向 dbo.projects 資料表
	CreatedAt         time.Time  `gorm:"type:timestamptz;not null" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"type:timestamptz;not null" json:"updatedAt"`
	PasswordExpiredAt *time.Time `gorm:"type:timestamptz" json:"passwordExpiredAt"`
	InitializedAt     *time.Time `gorm:"type:timestamptz" json:"initializedAt"`
}

func (ProjectView) TableName() string {
	return "dbo.vd_projects"
}
