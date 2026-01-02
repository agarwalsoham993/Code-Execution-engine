package models

import "time"

type ExecutionRequest struct {
	Language    string            `json:"language"`
	Code        string            `json:"code"`
	QuestionID  string            `json:"question_id"`
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

type QuestionMeta struct {
	ID    string `json:"id"`
	Title string `json:"title"` // inferred from folder name or ID
}

type Question struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// Database Model
type Submission struct {
	ID           string    `json:"id"`
	Language     string    `json:"language"`
	Code         string    `json:"code"`
	Status       string    `json:"status"`
	StdOut       string    `json:"stdout"`
	StdErr       string    `json:"stderr"`
	ExecTimeMS   int       `json:"exec_time_ms"`
	QuestionID   string    `json:"question_id"`
	PassedCount  int       `json:"passed_count"`
	TotalCount   int       `json:"total_count"`
	MemoryUsed   string    `json:"memory_used"`
	CreatedAt    time.Time `json:"created_at"`
}