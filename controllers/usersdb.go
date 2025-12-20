package controllers

type UsersDBControllerInterface interface {
	// RegisterGetRootClasses(api huma.API)
}

type UsersDBController struct{}

func NewUsersDBController() UsersDBControllerInterface {
	return &UsersDBController{}
}
