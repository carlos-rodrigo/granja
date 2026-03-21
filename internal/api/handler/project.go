package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"granja/internal/service"
)

type ProjectHandler struct {
	service *service.ProjectService
}

func NewProjectHandler(service *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{service: service}
}

type createProjectRequest struct {
	Name    string `json:"name"`
	RepoURL string `json:"repo_url"`
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.RepoURL) == "" {
		respondError(w, http.StatusBadRequest, "name and repo_url are required")
		return
	}

	project, err := h.service.Create(r.Context(), strings.TrimSpace(req.Name), strings.TrimSpace(req.RepoURL))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, projects)
}
