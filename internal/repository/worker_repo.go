package repository

import (
	"context"
	"database/sql"
	"time"

	"granja/internal/domain"
)

type WorkerRepository struct {
	db *sql.DB
}

func NewWorkerRepository(db *sql.DB) *WorkerRepository {
	return &WorkerRepository{db: db}
}

func (r *WorkerRepository) Create(ctx context.Context, w domain.Worker) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO workers (id, task_id, container_id, status, started_at)
		VALUES (?, ?, ?, ?, ?)
	`, w.ID, w.TaskID, w.ContainerID, w.Status, time.Now().UTC())
	return err
}

func (r *WorkerRepository) List(ctx context.Context) ([]domain.Worker, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, task_id, container_id, status, started_at, last_heartbeat
		FROM workers
		WHERE status IN ('starting', 'working', 'committing')
		ORDER BY started_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Worker
	for rows.Next() {
		var w domain.Worker
		if err := rows.Scan(&w.ID, &w.TaskID, &w.ContainerID, &w.Status, &w.StartedAt, &w.LastHeartbeat); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *WorkerRepository) CountActive(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM workers
		WHERE status IN ('starting', 'working', 'committing')
	`).Scan(&count)
	return count, err
}

func (r *WorkerRepository) UpdateStatus(ctx context.Context, id string, status domain.WorkerStatus) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE workers SET status = ?, last_heartbeat = ? WHERE id = ?
	`, status, time.Now().UTC(), id)
	return err
}

func (r *WorkerRepository) FindByContainer(ctx context.Context, containerID string) (*domain.Worker, error) {
	var w domain.Worker
	err := r.db.QueryRowContext(ctx, `
		SELECT id, task_id, container_id, status, started_at, last_heartbeat
		FROM workers WHERE container_id = ?
	`, containerID).Scan(&w.ID, &w.TaskID, &w.ContainerID, &w.Status, &w.StartedAt, &w.LastHeartbeat)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &w, nil
}
