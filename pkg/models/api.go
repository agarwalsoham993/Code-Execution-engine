package models

import "time"

type ExecutionRequest struct {
	Language    string            `json:"language"`
	Code        string            `json:"code"`
	Arguments   []string          `json:"arguments"`
	Environment map[string]string `json:"environment"`
}

type ExecutionResponse struct {
	SubmissionID string `json:"submission_id"`
	Status       string `json:"status"`
}

type ErrorModel struct {
	Error string `json:"error"`
}

// Database Model
type Submission struct {
	ID          string    `json:"id"`
	Language    string    `json:"language"`
	Code        string    `json:"code"`
	Status      string    `json:"status"`
	StdOut      string    `json:"stdout"`
	StdErr      string    `json:"stderr"`
	ExecTimeMS  int       `json:"exec_time_ms"`
	CreatedAt   time.Time `json:"created_at"`
}