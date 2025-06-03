package router

import (
	"baas-api/internal/configs"
	"baas-api/internal/controllers"
	"baas-api/internal/i3s"
	"baas-api/internal/repo"
	"fmt"
	"log"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8888"`
}

func NewAPI(appConfig *configs.Config, repo repo.ProjectRepository) humacli.CLI {
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		humaConfig := huma.DefaultConfig("WKE BaaS API", "0.1.0")
		humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
			"bearer": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		}
		humaConfig.Servers = []*huma.Server{
			{URL: fmt.Sprintf("http://localhost:%d", options.Port)},
		}

		app := fiber.New()
		app.Use(logger.New(logger.Config{
			Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
		}))

		i3s := i3s.NewI3S(appConfig)
		if err := i3s.Migrate(); err != nil {
			log.Fatalf("failed to migrate database: %v\n", err)
		}

		// if err := i3s.PostMetadata(); err != nil {
		// 	log.Fatalf("failed to post metadata: %v\n", err)
		// }

		////////// Register APIs //////////
		api := humafiber.New(app, humaConfig)
		v1Api := huma.NewGroup(api, "/api/v1")

		// Init controllers
		authController := controllers.NewAuthController()

		// Register controllers
		huma.AutoRegister(v1Api, authController)

		hooks.OnStart(func() {
			app.Listen(fmt.Sprintf(":%d", options.Port))
		})
		hooks.OnStop(func() {
			// service.DB.Close()
		})
	})

	return cli
}
