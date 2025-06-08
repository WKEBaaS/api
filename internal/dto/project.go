package dto

import "baas-api/internal/models"

type (
	GetProjectByRefInput struct {
		Ref string `query:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
	}
	GetProjectByRefOutput struct {
		Body models.ProjectView
	}
)

type (
	CreateProjectInput struct {
		Body struct {
			Name        string  `json:"name" maxLength:"100" example:"My Project" doc:"Project name"`
			Description *string `json:"description" maxLength:"4000" required:"false" example:"This is my project" doc:"Project description"`
			StorageSize string  `json:"storageSize" hidden:"true" default:"1Gi" example:"1Gi" doc:"Storage size for the project"`
		}
	}
	CreateProjectOutput struct {
		Body struct {
			ID  string `json:"id" doc:"Project ID (nanoid)"`
			Ref string `json:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
		}
	}
)

type (
	DeleteProjectByRefInput struct {
		Ref string `query:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
	}
	DeleteProjectByRefOutput struct {
		Body struct {
			Success bool `json:"success" doc:"Indicates if the project was successfully deleted"`
		}
	}
)

type (
	GetUsersProjectsInput  struct{}
	GetUsersProjectsOutput struct {
		Body struct {
			Projects []*models.ProjectView `json:"projects" doc:"List of projects for the user"`
		}
	}
)
