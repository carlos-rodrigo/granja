package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr              string
	DBPath            string
	DockerWorkerImage string
	MaxWorkers        int
	OrchestratorPoll  time.Duration
	AnthropicModel    string
	ReviewRepoPath    string
}

func Load() Config {
	return Config{
		Addr:              getEnv("GRANJA_ADDR", ":3000"),
		DBPath:            getEnv("GRANJA_DB_PATH", "granja.db"),
		DockerWorkerImage: getEnv("GRANJA_WORKER_IMAGE", "granja-worker:latest"),
		MaxWorkers:        getEnvInt("GRANJA_MAX_WORKERS", 3),
		OrchestratorPoll:  getEnvDuration("GRANJA_ORCH_POLL", 10*time.Second),
		AnthropicModel:    getEnv("GRANJA_ANTHROPIC_MODEL", "claude-sonnet-4-5"),
		ReviewRepoPath:    getEnv("GRANJA_REVIEW_REPO_PATH", "."),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return fallback
}
