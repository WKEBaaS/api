// Package router
//
// BaaS API Router
package router

import (
	"fmt"
	"log/slog"
	"net/http"

	"baas-api/internal/classfunc"
	"baas-api/internal/config"
	"baas-api/internal/project"
	"baas-api/internal/usersdb"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/samber/do/v2"
)

type BaaSRouter struct {
	config              *config.Config       `do:""`
	router              *chi.Mux             `do:""`
	v1API               *huma.Group          `do:"huma.api.v1"`
	projectController   project.Controller   `do:""`
	usersdbController   usersdb.Controller   `do:""`
	classfuncController classfunc.Controller `do:""`
}

func NewBaaSRouter(i do.Injector) (*BaaSRouter, error) {
	return &BaaSRouter{
		config:              do.MustInvoke[*config.Config](i),
		router:              do.MustInvoke[*chi.Mux](i),
		v1API:               do.MustInvokeNamed[*huma.Group](i, "huma.api.v1"),
		projectController:   do.MustInvokeAs[project.Controller](i),
		usersdbController:   do.MustInvokeAs[usersdb.Controller](i),
		classfuncController: do.MustInvokeAs[classfunc.Controller](i),
	}, nil
}

func (r *BaaSRouter) RegisterControllers() {
	huma.AutoRegister(r.v1API, r.projectController)
	huma.AutoRegister(r.v1API, r.usersdbController)
	huma.AutoRegister(r.v1API, r.classfuncController)
}

func (r *BaaSRouter) Start() {
	slog.Info("Starting server", "host", r.config.App.Host, "port", r.config.App.Port)
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", r.config.App.Host, r.config.App.Port), r.router)
	if err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
