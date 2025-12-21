package project

import "github.com/samber/do/v2"

var Package = do.Package(
	do.Lazy(NewRepository),
	do.Lazy(NewService),
	do.Lazy(NewController),
)
