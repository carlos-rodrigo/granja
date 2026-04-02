package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"granja/internal/domain"
	"granja/internal/service"
)

// EpicTaskLister is the interface the epic handler needs for listing tasks.
type EpicTaskLister interface {
	ListByEpic(ctx context.Context, epicID string) ([]domain.Task, error)
}

type EpicHandler struct {
	epicService *service.EpicService
	taskService EpicTaskLister
}

func NewEpicHandler(epicService *service.EpicService, taskService EpicTaskLister) *EpicHandler {
	return &EpicHandler{epicService: epicService, taskService: taskService}
}

type createEpicRequest struct {
	ProjectID string `json:"project_id"`
	PRD       string `json:"prd"`
	Design    string `json:"design"`
}

func (h *EpicHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createEpicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.ProjectID) == "" || strings.TrimSpace(req.PRD) == "" {
		respondError(w, http.StatusBadRequest, "project_id and prd are required")
		return
	}
	epic, err := h.epicService.Create(r.Context(), req.ProjectID, req.PRD, req.Design)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, epic)
}

func (h *EpicHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project")
	status := r.URL.Query().Get("status")
	epics, err := h.epicService.List(r.Context(), projectID, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, epics)
}

func (h *EpicHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	epic, tasks, err := h.epicService.GetWithTasks(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if epic == nil {
		respondError(w, http.StatusNotFound, "epic not found")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{
		"epic":  epic,
		"tasks": tasks,
	})
}

func (h *EpicHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tasks, err := h.taskService.ListByEpic(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, tasks)
}
