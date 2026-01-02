package models

import "time"

type TestCase struct {
	ID             string `json:"id"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
}

type TestResult struct {
	TestCaseID string `json:"test_case_id"`
	Status     string `json:"status"` // PASSED, FAILED, ERROR
	Actual     string `json:"actual"`
	Expected   string `json:"expected"`
}

type ExecutionRequest struct {
	Language    string            `json:"language"`
	Code        string            `json:"code"`
	QuestionID  string            `json:"question_id"`
	Arguments   []string          `json:"arguments"`
	Environment map[string]string `json:"environment"`
}

type ExecutionResponse struct {
	SubmissionID string       `json:"submission_id"`
	Status       string       `json:"status"`
	Results      []TestResult `json:"results,omitempty"`
}

type ErrorModel struct {
	Error string `json:"error"`
}

type Question struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type JobPayload struct {									// transfered to REDIS
	SubmissionID string `json:"submission_id"`
	Language     string `json:"language"`
	Code         string `json:"code"`
	QuestionID   string `json:"question_id"`
}

type Submission struct {									// transfered to Database
	ID          string       `json:"id"`
	Language    string       `json:"language"`
	Code        string       `json:"code"`
	QuestionID  string       `json:"question_id"`
	Status      string       `json:"status"`
	StdOut      string       `json:"stdout"`
	StdErr      string       `json:"stderr"`
	ExecTimeMS  int          `json:"exec_time_ms"`
	Results     []TestResult `json:"results,omitempty"`
	PassedCount int          `json:"passed_count"`
	TotalCount  int          `json:"total_count"`
	CreatedAt   time.Time    `json:"created_at"`
}