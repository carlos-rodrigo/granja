package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"granja/internal/domain"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	migration, err := os.ReadFile("../../migrations/001_init.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := db.Exec(string(migration)); err != nil {
		t.Fatalf("run migration: %v", err)
	}

	return db
}

func seedProject(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO projects (id, name, repo_url) VALUES (?, ?, ?)`, id, "test-proj", "https://example.com/repo")
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}
}

func seedEpic(t *testing.T, db *sql.DB, id, projectID string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO epics (id, project_id, title, status, prd_content) VALUES (?, ?, ?, ?, ?)`,
		id, projectID, "Test Epic", "growing", "# PRD")
	if err != nil {
		t.Fatalf("seed epic: %v", err)
	}
}

func seedTask(t *testing.T, db *sql.DB, id, epicID string, status domain.TaskStatus) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO tasks (id, epic_id, title, description, status, effort, relevant_files, container_id, worker_logs)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, epicID, "Task "+id, "", string(status), "", "", "", "")
	if err != nil {
		t.Fatalf("seed task %s: %v", id, err)
	}
}

func seedDep(t *testing.T, db *sql.DB, taskID, dependsOnID string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO task_deps (task_id, depends_on_id) VALUES (?, ?)`, taskID, dependsOnID)
	if err != nil {
		t.Fatalf("seed dep: %v", err)
	}
}

// --- FindReadyTasks tests ---

func TestFindReadyTasks_NoDeps_TodoOnly(t *testing.T) {
	db := setupTestDB(t)
	seedProject(t, db, "p1")
	seedEpic(t, db, "e1", "p1")

	seedTask(t, db, "t1", "e1", domain.TaskTodo)
	seedTask(t, db, "t2", "e1", domain.TaskDone)
	seedTask(t, db, "t3", "e1", domain.TaskInProgress)

	repo := NewTaskRepository(db)
	tasks, err := repo.FindReadyTasks(context.Background(), 10)
	if err != nil {
		t.Fatalf("FindReadyTasks: %v", err)
	}

	// t1 is todo but t3 is in_progress in the same epic, so no tasks should be ready
	// (the query excludes tasks in epics where another task is in_progress)
	if len(tasks) != 0 {
		t.Errorf("got %d ready tasks, want 0 (in_progress task blocks epic)", len(tasks))
	}
}

func TestFindReadyTasks_NoDeps_NoInProgress(t *testing.T) {
	db := setupTestDB(t)
	seedProject(t, db, "p1")
	seedEpic(t, db, "e1", "p1")

	seedTask(t, db, "t1", "e1", domain.TaskTodo)
	seedTask(t, db, "t2", "e1", domain.TaskDone)

	repo := NewTaskRepository(db)
	tasks, err := repo.FindReadyTasks(context.Background(), 10)
	if err != nil {
		t.Fatalf("FindReadyTasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("got %d ready tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "t1" {
		t.Errorf("ready task ID = %q, want %q", tasks[0].ID, "t1")
	}
}

func TestFindReadyTasks_WithDeps_AllMet(t *testing.T) {
	db := setupTestDB(t)
	seedProject(t, db, "p1")
	seedEpic(t, db, "e1", "p1")

	seedTask(t, db, "t1", "e1", domain.TaskDone) // dependency
	seedTask(t, db, "t2", "e1", domain.TaskTodo) // depends on t1

	seedDep(t, db, "t2", "t1")

	repo := NewTaskRepository(db)
	tasks, err := repo.FindReadyTasks(context.Background(), 10)
	if err != nil {
		t.Fatalf("FindReadyTasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("got %d ready tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "t2" {
		t.Errorf("ready task ID = %q, want %q", tasks[0].ID, "t2")
	}
}

func TestFindReadyTasks_WithDeps_NotMet(t *testing.T) {
	db := setupTestDB(t)
	seedProject(t, db, "p1")
	seedEpic(t, db, "e1", "p1")

	seedTask(t, db, "t1", "e1", domain.TaskTodo)       // dependency not done
	seedTask(t, db, "t2", "e1", domain.TaskTodo)       // depends on t1
	seedTask(t, db, "t3", "e1", domain.TaskInProgress) // dependency not done

	seedDep(t, db, "t2", "t1")
	seedDep(t, db, "t2", "t3")

	repo := NewTaskRepository(db)
	tasks, err := repo.FindReadyTasks(context.Background(), 10)
	if err != nil {
		t.Fatalf("FindReadyTasks: %v", err)
	}

	// t1 is todo but t3 is in_progress, blocking the entire epic
	// Also t2 has unmet deps. No tasks should be ready.
	if len(tasks) != 0 {
		t.Errorf("got %d ready tasks, want 0 (unmet deps + in_progress blocks epic)", len(tasks))
	}
}

func TestFindReadyTasks_CompletedNotReturned(t *testing.T) {
	db := setupTestDB(t)
	seedProject(t, db, "p1")
	seedEpic(t, db, "e1", "p1")

	seedTask(t, db, "t1", "e1", domain.TaskDone)
	seedTask(t, db, "t2", "e1", domain.TaskDone)

	repo := NewTaskRepository(db)
	tasks, err := repo.FindReadyTasks(context.Background(), 10)
	if err != nil {
		t.Fatalf("FindReadyTasks: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("got %d ready tasks, want 0 (all done)", len(tasks))
	}
}

func TestFindReadyTasks_LimitRespected(t *testing.T) {
	db := setupTestDB(t)
	seedProject(t, db, "p1")
	seedEpic(t, db, "e1", "p1")

	seedTask(t, db, "t1", "e1", domain.TaskTodo)
	seedTask(t, db, "t2", "e1", domain.TaskTodo)
	seedTask(t, db, "t3", "e1", domain.TaskTodo)

	repo := NewTaskRepository(db)
	tasks, err := repo.FindReadyTasks(context.Background(), 2)
	if err != nil {
		t.Fatalf("FindReadyTasks: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("got %d ready tasks, want 2 (limit=2)", len(tasks))
	}
}

func TestFindReadyTasks_MultipleEpics(t *testing.T) {
	db := setupTestDB(t)
	seedProject(t, db, "p1")
	seedEpic(t, db, "e1", "p1")
	seedEpic(t, db, "e2", "p1")

	// e1 has an in_progress task — blocks its todo tasks
	seedTask(t, db, "t1", "e1", domain.TaskInProgress)
	seedTask(t, db, "t2", "e1", domain.TaskTodo)

	// e2 has no in_progress — its todo task should be ready
	seedTask(t, db, "t3", "e2", domain.TaskTodo)

	repo := NewTaskRepository(db)
	tasks, err := repo.FindReadyTasks(context.Background(), 10)
	if err != nil {
		t.Fatalf("FindReadyTasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("got %d ready tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "t3" {
		t.Errorf("ready task ID = %q, want %q", tasks[0].ID, "t3")
	}
}

// --- CRUD tests ---

func TestTaskCreate_And_GetByID(t *testing.T) {
	db := setupTestDB(t)
	seedProject(t, db, "p1")
	seedEpic(t, db, "e1", "p1")

	// Seed using raw SQL so all nullable columns have explicit values.
	_, err := db.Exec(
		`INSERT INTO tasks (id, epic_id, title, description, status, effort, relevant_files, container_id, worker_logs)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"task_abc", "e1", "Build something", "A detailed description", "todo", "medium", "", "", "")
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	repo := NewTaskRepository(db)
	got, err := repo.GetByID(context.Background(), "task_abc")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("expected task, got nil")
	}
	if got.Title != "Build something" {
		t.Errorf("title = %q, want %q", got.Title, "Build something")
	}
	if got.Status != domain.TaskTodo {
		t.Errorf("status = %q, want %q", got.Status, domain.TaskTodo)
	}
}

func TestTaskGetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTaskRepository(db)

	got, err := repo.GetByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestTaskUpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	seedProject(t, db, "p1")
	seedEpic(t, db, "e1", "p1")
	seedTask(t, db, "t1", "e1", domain.TaskTodo)

	repo := NewTaskRepository(db)

	if err := repo.UpdateStatus(context.Background(), "t1", domain.TaskInProgress, "worker starting"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	got, err := repo.GetByID(context.Background(), "t1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Status != domain.TaskInProgress {
		t.Errorf("status = %q, want %q", got.Status, domain.TaskInProgress)
	}
	if got.StartedAt == nil {
		t.Error("expected started_at to be set")
	}

	// Transition to done
	if err := repo.UpdateStatus(context.Background(), "t1", domain.TaskDone, "all done"); err != nil {
		t.Fatalf("UpdateStatus to done: %v", err)
	}
	got, _ = repo.GetByID(context.Background(), "t1")
	if got.Status != domain.TaskDone {
		t.Errorf("status = %q, want %q", got.Status, domain.TaskDone)
	}
	if got.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestTaskListByEpic(t *testing.T) {
	db := setupTestDB(t)
	seedProject(t, db, "p1")
	seedEpic(t, db, "e1", "p1")
	seedEpic(t, db, "e2", "p1")

	seedTask(t, db, "t1", "e1", domain.TaskTodo)
	seedTask(t, db, "t2", "e1", domain.TaskDone)
	seedTask(t, db, "t3", "e2", domain.TaskTodo) // different epic

	repo := NewTaskRepository(db)
	tasks, err := repo.ListByEpic(context.Background(), "e1")
	if err != nil {
		t.Fatalf("ListByEpic: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}
	for _, task := range tasks {
		if task.EpicID != "e1" {
			t.Errorf("task %s has epic_id %q, want %q", task.ID, task.EpicID, "e1")
		}
	}
}

func TestTaskAddDependency(t *testing.T) {
	db := setupTestDB(t)
	seedProject(t, db, "p1")
	seedEpic(t, db, "e1", "p1")
	seedTask(t, db, "t1", "e1", domain.TaskDone)
	seedTask(t, db, "t2", "e1", domain.TaskTodo)

	repo := NewTaskRepository(db)
	if err := repo.AddDependency(context.Background(), "t2", "t1"); err != nil {
		t.Fatalf("AddDependency: %v", err)
	}

	// Verify via FindReadyTasks — t2 should be ready since t1 is done
	tasks, err := repo.FindReadyTasks(context.Background(), 10)
	if err != nil {
		t.Fatalf("FindReadyTasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != "t2" {
		t.Errorf("expected t2 to be ready, got %+v", tasks)
	}
}
