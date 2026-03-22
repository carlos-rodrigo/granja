package orchestrator

import (
	"context"
	"fmt"
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
	epicSvc     *service.EpicService
	dockerSvc   *service.DockerService
	reviewer    *Reviewer
	githubSvc   *service.GitHubService
}

func New(logger *slog.Logger, pollInterval time.Duration, maxWorkers int, apiBaseURL string, taskRepo *repository.TaskRepository, epicRepo *repository.EpicRepository, workerRepo *repository.WorkerRepository, projectRepo *repository.ProjectRepository, taskSvc *service.TaskService, epicSvc *service.EpicService, dockerSvc *service.DockerService, reviewer *Reviewer, githubSvc *service.GitHubService) *Orchestrator {
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
		epicSvc:      epicSvc,
		dockerSvc:    dockerSvc,
		reviewer:     reviewer,
		githubSvc:    githubSvc,
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

	if err := o.reviewReadyEpics(ctx); err != nil {
		o.logger.Error("review ready epics", "error", err)
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

func (o *Orchestrator) reviewReadyEpics(ctx context.Context) error {
	if o.reviewer == nil || o.epicSvc == nil {
		return nil
	}
	epics, err := o.epicRepo.List(ctx, "", string(domain.EpicReady))
	if err != nil {
		return err
	}
	for _, epic := range epics {
		isReady, err := o.epicSvc.IsReadyForReview(ctx, epic.ID)
		if err != nil {
			o.logger.Error("check review readiness", "epic_id", epic.ID, "error", err)
			continue
		}
		if !isReady {
			continue
		}
		project, err := o.projectRepo.GetByID(ctx, epic.ProjectID)
		if err != nil || project == nil {
			o.logger.Error("load project for review", "epic_id", epic.ID, "error", err)
			continue
		}

		reviewResult, err := o.reviewer.ReviewEpic(ctx, &epic, project)
		if err != nil {
			o.logger.Error("review epic", "epic_id", epic.ID, "error", err)
			_ = o.epicRepo.UpdateStatus(ctx, epic.ID, domain.EpicBlocked, err.Error())
			continue
		}

		if reviewResult.IsPass() {
			o.triggerMergeFlow(ctx, &epic)
		}
		if err := o.epicSvc.HandleReviewResult(ctx, epic.ID, *reviewResult); err != nil {
			o.logger.Error("handle review result", "epic_id", epic.ID, "error", err)
			_ = o.epicRepo.UpdateStatus(ctx, epic.ID, domain.EpicBlocked, err.Error())
		}
	}
	return nil
}

func (o *Orchestrator) triggerMergeFlow(ctx context.Context, epic *domain.Epic) {
	o.logger.Info("merge flow triggered", "epic_id", epic.ID, "branch", epic.BranchName)

	if o.githubSvc == nil {
		o.logger.Warn("GitHub service not configured, skipping merge flow", "epic_id", epic.ID)
		return
	}

	project, err := o.projectRepo.GetByID(ctx, epic.ProjectID)
	if err != nil || project == nil {
		o.logger.Error("load project for merge flow", "epic_id", epic.ID, "error", err)
		return
	}

	// Run merge flow in goroutine to avoid blocking orchestrator loop
	go o.runMergeFlow(context.Background(), epic, project)
}

func (o *Orchestrator) runMergeFlow(ctx context.Context, epic *domain.Epic, project *domain.Project) {
	owner, repo, err := service.ParseRepoURL(project.RepoURL)
	if err != nil {
		o.logger.Error("parse repo URL", "epic_id", epic.ID, "repo_url", project.RepoURL, "error", err)
		_ = o.epicRepo.UpdateStatus(ctx, epic.ID, domain.EpicBlocked, "invalid repo URL: "+err.Error())
		return
	}

	base := project.DefaultBranch
	if base == "" {
		base = "main"
	}

	// Extract PRD summary for PR body
	prdSummary := epic.PRDContent
	if len(prdSummary) > 2000 {
		prdSummary = prdSummary[:2000] + "..."
	}

	pr, err := o.githubSvc.CreatePR(ctx, owner, repo, epic.BranchName, base, epic.Title, prdSummary)
	if err != nil {
		o.logger.Error("create pull request", "epic_id", epic.ID, "error", err)
		_ = o.epicRepo.UpdateStatus(ctx, epic.ID, domain.EpicBlocked, "failed to create PR: "+err.Error())
		return
	}

	prNumber := pr.GetNumber()
	o.logger.Info("pull request created", "epic_id", epic.ID, "pr_number", prNumber, "url", pr.GetHTMLURL())

	// Poll for CI status (every 30s, max 10 min)
	const (
		pollInterval = 30 * time.Second
		maxWait      = 10 * time.Minute
	)

	deadline := time.Now().Add(maxWait)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		passed, pending, err := o.githubSvc.GetPRStatus(ctx, owner, repo, prNumber)
		if err != nil {
			o.logger.Error("get PR status", "epic_id", epic.ID, "pr_number", prNumber, "error", err)
			_ = o.epicRepo.UpdateStatus(ctx, epic.ID, domain.EpicBlocked, "failed to get PR status: "+err.Error())
			return
		}

		if passed {
			o.logger.Info("CI passed, merging PR", "epic_id", epic.ID, "pr_number", prNumber)
			commitMsg := fmt.Sprintf("%s (#%d)", epic.Title, prNumber)
			if err := o.githubSvc.MergePR(ctx, owner, repo, prNumber, commitMsg); err != nil {
				o.logger.Error("merge PR", "epic_id", epic.ID, "pr_number", prNumber, "error", err)
				_ = o.epicRepo.UpdateStatus(ctx, epic.ID, domain.EpicBlocked, "failed to merge PR: "+err.Error())
				return
			}

			o.logger.Info("PR merged, epic harvested", "epic_id", epic.ID, "pr_number", prNumber)
			_ = o.epicRepo.UpdateStatus(ctx, epic.ID, domain.EpicHarvested, "")
			return
		}

		if !pending {
			o.logger.Error("CI failed", "epic_id", epic.ID, "pr_number", prNumber)
			_ = o.epicRepo.UpdateStatus(ctx, epic.ID, domain.EpicBlocked, fmt.Sprintf("CI failed for PR #%d", prNumber))
			return
		}

		if time.Now().After(deadline) {
			o.logger.Error("CI timeout", "epic_id", epic.ID, "pr_number", prNumber)
			_ = o.epicRepo.UpdateStatus(ctx, epic.ID, domain.EpicBlocked, fmt.Sprintf("CI timeout waiting for PR #%d", prNumber))
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			o.logger.Debug("waiting for CI", "epic_id", epic.ID, "pr_number", prNumber)
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
