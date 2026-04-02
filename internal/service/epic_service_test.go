package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"granja/internal/domain"
)

// --- mock epic repo ---

type mockEpicRepo struct {
	createFn              func(ctx context.Context, e domain.Epic) error
	getByIDFn             func(ctx context.Context, id string) (*domain.Epic, error)
	listFn                func(ctx context.Context, projectID, status string) ([]domain.Epic, error)
	updateStatusFn        func(ctx context.Context, id string, status domain.EpicStatus, errorMsg string) error
	markReadyWhenAllDoneFn func(ctx context.Context) error
	setReviewResultFn     func(ctx context.Context, id, reviewResult string) error
}

func (m *mockEpicRepo) Create(ctx context.Context, e domain.Epic) error {
	if m.createFn != nil {
		return m.createFn(ctx, e)
	}
	return nil
}

func (m *mockEpicRepo) GetByID(ctx context.Context, id string) (*domain.Epic, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockEpicRepo) List(ctx context.Context, projectID, status string) ([]domain.Epic, error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, status)
	}
	return nil, nil
}

func (m *mockEpicRepo) UpdateStatus(ctx context.Context, id string, status domain.EpicStatus, errorMsg string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status, errorMsg)
	}
	return nil
}

func (m *mockEpicRepo) MarkReadyWhenAllDone(ctx context.Context) error {
	if m.markReadyWhenAllDoneFn != nil {
		return m.markReadyWhenAllDoneFn(ctx)
	}
	return nil
}

func (m *mockEpicRepo) SetReviewResult(ctx context.Context, id, reviewResult string) error {
	if m.setReviewResultFn != nil {
		return m.setReviewResultFn(ctx, id, reviewResult)
	}
	return nil
}

// --- mock task repo ---

type mockTaskRepo struct {
	createFn        func(ctx context.Context, t domain.Task) error
	listByEpicFn    func(ctx context.Context, epicID string) ([]domain.Task, error)
	getByIDFn       func(ctx context.Context, id string) (*domain.Task, error)
	updateStatusFn  func(ctx context.Context, id string, status domain.TaskStatus, logs string) error
	addDependencyFn func(ctx context.Context, taskID, dependsOnID string) error
}

func (m *mockTaskRepo) Create(ctx context.Context, t domain.Task) error {
	if m.createFn != nil {
		return m.createFn(ctx, t)
	}
	return nil
}

func (m *mockTaskRepo) ListByEpic(ctx context.Context, epicID string) ([]domain.Task, error) {
	if m.listByEpicFn != nil {
		return m.listByEpicFn(ctx, epicID)
	}
	return nil, nil
}

func (m *mockTaskRepo) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockTaskRepo) UpdateStatus(ctx context.Context, id string, status domain.TaskStatus, logs string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status, logs)
	}
	return nil
}

func (m *mockTaskRepo) AddDependency(ctx context.Context, taskID, dependsOnID string) error {
	if m.addDependencyFn != nil {
		return m.addDependencyFn(ctx, taskID, dependsOnID)
	}
	return nil
}

// --- tests ---

func TestEpicCreate_StatusIsPlanted(t *testing.T) {
	var createdEpic domain.Epic

	epicRepo := &mockEpicRepo{
		createFn: func(_ context.Context, e domain.Epic) error {
			createdEpic = e
			return nil
		},
		getByIDFn: func(_ context.Context, id string) (*domain.Epic, error) {
			return &createdEpic, nil
		},
	}
	taskRepo := &mockTaskRepo{}

	svc := NewEpicService(epicRepo, taskRepo, nil)
	epic, err := svc.Create(context.Background(), "proj_1", "# My Feature\nSome PRD content", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if epic.Status != domain.EpicPlanted {
		t.Errorf("status = %q, want %q", epic.Status, domain.EpicPlanted)
	}
	if !strings.HasPrefix(epic.ID, "epic_") {
		t.Errorf("id = %q, want prefix 'epic_'", epic.ID)
	}
	if epic.Title != "My Feature" {
		t.Errorf("title = %q, want %q", epic.Title, "My Feature")
	}
	if epic.BranchName != "epic/my-feature" {
		t.Errorf("branch = %q, want %q", epic.BranchName, "epic/my-feature")
	}
}

func TestEpicCreate_EmptyPRDReturnsError(t *testing.T) {
	epicRepo := &mockEpicRepo{}
	taskRepo := &mockTaskRepo{}
	svc := NewEpicService(epicRepo, taskRepo, nil)

	_, err := svc.Create(context.Background(), "proj_1", "", "")
	if err == nil {
		t.Fatal("expected error for empty PRD")
	}
	if !strings.Contains(err.Error(), "prd is required") {
		t.Errorf("error = %q, want 'prd is required'", err.Error())
	}
}

func TestEpicCreate_RepoErrorPropagates(t *testing.T) {
	epicRepo := &mockEpicRepo{
		createFn: func(_ context.Context, _ domain.Epic) error {
			return errors.New("db write failed")
		},
	}
	taskRepo := &mockTaskRepo{}
	svc := NewEpicService(epicRepo, taskRepo, nil)

	_, err := svc.Create(context.Background(), "proj_1", "# Title\ncontent", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "db write failed") {
		t.Errorf("error = %q, want 'db write failed'", err.Error())
	}
}

func TestEpicList(t *testing.T) {
	want := []domain.Epic{
		{ID: "epic_1", Title: "First", Status: domain.EpicPlanted},
		{ID: "epic_2", Title: "Second", Status: domain.EpicGrowing},
	}

	epicRepo := &mockEpicRepo{
		listFn: func(_ context.Context, projectID, status string) ([]domain.Epic, error) {
			return want, nil
		},
	}
	taskRepo := &mockTaskRepo{}
	svc := NewEpicService(epicRepo, taskRepo, nil)

	got, err := svc.List(context.Background(), "proj_1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("got %d epics, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i].ID {
			t.Errorf("epic[%d].ID = %q, want %q", i, got[i].ID, want[i].ID)
		}
	}
}

