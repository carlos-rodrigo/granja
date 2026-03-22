# Task: Create Local Granja Skill

Create a skill for OpenClaw agents to publish PRDs to Granja.

## Location: `skills/granja/`

### 1. SKILL.md
- Name: granja
- Description: Publish PRDs and designs to Granja orchestration system
- Trigger: "granja publish", "/publish", "publish to granja"

### 2. Commands

#### `publish` command
Usage: `granja publish --project <name> [--server <url>]`

Flow:
1. Find `.features/{feature}/prd.md` in current directory (or specified path)
2. Optionally find `design.md` in same directory
3. Validate PRD has required sections (title, user stories)
4. POST to Granja API: `POST /api/epics` with {project_id, prd, design}
5. Return epic ID and dashboard URL

Default server: `http://localhost:3000`

### 3. Instructions in SKILL.md
- How to use the skill
- Example commands
- What the agent should do step by step
- Error handling guidance

This skill will be used by OpenClaw agents to submit work to Granja.

Commit your changes when done.
