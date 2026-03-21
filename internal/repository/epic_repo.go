package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"granja/internal/domain"
)

type EpicRepository struct {
	db *sql.DB
}

func NewEpicRepository(db *sql.DB) *EpicRepository {
	return &EpicRepository{db: db}
}

func (r *EpicRepository) Create(ctx context.Context, e domain.Epic) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO epics (id, project_id, title, status, branch_name, prd_content, design_content)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, e.ID, e.ProjectID, e.Title, e.Status, e.BranchName, e.PRDContent, e.DesignContent)
	return err
}

func (r *EpicRepository) List(ctx context.Context, projectID, status string) ([]domain.Epic, error) {
	query := `
		SELECT id, project_id, title, status, branch_name, prd_content, design_content, review_result, error_message, created_at, updated_at
		FROM epics
	`
	var args []any
	var where []string
	if projectID != "" {
		where = append(where, "project_id = ?")
		args = append(args, projectID)
	}
	if status != "" {
		where = append(where, "status = ?")
		args = append(args, status)
	}
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Epic
	for rows.Next() {
		var e domain.Epic
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Title, &e.Status, &e.BranchName, &e.PRDContent, &e.DesignContent, &e.ReviewResult, &e.ErrorMessage, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *EpicRepository) GetByID(ctx context.Context, id string) (*domain.Epic, error) {
	var e domain.Epic
	err := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, title, status, branch_name, prd_content, design_content, review_result, error_message, created_at, updated_at
		FROM epics WHERE id = ?
	`, id).Scan(&e.ID, &e.ProjectID, &e.Title, &e.Status, &e.BranchName, &e.PRDContent, &e.DesignContent, &e.ReviewResult, &e.ErrorMessage, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *EpicRepository) UpdateStatus(ctx context.Context, id string, status domain.EpicStatus, errorMsg string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE epics
		SET status = ?, error_message = ?, updated_at = ?
		WHERE id = ?
	`, status, errorMsg, time.Now().UTC(), id)
	return err
}

func (r *EpicRepository) MarkReadyWhenAllDone(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE epics
		SET status = 'ready', updated_at = ?
		WHERE status = 'growing'
		  AND NOT EXISTS (
			SELECT 1 FROM tasks t
			WHERE t.epic_id = epics.id AND t.status != 'done'
		  )
	`, time.Now().UTC())
	return err
}
