package service

import (
	"context"
	"errors"
	"strings"

	"granja/internal/domain"
	"granja/internal/repository"
)

type TaskService struct {
	taskRepo *repository.TaskRepository
	epicRepo *repository.EpicRepository
}

func NewTaskService(taskRepo *repository.TaskRepository, epicRepo *repository.EpicRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo, epicRepo: epicRepo}
}

func (s *TaskService) ListByEpic(ctx context.Context, epicID string) ([]domain.Task, error) {
	return s.taskRepo.ListByEpic(ctx, epicID)
}

func (s *TaskService) UpdateStatus(ctx context.Context, taskID string, status domain.TaskStatus, logs string) error {
	switch status {
	case domain.TaskTodo, domain.TaskInProgress, domain.TaskDone, domain.TaskBlocked:
	default:
		return errors.New("invalid task status")
	}
	if err := s.taskRepo.UpdateStatus(ctx, taskID, status, strings.TrimSpace(logs)); err != nil {
		return err
	}
	return s.epicRepo.MarkReadyWhenAllDone(ctx)
}

func (s *TaskService) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	return s.taskRepo.GetByID(ctx, id)
}
