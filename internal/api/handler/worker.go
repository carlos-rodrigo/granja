package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"granja/internal/repository"
	"granja/internal/service"
)

type WorkerHandler struct {
	workerRepo *repository.WorkerRepository
	taskRepo   *repository.TaskRepository
	dockerSvc  *service.DockerService
}

func NewWorkerHandler(workerRepo *repository.WorkerRepository, taskRepo *repository.TaskRepository, dockerSvc *service.DockerService) *WorkerHandler {
	return &WorkerHandler{workerRepo: workerRepo, taskRepo: taskRepo, dockerSvc: dockerSvc}
}

func (h *WorkerHandler) List(w http.ResponseWriter, r *http.Request) {
	workers, err := h.workerRepo.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, workers)
}

func (h *WorkerHandler) Logs(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	worker, err := h.workerRepo.FindByContainer(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if worker == nil {
		respondError(w, http.StatusNotFound, "worker not found")
		return
	}
	task, err := h.taskRepo.GetByID(r.Context(), worker.TaskID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	logs := ""
	if task != nil {
		logs = task.WorkerLogs
	}
	if logs == "" {
		if dockerLogs, derr := h.dockerSvc.Logs(r.Context(), worker.ContainerID); derr == nil {
			logs = dockerLogs
		}
	}
	respondJSON(w, http.StatusOK, map[string]any{"logs": logs})
}
