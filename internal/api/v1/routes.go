package v1

import (
	"code-runner/internal/config"
	"code-runner/internal/database"
	"code-runner/internal/queue"
	"code-runner/internal/spec"
	"code-runner/pkg/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/xid"
	"os"
	"path/filepath"
	"strings"
)

func Setup(router fiber.Router, cfg *config.EnvProvider, sp *spec.BaseProvider, q *queue.RedisQueue, db *database.PostgresDB) {
	
	router.Get("/spec", func(c *fiber.Ctx) error {
		return c.JSON(sp.Spec())
	})

	router.Get("/questions", func(c *fiber.Ctx) error {
		entries, err := os.ReadDir("Questions")
		if err != nil {
			return c.JSON([]models.Question{})
		}
		var questions []models.Question
		for _, e := range entries {
			if e.IsDir() {
				title := "Unknown"
				data, _ := os.ReadFile(filepath.Join("Questions", e.Name(), "question.txt"))
				lines := strings.Split(string(data), "\n")
				if len(lines) > 0 {
					title = lines[0]
				}
				questions = append(questions, models.Question{ID: e.Name(), Title: title})
			}
		}
		return c.JSON(questions)
	})

	router.Get("/questions/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		path := filepath.Join("Questions", id, "question.txt")
		data, err := os.ReadFile(path)
		if err != nil {
			return c.Status(404).JSON(models.ErrorModel{Error: "Question not found"})
		}
		
		content := string(data)
		parts := strings.SplitN(content, "\n", 2)
		title := parts[0]
		desc := ""
		if len(parts) > 1 { desc = parts[1] }

		return c.JSON(models.Question{
			ID:          id,
			Title:       title,
			Description: desc,
		})
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