package repository

import (
	"context"
	"database/sql"

	"granja/internal/domain"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, p domain.Project) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO projects (id, name, repo_url, default_branch)
		VALUES (?, ?, ?, ?)
	`, p.ID, p.Name, p.RepoURL, p.DefaultBranch)
	return err
}

func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, repo_url, default_branch, created_at
		FROM projects
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.RepoURL, &p.DefaultBranch, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	var p domain.Project
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, repo_url, default_branch, created_at
		FROM projects WHERE id = ?
	`, id).Scan(&p.ID, &p.Name, &p.RepoURL, &p.DefaultBranch, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}
