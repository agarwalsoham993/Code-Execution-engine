package v1

import (
	"code-runner/internal/config"
	"code-runner/internal/database"
	"code-runner/internal/question"
	"code-runner/internal/queue"
	"code-runner/internal/spec"
	"code-runner/pkg/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/xid"
)

// Updated Setup signature to accept QuestionProvider
func Setup(router fiber.Router, cfg *config.EnvProvider, sp *spec.BaseProvider, q *queue.RedisQueue, db *database.PostgresDB, qp *question.Provider) {
	
	router.Get("/spec", func(c *fiber.Ctx) error {
		return c.JSON(sp.Spec())
	})

	// List Questions
	router.Get("/questions", func(c *fiber.Ctx) error {
		qs, err := qp.ListQuestions()
		if err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: err.Error()})
		}
		return c.JSON(qs)
	})

	// Get Question Details
	router.Get("/questions/:id", func(c *fiber.Ctx) error {
		q, err := qp.GetQuestion(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(models.ErrorModel{Error: "Question not found"})
		}
		return c.JSON(q)
	})

	router.Get("/submissions", func(c *fiber.Ctx) error {
		subs, err := db.GetAllSubmissions()
		if err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: err.Error()})
		}
		return c.JSON(subs)
	})

	router.Post("/exec", func(c *fiber.Ctx) error {
		req := new(models.ExecutionRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).JSON(models.ErrorModel{Error: "Invalid JSON"})
		}

		id := xid.New().String()

		sub := &models.Submission{
			ID:         id,
			Language:   req.Language,
			Code:       req.Code,
			QuestionID: req.QuestionID,
			Status:     "PENDING",
		}
		if err := db.CreateSubmission(sub); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: "Database Error"})
		}

		if err := q.Enqueue(id); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: "Queue Error"})
		}

		return c.JSON(models.ExecutionResponse{
			SubmissionID: id,
			Status:       "PENDING",
		})
	})
}