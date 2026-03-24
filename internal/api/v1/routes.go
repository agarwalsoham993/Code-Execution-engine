package v1

import (
	"code-runner/internal/config"
	"code-runner/internal/database"
	"code-runner/internal/queue"
	"code-runner/internal/spec"
	"code-runner/internal/util"
	"code-runner/pkg/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/xid"
)

func Setup(router fiber.Router, cfg *config.EnvProvider, sp *spec.BaseProvider, q *queue.RedisQueue, db *database.PostgresDB) {
	
	router.Get("/spec", func(c *fiber.Ctx) error {
		return c.JSON(sp.Spec())
	})

	router.Get("/questions", func(c *fiber.Ctx) error {
		questions, err := db.GetAllQuestions()
		if err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: err.Error()})
		}
		return c.JSON(questions)
	})

	router.Get("/questions/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		q, err := db.GetQuestion(id)
		if err != nil {
			return c.Status(404).JSON(models.ErrorModel{Error: "Question not found"})
		}
		return c.JSON(q)
	})

	router.Post("/admin/questions", func(c *fiber.Ctx) error {
		var q models.Question
		if err := c.BodyParser(&q); err != nil {
			return c.Status(400).JSON(models.ErrorModel{Error: "Invalid JSON"})
		}
		if q.ID == "" {
			q.ID = xid.New().String()
		}
		if err := db.CreateQuestion(&q); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: err.Error()})
		}
		return c.JSON(q)
	})

	router.Put("/admin/questions/:id", func(c *fiber.Ctx) error {
		var q models.Question
		if err := c.BodyParser(&q); err != nil {
			return c.Status(400).JSON(models.ErrorModel{Error: "Invalid JSON"})
		}
		q.ID = c.Params("id")
		if err := db.UpdateQuestion(&q); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: err.Error()})
		}
		return c.JSON(q)
	})

	router.Delete("/admin/questions/:id", func(c *fiber.Ctx) error {
		if err := db.DeleteQuestion(c.Params("id")); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: err.Error()})
		}
		return c.JSON(fiber.Map{"success": true})
	})

	router.Post("/admin/generate-inputs", func(c *fiber.Ctx) error {
		var req models.GenerateRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(models.ErrorModel{Error: "Invalid JSON"})
		}
		id := xid.New().String()
		sub := &models.Submission{
			ID:         id,
			Language:   req.Language,
			Code:       req.Code,
			Status:     "PENDING",
			IsAdmin:    true,
		}
		if err := db.CreateSubmission(sub); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: "Database Error"})
		}
		payload := models.JobPayload{
			SubmissionID:     id,
			Language:         req.Language,
			Code:             req.Code,
			IsInputGenerator: true,
		}
		if err := q.Enqueue(payload); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: "Queue Error"})
		}
		return c.JSON(fiber.Map{"submission_id": id})
	})

	router.Post("/admin/generate", func(c *fiber.Ctx) error {
		var req models.GenerateRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(models.ErrorModel{Error: "Invalid JSON"})
		}
		id := xid.New().String()
		sub := &models.Submission{
			ID:         id,
			Language:   req.Language,
			Code:       req.Code,
			QuestionID: req.QuestionID,
			Status:     "PENDING",
			IsAdmin:    true,
		}
		if err := db.CreateSubmission(sub); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: "Database Error"})
		}
		payload := models.JobPayload{
			SubmissionID: id,
			Language:     req.Language,
			Code:         req.Code,
			QuestionID:   req.QuestionID,
			AdminInputs:  req.AdminInputs,
		}
		if err := q.Enqueue(payload); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: "Queue Error"})
		}
		return c.JSON(fiber.Map{"submission_id": id})
	})

	router.Get("/admin/logs", func(c *fiber.Ctx) error {
		return c.JSON(util.GlobalRingLogger.GetLogs())
	})

	router.Get("/submissions", func(c *fiber.Ctx) error {
		subs, err := db.GetAllSubmissions()
		if err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: err.Error()})
		}
		return c.JSON(subs)
	})

	router.Get("/submissions/:id", func(c *fiber.Ctx) error {
		sub, err := db.GetSubmission(c.Params("id"))
		if err != nil {
			return c.Status(404).JSON(models.ErrorModel{Error: "Not found"})
		}
		return c.JSON(sub)
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

		payload := models.JobPayload{
			SubmissionID: id,
			Language:     req.Language,
			Code:         req.Code,
			QuestionID:   req.QuestionID,
		}

		if err := q.Enqueue(payload); err != nil {
			return c.Status(500).JSON(models.ErrorModel{Error: "Queue Error"})
		}

		return c.JSON(models.ExecutionResponse{
			SubmissionID: id,
			Status:       "PENDING",
		})
	})
}