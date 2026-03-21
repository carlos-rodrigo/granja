package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"

	"granja/internal/domain"
	"granja/internal/repository"
)

type EpicService struct {
	epicRepo *repository.EpicRepository
	taskRepo *repository.TaskRepository
}

func NewEpicService(epicRepo *repository.EpicRepository, taskRepo *repository.TaskRepository) *EpicService {
	return &EpicService{epicRepo: epicRepo, taskRepo: taskRepo}
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
