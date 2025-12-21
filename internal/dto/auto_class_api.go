package dto

import "baas-api/internal/models"

type CreateClassFunctionInput struct {
	Body struct {
		ProjectID     string   `json:"project_id"`
		ProjectRef    string   `json:"project_ref"`
		Name          string   `json:"name"`
		Version       string   `json:"version"`
		Description   string   `json:"description"`
		Authenticated bool     `json:"authenticated"`
		RootNode      RootNode `json:"root_node"`
		Nodes         Node     `json:"nodes"`
	}
}

type RootNode struct {
	ClassID         string `json:"class_id"`
	CheckPermission bool   `json:"check_permission"`
	CheckBits       int16  `json:"check_bits"`
}

type Node struct {
	// Fields is a map because the keys are dynamic.
	// The value is interface{} because the JSON has mixed types
	// (sometimes an object, sometimes a string).
	Fields      NodeFields          `json:"fields"`
	Permissions []models.Permission `json:"permissions,omitempty"`
	Children    []Node              `json:"children,omitempty"`
	// Used to identify the top-level node in the class tree
	Top bool `json:"-"`
}

type FieldConfig struct {
	ParamName *string `json:"param_name,omitempty"`
	Value     *string `json:"value,omitempty"`
}

type NodeFields struct {
	ChineseName        *FieldConfig `json:"chinese_name,omitempty"`
	ChineseDescription *FieldConfig `json:"chinese_description,omitempty"`
	EnglishName        *FieldConfig `json:"english_name,omitempty"`
	EnglishDescription *FieldConfig `json:"english_description,omitempty"`
	EntityID           *FieldConfig `json:"entity_id,omitempty"`
}
