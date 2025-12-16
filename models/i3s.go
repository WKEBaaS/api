package models

import (
	"database/sql/driver"
	"time"

	"github.com/lib/pq"
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

type Class struct {
	// Primary Key with nanoid default
	ID string `gorm:"primaryKey;type:varchar(21);default:nanoid();not null" json:"id"`

	// Foreign Key: dbo.entities
	EntityID *int `gorm:"type:integer" json:"entity_id"`

	// Descriptive Fields (Pointers allow for NULL values)
	ChineseName        *string `gorm:"type:varchar(256)" json:"chinese_name"`
	ChineseDescription *string `gorm:"type:varchar(4000)" json:"chinese_description"`
	EnglishName        *string `gorm:"type:varchar(256)" json:"english_name"`
	EnglishDescription *string `gorm:"type:varchar(4000)" json:"english_description"`

	// Unique Constraints
	IDPath   string `gorm:"type:varchar(2300);not null;uniqueIndex:qu_dbo_class_id_path" json:"id_path"`
	NamePath string `gorm:"type:varchar(2300);not null;uniqueIndex:qu_dbo_class_name_path" json:"name_path"`

	// Timestamps & Soft Delete
	CreatedAt time.Time      `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP;not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:timestamp with time zone" json:"deleted_at"`

	// Integer Metrics & Ranks
	ObjectCount    int   `gorm:"type:integer;default:0;not null" json:"object_count"`
	ClassRank      int16 `gorm:"type:smallint;default:0;not null" json:"class_rank"`
	ObjectRank     int16 `gorm:"type:smallint;default:0;not null" json:"object_rank"`
	HierarchyLevel int16 `gorm:"type:smallint;not null" json:"hierarchy_level"`
	ClickCount     int   `gorm:"type:integer;default:0;not null" json:"click_count"`

	// Postgres Array (Requires lib/pq)
	Keywords pq.StringArray `gorm:"type:text[];default:'{}'::text[];not null" json:"keywords"`

	// Owner (Foreign Key to auth.users)
	OwnerID *string `gorm:"type:uuid" json:"owner_id"`

	// Booleans
	IsHidden bool `gorm:"type:boolean;default:false;not null" json:"is_hidden"`
	IsChild  bool `gorm:"type:boolean;default:false;not null" json:"is_child"`
}

type ClassWithPCID struct {
	ID          string  `gorm:"primaryKey;type:varchar(21);default:nanoid();not null" json:"id"`
	ChineseName *string `gorm:"type:varchar(256)" json:"chinese_name"`
	PCID        string  `gorm:"type:varchar(21);column:pcid;not null" json:"pcid"`
}

// TableName specifies the schema and table name strictly
func (Class) TableName() string {
	return "dbo.classes"
}

type PermissionRoleType string

const (
	RoleTypeUser  PermissionRoleType = "USER"
	RoleTypeGroup PermissionRoleType = "GROUP"
)

// Scan enables GORM to read the custom ENUM type from the DB
func (p *PermissionRoleType) Scan(value any) error {
	*p = PermissionRoleType(value.(string))
	return nil
}

// Value enables GORM to write the custom ENUM type to the DB
func (p PermissionRoleType) Value() (driver.Value, error) {
	return string(p), nil
}

type Permission struct {
	ClassID        string             `gorm:"type:varchar(21);not null;uniqueIndex:uq_dbo_permissions,priority:1" json:"class_id"`
	RoleType       PermissionRoleType `gorm:"type:dbo.permission_role_type;not null;uniqueIndex:uq_dbo_permissions,priority:2" json:"role_type"`
	RoleID         string             `gorm:"type:uuid;not null;uniqueIndex:uq_dbo_permissions,priority:3" json:"role_id"`
	PermissionBits int16              `gorm:"type:smallint;default:1;not null" json:"permission_bits"`
}

type PermissionWithRoleName struct {
	Permission
	RoleName string `json:"role_name"`
}

func (Permission) TableName() string {
	return "dbo.permissions"
}

type Inheritance struct {
	PCID string `gorm:"primaryKey;type:varchar(21);not null" json:"pcid"`
	CCID string `gorm:"primaryKey;type:varchar(21);not null" json:"ccid"`

	// relationships
	ParentClass Class `gorm:"foreignKey:PCID;constraint:OnDelete:CASCADE" json:"parent_class"`
	ChildClass  Class `gorm:"foreignKey:CCID;constraint:OnDelete:CASCADE" json:"child_class"`

	Rank            *int16 `gorm:"type:smallint" json:"rank"`
	MembershipGrade *int   `gorm:"type:integer" json:"membership_grade"`
}

// TableName specifies the schema and table name strictly
func (Inheritance) TableName() string {
	return "dbo.inheritances"
}
