package controllers

import (
	"baas-api/controllers/middlewares"

	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(middlewares.NewAuthMiddleware),
	do.Lazy(NewProjectController),
	do.Lazy(NewUsersDBController),
)
