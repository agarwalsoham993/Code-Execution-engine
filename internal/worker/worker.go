package worker

import (
	"code-runner/internal/database"
	"code-runner/internal/queue"
	"code-runner/internal/sandbox"
	"code-runner/internal/util"
	"code-runner/pkg/cappedbuffer"
	"code-runner/pkg/models"
	"github.com/zekrotja/rogu/log"
)

type Worker struct {
	id      int
	queue   *queue.RedisQueue
	db      *database.PostgresDB
	manager *sandbox.Manager
}

func NewWorker(id int, q *queue.RedisQueue, db *database.PostgresDB, mgr *sandbox.Manager) *Worker {
	return &Worker{id: id, queue: q, db: db, manager: mgr}
}

func (w *Worker) Start() {
	log.Info().Field("worker_id", w.id).Msg("Worker started, waiting for jobs...")
	for {
		// 1. Dequeue (Blocking)
		submissionID, err := w.queue.Dequeue()
		if err != nil {
			log.Error().Err(err).Msg("Redis error")
			continue
		}

		log.Info().Field("worker_id", w.id).Field("job_id", submissionID).Msg("Processing job")

		// 2. Fetch Job Details
		sub, err := w.db.GetSubmission(submissionID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get submission from DB")
			continue
		}

		// 3. Mark as Processing
		w.db.UpdateResult(submissionID, "PROCESSING", "", "", 0)

		// 4. Prepare execution
		req := &models.ExecutionRequest{
			Language: sub.Language,
			Code:     sub.Code,
		}

		cStdOut := make(chan []byte)
		cStdErr := make(chan []byte)
		cStop := make(chan bool, 1)

		stdOutBuf := cappedbuffer.New([]byte{}, 1024*1024)
		stdErrBuf := cappedbuffer.New([]byte{}, 1024*1024)

		go func() {
			for {
				select {
				case <-cStop:
					return
				case p := <-cStdOut:
					stdOutBuf.Write(p)
				case p := <-cStdErr:
					stdErrBuf.Write(p)
				}
			}
		}()

		// 5. Execute
		execTime := util.MeasureTime(func() {
			err = w.manager.RunInSandbox(submissionID, req, cStdOut, cStdErr, cStop)
		})

		status := "SUCCESS"
		if err != nil {
			if err.Error() == "execution timed out" {
				status = "TIMEOUT"
			} else {
				status = "ERROR"
			}
		}

		// 6. Update Result
		w.db.UpdateResult(
			submissionID,
			status,
			stdOutBuf.String(),
			stdErrBuf.String(),
			int(execTime.Milliseconds()),
		)

		log.Info().Field("job_id", submissionID).Field("status", status).Msg("Job finished")
	}
}
