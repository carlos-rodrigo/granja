package domain

import "time"

type TaskStatus string

const (
	TaskTodo       TaskStatus = "todo"
	TaskInProgress TaskStatus = "in_progress"
	TaskDone       TaskStatus = "done"
	TaskBlocked    TaskStatus = "blocked"
)

type Task struct {
	ID            string     `json:"id" db:"id"`
	EpicID        string     `json:"epic_id" db:"epic_id"`
	Title         string     `json:"title" db:"title"`
	Description   string     `json:"description,omitempty" db:"description"`
	Status        TaskStatus `json:"status" db:"status"`
	Effort        string     `json:"effort,omitempty" db:"effort"`
	RelevantFiles string     `json:"relevant_files,omitempty" db:"relevant_files"`
	ContainerID   string     `json:"container_id,omitempty" db:"container_id"`
	WorkerLogs    string     `json:"worker_logs,omitempty" db:"worker_logs"`
	StartedAt     *time.Time `json:"started_at,omitempty" db:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

type TaskDependency struct {
	TaskID      string `json:"task_id" db:"task_id"`
	DependsOnID string `json:"depends_on_id" db:"depends_on_id"`
}
