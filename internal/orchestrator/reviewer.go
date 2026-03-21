package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"

	"granja/internal/domain"
	"granja/internal/service"
)

type Reviewer struct {
	client  anthropic.Client
	model   anthropic.Model
	repoDir string
}

func NewReviewer(model anthropic.Model, repoDir string) *Reviewer {
	if model == "" {
		model = anthropic.ModelClaudeSonnet4_5
	}
	if strings.TrimSpace(repoDir) == "" {
		repoDir = "."
	}
	return &Reviewer{
		client:  anthropic.NewClient(),
		model:   model,
		repoDir: repoDir,
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

	prompt := fmt.Sprintf("PRD:\n%s\n\nDESIGN:\n%s\n\nDIFF:\n%s", epic.PRDContent, epic.DesignContent, diff)
	resp, err := r.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     r.model,
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{{
			Text: "Review the implementation strictly against PRD and design. Always call emit_review exactly once. Return FAIL when any required behavior is missing or broken.",
		}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Tools: []anthropic.ToolUnionParam{
			anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"result":  map[string]any{"type": "string", "enum": []string{"PASS", "FAIL"}},
					"summary": map[string]any{"type": "string"},
					"issues": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"title":       map[string]any{"type": "string"},
								"description": map[string]any{"type": "string"},
							},
							"required":             []string{"title", "description"},
							"additionalProperties": false,
						},
					},
				},
				Required: []string{"result", "summary", "issues"},
			}, "emit_review"),
		},
		ToolChoice: anthropic.ToolChoiceParamOfTool("emit_review"),
	})
	if err != nil {
		return nil, err
	}

	for _, block := range resp.Content {
		if block.Type != "tool_use" || block.Name != "emit_review" {
			continue
		}
		var out struct {
			Result  string                    `json:"result"`
			Summary string                    `json:"summary"`
			Issues  []service.EpicReviewIssue `json:"issues"`
		}
		if err := json.Unmarshal(block.Input, &out); err != nil {
			return nil, fmt.Errorf("decode review output: %w", err)
		}
		return &service.EpicReviewResult{
			Result:  strings.ToUpper(strings.TrimSpace(out.Result)),
			Summary: strings.TrimSpace(out.Summary),
			Issues:  out.Issues,
		}, nil
	}

	return nil, errors.New("reviewer did not return tool_use output")
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
