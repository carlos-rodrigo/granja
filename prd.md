# PRD: Granja

## Introduction

Granja is an AI agent orchestration system that automates software development workflows. You submit a PRD + Technical Design, and Granja breaks it into tasks, spins up isolated Docker containers with coding agents (Pi), executes tasks respecting dependencies, and automatically reviews completed work before merging.

The system runs on a dedicated server, with a local skill to publish work from your development machine.

## Goals

- Receive PRDs and Technical Designs from a local skill
- Parse designs into executable tasks with dependency tracking
- Display projects, epics, and tasks in a Kanban web UI
- Spawn isolated Docker containers for each worker
- Run Pi coding agent with user's config inside containers
- Execute tasks sequentially within an epic (respecting dependencies)
- Parallelize across independent epics
- Automatically review completed epics against PRD requirements
- Merge to main on successful review, or create fix tasks on failure
- Provide real-time visibility into worker status and progress

## User Stories

### US-001: Publish PRD and Design from local machine

**Description:** As a developer, I want to publish a PRD and Technical Design to Granja so that work can begin.

**BDD Spec:**
- Given: I have a PRD and Design in `.features/{feature}/`
- When: I run `granja publish --project hippo`
- Then: The epic appears in Granja's web UI with status "planted"

**Acceptance Criteria:**

- [ ] Skill reads `prd.md` and `design.md` from `.features/{feature}/`
- [ ] Sends both files to Granja API via HTTP POST
- [ ] API validates required fields (title, user stories)
- [ ] Returns epic ID and URL to view in dashboard
- [ ] Epic appears in Kanban with "Planted" status

**Feedback Loop:**

Setup: Granja server running on `http://granja.local:3000`

Verification:
1. Create `.features/test-epic/prd.md` and `design.md` with valid content
2. Run skill: `granja publish --project test-project` → returns epic ID
3. `curl http://granja.local:3000/api/epics/{id}` → returns epic with status "planted"
4. Open browser → epic visible in Kanban under "Planted" column

Edge cases:
- Missing `design.md` → error with clear message
- Invalid PRD (no user stories) → 400 error with validation details
- Server unreachable → timeout with retry suggestion

---

### US-002: Parse design into tasks

**Description:** As a system, I want to automatically break a Technical Design into tasks so that workers can execute them.

**BDD Spec:**
- Given: An epic with PRD and Design is created
- When: The parser processes it
- Then: Individual tasks are created with dependencies

**Acceptance Criteria:**

- [ ] Parser extracts user stories from PRD
- [ ] Parser identifies implementation steps from Design
- [ ] Each task has: title, description, estimated effort, relevant files, dependencies
- [ ] Dependencies are inferred from Design's "Implementation Plan" section
- [ ] Tasks are created in "todo" status
- [ ] Epic moves from "planted" to "growing"

**Feedback Loop:**

Setup: Granja server running, epic published

Verification:
1. Publish epic with 3 user stories and clear implementation plan
2. Wait for parser (webhook or poll) → epic status changes to "growing"
3. `curl http://granja.local:3000/api/epics/{id}/tasks` → returns array of tasks
4. Verify tasks have correct dependencies (task 2 depends on task 1, etc.)

Edge cases:
- Design with no clear implementation plan → parser requests clarification or creates single task per story
- Circular dependencies detected → error, epic stays "planted"

---

### US-003: Display Kanban board

**Description:** As a developer, I want to see all epics and tasks in a Kanban board so I can track progress.

**BDD Spec:**
- Given: Multiple epics exist with various statuses
- When: I open the dashboard
- Then: I see columns: Planted | Growing | Ready | Harvested

**Acceptance Criteria:**

- [ ] Web UI shows Kanban with 4 columns
- [ ] Each epic card shows: title, project, task progress (3/5 done)
- [ ] Click epic → expands to show individual tasks
- [ ] Tasks show their own status: todo | in_progress | done | review
- [ ] PRD and Design viewable from epic detail
- [ ] Real-time updates when status changes (polling every 5s or WebSocket)

**Feedback Loop:**

Setup: Granja server with seeded data (multiple epics/tasks)

