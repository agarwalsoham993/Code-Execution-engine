package api

import (
	v1 "code-runner/internal/api/v1"
	"code-runner/internal/config"
	"code-runner/internal/database"
	"code-runner/internal/question"
	"code-runner/internal/queue"
	"code-runner/internal/spec"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type RestAPI struct {
	bindAddress string
	app         *fiber.App
}

func NewRestAPI(cfg *config.EnvProvider, sp *spec.BaseProvider, q *queue.RedisQueue, db *database.PostgresDB, qp *question.Provider) (*RestAPI, error) {
	r := &RestAPI{
		bindAddress: cfg.Config().API.BindAddress,
	}

	r.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	r.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Pass all dependencies to routes
	v1.Setup(r.app.Group("/v1"), cfg, sp, q, db, qp)

	return r, nil
}

func (r *RestAPI) ListenAndServeBlocking() error {
	return r.app.Listen(r.bindAddress)
}