package database

import (
	"code-runner/pkg/models"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Init Schema
	query := `
	CREATE TABLE IF NOT EXISTS submissions (
		id TEXT PRIMARY KEY,
		language TEXT,
		code TEXT,
		status TEXT,
		stdout TEXT DEFAULT '',
		stderr TEXT DEFAULT '',
		exec_time_ms INT DEFAULT 0,
		created_at TIMESTAMP
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	// Migrations for new columns
	migrations := []string{
		"ALTER TABLE submissions ADD COLUMN IF NOT EXISTS question_id TEXT DEFAULT '';",
		"ALTER TABLE submissions ADD COLUMN IF NOT EXISTS passed_count INT DEFAULT 0;",
		"ALTER TABLE submissions ADD COLUMN IF NOT EXISTS total_count INT DEFAULT 0;",
		"ALTER TABLE submissions ADD COLUMN IF NOT EXISTS memory_used TEXT DEFAULT '';",
	}

	for _, mig := range migrations {
		if _, err := db.Exec(mig); err != nil {
			fmt.Printf("Migration warning: %v\n", err)
		}
	}

	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) CreateSubmission(sub *models.Submission) error {
	query := `INSERT INTO submissions (id, language, code, status, stdout, stderr, exec_time_ms, created_at, question_id, passed_count, total_count, memory_used) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := p.db.Exec(query, sub.ID, sub.Language, sub.Code, sub.Status, "", "", 0, time.Now(), sub.QuestionID, 0, 0, "")
	return err
}

func (p *PostgresDB) UpdateResult(id string, status string, stdout, stderr string, timeMs int, passed, total int, memory string) error {
	query := `UPDATE submissions SET status=$1, stdout=$2, stderr=$3, exec_time_ms=$4, passed_count=$5, total_count=$6, memory_used=$7 WHERE id=$8`
	_, err := p.db.Exec(query, status, stdout, stderr, timeMs, passed, total, memory, id)
	return err
}

func (p *PostgresDB) GetSubmission(id string) (*models.Submission, error) {
	s := &models.Submission{}
	query := `SELECT id, language, code, status, 
              COALESCE(stdout, ''), COALESCE(stderr, ''), COALESCE(exec_time_ms, 0), 
              created_at, COALESCE(question_id, ''), COALESCE(passed_count, 0), COALESCE(total_count, 0), COALESCE(memory_used, '')
              FROM submissions WHERE id=$1`
	err := p.db.QueryRow(query, id).
		Scan(&s.ID, &s.Language, &s.Code, &s.Status, &s.StdOut, &s.StdErr, &s.ExecTimeMS, &s.CreatedAt, &s.QuestionID, &s.PassedCount, &s.TotalCount, &s.MemoryUsed)
	return s, err
}

func (p *PostgresDB) GetAllSubmissions() ([]models.Submission, error) {
	query := `SELECT id, language, code, status, 
              COALESCE(stdout, ''), COALESCE(stderr, ''), COALESCE(exec_time_ms, 0), 
              created_at, COALESCE(question_id, ''), COALESCE(passed_count, 0), COALESCE(total_count, 0), COALESCE(memory_used, '')
              FROM submissions ORDER BY created_at DESC LIMIT 50`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.Submission
	for rows.Next() {
		var s models.Submission
		if err := rows.Scan(&s.ID, &s.Language, &s.Code, &s.Status, &s.StdOut, &s.StdErr, &s.ExecTimeMS, &s.CreatedAt, &s.QuestionID, &s.PassedCount, &s.TotalCount, &s.MemoryUsed); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}