package router

import (
	"baas-api/internal/config"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/samber/do/v2"
)

func NewChiRouter(i do.Injector) (*chi.Mux, error) {
	cfg := do.MustInvoke[*config.Config](i)
	router := chi.NewMux()
	router.Use(middleware.Logger)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.App.TrustedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	return router, nil
}

func NewHumaAPI(i do.Injector) (huma.API, error) {
	router := do.MustInvoke[*chi.Mux](i)

	humaConfig := huma.DefaultConfig("WKE BaaS API", "0.2.0")
	humaConfig.OpenAPIPath = "/api/docs/openapi"

	api := humachi.New(router, humaConfig)
	return api, nil
}

func NewV1Group(i do.Injector) (*huma.Group, error) {
	api := do.MustInvoke[huma.API](i)
	group := huma.NewGroup(api, "/v1")
	return group, nil
}

var Package = do.Package(
	do.Lazy(NewChiRouter),
	do.Lazy(NewHumaAPI),
	do.LazyNamed("huma.api.v1", NewV1Group),
	do.Lazy(NewBaaSRouter),
)
