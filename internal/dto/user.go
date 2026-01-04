package dto

type GetRolesInput struct {
	Ref      string `query:"ref" example:"hisqrzwgndjcycmkwpnj" doc:"Project reference (20 lower characters [a-z])"`
	RoleType string `query:"role_type" enum:"USER,GROUP"`
	Query    string `query:"query" doc:"Search query for filtering users by name or email or ID"`
	Limit    int    `query:"limit" required:"false" default:"10" doc:"Maximum number of users to return"`
}

type Role struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email *string `json:"email,omitempty"`
}

type GetRolesOutput struct {
	Body struct {
		Roles []Role `json:"roles" doc:"List of roles in the project's database"`
	}
}
