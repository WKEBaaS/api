package dto

import (
	"baas-api/internal/models"
)

type AuthProvider struct {
	Enabled      bool    `json:"enabled" doc:"Enable this OAuth provider"`
	ClientID     *string `json:"clientId,omitempty" doc:"OAuth Client ID"`
	ClientSecret *string `json:"clientSecret,omitempty" doc:"OAuth Clien Secret"`
}

type GetProjectByRefInput struct {
	Ref string `query:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
}

type GetProjectByRefOutput struct {
	Body models.ProjectView
}

type CreateProjectInput struct {
	Body struct {
		Name        string  `json:"name" maxLength:"100" example:"My Project" doc:"Project name"`
		Description *string `json:"description" maxLength:"4000" required:"false" example:"This is my project" doc:"Project description"`
		StorageSize string  `json:"storageSize" hidden:"true" default:"1Gi" example:"1Gi" doc:"Storage size for the project"`
	}
}

type CreateProjectOutput struct {
	Body struct {
		ID        string `json:"id" doc:"Project ID (nanoid)"`
		Reference string `json:"reference" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
	}
}

// UpdateProjectInput 是更新專案設定的 payload
type UpdateProjectInput struct {
	Body struct {
		ID             string   `json:"id" format:"uuid"`
		Name           *string  `json:"name,omitempty" maxLength:"100"`
		Description    *string  `json:"description,omitempty" maxLength:"4000"`
		TrustedOrigins []string `json:"trustedOrigins,omitempty"`
		ProxyURL       *string  `json:"proxyUrl,omitempty"`
		// 使用 map 來動態接收任意數量的 auth provider
		Auth map[string]AuthProvider `json:"auth,omitempty"`
	}
}

type DeleteProjectByIDInput struct {
	ID string `query:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440000" doc:"Project ID (UUID)"`
}
type DeleteProjectByIDOutput struct {
	Body struct {
		Success bool `json:"success" doc:"Indicates if the project was successfully deleted"`
	}
}

type GetUsersProjectsInput struct{}

type GetUsersProjectsOutput struct {
	Body struct {
		Projects []*models.ProjectView `json:"projects" doc:"List of projects for the user"`
	}
}

type ResetDatabasePasswordInput struct {
	Body struct {
		Reference string `json:"reference" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
		Password  string `json:"password" example:"newpassword123" doc:"New password for the project's database"`
	}
}

type ResetDatabasePasswordOutput struct {
	Body struct {
		Success bool `json:"success" doc:"Indicates if the password was successfully reset"`
	}
}

type GetProjectSettingsInput struct {
	Ref string `query:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
}

type ProjectAuthSettings struct {
	Email   *ProjectAuthProviderInfo `json:"email,omitempty" doc:"Email and password authentication settings"`
	Google  *ProjectAuthProviderInfo `json:"google,omitempty" doc:"Google OAuth provider settings"`
	GitHub  *ProjectAuthProviderInfo `json:"github,omitempty" doc:"GitHub OAuth provider settings"`
	Discord *ProjectAuthProviderInfo `json:"discord,omitempty" doc:"Discord OAuth provider settings"`
}

type ProjectAuthProviderInfo struct {
	Enabled      bool    `json:"enabled" doc:"Whether this OAuth provider is enabled"`
	ClientID     *string `json:"clientId,omitempty" doc:"OAuth Client ID"`
	ClientSecret *string `json:"clientSecret,omitempty" doc:"OAuth Client Secret"`
}

type GetProjectSettingsOutput struct {
	Body struct {
		ID             string              `json:"id" doc:"Project ID"`
		ProjectID      string              `json:"projectId" doc:"Project unique identifier"`
		TrustedOrigins []string            `json:"trustedOrigins" doc:"List of trusted origins for CORS"`
		ProxyURL       *string             `json:"proxyURL,omitempty" doc:"Proxy URL if set"`
		Auth           ProjectAuthSettings `json:"auth" doc:"Authentication settings for the project"`
		CreatedAt      string              `json:"createdAt" doc:"Project creation timestamp"`
		UpdatedAt      string              `json:"updatedAt" doc:"Project last update timestamp"`
	}
}

type GetUsersRootClassOutput struct {
	Body struct {
		Class models.Class `json:"class" doc:"Root class details"`
	}
}

type GetUsersRootClassesOutput struct {
	Body struct {
		Classes []models.Class `json:"classes" doc:"List of first-level classes"`
	}
}

type GetUsersChildClassesInput struct {
	Ref  string `query:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
	PCID string `json:"pcid" query:"pcid" doc:"Parent Class ID"`
}

type GetUsersChildClassesOutput struct {
	Body struct {
		Classes []models.Class `json:"classes" doc:"List of child classes"`
	}
}

type GetUsersClassByIDInput struct {
	Ref     string `query:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
	ClassID string `query:"class_id" doc:"Class ID to retrieve"`
}

type GetUsersClassByIDOutput struct {
	Body struct {
		Class models.Class `json:"class" doc:"Class details"`
	}
}

type GetUsersClassPermissionsInput struct {
	Ref     string `query:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
	ClassID string `query:"class_id" doc:"Class ID to retrieve permissions for"`
}

type GetUsersClassPermissionsOutput struct {
	Body struct {
		Permissions []models.PermissionWithRoleName `json:"permissions" doc:"List of permissions for the class"`
	}
}

type UpdateUsersClassPermissionsInput struct {
	Body struct {
		Ref         string              `json:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
		ClassID     string              `json:"class_id" doc:"Class ID to update permissions for"`
		Permissions []models.Permission `json:"permissions" doc:"List of permissions to set for the class"`
	}
}

type GetUsersClassesChildInput struct {
	Ref      string   `query:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
	ClassIDs []string `query:"class_ids" doc:"List of Class IDs to retrieve children for"`
}

type GetUsersClassesChildOutput struct {
	Body struct {
		Classes []models.ClassWithPCID `json:"classes" doc:"List of classes with their parent class IDs"`
	}
}
