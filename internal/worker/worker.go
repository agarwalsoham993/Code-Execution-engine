package worker

import (
	"code-runner/internal/database"
	"code-runner/internal/question"
	"code-runner/internal/queue"
	"code-runner/internal/sandbox"
	"code-runner/internal/util"
	"code-runner/pkg/cappedbuffer"
	"code-runner/pkg/models"
	"fmt"
	"github.com/zekrotja/rogu/log"
	"strings"
	"time"
)

type Worker struct {
	id       int
	queue    *queue.RedisQueue
	db       *database.PostgresDB
	manager  *sandbox.Manager
	qProvider *question.Provider
}

func NewWorker(id int, q *queue.RedisQueue, db *database.PostgresDB, mgr *sandbox.Manager, qp *question.Provider) *Worker {
	return &Worker{id: id, queue: q, db: db, manager: mgr, qProvider: qp}
}

func (w *Worker) Start() {
	log.Info().Field("worker_id", w.id).Msg("Worker started, waiting for jobs...")
	for {
		submissionID, err := w.queue.Dequeue()
		if err != nil {
			log.Error().Err(err).Msg("Redis error")
			continue
		}

		log.Info().Field("worker_id", w.id).Field("job_id", submissionID).Msg("Processing job")

		sub, err := w.db.GetSubmission(submissionID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get submission from DB")
			continue
		}

		w.db.UpdateResult(submissionID, "PROCESSING", "", "", 0, 0, 0, "")

		req := &models.ExecutionRequest{
			Language:   sub.Language,
			Code:       sub.Code,
			QuestionID: sub.QuestionID,
		}

		// Determine inputs: If question ID exists, fetch test cases. Otherwise single run (playground).
		var testCases []question.TestCase
		if req.QuestionID != "" {
			testCases, err = w.qProvider.GetTestCases(req.QuestionID)
			if err != nil {
				w.db.UpdateResult(submissionID, "ERROR", "", fmt.Sprintf("Failed to load question: %v", err), 0, 0, 0, "")
				continue
			}
		} else {
			// Playground mode: 1 empty test case
			testCases = []question.TestCase{{ID: "1", Input: "", Expected: ""}}
		}

		totalTests := len(testCases)
		passedTests := 0
		finalStatus := "SUCCESS"
		var finalStdOut, finalStdErr strings.Builder
		totalTime := 0

		for _, tc := range testCases {
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

			var execTime time.Duration
			execTime = util.MeasureTime(func() {
				// Pass Test Case Input here
				err = w.manager.RunInSandbox(submissionID, req, []byte(tc.Input), cStdOut, cStdErr, cStop)
			})
			totalTime += int(execTime.Milliseconds())

			// Check runtime/compile error
			if err != nil {
				finalStatus = "ERROR"
				if err.Error() == "execution timed out" {
					finalStatus = "TIMEOUT"
				}
				
				// Reset buffers to ensure we only show the relevant error
				finalStdErr.Reset()
				finalStdErr.WriteString(fmt.Sprintf("Error on Test Case %s: %v\n%s\n", tc.ID, err, stdErrBuf.String()))
				
				finalStdOut.Reset()
				finalStdOut.WriteString("Runtime/Compilation Error")
				break // Stop on first fatal error
			}

			runOut := strings.TrimSpace(stdOutBuf.String())
			
			// Verify logic (only if not playground)
			if req.QuestionID != "" {
				expected := strings.TrimSpace(tc.Expected)
				if runOut == expected {
					passedTests++
				} else {
					finalStatus = "FAILURE"
					
					// User Requirement: Only show the failed test case number in main message
					finalStdOut.Reset()
					finalStdOut.WriteString(fmt.Sprintf("Failed Test case %s", tc.ID))

					// User Requirement: Details for the "view wrong test case" button (stored in StdErr)
					finalStdErr.Reset()
					finalStdErr.WriteString(fmt.Sprintf("Input:\n%s\n\nExpected:\n%s\n\nGot:\n%s", tc.Input, expected, runOut))
					
					break // User Requirement: Stop finding failed test cases after the first one
				}
			} else {
				// Playground: just record output
				finalStdOut.WriteString(runOut)
				finalStdErr.WriteString(stdErrBuf.String())
			}
		}

		// Final Status Logic for Questions
		if req.QuestionID != "" {
			if finalStatus == "SUCCESS" {
				finalStdOut.Reset()
				finalStdOut.WriteString("Success")
			} 
			// If FAILURE or ERROR, finalStdOut is already correctly set inside the loop
		}

		w.db.UpdateResult(
			submissionID,
			finalStatus,
			finalStdOut.String(),
			finalStdErr.String(),
			totalTime,
			passedTests,
			totalTests,
			"100M", // Placeholder
		)

		log.Info().Field("job_id", submissionID).Field("status", finalStatus).Msg("Job finished")
	}
}