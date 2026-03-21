package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"granja/internal/api/handler"
	"granja/internal/api/middleware"
)

type Handlers struct {
	Project *handler.ProjectHandler
	Epic    *handler.EpicHandler
	Task    *handler.TaskHandler
	Worker  *handler.WorkerHandler
}

func NewRouter(logger *slog.Logger, h Handlers) http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logging(logger))

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		handlerPayload := map[string]string{"status": "ok"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
		_ = handlerPayload
	})

	r.Route("/api", func(api chi.Router) {
		api.Get("/projects", h.Project.List)
		api.Post("/projects", h.Project.Create)

		api.Get("/epics", h.Epic.List)
		api.Post("/epics", h.Epic.Create)
		api.Get("/epics/{id}", h.Epic.Get)
		api.Get("/epics/{id}/tasks", h.Epic.ListTasks)

		api.Patch("/tasks/{id}", h.Task.Patch)

		api.Get("/workers", h.Worker.List)
		api.Get("/workers/{id}/logs", h.Worker.Logs)
	})

	return r
}
