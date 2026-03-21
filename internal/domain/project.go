package domain

import "time"

type Project struct {
	ID            string    `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	RepoURL       string    `json:"repo_url" db:"repo_url"`
	DefaultBranch string    `json:"default_branch" db:"default_branch"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
