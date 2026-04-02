package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"granja/internal/domain"
)

// --- mock ---

type mockTaskService struct {
	updateStatusFn func(ctx context.Context, id string, status domain.TaskStatus, logs string) error
	getByIDFn      func(ctx context.Context, id string) (*domain.Task, error)
}

func (m *mockTaskService) UpdateStatus(ctx context.Context, id string, status domain.TaskStatus, logs string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status, logs)
	}
	return nil
}

func (m *mockTaskService) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

// --- helpers ---

func newChiRequest(method, path string, body string, urlParams map[string]string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	for k, v := range urlParams {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return m
}

// --- tests ---

func TestComplete(t *testing.T) {
	task := &domain.Task{ID: "task_1", EpicID: "epic_1", Title: "Do stuff", Status: domain.TaskDone}

	tests := []struct {
		name           string
		id             string
		updateErr      error
		getResult      *domain.Task
		getErr         error
		wantStatusCode int
		wantError      bool
	}{
		{
			name:           "happy path",
			id:             "task_1",
			updateErr:      nil,
			getResult:      task,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "update status fails",
			id:             "bad_id",
			updateErr:      errors.New("task not found"),
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name:           "get returns nil (not found after update)",
			id:             "task_1",
			updateErr:      nil,
			getResult:      nil,
			wantStatusCode: http.StatusNotFound,
			wantError:      true,
		},
		{
			name:           "get returns error",
			id:             "task_1",
			updateErr:      nil,
			getErr:         errors.New("db error"),
			wantStatusCode: http.StatusInternalServerError,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedStatus domain.TaskStatus
			svc := &mockTaskService{
				updateStatusFn: func(_ context.Context, id string, status domain.TaskStatus, _ string) error {
					capturedStatus = status
					return tt.updateErr
				},
				getByIDFn: func(_ context.Context, id string) (*domain.Task, error) {
					return tt.getResult, tt.getErr
				},
			}

			h := NewTaskHandler(svc)
			rr := httptest.NewRecorder()
			req := newChiRequest(http.MethodPost, "/api/tasks/"+tt.id+"/complete", "", map[string]string{"id": tt.id})

			h.Complete(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatusCode)
			}
			if tt.wantError {
				body := decodeBody(t, rr)
				if _, ok := body["error"]; !ok {
					t.Error("expected error in response body")
				}
			}
			if tt.updateErr == nil && capturedStatus != domain.TaskDone {
				t.Errorf("expected status %q, got %q", domain.TaskDone, capturedStatus)
			}
		})
	}
}

func TestFail(t *testing.T) {
	task := &domain.Task{ID: "task_1", EpicID: "epic_1", Title: "Do stuff", Status: domain.TaskBlocked}

	tests := []struct {
		name           string
		id             string
		updateErr      error
		getResult      *domain.Task
		getErr         error
		wantStatusCode int
		wantError      bool
	}{
		{
			name:           "happy path",
			id:             "task_1",
			getResult:      task,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "update fails",
			id:             "bad_id",
			updateErr:      errors.New("not found"),
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name:           "get returns nil",
			id:             "task_1",
			getResult:      nil,
			wantStatusCode: http.StatusNotFound,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedStatus domain.TaskStatus
			svc := &mockTaskService{
				updateStatusFn: func(_ context.Context, _ string, status domain.TaskStatus, _ string) error {
					capturedStatus = status
					return tt.updateErr
				},
				getByIDFn: func(_ context.Context, _ string) (*domain.Task, error) {
					return tt.getResult, tt.getErr
				},
			}

			h := NewTaskHandler(svc)
			rr := httptest.NewRecorder()
			req := newChiRequest(http.MethodPost, "/api/tasks/"+tt.id+"/fail", "", map[string]string{"id": tt.id})

			h.Fail(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatusCode)
			}
			if tt.wantError {
				body := decodeBody(t, rr)
				if _, ok := body["error"]; !ok {
					t.Error("expected error in response body")
				}
			}
			if tt.updateErr == nil && capturedStatus != domain.TaskBlocked {
				t.Errorf("expected status %q, got %q", domain.TaskBlocked, capturedStatus)
			}
		})
	}
}

func TestPatch(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		body           string
		updateErr      error
		getResult      *domain.Task
		getErr         error
		wantStatusCode int
		wantError      bool
	}{
		{
			name:           "happy path with status update",
			id:             "task_1",
			body:           `{"status":"in_progress","logs":"starting"}`,
			getResult:      &domain.Task{ID: "task_1", Status: domain.TaskInProgress},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			id:             "task_1",
			body:           `{bad json`,
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name:           "service rejects invalid status",
			id:             "task_1",
			body:           `{"status":"invalid_status"}`,
			updateErr:      errors.New("invalid task status"),
			wantStatusCode: http.StatusBadRequest,
			wantError:      true,
		},
		{
			name:           "task not found after update",
			id:             "task_1",
			body:           `{"status":"done"}`,
			getResult:      nil,
			wantStatusCode: http.StatusNotFound,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTaskService{
				updateStatusFn: func(_ context.Context, _ string, _ domain.TaskStatus, _ string) error {
					return tt.updateErr
				},
				getByIDFn: func(_ context.Context, _ string) (*domain.Task, error) {
					return tt.getResult, tt.getErr
				},
			}

			h := NewTaskHandler(svc)
			rr := httptest.NewRecorder()
			req := newChiRequest(http.MethodPatch, "/api/tasks/"+tt.id, tt.body, map[string]string{"id": tt.id})

			h.Patch(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatusCode)
			}
			if tt.wantError {
				body := decodeBody(t, rr)
				if _, ok := body["error"]; !ok {
					t.Error("expected error in response body")
				}
			}
			if !tt.wantError && rr.Code == http.StatusOK {
				body := decodeBody(t, rr)
				if body["id"] != tt.id {
					t.Errorf("response id = %v, want %v", body["id"], tt.id)
				}
			}
		})
	}
}
