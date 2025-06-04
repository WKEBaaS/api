package dto

type CreateProjectInput struct {
	Body struct {
		Name        string `json:"name" example:"My Project" doc:"Project name" maxLength:"100"`
		Description string `json:"description" example:"This is my project" doc:"Project description" maxLength:"4000"`
		StorageSize string `json:"storageSize" example:"1Gi" doc:"Storage size for the project" maxLength:"100"`
	}
}

type CreateProjectOutput struct {
	Body struct {
		ID  string `json:"id" doc:"Project ID (nanoid)"`
		Ref string `json:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
	}
}

type DeleteProjectByRefInput struct {
	Ref string `query:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
}

type DeleteProjectByRefOutput struct {
	Body struct {
		Success bool `json:"success" doc:"Indicates if the project was successfully deleted"`
	}
}
