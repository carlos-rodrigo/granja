package service

import (
	"context"

	"github.com/google/uuid"

	"granja/internal/domain"
	"granja/internal/repository"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(ctx context.Context, name, repoURL string) (*domain.Project, error) {
	project := domain.Project{
		ID:            "proj_" + uuid.NewString(),
		Name:          name,
		RepoURL:       repoURL,
		DefaultBranch: "main",
	}
	if err := s.repo.Create(ctx, project); err != nil {
		return nil, err
	}
	stored, err := s.repo.GetByID(ctx, project.ID)
	if err != nil {
		return nil, err
	}
	return stored, nil
}

func (s *ProjectService) List(ctx context.Context) ([]domain.Project, error) {
	return s.repo.List(ctx)
}