func TestEpicList_FilterByStatus(t *testing.T) {
	var capturedStatus string
	epicRepo := &mockEpicRepo{
		listFn: func(_ context.Context, _ string, status string) ([]domain.Epic, error) {
			capturedStatus = status
			return nil, nil
		},
	}
	taskRepo := &mockTaskRepo{}
	svc := NewEpicService(epicRepo, taskRepo, nil)

	_, _ = svc.List(context.Background(), "", "growing")
	if capturedStatus != "growing" {
		t.Errorf("status filter = %q, want %q", capturedStatus, "growing")
	}
}

func TestEpicGetWithTasks(t *testing.T) {
	epic := &domain.Epic{ID: "epic_1", Title: "Test", Status: domain.EpicGrowing}
	tasks := []domain.Task{
		{ID: "task_1", EpicID: "epic_1", Title: "Task One", Status: domain.TaskTodo},
		{ID: "task_2", EpicID: "epic_1", Title: "Task Two", Status: domain.TaskDone},
	}

	epicRepo := &mockEpicRepo{
		getByIDFn: func(_ context.Context, id string) (*domain.Epic, error) {
			if id == "epic_1" {
				return epic, nil
			}
			return nil, nil
		},
	}
	taskRepo := &mockTaskRepo{
		listByEpicFn: func(_ context.Context, epicID string) ([]domain.Task, error) {
			if epicID == "epic_1" {
				return tasks, nil
			}
			return nil, nil
		},
	}
	svc := NewEpicService(epicRepo, taskRepo, nil)

	gotEpic, gotTasks, err := svc.GetWithTasks(context.Background(), "epic_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotEpic == nil {
		t.Fatal("expected epic, got nil")
	}
	if gotEpic.ID != "epic_1" {
		t.Errorf("epic.ID = %q, want %q", gotEpic.ID, "epic_1")
	}
	if len(gotTasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(gotTasks))
	}
}

func TestEpicGetWithTasks_NotFound(t *testing.T) {
	epicRepo := &mockEpicRepo{
		getByIDFn: func(_ context.Context, id string) (*domain.Epic, error) {
			return nil, nil
		},
	}
	taskRepo := &mockTaskRepo{}
	svc := NewEpicService(epicRepo, taskRepo, nil)

	gotEpic, gotTasks, err := svc.GetWithTasks(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotEpic != nil {
		t.Errorf("expected nil epic, got %+v", gotEpic)
	}
	if gotTasks != nil {
		t.Errorf("expected nil tasks, got %+v", gotTasks)
	}
}

func TestEpicIsReadyForReview(t *testing.T) {
	tests := []struct {
		name      string
		epic      *domain.Epic
		tasks     []domain.Task
		wantReady bool
	}{
		{
			name: "all tasks done",
			epic: &domain.Epic{ID: "e1", Status: domain.EpicGrowing},
			tasks: []domain.Task{
				{ID: "t1", Status: domain.TaskDone},
				{ID: "t2", Status: domain.TaskDone},
			},
			wantReady: true,
		},
		{
			name: "some tasks not done",
			epic: &domain.Epic{ID: "e1", Status: domain.EpicGrowing},
			tasks: []domain.Task{
				{ID: "t1", Status: domain.TaskDone},
				{ID: "t2", Status: domain.TaskInProgress},
			},
			wantReady: false,
		},
		{
			name:      "no tasks",
			epic:      &domain.Epic{ID: "e1", Status: domain.EpicGrowing},
			tasks:     []domain.Task{},
			wantReady: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			epicRepo := &mockEpicRepo{
				getByIDFn: func(_ context.Context, _ string) (*domain.Epic, error) {
					return tt.epic, nil
				},
				updateStatusFn: func(_ context.Context, _ string, _ domain.EpicStatus, _ string) error {
					return nil
				},
			}
			taskRepo := &mockTaskRepo{
				listByEpicFn: func(_ context.Context, _ string) ([]domain.Task, error) {
					return tt.tasks, nil
				},
			}
			svc := NewEpicService(epicRepo, taskRepo, nil)

			ready, err := svc.IsReadyForReview(context.Background(), "e1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ready != tt.wantReady {
				t.Errorf("ready = %v, want %v", ready, tt.wantReady)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		prd  string
		want string
	}{
		{"# My Feature\nSome content", "My Feature"},
		{"## Sub heading\ncontent", "Sub heading"},
		{"no heading here", "Untitled Epic"},
		{"# \n## Real Title", "Real Title"},
	}
	for _, tt := range tests {
		got := extractTitle(tt.prd)
		if got != tt.want {
			t.Errorf("extractTitle(%q) = %q, want %q", tt.prd, got, tt.want)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My Feature", "my-feature"},
		{"Hello World 123", "hello-world-123"},
		{"Special!@#Chars", "special-chars"},
		{"", "epic"},
		{"---", "epic"},
	}
	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
