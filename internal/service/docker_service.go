package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

type DockerService struct {
	cli         *client.Client
	workerImage string
}

type SpawnInput struct {
	TaskID      string
	TaskTitle   string
	TaskPrompt  string
	ProjectRepo string
	Branch      string
	APIBaseURL  string
}

type ContainerState struct {
	ID       string
	Status   string
	ExitCode int
}

func NewDockerService(workerImage string) (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerService{cli: cli, workerImage: workerImage}, nil
}

func (s *DockerService) Ping(ctx context.Context) error {
	_, err := s.cli.Ping(ctx)
	return err
}

func (s *DockerService) EnsureImage(ctx context.Context) error {
	reader, err := s.cli.ImagePull(ctx, s.workerImage, image.PullOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "pull access denied") {
			return nil
		}
		return err
	}
	defer reader.Close()
	_, _ = io.Copy(io.Discard, reader)
	return nil
}

func (s *DockerService) SpawnWorker(ctx context.Context, in SpawnInput) (string, error) {
	env := []string{
		"TASK_ID=" + in.TaskID,
		"TASK_TITLE=" + in.TaskTitle,
		"TASK_PROMPT=" + in.TaskPrompt,
		"REPO_URL=" + in.ProjectRepo,
		"BRANCH=" + in.Branch,
		"GRANJA_API=" + in.APIBaseURL,
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}

	mounts := []mount.Mount{
		{
			Type:     mount.TypeBind,
			Source:   filepath.Join(homeDir, ".pi"),
			Target:   "/root/.pi",
			ReadOnly: false,
		},
	}

	gitconfigPath := filepath.Join(homeDir, ".gitconfig")
	if _, err := os.Stat(gitconfigPath); err == nil {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   gitconfigPath,
			Target:   "/root/.gitconfig",
			ReadOnly: true,
		})
	}

	resp, err := s.cli.ContainerCreate(ctx, &container.Config{
		Image: s.workerImage,
		Env:   env,
		Labels: map[string]string{
			"granja.task_id": in.TaskID,
		},
	}, &container.HostConfig{
		Mounts: mounts,
	}, nil, nil, fmt.Sprintf("granja-%s", in.TaskID))
	if err != nil {
		return "", err
	}
	if err := s.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (s *DockerService) ListGranjaContainers(ctx context.Context) ([]ContainerState, error) {
	items, err := s.cli.ContainerList(ctx, container.ListOptions{All: true, Filters: filters.NewArgs(filters.Arg("label", "granja.task_id"))})
	if err != nil {
		return nil, err
	}
	out := make([]ContainerState, 0, len(items))
	for _, c := range items {
		inspect, err := s.cli.ContainerInspect(ctx, c.ID)
		if err != nil {
			continue
		}
		exit := 0
		status := "unknown"
		if inspect.State != nil {
			exit = inspect.State.ExitCode
			status = inspect.State.Status
		}
		out = append(out, ContainerState{ID: c.ID, Status: status, ExitCode: exit})
	}
	return out, nil
}

func (s *DockerService) Logs(ctx context.Context, containerID string) (string, error) {
	reader, err := s.cli.ContainerLogs(ctx, containerID, container.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: "500"})
	if err != nil {
		return "", err
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *DockerService) RemoveContainer(ctx context.Context, containerID string) error {
	return s.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
}
