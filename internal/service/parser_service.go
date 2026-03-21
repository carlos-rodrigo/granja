package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

type ParsedTask struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Effort        string   `json:"effort"`
	Dependencies  []string `json:"dependencies"`
	RelevantFiles []string `json:"relevant_files"`
}

type ParserService struct {
	client anthropic.Client
	model  anthropic.Model
}

func NewParserService(model anthropic.Model) *ParserService {
	if model == "" {
		model = anthropic.ModelClaudeSonnet4_5
	}
	return &ParserService{
		client: anthropic.NewClient(),
		model:  model,
	}
}

func (s *ParserService) ParseTasks(ctx context.Context, prd, design string) ([]ParsedTask, error) {
	if strings.TrimSpace(prd) == "" {
		return nil, errors.New("prd is required")
	}

	prompt := fmt.Sprintf("PRD:\n%s\n\nDESIGN:\n%s", prd, design)
	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     s.model,
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{{
			Text: "Extract an implementation task DAG from PRD + design. Always call the emit_tasks tool exactly once. Keep tasks actionable, concrete, and small enough for one coding agent iteration.",
		}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
		Tools: []anthropic.ToolUnionParam{
			anthropic.ToolUnionParamOfTool(anthropic.ToolInputSchemaParam{
				Properties: map[string]any{
					"tasks": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"title":          map[string]any{"type": "string"},
								"description":    map[string]any{"type": "string"},
								"effort":         map[string]any{"type": "string", "enum": []string{"small", "medium", "large"}},
								"dependencies":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
								"relevant_files": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
							},
							"required":             []string{"title", "description", "effort", "dependencies", "relevant_files"},
							"additionalProperties": false,
						},
					},
				},
				Required: []string{"tasks"},
			}, "emit_tasks"),
		},
		ToolChoice: anthropic.ToolChoiceParamOfTool("emit_tasks"),
	})
	if err != nil {
		return nil, err
	}

	type toolPayload struct {
		Tasks []ParsedTask `json:"tasks"`
	}

	for _, block := range resp.Content {
		if block.Type != "tool_use" || block.Name != "emit_tasks" {
			continue
		}
		var payload toolPayload
		if err := json.Unmarshal(block.Input, &payload); err != nil {
			return nil, fmt.Errorf("decode parser tool output: %w", err)
		}
		if len(payload.Tasks) == 0 {
			return nil, errors.New("parser returned no tasks")
		}
		return payload.Tasks, nil
	}

	return nil, errors.New("parser did not return tool_use output")
}
