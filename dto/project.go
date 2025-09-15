package dto

import (
	"baas-api/models"
)

type AuthProvider struct {
	Enabled      bool   `json:"enabled" doc:"Enable this OAuth provider"`
	ClientID     string `json:"clientId,omitempty" doc:"OAuth Client ID"`
	ClientSecret string `json:"clientSecret,omitempty" doc:"OAuth Clien Secret"`
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

type PatchProjectSettingInput struct {
	Body struct {
		Ref            string   `json:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
		Name           *string  `json:"name,omitempty" maxLength:"100" example:"My Project" doc:"Project name"`
		Description    *string  `json:"description,omitempty" maxLength:"4000" required:"false" example:"This is my project" doc:"Project description"`
		TrustedOrigins []string `json:"trustedOrigins,omitempty" example:"https://example.com" doc:"List of trusted origins for CORS"`
		Auth           *struct {
			Email   *AuthProvider `json:"email,omitempty" doc:"Enable email and password authentication"`
			Google  *AuthProvider `json:"google,omitempty" doc:"Google OAuth provider settings"`
			GitHub  *AuthProvider `json:"github,omitempty" doc:"GitHub OAuth provider settings"`
			Discord *AuthProvider `json:"discord,omitempty" doc:"Discord OAuth provider settings"`
		} `json:"auth,omitempty" doc:"Authentication settings for the project"`
	}
}

type DeleteProjectByRefInput struct {
	ID string `query:"id" format:"uuid" example:"550e8400-e29b-41d4-a716-446655440000" doc:"Project ID (UUID)"`
}
type DeleteProjectByRefOutput struct {
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
	Enabled      bool   `json:"enabled" doc:"Whether this OAuth provider is enabled"`
	ClientID     string `json:"clientId,omitempty" doc:"OAuth Client ID"`
	ClientSecret string `json:"clientSecret,omitempty" doc:"OAuth Client Secret"`
}

type GetProjectSettingsOutput struct {
	Body struct {
		ID             string              `json:"id" doc:"Project ID"`
		TrustedOrigins []string            `json:"trustedOrigins" doc:"List of trusted origins for CORS"`
		Auth           ProjectAuthSettings `json:"auth" doc:"Authentication settings for the project"`
		CreatedAt      string              `json:"createdAt" doc:"Project creation timestamp"`
		UpdatedAt      string              `json:"updatedAt" doc:"Project last update timestamp"`
	}
}
