package v1

import (
	"code-runner/internal/config"
	"code-runner/internal/sandbox"
	"code-runner/internal/spec"
	"code-runner/internal/util"
	"code-runner/pkg/cappedbuffer"
	"code-runner/pkg/models"
	"github.com/gofiber/fiber/v2"
)

func Setup(router fiber.Router, cfg *config.EnvProvider, sp *spec.BaseProvider, mgr *sandbox.Manager) {
	router.Get("/spec", func(c *fiber.Ctx) error {
		return c.JSON(sp.Spec())
	})

	router.Post("/exec", func(c *fiber.Ctx) error {
		req := new(models.ExecutionRequest)
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).JSON(models.ErrorModel{Error: "Invalid JSON"})
		}

		cStdOut := make(chan []byte)
		cStdErr := make(chan []byte)
		cStop := make(chan bool, 1)

		stdOut := cappedbuffer.New([]byte{}, 1024*1024) // 1MB limit
		stdErr := cappedbuffer.New([]byte{}, 1024*1024)

		go func() {
			for {
				select {
				case <-cStop:
					return
				case p := <-cStdOut:
					stdOut.Write(p)
				case p := <-cStdErr:
					stdErr.Write(p)
				}
			}
		}()

		execTime := util.MeasureTime(func() {
			mgr.RunInSandbox(req, cStdOut, cStdErr, cStop)
		})

		return c.JSON(models.ExecutionResponse{
			StdOut:     stdOut.String(),
			StdErr:     stdErr.String(),
			ExecTimeMS: int(execTime.Milliseconds()),
		})
	})
}
