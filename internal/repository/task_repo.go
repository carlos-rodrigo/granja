package repository

import (
	"context"
	"database/sql"
	"time"

	"granja/internal/domain"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, t domain.Task) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (id, epic_id, title, description, status, effort, relevant_files)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, t.ID, t.EpicID, t.Title, t.Description, t.Status, t.Effort, t.RelevantFiles)
	return err
}

func (r *TaskRepository) AddDependency(ctx context.Context, taskID, dependsOnID string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO task_deps (task_id, depends_on_id) VALUES (?, ?)
	`, taskID, dependsOnID)
	return err
}

func (r *TaskRepository) ListByEpic(ctx context.Context, epicID string) ([]domain.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, epic_id, title, description, status, effort, relevant_files, container_id, worker_logs, started_at, completed_at, created_at
		FROM tasks WHERE epic_id = ?
		ORDER BY created_at ASC
	`, epicID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.EpicID, &t.Title, &t.Description, &t.Status, &t.Effort, &t.RelevantFiles, &t.ContainerID, &t.WorkerLogs, &t.StartedAt, &t.CompletedAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	var t domain.Task
	err := r.db.QueryRowContext(ctx, `
		SELECT id, epic_id, title, description, status, effort, relevant_files, container_id, worker_logs, started_at, completed_at, created_at
		FROM tasks WHERE id = ?
	`, id).Scan(&t.ID, &t.EpicID, &t.Title, &t.Description, &t.Status, &t.Effort, &t.RelevantFiles, &t.ContainerID, &t.WorkerLogs, &t.StartedAt, &t.CompletedAt, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id string, status domain.TaskStatus, logs string) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET status = ?,
			worker_logs = CASE WHEN ? != '' THEN ? ELSE worker_logs END,
			started_at = CASE WHEN ? = 'in_progress' THEN ? ELSE started_at END,
			completed_at = CASE WHEN ? IN ('done', 'blocked') THEN ? ELSE completed_at END
		WHERE id = ?
	`, status, logs, logs, status, now, status, now, id)
	return err
}

func (r *TaskRepository) AssignContainer(ctx context.Context, taskID, containerID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tasks SET container_id = ? WHERE id = ?
	`, containerID, taskID)
	return err
}

func (r *TaskRepository) FindReadyTasks(ctx context.Context, limit int) ([]domain.Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.epic_id, t.title, t.description, t.status, t.effort, t.relevant_files, t.container_id, t.worker_logs, t.started_at, t.completed_at, t.created_at
		FROM tasks t
		WHERE t.status = 'todo'
		  AND NOT EXISTS (
			SELECT 1
			FROM task_deps d
			JOIN tasks dep ON dep.id = d.depends_on_id
			WHERE d.task_id = t.id AND dep.status != 'done'
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM tasks t2
			WHERE t2.epic_id = t.epic_id AND t2.status = 'in_progress'
		  )
		ORDER BY t.created_at ASC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.EpicID, &t.Title, &t.Description, &t.Status, &t.Effort, &t.RelevantFiles, &t.ContainerID, &t.WorkerLogs, &t.StartedAt, &t.CompletedAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
