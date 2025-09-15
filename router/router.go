// Package router
//
// BaaS API Router
package router

import (
	"baas-api/config"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Options struct {
	Debug bool `help:"Enable debug mode" short:"d" default:"false"`
}

func NewAPICli(config *config.Config, controllers ...any) humacli.CLI {
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		slog.Info("Option.Debug", "debug", options.Debug)
		if options.Debug {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}

		humaConfig := huma.DefaultConfig("WKE BaaS API", "0.1.0")
		humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
			"BearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		}

		huma.NewError = NewCustomError

		// app := fiber.New()
		router := chi.NewMux()
		router.Use(middleware.Logger)
		router.Use(cors.Handler(cors.Options{
			AllowedOrigins:   config.App.TrustedOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			AllowCredentials: true,
		}))

		////////// Register APIs //////////
		api := humachi.New(router, humaConfig)
		v1Api := huma.NewGroup(api, "/v1")

		// Register controllers
		for _, controller := range controllers {
			huma.AutoRegister(v1Api, controller)
		}

		hooks.OnStart(func() {
			slog.Info("Starting server", "host", config.App.Host, "port", config.App.Port)
			err := http.ListenAndServe(fmt.Sprintf("%s:%s", config.App.Host, config.App.Port), router)
			if err != nil {
				slog.Error("Failed to start server", "error", err)
			}
		})
	})

	return cli
}
