package router

import (
	"github.com/samber/do/v2"
)

var Package = do.Package(
	do.Lazy(NewChiRouter),
	do.Lazy(NewHumaAPI),
	do.LazyNamed("huma.api.v1", NewV1Group),
	do.Lazy(NewBaaSRouter),
)
