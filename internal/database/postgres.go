package database

import (
	"code-runner/pkg/models"
	"database/sql"
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

	// Init Schema with DEFAULT values to prevent NULL scan errors
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

	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) CreateSubmission(sub *models.Submission) error {
	// Explicitly insert empty strings and 0 for the output fields
	query := `INSERT INTO submissions (id, language, code, status, stdout, stderr, exec_time_ms, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := p.db.Exec(query, sub.ID, sub.Language, sub.Code, sub.Status, "", "", 0, time.Now())
	return err
}

func (p *PostgresDB) UpdateResult(id string, status string, stdout, stderr string, timeMs int) error {
	query := `UPDATE submissions SET status=$1, stdout=$2, stderr=$3, exec_time_ms=$4 WHERE id=$5`
	_, err := p.db.Exec(query, status, stdout, stderr, timeMs, id)
	return err
}

func (p *PostgresDB) GetSubmission(id string) (*models.Submission, error) {
	s := &models.Submission{}
	// Use COALESCE as an extra safety measure to handle existing NULL rows if any
	query := `SELECT id, language, code, status, 
              COALESCE(stdout, ''), COALESCE(stderr, ''), COALESCE(exec_time_ms, 0), 
              created_at FROM submissions WHERE id=$1`
	err := p.db.QueryRow(query, id).
		Scan(&s.ID, &s.Language, &s.Code, &s.Status, &s.StdOut, &s.StdErr, &s.ExecTimeMS, &s.CreatedAt)
	return s, err
}

func (p *PostgresDB) GetAllSubmissions() ([]models.Submission, error) {
	query := `SELECT id, language, code, status, 
              COALESCE(stdout, ''), COALESCE(stderr, ''), COALESCE(exec_time_ms, 0), 
              created_at FROM submissions ORDER BY created_at DESC LIMIT 50`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.Submission
	for rows.Next() {
		var s models.Submission
		if err := rows.Scan(&s.ID, &s.Language, &s.Code, &s.Status, &s.StdOut, &s.StdErr, &s.ExecTimeMS, &s.CreatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}