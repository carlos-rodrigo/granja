package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"granja/internal/domain"
	"granja/internal/service"
)

type Reviewer struct {
	piModel    string
	piThinking string
	repoDir    string
}

func NewReviewer(piModel, piThinking, repoDir string) *Reviewer {
	if piModel == "" {
		piModel = "openai-codex/gpt-5.3"
	}
	if piThinking == "" {
		piThinking = "high"
	}
	if strings.TrimSpace(repoDir) == "" {
		repoDir = "."
	}
	return &Reviewer{
		piModel:    piModel,
		piThinking: piThinking,
		repoDir:    repoDir,
	}
}

func (r *Reviewer) ReviewEpic(ctx context.Context, epic *domain.Epic, project *domain.Project) (*service.EpicReviewResult, error) {
	if epic == nil || project == nil {
		return nil, errors.New("epic and project are required")
	}

	diff, err := r.gitDiff(ctx, project, epic)
	if err != nil {
		return nil, err
	}

	// Create temp directory for review
	tmpDir, err := os.MkdirTemp("", "granja-review-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write context files
	if err := os.WriteFile(filepath.Join(tmpDir, "prd.md"), []byte(epic.PRDContent), 0644); err != nil {
		return nil, fmt.Errorf("write prd: %w", err)
	}
	if epic.DesignContent != "" {
		if err := os.WriteFile(filepath.Join(tmpDir, "design.md"), []byte(epic.DesignContent), 0644); err != nil {
			return nil, fmt.Errorf("write design: %w", err)
		}
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "diff.patch"), []byte(diff), 0644); err != nil {
		return nil, fmt.Errorf("write diff: %w", err)
	}

	// Build prompt for Pi
	prompt := `You are a code reviewer. Review the implementation against the PRD and design.

Read:
- prd.md - The product requirements
- design.md - The technical design (if exists)
- diff.patch - The code changes to review

Analyze whether the implementation correctly fulfills all requirements in the PRD.

After reviewing, create a file called review.json with this exact format:

{
  "result": "PASS" or "FAIL",
  "summary": "Brief summary of your review",
  "issues": [
    {"title": "Issue title", "description": "Description of the problem"}
  ]
}

Rules:
- Return "PASS" only if ALL user stories are correctly implemented
- Return "FAIL" if any required behavior is missing or broken
- List specific issues if failing
- Be strict but fair

Read the files and create review.json.`

	// Run Pi
	cmd := exec.CommandContext(ctx, "pi",
		"--print",
		"--dangerously-skip-permissions",
		"--model", r.piModel,
		"--thinking", r.piThinking,
		"-p", prompt)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "HOME="+os.Getenv("HOME"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pi review failed: %w\noutput: %s", err, string(output))
	}

	// Read review.json
	reviewPath := filepath.Join(tmpDir, "review.json")
	reviewData, err := os.ReadFile(reviewPath)
	if err != nil {
		return nil, fmt.Errorf("read review.json: %w (pi may not have created it)", err)
	}

	var result struct {
		Result  string                    `json:"result"`
		Summary string                    `json:"summary"`
		Issues  []service.EpicReviewIssue `json:"issues"`
	}
	if err := json.Unmarshal(reviewData, &result); err != nil {
		return nil, fmt.Errorf("parse review.json: %w", err)
	}

	return &service.EpicReviewResult{
		Result:  strings.ToUpper(strings.TrimSpace(result.Result)),
		Summary: strings.TrimSpace(result.Summary),
		Issues:  result.Issues,
	}, nil
}

func (r *Reviewer) gitDiff(ctx context.Context, project *domain.Project, epic *domain.Epic) (string, error) {
	repoDir := r.repoDir
	if project != nil && project.RepoURL != "" {
		if st, err := os.Stat(project.RepoURL); err == nil && st.IsDir() {
			repoDir = project.RepoURL
		}
	}
	base := project.DefaultBranch
	if strings.TrimSpace(base) == "" {
		base = "main"
	}
	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "diff", fmt.Sprintf("%s...%s", base, epic.BranchName))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w: %s", err, string(out))
	}
	if strings.TrimSpace(string(out)) == "" {
		return "(empty diff)", nil
	}
	return string(out), nil
}
