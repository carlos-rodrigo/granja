package domain

import "time"

type EpicStatus string

const (
	EpicPlanted   EpicStatus = "planted"
	EpicGrowing   EpicStatus = "growing"
	EpicReady     EpicStatus = "ready"
	EpicHarvested EpicStatus = "harvested"
	EpicBlocked   EpicStatus = "blocked"
)

type Epic struct {
	ID            string     `json:"id" db:"id"`
	ProjectID     string     `json:"project_id" db:"project_id"`
	Title         string     `json:"title" db:"title"`
	Status        EpicStatus `json:"status" db:"status"`
	BranchName    string     `json:"branch_name" db:"branch_name"`
	PRDContent    string     `json:"prd_content" db:"prd_content"`
	DesignContent string     `json:"design_content,omitempty" db:"design_content"`
	ReviewResult  string     `json:"review_result,omitempty" db:"review_result"`
	ErrorMessage  string     `json:"error_message,omitempty" db:"error_message"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}
