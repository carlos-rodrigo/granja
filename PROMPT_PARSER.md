# Task: Add Parser and Reviewer Services

Read the existing codebase (especially `design.md`) and add:

## 1. Parser Service (`internal/service/parser_service.go`)
- Calls Claude API (Anthropic) to parse PRD + Design into structured tasks
- Uses function calling / tool_use for structured JSON output
- Input: PRD content, Design content
- Output: Array of tasks with title, description, effort, dependencies, relevant_files
- Should be called when an epic is created (status "planted")

## 2. Reviewer Service (`internal/orchestrator/reviewer.go`)
- Runs when all tasks in an epic are "done" (epic moves to "ready")
- Spawns a review container OR calls Claude directly
- Input: PRD, Design, git diff of all changes
- Output: PASS/FAIL with summary and issues
- If FAIL: creates fix tasks and moves epic back to "growing"
- If PASS: triggers merge flow

## 3. Update Epic Service
- After epic creation, trigger parser
- Add method to check if epic is ready for review
- Add method to handle review results

## 4. Integration
- Wire parser into epic creation flow
- Add reviewer check in orchestrator tick
- Update epic status transitions

Use the Anthropic Go SDK: `github.com/anthropics/anthropic-sdk-go`

Commit your changes when done.
