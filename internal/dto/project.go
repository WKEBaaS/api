package dto

type CreateProjectInput struct {
	Body struct {
		Name        string `json:"name" example:"My Project" doc:"Project name" maxLength:"100"`
		Description string `json:"description" example:"This is my project" doc:"Project description" maxLength:"4000"`
		StorageSize string `json:"storage_size" example:"1Gi" doc:"Storage size for the project" maxLength:"100"`
	}
}

type CreateProjectOutput struct {
	Body struct {
		ID  string `json:"id" doc:"Project ID (nanoid)"`
		Ref string `json:"ref" doc:"Project reference (20 characters [a-zA-Z)"`
	}
}
