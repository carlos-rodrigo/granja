package domain

import "time"

type WorkerStatus string

const (
	WorkerStarting   WorkerStatus = "starting"
	WorkerWorking    WorkerStatus = "working"
	WorkerCommitting WorkerStatus = "committing"
	WorkerDone       WorkerStatus = "done"
	WorkerError      WorkerStatus = "error"
)

type Worker struct {
	ID            string       `json:"id" db:"id"`
	TaskID        string       `json:"task_id" db:"task_id"`
	ContainerID   string       `json:"container_id" db:"container_id"`
	Status        WorkerStatus `json:"status" db:"status"`
	StartedAt     time.Time    `json:"started_at" db:"started_at"`
	LastHeartbeat *time.Time   `json:"last_heartbeat,omitempty" db:"last_heartbeat"`
}
