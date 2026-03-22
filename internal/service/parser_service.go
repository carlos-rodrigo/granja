package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type ParsedTask struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Effort        string   `json:"effort"`
	Dependencies  []string `json:"dependencies"`
	RelevantFiles []string `json:"relevant_files"`
}

type ParserService struct {
	piModel    string
	piThinking string
}

func NewParserService(model string) *ParserService {
	if model == "" {
		model = "openai-codex/gpt-5.3"
	}
	return &ParserService{
		piModel:    model,
		piThinking: "high",
	}
}

func (s *ParserService) ParseTasks(ctx context.Context, prd, design string) ([]ParsedTask, error) {
	if strings.TrimSpace(prd) == "" {
		return nil, errors.New("prd is required")
	}

	// Create temp directory for the feature
	tmpDir, err := os.MkdirTemp("", "granja-parse-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .features/parse/tasks structure
	featureDir := filepath.Join(tmpDir, ".features", "parse")
	tasksDir := filepath.Join(featureDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return nil, fmt.Errorf("create tasks dir: %w", err)
	}

	// Write PRD and design
	if err := os.WriteFile(filepath.Join(featureDir, "prd.md"), []byte(prd), 0644); err != nil {
		return nil, fmt.Errorf("write prd: %w", err)
	}
	if design != "" {
		if err := os.WriteFile(filepath.Join(featureDir, "design.md"), []byte(design), 0644); err != nil {
			return nil, fmt.Errorf("write design: %w", err)
		}
	}

	// Build prompt for Pi
	prompt := fmt.Sprintf(`You are a task planner. Read the PRD and design, then create task files.

Read:
- .features/parse/prd.md
- .features/parse/design.md (if exists)

Create task files in .features/parse/tasks/ following this format:

Each task file: NNN-kebab-name.md with YAML frontmatter:

---
id: NNN
status: open
depends: []        # IDs of tasks that must complete first (as strings like "001")
---

# Task Title

Description of what to do.

## What to do

- Step 1
- Step 2

## Files

- relevant/file/paths

## Verify

` + "```bash" + `
# command to verify task is complete
` + "```" + `

Rules:
- Create 3-10 small, focused tasks
- Use zero-padded IDs: 001, 002, 003
- Set depends correctly for task ordering
- Keep tasks small enough for one coding iteration
- Include verify commands when possible

Start by reading the PRD, then create the task files.`)

	// Run Pi
	cmd := exec.CommandContext(ctx, "pi",
		"--model", s.piModel,
		"--thinking", s.piThinking,
		"-p", prompt)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "HOME="+os.Getenv("HOME"))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("pi failed: %w\noutput: %s", err, string(output))
	}

	// Read generated task files
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil, fmt.Errorf("read tasks dir: %w", err)
	}

	var tasks []ParsedTask
	taskIDPattern := regexp.MustCompile(`^(\d+)-.*\.md$`)

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), "_") {
			continue
		}
		if !taskIDPattern.MatchString(entry.Name()) {
			continue
		}

		taskPath := filepath.Join(tasksDir, entry.Name())
		task, err := parseTaskFile(taskPath)
		if err != nil {
			continue // Skip malformed tasks
		}
		tasks = append(tasks, task)
	}

	if len(tasks) == 0 {
		return nil, errors.New("pi did not create any task files")
	}

	return tasks, nil
}

func parseTaskFile(path string) (ParsedTask, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return ParsedTask{}, err
	}

	lines := strings.Split(string(content), "\n")
	
	var task ParsedTask
	var inFrontmatter bool
	var frontmatterLines []string
	var bodyLines []string
	var frontmatterDone bool

	for _, line := range lines {
		if line == "---" {
			if !inFrontmatter && !frontmatterDone {
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				inFrontmatter = false
				frontmatterDone = true
				continue
			}
		}
		if inFrontmatter {
			frontmatterLines = append(frontmatterLines, line)
		} else if frontmatterDone {
			bodyLines = append(bodyLines, line)
		}
	}

	// Parse frontmatter
	for _, line := range frontmatterLines {
		if strings.HasPrefix(line, "depends:") {
			// Parse depends array
			depsStr := strings.TrimPrefix(line, "depends:")
			depsStr = strings.TrimSpace(depsStr)
			if depsStr != "[]" && depsStr != "" {
				// Simple parsing for [001, 002] or ["001", "002"]
				depsStr = strings.Trim(depsStr, "[]")
				if depsStr != "" {
					for _, dep := range strings.Split(depsStr, ",") {
						dep = strings.TrimSpace(dep)
						dep = strings.Trim(dep, `"'`)
						if dep != "" {
							task.Dependencies = append(task.Dependencies, dep)
						}
					}
				}
			}
		}
	}

	// Extract title from first # heading
	for _, line := range bodyLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			task.Title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Build description from body
	task.Description = strings.TrimSpace(strings.Join(bodyLines, "\n"))

	// Extract files section
	inFilesSection := false
	for _, line := range bodyLines {
		if strings.HasPrefix(line, "## Files") {
			inFilesSection = true
			continue
		}
		if strings.HasPrefix(line, "## ") && inFilesSection {
			break
		}
		if inFilesSection && strings.HasPrefix(line, "- ") {
			file := strings.TrimPrefix(line, "- ")
			file = strings.TrimSpace(file)
			if file != "" {
				task.RelevantFiles = append(task.RelevantFiles, file)
			}
		}
	}

	// Estimate effort based on description length and steps
	stepCount := strings.Count(task.Description, "- ")
	if stepCount <= 3 {
		task.Effort = "small"
	} else if stepCount <= 6 {
		task.Effort = "medium"
	} else {
		task.Effort = "large"
	}

	if task.Title == "" {
		return ParsedTask{}, errors.New("task has no title")
	}

	return task, nil
}

// ParseTaskID extracts the numeric ID from a task filename like "001-setup.md"
func ParseTaskID(filename string) (int, error) {
	pattern := regexp.MustCompile(`^(\d+)-`)
	matches := pattern.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return 0, errors.New("invalid task filename")
	}
	return strconv.Atoi(matches[1])
}
