package api

import (
	v1 "code-runner/internal/api/v1"
	"code-runner/internal/config"
	"code-runner/internal/sandbox"
	"code-runner/internal/spec"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type RestAPI struct {
	bindAddress string
	app         *fiber.App
}

func NewRestAPI(cfg *config.EnvProvider, sp *spec.BaseProvider, mgr *sandbox.Manager) (*RestAPI, error) {
	r := &RestAPI{
		bindAddress: cfg.Config().API.BindAddress,
	}

	r.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	r.app.Use(cors.New(cors.Config{			// ENABLE CORS 
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	v1.Setup(r.app.Group("/v1"), cfg, sp, mgr)

	return r, nil
}

func (r *RestAPI) ListenAndServeBlocking() error {
	return r.app.Listen(r.bindAddress)
}
