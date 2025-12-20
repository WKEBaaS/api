// Package router
//
// BaaS API Router
package router

import (
	"fmt"
	"log/slog"
	"net/http"

	"baas-api/config"
	"baas-api/controllers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
)

type BaaSRouter struct {
	config            *config.Config                         `do:""`
	router            *chi.Mux                               `do:""`
	v1API             *huma.Group                            `do:"huma.api.v1"`
	projectController controllers.ProjectControllerInterface `do:""`
}

func NewBaaSRouter(i do.Injector) (*BaaSRouter, error) {
	return &BaaSRouter{
		config:            do.MustInvoke[*config.Config](i),
		router:            do.MustInvoke[*chi.Mux](i),
		v1API:             do.MustInvokeNamed[*huma.Group](i, "huma.api.v1"),
		projectController: do.MustInvoke[controllers.ProjectControllerInterface](i),
	}, nil
}

func (r *BaaSRouter) RegisterControllers() {
	huma.AutoRegister(r.v1API, r.projectController)
}

func (r *BaaSRouter) Start() {
	slog.Info("Starting server", "host", r.config.App.Host, "port", r.config.App.Port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", r.config.App.Host, r.config.App.Port), r.router)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
