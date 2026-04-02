package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"granja/internal/api"
	"granja/internal/api/handler"
	"granja/internal/config"
	"granja/internal/orchestrator"
	"granja/internal/repository"
	"granja/internal/service"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := repository.OpenSQLite(cfg.DBPath)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()
	if err := repository.RunMigrations(ctx, db, "migrations"); err != nil {
		logger.Error("run migrations", "error", err)
		os.Exit(1)
	}

	projectRepo := repository.NewProjectRepository(db)
	epicRepo := repository.NewEpicRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	workerRepo := repository.NewWorkerRepository(db)

	projectSvc := service.NewProjectService(projectRepo)
	parserSvc := service.NewParserService(cfg.PiModel)
	epicSvc := service.NewEpicService(epicRepo, taskRepo, parserSvc)
	taskSvc := service.NewTaskService(taskRepo, epicRepo)

	dockerSvc, err := service.NewDockerService(cfg.DockerWorkerImage)
	if err != nil {
		logger.Error("init docker client", "error", err)
		os.Exit(1)
	}
	if err := dockerSvc.Ping(ctx); err != nil {
		logger.Warn("docker is not reachable, orchestrator will retry on tick", "error", err)
	}

	router := api.NewRouter(logger, api.Handlers{
		Project: handler.NewProjectHandler(projectSvc),
		Epic:    handler.NewEpicHandler(epicSvc, taskSvc),
		Task:    handler.NewTaskHandler(taskSvc),
		Worker:  handler.NewWorkerHandler(workerRepo, taskRepo, dockerSvc),
	})

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	reviewer := orchestrator.NewReviewer(cfg.PiModel, cfg.PiThinking, cfg.ReviewRepoPath)

	var githubSvc *service.GitHubService
	if cfg.GitHubToken != "" {
		githubSvc = service.NewGitHubService(cfg.GitHubToken)
		logger.Info("GitHub service initialized")
	} else {
		logger.Warn("GITHUB_TOKEN not set, merge flow will be disabled")
	}

	orch := orchestrator.New(logger, cfg.OrchestratorPoll, cfg.MaxWorkers, "http://host.docker.internal"+cfg.Addr, taskRepo, epicRepo, workerRepo, projectRepo, taskSvc, epicSvc, dockerSvc, reviewer, githubSvc)
	orchCtx, cancelOrch := context.WithCancel(context.Background())
	defer cancelOrch()
	go orch.Run(orchCtx)

	go func() {
		logger.Info("server listening", "addr", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	cancelOrch()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
