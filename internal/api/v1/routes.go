package v1

import (
	"code-runner/internal/config"
	"code-runner/internal/database"
	"code-runner/internal/queue"
	"code-runner/internal/spec"
	"code-runner/pkg/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/xid"
)

func Setup(router fiber.Router, cfg *config.EnvProvider, sp *spec.BaseProvider, q *queue.RedisQueue, db *database.PostgresDB) {
	
	router.Get("/spec", func(c *fiber.Ctx) error {
		return c.JSON(sp.Spec())
	})

	// Get History
	router.Get("/submissions", func(c *fiber.Ctx) error {
		subs, err := db.GetAllSubmissions()
		if err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: err.Error()})
		}
		return c.JSON(subs)
	})

	// Submit Code
	router.Post("/exec", func(c *fiber.Ctx) error {
		req := new(models.ExecutionRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).JSON(models.ErrorModel{Error: "Invalid JSON"})
		}

		// 1. Generate ID
		id := xid.New().String()

		// 2. Save to DB (PENDING)
		sub := &models.Submission{
			ID:       id,
			Language: req.Language,
			Code:     req.Code,
			Status:   "PENDING",
		}
		if err := db.CreateSubmission(sub); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: "Database Error"})
		}

		// 3. Push to Redis
		if err := q.Enqueue(id); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: "Queue Error"})
		}

		return c.JSON(models.ExecutionResponse{
			SubmissionID: id,
			Status:       "PENDING",
		})
	})
}
