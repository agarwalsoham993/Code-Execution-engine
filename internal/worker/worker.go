package worker

import (
	"code-runner/internal/database"
	"code-runner/internal/queue"
	"code-runner/internal/sandbox"
	"code-runner/internal/util"
	"code-runner/pkg/cappedbuffer"
	"code-runner/pkg/models"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"github.com/redis/go-redis/v9"
	"github.com/zekrotja/rogu/log"
)

type Worker struct {
	id      int
	queue   *queue.RedisQueue
	db      *database.PostgresDB
	manager *sandbox.Manager
	quit    chan bool
}

func NewWorker(id int, q *queue.RedisQueue, db *database.PostgresDB, mgr *sandbox.Manager) *Worker {
	return &Worker{
		id:      id,
		queue:   q,
		db:      db,
		manager: mgr,
		quit:    make(chan bool),
	}
}

// Stop signals the worker to finish the current loop and exit
func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func (w *Worker) Start() {
	log.Info().Field("worker_id", w.id).Msg("Worker started")
	
	for {
		// Check for stop signal before polling
		select {
		case <-w.quit:
			log.Info().Field("worker_id", w.id).Msg("Worker stopping (signal received)")
			return
		default:
		}
 
		payload, err := w.queue.Dequeue(2 * time.Second)
		if err != nil {
			if err == redis.Nil {
				continue
			}
			log.Error().Err(err).Msg("Redis error")
			time.Sleep(1 * time.Second) // Backoff on error
			continue
		}

		log.Info().Field("worker_id", w.id).Field("job_id", payload.SubmissionID).Msg("Processing job")

		files, totalTestCases, err := w.generateFiles(payload)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate files")
			w.db.UpdateResult(payload.SubmissionID, "ERROR", "", "Failed to generate runner: "+err.Error(), 0, 0, 0)
			continue
		}

		cStdOut := make(chan []byte)
		cStdErr := make(chan []byte)
		cStop := make(chan bool, 1)

		stdOutBuf := cappedbuffer.New([]byte{}, 10*1024) 
		stdErrBuf := cappedbuffer.New([]byte{}, 10*1024)

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

		execTime := util.MeasureTime(func() {
			err = w.manager.RunInSandbox(payload.SubmissionID, payload.Language, files, nil, cStdOut, cStdErr, cStop)
		})

		status := "SUCCESS"
		output := stdOutBuf.String()
		stderr := stdErrBuf.String()
		passedCount := 0

		if err != nil {
			if err.Error() == "execution timed out" {
				status = "TIMEOUT"
			} else {
				status = "ERROR"
			}
		}

		if status == "SUCCESS" {
			if payload.IsInputGenerator {
				var rawInputs []string
				if jsonErr := json.Unmarshal([]byte(output), &rawInputs); jsonErr == nil {
					status = "SUCCESS"
					passedCount = len(rawInputs)
				} else {
					status = "ERROR"
					stderr = "Failed to parse generated inputs as JSON array of strings.\nOutput was:\n" + output
				}
			} else {
				var genResult struct {
					Generated []models.TestCase `json:"generated"`
				}
				if jsonErr := json.Unmarshal([]byte(output), &genResult); jsonErr == nil && len(genResult.Generated) > 0 {
					status = "SUCCESS"
					passedCount = totalTestCases
				} else {
					var results []models.TestResult
					if jsonErr := json.Unmarshal([]byte(output), &results); jsonErr == nil {
						if len(results) == 0 {
							status = "SUCCESS"
							passedCount = totalTestCases
						} else {
							failure := results[0]
							status = "FAILURE"
							if idVal, err := strconv.Atoi(failure.TestCaseID); err == nil {
								passedCount = idVal - 1
							}
							stderr = fmt.Sprintf("Failed Case %s:\n\nExpected Output:\n%s\n\nActual Output:\n%s", 
								failure.TestCaseID, failure.Expected, failure.Actual)
						}
					} else {
						status = "ERROR"
						stderr += "\nJudge Error: Output format invalid (Output limit might be exceeded)."
					}
				}
			}
		}

		w.db.UpdateResult(payload.SubmissionID, status, output, stderr, int(execTime.Milliseconds()), passedCount, totalTestCases)
		log.Info().Field("job_id", payload.SubmissionID).Field("status", status).Msg("Job finished")
	}
}

