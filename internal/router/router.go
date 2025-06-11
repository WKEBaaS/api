package router

import (
	"baas-api/internal/configs"
	"fmt"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Options struct {
	Port  int  `help:"Port to listen on" short:"p" default:"8888"`
	Debug bool `help:"Enable debug mode" short:"d" default:"false"`
}

func NewApiCli(appConfig *configs.Config, controllers ...any) humacli.CLI {
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

		app := fiber.New()
		app.Use(logger.New(logger.Config{
			Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
		}))

		////////// Register APIs //////////
		api := humafiber.New(app, humaConfig)
		v1Api := huma.NewGroup(api, "/v1")

		// Register controllers
		for _, controller := range controllers {
			huma.AutoRegister(v1Api, controller)
		}

		hooks.OnStart(func() {
			if err := app.Listen(fmt.Sprintf(":%d", options.Port)); err != nil {
				slog.Error("Failed to start server", "error", err)
			}
		})
	})

	return cli
}
