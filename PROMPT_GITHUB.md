# Task: Implement GitHub Merge Flow

## Context
Granja is an AI agent orchestration system. The backend is in Go. 
When all tasks in an epic complete and pass review, we need to:
1. Create a PR on GitHub
2. Monitor CI status
3. Auto-merge when CI passes
4. Update epic status to "harvested"

## Current State
- `internal/orchestrator/orchestrator.go` has `triggerMergeFlow()` that only logs
- `internal/service/github_service.go` does NOT exist
- Config needs GitHub token field

## What to Implement

### 1. Update config.go
Add:
```go
GitHubToken string  // GITHUB_TOKEN env var
```

### 2. Create internal/service/github_service.go
Use the `google/go-github` package. Implement:

```go
type GitHubService struct {
    client *github.Client
}

func NewGitHubService(token string) *GitHubService

// CreatePR creates a pull request
func (s *GitHubService) CreatePR(ctx context.Context, owner, repo, head, base, title, body string) (*github.PullRequest, error)

// GetPRStatus checks if CI passed (combines check runs and commit statuses)
func (s *GitHubService) GetPRStatus(ctx context.Context, owner, repo string, prNumber int) (passed bool, pending bool, err error)

// MergePR merges the PR using squash merge
func (s *GitHubService) MergePR(ctx context.Context, owner, repo string, prNumber int, commitMessage string) error
```

### 3. Update orchestrator.go
- Inject GitHubService
- Implement `triggerMergeFlow()`:
  1. Parse owner/repo from project.RepoURL (handle both HTTPS and SSH URLs)
  2. Create PR with epic title and PRD summary as body
  3. Poll for CI status (every 30s, max 10 min)
  4. If CI passes, merge and update epic to "harvested"
  5. If CI fails, update epic to "blocked" with error message

### 4. Update cmd/server/main.go
- Initialize GitHubService with token from config
- Pass to orchestrator

## Constraints
- Use `github.com/google/go-github/v60/github` (already in go ecosystem)
- Handle missing token gracefully (skip merge flow, just log warning)
- Parse repo URLs: `git@github.com:user/repo.git` and `https://github.com/user/repo.git`
- Use squash merge
- Don't block orchestrator loop - run merge flow in goroutine

## Files to modify/create:
1. `internal/config/config.go` - add GitHubToken
2. `internal/service/github_service.go` - NEW
3. `internal/orchestrator/orchestrator.go` - implement triggerMergeFlow
4. `cmd/server/main.go` - wire up GitHubService

When finished, run: `go build ./cmd/server` to verify it compiles.

When completely finished, run this command to notify me:
openclaw system event --text "Done: GitHub merge flow implemented - PR creation, CI monitoring, auto-merge" --mode now