func (w *Worker) generateFiles(payload *models.JobPayload) (map[string]string, int, error) {
	files := make(map[string]string)
	
	if payload.IsInputGenerator {
		switch payload.Language {
		case "python3", "python":
			files["driver.py"] = payload.Code
		case "node", "javascript":
			files["driver.js"] = payload.Code
		default:
			files["main.code"] = payload.Code
			return nil, 0, fmt.Errorf("language %s not fully supported for input generation", payload.Language)
		}
		return files, 0, nil
	}

	var testsJSON []byte
	var tests []models.TestCase

	if len(payload.AdminInputs) > 0 {
		tests = make([]models.TestCase, len(payload.AdminInputs))
		for i, inp := range payload.AdminInputs {
			tests[i] = models.TestCase{
				ID:             fmt.Sprintf("%d", i+1),
				Input:          inp,
				ExpectedOutput: "__GENERATE__",
			}
		}
		testsJSON, _ = json.Marshal(tests)
	} else if payload.QuestionID != "" {
		q, err := w.db.GetQuestion(payload.QuestionID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load tests for question %s: %v", payload.QuestionID, err)
		}
		testsJSON, err = json.Marshal(q.TestCases)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to encode tests for question %s: %v", payload.QuestionID, err)
		}
	} else {
		defaultTests := []models.TestCase{{ID: "1", Input: "test", ExpectedOutput: "test"}}
		testsJSON, _ = json.Marshal(defaultTests)
	}

	if err := json.Unmarshal(testsJSON, &tests); err != nil {
		return nil, 0, err
	}
	
	files["tests.json"] = string(testsJSON)

	switch payload.Language {
	case "python3", "python":
		files["solution.py"] = payload.Code
		files["driver.py"] = pythonDriverTemplate
	case "node", "javascript":
		files["solution.js"] = payload.Code
		files["driver.js"] = nodeDriverTemplate
	default:
		files["main.code"] = payload.Code
		return nil, 0, fmt.Errorf("language %s not fully supported", payload.Language)
	}

	return files, len(tests), nil
}

const pythonDriverTemplate = `
import json
import sys

def solve(i): return str(i) # Default mock

try:
    import solution
    if hasattr(solution, 'solve'):
        solve = solution.solve
except ImportError:
    pass
except Exception:
    pass

def run():
    try:
        with open("tests.json") as f: tests = json.load(f)
        results = []
        is_gen = len(tests) > 0 and tests[0].get("expected_output") == "__GENERATE__"
        failed_result = None
        for t in tests:
            res = {"id": t["id"], "status": "FAILED", "expected": t["expected_output"], "actual": ""}
            try:
                inp = t["input"]
                try: 
                    if "." not in inp: inp = int(inp)
                except: pass
                val = solve(inp)
                actual = str(val)
                if is_gen:
                    results.append({"input": str(t["input"]), "expected_output": actual})
                    continue
                res["actual"] = actual
                if actual.strip() == t["expected_output"].strip(): res["status"] = "PASSED"
            except Exception as e:
                res["status"] = "ERROR"
                res["actual"] = str(e)
            if not is_gen and res["status"] != "PASSED":
                failed_result = res
                break
        
        if is_gen: print(json.dumps({"generated": results}))
        elif failed_result: print(json.dumps([failed_result]))
        else: print(json.dumps([]))
    except Exception as e:
        print(json.dumps([{"id": "0", "status": "ERROR", "actual": str(e), "expected": ""}]))
if __name__ == "__main__": run()
`

const nodeDriverTemplate = `
const fs = require('fs');
let solve = (i) => i; 
try {
    const userMod = require('./solution');
    if (typeof userMod === 'function') solve = userMod;
} catch (e) {}
try {
    const tests = JSON.parse(fs.readFileSync('tests.json', 'utf8'));
    let failedResult = null;
    let isGen = tests.length > 0 && tests[0].expected_output === "__GENERATE__";
    let results = [];
    for (const t of tests) {
        const res = { id: t.id, status: "FAILED", expected: t.expected_output, actual: "" };
        try {
            let inp = t.input;
            if(!isNaN(inp)) inp = Number(inp);
            const val = solve(inp);
            res.actual = String(val);
            if (isGen) {
                results.push({input: String(t.input), expected_output: res.actual});
                continue;
            }
            if (res.actual.trim() === t.expected_output.trim()) res.status = "PASSED";
        } catch (e) { res.status = "ERROR"; res.actual = e.message; }
        if (!isGen && res.status !== "PASSED") { failedResult = res; break; }
    }
    if (isGen) console.log(JSON.stringify({generated: results}));
    else if (failedResult) console.log(JSON.stringify([failedResult]));
    else console.log(JSON.stringify([])); 
} catch (e) { console.log(JSON.stringify([{id: "0", status: "ERROR", actual: e.message, expected: ""}])); }
`