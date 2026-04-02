package service

import (
	"context"

	"granja/internal/domain"
)

// TaskRepo is the interface for task persistence used by services.
type TaskRepo interface {
	Create(ctx context.Context, t domain.Task) error
	ListByEpic(ctx context.Context, epicID string) ([]domain.Task, error)
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	UpdateStatus(ctx context.Context, id string, status domain.TaskStatus, logs string) error
	AddDependency(ctx context.Context, taskID, dependsOnID string) error
}

// EpicRepo is the interface for epic persistence used by services.
type EpicRepo interface {
	Create(ctx context.Context, e domain.Epic) error
	GetByID(ctx context.Context, id string) (*domain.Epic, error)
	List(ctx context.Context, projectID, status string) ([]domain.Epic, error)
	UpdateStatus(ctx context.Context, id string, status domain.EpicStatus, errorMsg string) error
	MarkReadyWhenAllDone(ctx context.Context) error
	SetReviewResult(ctx context.Context, id, reviewResult string) error
}
