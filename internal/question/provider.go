package question

import (
	"code-runner/pkg/models"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Provider struct {
	baseDir string
}

type TestCase struct {
	ID       string
	Input    string
	Expected string
}

func NewProvider(baseDir string) *Provider {
	return &Provider{baseDir: baseDir}
}

func (p *Provider) ListQuestions() ([]models.QuestionMeta, error) {
	entries, err := os.ReadDir(p.baseDir)
	if err != nil {
		return nil, err
	}

	var questions []models.QuestionMeta
	for _, e := range entries {
		if e.IsDir() {
			questions = append(questions, models.QuestionMeta{
				ID:    e.Name(),
				Title: fmt.Sprintf("Problem %s", e.Name()),
			})
		}
	}
	return questions, nil
}

func (p *Provider) GetQuestion(id string) (models.Question, error) {
	// Assumes question text is in "Questions/{id}/question.txt" or just the folder structure
	// We will look for a .txt file in the folder that isn't in input/output
	path := filepath.Join(p.baseDir, id)
	entries, err := os.ReadDir(path)
	if err != nil {
		return models.Question{}, err
	}

	desc := "No description available."
	
	// Try to find a description file
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			b, err := os.ReadFile(filepath.Join(path, e.Name()))
			if err == nil {
				desc = string(b)
				break
			}
		}
	}

	return models.Question{
		ID:          id,
		Description: desc,
	}, nil
}

func (p *Provider) GetTestCases(questionID string) ([]TestCase, error) {
	inDir := filepath.Join(p.baseDir, questionID, "input")
	outDir := filepath.Join(p.baseDir, questionID, "output")

	inFiles, err := os.ReadDir(inDir)
	if err != nil {
		return nil, err
	}

	var cases []TestCase

	for _, f := range inFiles {
		if f.IsDir() {
			continue
		}
		
		// Map input_X.txt to output_X.txt
		// Assuming format input_{id}.txt and output_{id}.txt
		idStr := strings.TrimPrefix(strings.TrimSuffix(f.Name(), ".txt"), "input_")
		
		inPath := filepath.Join(inDir, f.Name())
		outPath := filepath.Join(outDir, "output_"+idStr+".txt")

		inputBytes, err := os.ReadFile(inPath)
		if err != nil {
			continue
		}
		outputBytes, err := os.ReadFile(outPath)
		if err != nil {
			// fallback or skip? skip for now
			continue
		}

		cases = append(cases, TestCase{
			ID:       idStr,
			Input:    string(inputBytes),
			Expected: string(outputBytes),
		})
	}
	
	// Sort by ID for consistent execution order
	sort.Slice(cases, func(i, j int) bool {
		return cases[i].ID < cases[j].ID
	})

	return cases, nil
}