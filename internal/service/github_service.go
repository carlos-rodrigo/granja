package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/go-github/v60/github"
)

type GitHubService struct {
	client *github.Client
}

func NewGitHubService(token string) *GitHubService {
	client := github.NewClient(nil).WithAuthToken(token)
	return &GitHubService{client: client}
}

// CreatePR creates a pull request.
func (s *GitHubService) CreatePR(ctx context.Context, owner, repo, head, base, title, body string) (*github.PullRequest, error) {
	pr, _, err := s.client.PullRequests.Create(ctx, owner, repo, &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
		Body:  github.String(body),
	})
	if err != nil {
		return nil, fmt.Errorf("create pull request: %w", err)
	}
	return pr, nil
}

// GetPRStatus checks if CI passed by combining check runs and commit statuses.
func (s *GitHubService) GetPRStatus(ctx context.Context, owner, repo string, prNumber int) (passed bool, pending bool, err error) {
	pr, _, err := s.client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return false, false, fmt.Errorf("get pull request: %w", err)
	}

	sha := pr.GetHead().GetSHA()

	// Get check runs
	checkRuns, _, err := s.client.Checks.ListCheckRunsForRef(ctx, owner, repo, sha, nil)
	if err != nil {
		return false, false, fmt.Errorf("list check runs: %w", err)
	}

	// Get commit statuses
	combinedStatus, _, err := s.client.Repositories.GetCombinedStatus(ctx, owner, repo, sha, nil)
	if err != nil {
		return false, false, fmt.Errorf("get combined status: %w", err)
	}

	// Analyze check runs
	for _, run := range checkRuns.CheckRuns {
		status := run.GetStatus()
		conclusion := run.GetConclusion()

		if status != "completed" {
			return false, true, nil // Still pending
		}
		if conclusion != "success" && conclusion != "skipped" && conclusion != "neutral" {
			return false, false, nil // Failed
		}
	}

	// Analyze commit statuses
	switch combinedStatus.GetState() {
	case "pending":
		return false, true, nil
	case "failure", "error":
		return false, false, nil
	}

	// All checks passed (or no checks configured)
	return true, false, nil
}

// MergePR merges the PR using squash merge.
func (s *GitHubService) MergePR(ctx context.Context, owner, repo string, prNumber int, commitMessage string) error {
	_, _, err := s.client.PullRequests.Merge(ctx, owner, repo, prNumber, commitMessage, &github.PullRequestOptions{
		MergeMethod: "squash",
	})
	if err != nil {
		return fmt.Errorf("merge pull request: %w", err)
	}
	return nil
}

var (
	sshRepoRegex   = regexp.MustCompile(`^git@github\.com:([^/]+)/([^/]+?)(?:\.git)?$`)
	httpsRepoRegex = regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)
)

// ParseRepoURL extracts owner and repo from GitHub URL.
// Supports both SSH (git@github.com:user/repo.git) and HTTPS (https://github.com/user/repo.git) formats.
func ParseRepoURL(repoURL string) (owner, repo string, err error) {
	repoURL = strings.TrimSpace(repoURL)

	if matches := sshRepoRegex.FindStringSubmatch(repoURL); len(matches) == 3 {
		return matches[1], matches[2], nil
	}

	if matches := httpsRepoRegex.FindStringSubmatch(repoURL); len(matches) == 3 {
		return matches[1], matches[2], nil
	}

	return "", "", fmt.Errorf("invalid GitHub repo URL: %s", repoURL)
}
