package orchestrator

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"granja/internal/domain"
	"granja/internal/repository"
	"granja/internal/service"
)

type Orchestrator struct {
	logger       *slog.Logger
	pollInterval time.Duration
	maxWorkers   int
	apiBaseURL   string

	taskRepo    *repository.TaskRepository
	epicRepo    *repository.EpicRepository
	workerRepo  *repository.WorkerRepository
	projectRepo *repository.ProjectRepository
	taskSvc     *service.TaskService
	dockerSvc   *service.DockerService
}

func New(logger *slog.Logger, pollInterval time.Duration, maxWorkers int, apiBaseURL string, taskRepo *repository.TaskRepository, epicRepo *repository.EpicRepository, workerRepo *repository.WorkerRepository, projectRepo *repository.ProjectRepository, taskSvc *service.TaskService, dockerSvc *service.DockerService) *Orchestrator {
	return &Orchestrator{
		logger:       logger,
		pollInterval: pollInterval,
		maxWorkers:   maxWorkers,
		apiBaseURL:   apiBaseURL,
		taskRepo:     taskRepo,
		epicRepo:     epicRepo,
		workerRepo:   workerRepo,
		projectRepo:  projectRepo,
		taskSvc:      taskSvc,
		dockerSvc:    dockerSvc,
	}
}

func (o *Orchestrator) Run(ctx context.Context) {
	ticker := time.NewTicker(o.pollInterval)
	defer ticker.Stop()
	o.logger.Info("orchestrator started", "poll_interval", o.pollInterval)

	o.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			o.logger.Info("orchestrator stopped")
			return
		case <-ticker.C:
			o.tick(ctx)
		}
	}
}

func (o *Orchestrator) tick(ctx context.Context) {
	if err := o.reconcileWorkers(ctx); err != nil {
		o.logger.Error("reconcile workers", "error", err)
	}

	active, err := o.workerRepo.CountActive(ctx)
	if err != nil {
		o.logger.Error("count active workers", "error", err)
		return
	}
	if active >= o.maxWorkers {
		return
	}

	ready, err := o.taskRepo.FindReadyTasks(ctx, o.maxWorkers-active)
	if err != nil {
		o.logger.Error("find ready tasks", "error", err)
		return
	}

	for _, task := range ready {
		epic, err := o.epicRepo.GetByID(ctx, task.EpicID)
		if err != nil || epic == nil {
			o.logger.Error("load epic for task", "task_id", task.ID, "error", err)
			continue
		}
		project, err := o.projectRepo.GetByID(ctx, epic.ProjectID)
		if err != nil || project == nil {
			o.logger.Error("load project for task", "task_id", task.ID, "error", err)
			continue
		}

		containerID, err := o.dockerSvc.SpawnWorker(ctx, service.SpawnInput{
			TaskID:      task.ID,
			TaskTitle:   task.Title,
			TaskPrompt:  task.Description,
			ProjectRepo: project.RepoURL,
			Branch:      epic.BranchName,
			APIBaseURL:  o.apiBaseURL,
		})
		if err != nil {
			o.logger.Error("spawn worker", "task_id", task.ID, "error", err)
			continue
		}

		if err := o.taskRepo.AssignContainer(ctx, task.ID, containerID); err != nil {
			o.logger.Error("assign container", "task_id", task.ID, "error", err)
		}
		if err := o.taskSvc.UpdateStatus(ctx, task.ID, domain.TaskInProgress, ""); err != nil {
			o.logger.Error("mark task in progress", "task_id", task.ID, "error", err)
		}
		if err := o.workerRepo.Create(ctx, domain.Worker{
			ID:          "wrk_" + uuid.NewString(),
			TaskID:      task.ID,
			ContainerID: containerID,
			Status:      domain.WorkerWorking,
		}); err != nil {
			o.logger.Error("create worker record", "task_id", task.ID, "error", err)
		}
	}
}

func (o *Orchestrator) reconcileWorkers(ctx context.Context) error {
	containers, err := o.dockerSvc.ListGranjaContainers(ctx)
	if err != nil {
		return err
	}
	for _, c := range containers {
		if c.Status == "running" || c.Status == "created" {
			continue
		}
		worker, err := o.workerRepo.FindByContainer(ctx, c.ID)
		if err != nil || worker == nil {
			continue
		}
		logs, _ := o.dockerSvc.Logs(ctx, c.ID)
		if c.ExitCode == 0 {
			_ = o.taskSvc.UpdateStatus(ctx, worker.TaskID, domain.TaskDone, logs)
			_ = o.workerRepo.UpdateStatus(ctx, worker.ID, domain.WorkerDone)
		} else {
			_ = o.taskSvc.UpdateStatus(ctx, worker.TaskID, domain.TaskBlocked, logs)
			_ = o.workerRepo.UpdateStatus(ctx, worker.ID, domain.WorkerError)
		}
		_ = o.dockerSvc.RemoveContainer(ctx, c.ID)
	}
	return nil
}
