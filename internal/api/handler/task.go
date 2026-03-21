package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"granja/internal/domain"
	"granja/internal/service"
)

type TaskHandler struct {
	service *service.TaskService
}

func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

type updateTaskRequest struct {
	Status string `json:"status"`
	Logs   string `json:"logs"`
}

func (h *TaskHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	status := domain.TaskStatus(strings.TrimSpace(req.Status))
	if err := h.service.UpdateStatus(r.Context(), id, status, req.Logs); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	task, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if task == nil {
		respondError(w, http.StatusNotFound, "task not found")
		return
	}
	respondJSON(w, http.StatusOK, task)
}
