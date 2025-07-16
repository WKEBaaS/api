package router

import (
	"baas-api/internal/configs"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Options struct {
	Port  int  `help:"Port to listen on" short:"p" default:"8888"`
	Debug bool `help:"Enable debug mode" short:"d" default:"false"`
}

func NewAPICli(appConfig *configs.Config, controllers ...any) humacli.CLI {
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		slog.Info("Option.Debug", "debug", options.Debug)
		if options.Debug {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}

		humaConfig := huma.DefaultConfig("WKE BaaS API", "0.1.0")
		humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
			"baasAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		}

		huma.NewError = NewCustomError

		// app := fiber.New()
		router := chi.NewMux()
		router.Use(middleware.Logger)

		////////// Register APIs //////////
		api := humachi.New(router, humaConfig)
		v1Api := huma.NewGroup(api, "/v1")

		// Register controllers
		for _, controller := range controllers {
			huma.AutoRegister(v1Api, controller)
		}

		hooks.OnStart(func() {
			slog.Info("Starting server", "port", options.Port)
			err := http.ListenAndServe(fmt.Sprintf(":%d", options.Port), router)
			if err != nil {
				slog.Error("Failed to start server", "error", err)
			}
		})
	})

	return cli
}
