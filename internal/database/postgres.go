package database

import (
	"code-runner/pkg/models"
	"database/sql"
	"encoding/json"
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
		question_id TEXT DEFAULT '',
		status TEXT,
		stdout TEXT DEFAULT '',
		stderr TEXT DEFAULT '',
		exec_time_ms INT DEFAULT 0,
		passed_count INT DEFAULT 0,
		total_count INT DEFAULT 0,
		created_at TIMESTAMP
	);
	
	CREATE TABLE IF NOT EXISTS test_questions (
		id TEXT PRIMARY KEY,
		title TEXT,
		description TEXT,
		test_cases JSONB DEFAULT '[]'
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	alterQuery := `
		ALTER TABLE submissions ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT false;
		ALTER TABLE test_questions ADD COLUMN IF NOT EXISTS solution_code TEXT DEFAULT '';
		ALTER TABLE test_questions ADD COLUMN IF NOT EXISTS solution_lang TEXT DEFAULT '';
		ALTER TABLE test_questions ADD COLUMN IF NOT EXISTS generator_config TEXT DEFAULT '{}';
	`
	if _, err := db.Exec(alterQuery); err != nil {
		return nil, err
	}

	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) CreateSubmission(sub *models.Submission) error {
	query := `INSERT INTO submissions (id, language, code, question_id, status, stdout, stderr, exec_time_ms, passed_count, total_count, created_at, is_admin) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := p.db.Exec(query, sub.ID, sub.Language, sub.Code, sub.QuestionID, sub.Status, "", "", 0, 0, 0, time.Now(), sub.IsAdmin)
	return err
}

func (p *PostgresDB) UpdateResult(id string, status string, stdout, stderr string, timeMs, passed, total int) error {
	query := `UPDATE submissions SET status=$1, stdout=$2, stderr=$3, exec_time_ms=$4, passed_count=$5, total_count=$6 WHERE id=$7`
	_, err := p.db.Exec(query, status, stdout, stderr, timeMs, passed, total, id)
	return err
}

func (p *PostgresDB) GetSubmission(id string) (*models.Submission, error) {
	s := &models.Submission{}
	query := `SELECT id, language, code, COALESCE(question_id,''), status, 
              COALESCE(stdout, ''), COALESCE(stderr, ''), COALESCE(exec_time_ms, 0),
              COALESCE(passed_count, 0), COALESCE(total_count, 0), created_at, COALESCE(is_admin, false) 
              FROM submissions WHERE id=$1`
	err := p.db.QueryRow(query, id).
		Scan(&s.ID, &s.Language, &s.Code, &s.QuestionID, &s.Status, &s.StdOut, &s.StdErr, &s.ExecTimeMS, &s.PassedCount, &s.TotalCount, &s.CreatedAt, &s.IsAdmin)
	return s, err
}

func (p *PostgresDB) GetAllSubmissions() ([]models.Submission, error) {
	query := `SELECT id, language, code, COALESCE(question_id,''), status, 
              COALESCE(stdout, ''), COALESCE(stderr, ''), COALESCE(exec_time_ms, 0),
              COALESCE(passed_count, 0), COALESCE(total_count, 0), created_at, COALESCE(is_admin, false) 
              FROM submissions 
              WHERE is_admin = false OR is_admin IS NULL
              ORDER BY created_at DESC LIMIT 50`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []models.Submission
	for rows.Next() {
		var s models.Submission
		if err := rows.Scan(&s.ID, &s.Language, &s.Code, &s.QuestionID, &s.Status, &s.StdOut, &s.StdErr, &s.ExecTimeMS, &s.PassedCount, &s.TotalCount, &s.CreatedAt, &s.IsAdmin); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func (p *PostgresDB) CreateQuestion(q *models.Question) error {
	casesJSON, _ := json.Marshal(q.TestCases)
	query := `INSERT INTO test_questions (id, title, description, test_cases, solution_code, solution_lang, generator_config) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := p.db.Exec(query, q.ID, q.Title, q.Description, casesJSON, q.SolutionCode, q.SolutionLang, q.GeneratorConfig)
	return err
}

func (p *PostgresDB) UpdateQuestion(q *models.Question) error {
	casesJSON, _ := json.Marshal(q.TestCases)
	query := `UPDATE test_questions SET title=$1, description=$2, test_cases=$3, solution_code=$4, solution_lang=$5, generator_config=$6 WHERE id=$7`
	_, err := p.db.Exec(query, q.Title, q.Description, casesJSON, q.SolutionCode, q.SolutionLang, q.GeneratorConfig, q.ID)
	return err
}

func (p *PostgresDB) DeleteQuestion(id string) error {
	query := `DELETE FROM test_questions WHERE id=$1`
	_, err := p.db.Exec(query, id)
	return err
}

func (p *PostgresDB) GetQuestion(id string) (*models.Question, error) {
	q := &models.Question{}
	var casesJSON []byte
	query := `SELECT id, title, description, test_cases, COALESCE(solution_code, ''), COALESCE(solution_lang, ''), COALESCE(generator_config, '{}') FROM test_questions WHERE id=$1`
	err := p.db.QueryRow(query, id).Scan(&q.ID, &q.Title, &q.Description, &casesJSON, &q.SolutionCode, &q.SolutionLang, &q.GeneratorConfig)
	if err != nil {
		return nil, err
	}
	if len(casesJSON) > 0 {
		json.Unmarshal(casesJSON, &q.TestCases)
	}
	return q, nil
}

func (p *PostgresDB) GetAllQuestions() ([]models.Question, error) {
	query := `SELECT id, title, description, test_cases, COALESCE(solution_code, ''), COALESCE(solution_lang, ''), COALESCE(generator_config, '{}') FROM test_questions ORDER BY id ASC`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		var casesJSON []byte
		if err := rows.Scan(&q.ID, &q.Title, &q.Description, &casesJSON, &q.SolutionCode, &q.SolutionLang, &q.GeneratorConfig); err != nil {
			return nil, err
		}
		if len(casesJSON) > 0 {
			json.Unmarshal(casesJSON, &q.TestCases)
		}
		questions = append(questions, q)
	}
	return questions, nil
}