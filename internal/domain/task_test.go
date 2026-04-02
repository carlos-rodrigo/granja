package domain

import "testing"

func TestTaskStatusConstants(t *testing.T) {
	tests := []struct {
		status TaskStatus
		want   string
	}{
		{TaskTodo, "todo"},
		{TaskInProgress, "in_progress"},
		{TaskDone, "done"},
		{TaskBlocked, "blocked"},
	}
	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("TaskStatus = %q, want %q", tt.status, tt.want)
		}
	}
}

func TestEpicStatusConstants(t *testing.T) {
	tests := []struct {
		status EpicStatus
		want   string
	}{
		{EpicPlanted, "planted"},
		{EpicGrowing, "growing"},
		{EpicReady, "ready"},
		{EpicHarvested, "harvested"},
		{EpicBlocked, "blocked"},
	}
	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("EpicStatus = %q, want %q", tt.status, tt.want)
		}
	}
}

func TestWorkerStatusConstants(t *testing.T) {
	tests := []struct {
		status WorkerStatus
		want   string
	}{
		{WorkerStarting, "starting"},
		{WorkerWorking, "working"},
		{WorkerCommitting, "committing"},
		{WorkerDone, "done"},
		{WorkerError, "error"},
	}
	for _, tt := range tests {
		if string(tt.status) != tt.want {
			t.Errorf("WorkerStatus = %q, want %q", tt.status, tt.want)
		}
	}
}

func TestTaskStatusValidTransitions(t *testing.T) {
	// Verify that status values can be used for comparisons and switches
	valid := map[TaskStatus]bool{
		TaskTodo:       true,
		TaskInProgress: true,
		TaskDone:       true,
		TaskBlocked:    true,
	}

	check := func(s TaskStatus) bool {
		return valid[s]
	}

	if !check(TaskTodo) {
		t.Error("TaskTodo should be valid")
	}
	if !check(TaskDone) {
		t.Error("TaskDone should be valid")
	}
	if check(TaskStatus("invalid")) {
		t.Error("invalid status should not be valid")
	}
}

func TestTaskDependencyFields(t *testing.T) {
	dep := TaskDependency{
		TaskID:      "task_1",
		DependsOnID: "task_2",
	}
	if dep.TaskID != "task_1" {
		t.Errorf("TaskID = %q, want %q", dep.TaskID, "task_1")
	}
	if dep.DependsOnID != "task_2" {
		t.Errorf("DependsOnID = %q, want %q", dep.DependsOnID, "task_2")
	}
}

func TestTaskZeroValue(t *testing.T) {
	var task Task
	if task.Status != "" {
		t.Errorf("zero value Status = %q, want empty", task.Status)
	}
	if task.StartedAt != nil {
		t.Error("zero value StartedAt should be nil")
	}
	if task.CompletedAt != nil {
		t.Error("zero value CompletedAt should be nil")
	}
}

func TestEpicZeroValue(t *testing.T) {
	var epic Epic
	if epic.Status != "" {
		t.Errorf("zero value Status = %q, want empty", epic.Status)
	}
	if epic.ProjectID != "" {
		t.Errorf("zero value ProjectID = %q, want empty", epic.ProjectID)
	}
}
