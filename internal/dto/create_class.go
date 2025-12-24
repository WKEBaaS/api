package dto

import "baas-api/internal/models"

type CreateClassInput struct {
	Body struct {
		ProjectRef    string  `json:"project_ref"`
		ParentClassID string  `json:"parent_class_id"`
		EntityID      *int    `json:"entity_id"`
		ChineseName   string  `json:"chinese_name"`
		ChineseDesc   *string `json:"chinese_description"`
		EnglishName   *string `json:"english_name"`
		EnglishDesc   *string `json:"english_description"`
	}
}

type CreateClassOutput struct {
	Body struct {
		Class models.Class `json:"class"`
	}
}

type DeleteClassInput struct {
	Body struct {
		ProjectRef string `json:"project_ref"`
		ClassID    string `json:"class_id"`
		Recursive  bool   `json:"recursive"`
	}
}