Verification:
1. Open `http://granja.local:3000` → Kanban renders with columns
2. Verify epic cards show correct project and progress
3. Click an epic → task list expands inline
4. In another terminal, update a task status via API → UI reflects change within 5s

Edge cases:
- No epics → empty state with "No work yet" message
- 50+ epics → pagination or virtual scroll
- Very long epic title → truncated with tooltip

---

### US-004: Spawn worker container for task

**Description:** As an orchestrator, I want to spawn a Docker container when a task is ready so that work can begin.

**BDD Spec:**
- Given: A task is in "todo" status with all dependencies "done"
- When: The orchestrator detects it
- Then: A Docker container is spawned with the coding agent

**Acceptance Criteria:**

- [ ] Orchestrator polls for ready tasks every 10s
- [ ] Worker image includes: git, Pi agent, user's Pi config
- [ ] Container clones the project repo
- [ ] Creates git worktree for the epic branch
- [ ] Task moves to "in_progress"
- [ ] Worker starts Pi with task description as prompt

**Feedback Loop:**

Setup: Granja server, Docker daemon running, task with no dependencies

Verification:
1. Create epic with single task (no deps)
2. Wait for orchestrator → `docker ps` shows new container
3. Check task status → "in_progress"
4. Container logs show Pi starting with task prompt

Edge cases:
- Docker daemon not running → error logged, task stays "todo", alert sent
- Repo clone fails (auth) → container exits with error, task marked "blocked"
- All workers busy (max containers reached) → task queued, dashboard shows "waiting"

---

### US-005: Execute task with Pi agent

**Description:** As a worker, I want to run Pi agent to implement the task and commit the result.

**BDD Spec:**
- Given: A container is running with task context
- When: Pi completes the task
- Then: Changes are committed and pushed to the epic branch

**Acceptance Criteria:**

- [ ] Pi runs with `--print` mode (non-interactive)
- [ ] Task description includes: user story, acceptance criteria, relevant files from design
- [ ] Pi has access to full repo in worktree
- [ ] On completion, Pi commits with message: `feat(epic): task title`
- [ ] Pushes to epic branch
- [ ] Reports completion to orchestrator API
- [ ] Container terminates
- [ ] Task moves to "done"

**Feedback Loop:**

Setup: Worker container running with valid task

Verification:
1. Watch container logs → Pi working on task
2. On completion, check git log on epic branch → new commit exists
3. `curl http://granja.local:3000/api/tasks/{id}` → status "done"
4. `docker ps` → container no longer running

Edge cases:
- Pi fails (syntax error, test failure) → task marked "blocked", logs captured
- Pi timeout (>30 min) → container killed, task "blocked"
- Partial commit (Pi crashed mid-work) → rollback, task "blocked"

---

### US-006: Parallelize independent epics

**Description:** As an orchestrator, I want to run multiple epics in parallel when they don't depend on each other.

**BDD Spec:**
- Given: Epic A and Epic B exist with no inter-dependencies
- When: Both have tasks ready
- Then: Two containers run simultaneously

**Acceptance Criteria:**

- [ ] Orchestrator detects independent epics
- [ ] Spawns up to N containers (configurable, default 3)
- [ ] Each container works on a different epic
- [ ] Same epic tasks run sequentially (respecting dependencies)
- [ ] Dashboard shows multiple "in_progress" tasks across epics

**Feedback Loop:**

Setup: Two epics with independent tasks

Verification:
1. Publish Epic A (2 tasks) and Epic B (2 tasks) to different projects
2. Wait for orchestrator → `docker ps` shows 2 containers
3. Dashboard shows Task A1 and Task B1 both "in_progress"
4. Task A1 completes → Task A2 starts, Task B1 still running

Edge cases:
- Max containers reached → new tasks queue in priority order
- One epic fails → other epic continues unaffected
- Same repo, different epics → worktrees prevent conflicts

---

### US-007: Review completed epic

**Description:** As a system, I want to automatically review a completed epic against its PRD to ensure quality.

**BDD Spec:**
- Given: All tasks in an epic are "done"
- When: The reviewer runs
- Then: Code is validated against PRD requirements

**Acceptance Criteria:**

