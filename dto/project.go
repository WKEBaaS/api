package dto

import (
	"baas-api/models"
)

type OAuthProvider struct {
	Enabled      bool   `json:"enabled" doc:"Enable this OAuth provider"`
	ClientID     string `json:"clientId" doc:"OAuth Client ID"`
	ClientSecret string `json:"clientSecret" doc:"OAuth Clien Secret"`
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
			EmailAndPasswordEnabled *bool          `json:"emailAndPasswordEnabled,omitempty" doc:"Enable email and password authentication"`
			Google                  *OAuthProvider `json:"google,omitempty" doc:"Google OAuth provider settings"`
			GitHub                  *OAuthProvider `json:"github,omitempty" doc:"GitHub OAuth provider settings"`
			Discord                 *OAuthProvider `json:"discord,omitempty" doc:"Discord OAuth provider settings"`
		} `json:"auth,omitempty" doc:"Authentication settings for the project"`
	}
}

type DeleteProjectByRefInput struct {
	Reference string `query:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
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
