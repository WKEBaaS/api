package dto

import (
	"baas-api/internal/models"
	"net/http"
)

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
	InitPasswordCookie *http.Cookie `header:"Set-Cookie" doc:"Initial password cookie for database of the project"`
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