- [ ] Epic moves to "ready" status when all tasks complete
- [ ] Orchestrator spawns reviewer container
- [ ] Reviewer has: PRD, Design, diff of all changes
- [ ] Pi runs in review mode with prompt: "Review this code against the PRD. Check each user story."
- [ ] Reviewer outputs: PASS (with summary) or FAIL (with specific issues)
- [ ] Results stored in epic metadata

**Feedback Loop:**

Setup: Epic with all tasks "done"

Verification:
1. Complete all tasks in an epic
2. Epic status → "ready"
3. `docker ps` → reviewer container running
4. Wait for completion → check epic review field populated
5. If PASS: epic shows "harvested"
6. If FAIL: new tasks created from issues

Edge cases:
- Review timeout → marked "needs_manual_review"
- Ambiguous PRD → reviewer asks for clarification (creates task)

---

### US-008: Create fix tasks on failed review

**Description:** As a system, I want to create new tasks when a review fails so issues are addressed.

**BDD Spec:**
- Given: A review identified issues
- When: The review completes with FAIL
- Then: New tasks are created for each issue

**Acceptance Criteria:**

- [ ] Each review issue becomes a task
- [ ] Task has: issue description, suggested fix, relevant files
- [ ] Tasks depend on all previously completed tasks
- [ ] Epic moves back to "growing"
- [ ] Fix tasks appear in dashboard immediately

**Feedback Loop:**

Setup: Epic ready for review, code has intentional issues

Verification:
1. Submit epic where implementation doesn't match PRD
2. Review runs → identifies 2 issues
3. Check tasks → 2 new tasks created with "todo" status
4. Epic status back to "growing"
5. New tasks have dependencies on original tasks

Edge cases:
- Review finds 10+ issues → consolidate into reasonable task count
- Same issue found multiple times → deduplicate

---

### US-009: Merge to main on successful review

**Description:** As a system, I want to merge the epic branch to main when review passes.

**BDD Spec:**
- Given: Review passed
- When: The merge runs
- Then: Epic branch is merged and epic marked "harvested"

**Acceptance Criteria:**

- [ ] Creates PR from epic branch to main
- [ ] PR description includes: epic title, PRD summary, tasks completed
- [ ] If CI passes, auto-merge
- [ ] If CI fails, mark epic "blocked" with CI logs
- [ ] On successful merge, epic moves to "harvested"
- [ ] Notification sent (webhook or dashboard)

**Feedback Loop:**

Setup: Epic with passed review

Verification:
1. Review completes with PASS
2. Check GitHub → PR created
3. CI runs (if configured)
4. PR merged automatically
5. Epic status → "harvested"
6. Epic branch deleted (cleanup)

Edge cases:
- Merge conflict → mark "blocked", create "resolve conflict" task
- CI fails → mark "blocked", show logs in dashboard
- No CI configured → merge directly

---

### US-010: View worker logs and status

**Description:** As a developer, I want to see what workers are doing in real-time so I can monitor progress.

**BDD Spec:**
- Given: A worker container is running
- When: I view its details in dashboard
- Then: I see live logs and status

**Acceptance Criteria:**

- [ ] Dashboard shows active workers with current task
- [ ] Click worker → shows live log stream
- [ ] Logs show Pi's output (what it's reading, editing, running)
- [ ] Status indicators: starting | working | committing | done | error
- [ ] Historical logs kept for completed tasks

**Feedback Loop:**

Setup: Worker actively processing a task

Verification:
1. Open dashboard while container running
2. Click on active worker → log panel opens
3. See real-time Pi output
4. Worker completes → logs preserved
5. Status updates reflect Pi progress

Edge cases:
- Container crashes → last 1000 log lines preserved
- Very verbose logs → truncate with "show more" option
- Multiple viewers → all see same stream

---

## Functional Requirements

