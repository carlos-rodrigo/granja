package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"granja/internal/domain"
)

type EpicReviewIssue struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type EpicReviewResult struct {
	Result  string            `json:"result"`
	Summary string            `json:"summary"`
	Issues  []EpicReviewIssue `json:"issues,omitempty"`
}

func (r EpicReviewResult) IsPass() bool {
	return strings.EqualFold(strings.TrimSpace(r.Result), "PASS")
}

type EpicService struct {
	epicRepo EpicRepo
	taskRepo TaskRepo
	parser   *ParserService
}

func NewEpicService(epicRepo EpicRepo, taskRepo TaskRepo, parser *ParserService) *EpicService {
	return &EpicService{epicRepo: epicRepo, taskRepo: taskRepo, parser: parser}
}

func (s *EpicService) Create(ctx context.Context, projectID, prd, design string) (*domain.Epic, error) {
	if strings.TrimSpace(prd) == "" {
		return nil, errors.New("prd is required")
	}
	title := extractTitle(prd)
	branch := "epic/" + slugify(title)
	e := domain.Epic{
		ID:            "epic_" + uuid.NewString(),
		ProjectID:     projectID,
		Title:         title,
		Status:        domain.EpicPlanted,
		BranchName:    branch,
		PRDContent:    prd,
		DesignContent: design,
	}
	if err := s.epicRepo.Create(ctx, e); err != nil {
		return nil, err
	}
	if err := s.parseAndCreateTasks(ctx, e.ID, prd, design); err != nil {
		_ = s.epicRepo.UpdateStatus(ctx, e.ID, domain.EpicBlocked, err.Error())
	}
	stored, err := s.epicRepo.GetByID(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	return stored, nil
}

func (s *EpicService) List(ctx context.Context, projectID, status string) ([]domain.Epic, error) {
	return s.epicRepo.List(ctx, projectID, status)
}

func (s *EpicService) GetWithTasks(ctx context.Context, id string) (*domain.Epic, []domain.Task, error) {
	e, err := s.epicRepo.GetByID(ctx, id)
	if err != nil || e == nil {
		return e, nil, err
	}
	tasks, err := s.taskRepo.ListByEpic(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return e, tasks, nil
}

func (s *EpicService) IsReadyForReview(ctx context.Context, epicID string) (bool, error) {
	epic, err := s.epicRepo.GetByID(ctx, epicID)
	if err != nil {
		return false, err
	}
	if epic == nil {
		return false, errors.New("epic not found")
	}
	tasks, err := s.taskRepo.ListByEpic(ctx, epicID)
	if err != nil {
		return false, err
	}
	if len(tasks) == 0 {
		return false, nil
	}
	for _, t := range tasks {
		if t.Status != domain.TaskDone {
			return false, nil
		}
	}
	if epic.Status != domain.EpicReady {
		if err := s.epicRepo.UpdateStatus(ctx, epicID, domain.EpicReady, ""); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (s *EpicService) HandleReviewResult(ctx context.Context, epicID string, result EpicReviewResult) error {
	if strings.TrimSpace(result.Result) == "" {
		return errors.New("review result is required")
	}
	result.Result = strings.ToUpper(strings.TrimSpace(result.Result))
	if result.Result != "PASS" && result.Result != "FAIL" {
		return fmt.Errorf("invalid review result: %s", result.Result)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if err := s.epicRepo.SetReviewResult(ctx, epicID, string(payload)); err != nil {
		return err
	}

	if result.Result == "PASS" {
		return s.epicRepo.UpdateStatus(ctx, epicID, domain.EpicHarvested, "")
	}

	if len(result.Issues) == 0 {
		result.Issues = []EpicReviewIssue{{
			Title:       "Address review feedback",
			Description: strings.TrimSpace(result.Summary),
		}}
	}
	for _, issue := range result.Issues {
		title := strings.TrimSpace(issue.Title)
		if title == "" {
			title = "Address review feedback"
		}
		desc := strings.TrimSpace(issue.Description)
		if desc == "" {
			desc = strings.TrimSpace(result.Summary)
		}
		task := domain.Task{
			ID:          "task_" + uuid.NewString(),
			EpicID:      epicID,
			Title:       "Fix: " + title,
			Description: desc,
			Status:      domain.TaskTodo,
			Effort:      "small",
		}
		if err := s.taskRepo.Create(ctx, task); err != nil {
			return err
		}
	}
	return s.epicRepo.UpdateStatus(ctx, epicID, domain.EpicGrowing, "")
}

func (s *EpicService) parseAndCreateTasks(ctx context.Context, epicID, prd, design string) error {
	if s.parser == nil {
		return nil
	}
	parsedTasks, err := s.parser.ParseTasks(ctx, prd, design)
	if err != nil {
		return err
	}
	if len(parsedTasks) == 0 {
		return errors.New("parser returned no tasks")
	}

	titleToID := make(map[string]string, len(parsedTasks))
	for _, pt := range parsedTasks {
		title := strings.TrimSpace(pt.Title)
		if title == "" {
			return errors.New("task title is required")
		}
		normTitle := strings.ToLower(title)
		if _, exists := titleToID[normTitle]; exists {
			return fmt.Errorf("duplicate task title: %s", title)
		}
		relevantFiles, _ := json.Marshal(pt.RelevantFiles)
		taskID := "task_" + uuid.NewString()
		task := domain.Task{
			ID:            taskID,
			EpicID:        epicID,
			Title:         title,
			Description:   strings.TrimSpace(pt.Description),
			Status:        domain.TaskTodo,
			Effort:        strings.TrimSpace(pt.Effort),
			RelevantFiles: string(relevantFiles),
		}
		if err := s.taskRepo.Create(ctx, task); err != nil {
			return err
		}
		titleToID[normTitle] = taskID
	}

	for _, pt := range parsedTasks {
		taskID := titleToID[strings.ToLower(strings.TrimSpace(pt.Title))]
		for _, dep := range pt.Dependencies {
			depID, ok := titleToID[strings.ToLower(strings.TrimSpace(dep))]
			if !ok {
				return fmt.Errorf("unknown dependency %q for task %q", dep, pt.Title)
			}
			if err := s.taskRepo.AddDependency(ctx, taskID, depID); err != nil {
				return err
			}
		}
	}
	return s.epicRepo.UpdateStatus(ctx, epicID, domain.EpicGrowing, "")
}

func extractTitle(prd string) string {
	for _, line := range strings.Split(prd, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			clean := strings.TrimSpace(strings.TrimLeft(line, "#"))
			if clean != "" {
				return clean
			}
		}
	}
	return "Untitled Epic"
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(v string) string {
	lower := strings.ToLower(v)
	slug := strings.Trim(nonSlug.ReplaceAllString(lower, "-"), "-")
	if slug == "" {
		return "epic"
	}
	return slug
}