- FR-1: Local skill (`granja`) publishes PRD + Design to server via HTTP
- FR-2: Server parses designs into tasks with AI (Claude Haiku)
- FR-3: Tasks have status: todo | in_progress | done | blocked | review
- FR-4: Epics have status: planted | growing | ready | harvested | blocked
- FR-5: Web UI displays Kanban board with real-time updates
- FR-6: Orchestrator spawns Docker containers for workers
- FR-7: Workers run Pi agent with user's config from mounted volume
- FR-8: Workers use git worktrees for parallel work on same repo
- FR-9: Maximum concurrent containers is configurable (default: 3)
- FR-10: Reviewer validates code against PRD requirements
- FR-11: Failed reviews create fix tasks automatically
- FR-12: Successful reviews trigger merge to main
- FR-13: All container logs are captured and viewable
- FR-14: Projects are isolated (different repos, different queues)

## Non-Goals

- No multi-user authentication (single-user system for now)
- No billing or usage tracking
- No support for non-Pi agents (Pi only for MVP)
- No distributed Docker (single server only)
- No manual task creation in UI (all tasks from parser)
- No editing PRD/Design after submission (create new epic instead)

## Technical Considerations

### Architecture

```
┌────────────────┐     HTTP      ┌─────────────────────────────────────────┐
│  Local Machine │ ───────────►  │              GRANJA SERVER              │
│  (granja skill)│               │                                         │
└────────────────┘               │  ┌─────────────┐   ┌─────────────────┐  │
                                 │  │  Next.js    │   │  Go Backend     │  │
                                 │  │  Frontend   │   │  (API + Orch)   │  │
                                 │  └──────┬──────┘   └────────┬────────┘  │
                                 │         │                   │           │
                                 │         └─────────┬─────────┘           │
                                 │                   │                     │
                                 │              ┌────┴────┐                │
                                 │              │ SQLite  │                │
                                 │              └────┬────┘                │
                                 │                   │                     │
                                 │         ┌─────────┴─────────┐           │
                                 │         │  Docker Daemon    │           │
                                 │         │  ┌─────┐ ┌─────┐  │           │
                                 │         │  │ W1  │ │ W2  │  │           │
                                 │         │  └─────┘ └─────┘  │           │
                                 │         └───────────────────┘           │
                                 └─────────────────────────────────────────┘
```

### Worker Docker Image

```dockerfile
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y git curl
RUN curl -fsSL https://get.pi.dev | bash
COPY pi-config/ /root/.pi/agent/
VOLUME /workspace
ENTRYPOINT ["/entrypoint.sh"]
```

### API Endpoints

- `POST /api/epics` - Create epic from PRD + Design
- `GET /api/epics` - List epics (filterable by project, status)
- `GET /api/epics/:id` - Get epic with tasks
- `GET /api/epics/:id/tasks` - List tasks for epic
- `PATCH /api/tasks/:id` - Update task status
- `GET /api/workers` - List active workers
- `GET /api/workers/:id/logs` - Stream worker logs
- `GET /api/projects` - List projects

### Database Schema (SQLite)

```sql
-- Projects
CREATE TABLE projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  repo_url TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Epics
CREATE TABLE epics (
  id TEXT PRIMARY KEY,
  project_id TEXT REFERENCES projects(id),
  title TEXT NOT NULL,
  status TEXT DEFAULT 'planted',
  prd_content TEXT,
  design_content TEXT,
  branch_name TEXT,
  review_result TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tasks
CREATE TABLE tasks (
  id TEXT PRIMARY KEY,
  epic_id TEXT REFERENCES epics(id),
  title TEXT NOT NULL,
  description TEXT,
  status TEXT DEFAULT 'todo',
  effort TEXT,
  relevant_files TEXT, -- JSON array
  container_id TEXT,
  started_at DATETIME,
  completed_at DATETIME,
  logs TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Task dependencies
CREATE TABLE task_dependencies (
  task_id TEXT REFERENCES tasks(id),
  depends_on TEXT REFERENCES tasks(id),
  PRIMARY KEY (task_id, depends_on)
);
```

## Success Metrics

- Epic from submission to harvest in <4 hours (for typical 5-task epic)
- 90% of reviews pass on first attempt
- Zero merge conflicts (worktree isolation)
- Dashboard loads in <1s
- Worker logs visible within 2s of generation

## Open Questions

- Should we support task priorities within an epic?
- How to handle secrets/env vars for different projects?
- Should there be a "pause" feature for epics?
- Notification preferences (email, webhook, Discord)?
