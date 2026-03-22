# Granja 🌾

AI agent orchestration system that automates the full software development cycle. Submit a PRD + Technical Design, and Granja breaks it into tasks, spins up isolated Docker containers with [Pi coding agents](https://github.com/mariozechner/pi), executes tasks respecting dependencies, reviews completed work, and merges to main.

## How It Works

```
Local Machine                         Granja Server
┌────────────────┐                   ┌─────────────────────────────────┐
│ .features/     │                   │                                 │
│ └─ my-feature/ │    POST /epics    │  Parser (Pi) → Tasks            │
│    ├─ prd.md   │ ───────────────►  │       ↓                         │
│    └─ design.md│                   │  Orchestrator (polling)         │
└────────────────┘                   │       ↓                         │
                                     │  Docker Workers (Pi agents)     │
                                     │       ↓                         │
                                     │  Reviewer (Pi)                  │
                                     │       ↓                         │
                                     │  GitHub PR → CI → Merge         │
                                     └─────────────────────────────────┘
```

### The Flow

1. **Publish** — Developer runs `granja publish --project hippo` from local machine
2. **Parse** — Pi agent reads PRD + Design, creates task files with dependencies
3. **Grow** — Orchestrator spawns Docker workers for ready tasks (respecting DAG)
4. **Execute** — Each worker (Pi agent) implements one task, commits, pushes to epic branch
5. **Review** — When all tasks complete, Pi reviews implementation against PRD
6. **Harvest** — On successful review: create PR, wait for CI, auto-merge to main

### Epic Lifecycle

```
planted → growing → ready → harvested
   │         │        │         │
   │         │        │         └─ PR merged to main
   │         │        └─ All tasks done, review passed
   │         └─ Tasks in progress
   └─ PRD received, parsing into tasks

blocked ← (any stage can fail, creates fix tasks or requires intervention)
```

## Quick Start

### Prerequisites

- Go 1.21+
- Docker
- Node.js 20+ (for Pi agent in workers)
- GitHub token (for PR/merge flow)

### Setup

```bash
# 1. Clone and build
git clone https://github.com/carlos-rodrigo/granja.git
cd granja
go build -o server ./cmd/server

# 2. Build worker Docker image
cd docker && ./build.sh

# 3. Configure and run
export GITHUB_TOKEN="ghp_..."  # Required for merge flow
./server  # Starts on :3000
```

### Create a Project

```bash
curl -X POST http://localhost:3000/api/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-app",
    "repo_url": "git@github.com:user/my-app.git"
  }'
```

### Publish a PRD

From your project directory:

```bash
# Option 1: Using the skill (if installed)
granja publish --project my-app

# Option 2: Direct API call
curl -X POST http://localhost:3000/api/epics \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "proj_xxx",
    "prd": "# User Auth\n\n## User Stories\n\n### US-001\nAs a user...",
    "design": "# Technical Design\n\n## Implementation\n1. Create model..."
  }'
```

### Watch Progress

- **Dashboard**: `http://localhost:3000` — Kanban board with epic/task status
- **API**: `GET /api/epics/{id}` — Epic details with tasks
- **Logs**: `GET /api/workers/{id}/logs` — Real-time worker logs (SSE)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Granja Server                          │
│                                                             │
│  ┌──────────┐  ┌──────────────┐  ┌─────────┐  ┌──────────┐ │
│  │ HTTP API │  │ Orchestrator │  │ Parser  │  │ Reviewer │ │
│  │  (chi)   │  │  (10s poll)  │  │  (Pi)   │  │   (Pi)   │ │
│  └──────────┘  └──────────────┘  └─────────┘  └──────────┘ │
│        │              │                                     │
│        └──────────────┼─────────────────────────────────────┤
│                       │                                     │
│               ┌───────┴───────┐         ┌────────────────┐ │
│               │    SQLite     │         │ GitHub Service │ │
│               │  (projects,   │         │  (PR, CI,      │ │
│               │   epics,      │         │   merge)       │ │
│               │   tasks)      │         └────────────────┘ │
│               └───────────────┘                            │
└─────────────────────────────────────────────────────────────┘
                        │ spawns
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                   Docker Workers                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ granja-worker:latest                                │   │
│  │ - Clones repo, checks out epic branch              │   │
│  │ - Runs: pi --model openai-codex/gpt-5.3 -p "task"  │   │
│  │ - Commits changes, pushes to branch                │   │
│  │ - Reports completion via callback                   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Up to GRANJA_MAX_WORKERS (default: 3) concurrent workers  │
│  Tasks within same epic run sequentially (dependencies)    │
│  Tasks across different epics can run in parallel          │
└─────────────────────────────────────────────────────────────┘
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `GRANJA_ADDR` | `:3000` | Server listen address |
| `GRANJA_DB_PATH` | `granja.db` | SQLite database path |
| `GRANJA_WORKER_IMAGE` | `granja-worker:latest` | Docker image for workers |
| `GRANJA_MAX_WORKERS` | `3` | Max concurrent workers |
| `GRANJA_ORCH_POLL` | `10s` | Orchestrator poll interval |
| `GRANJA_PI_MODEL` | `openai-codex/gpt-5.3` | Model for Pi agents |
| `GRANJA_PI_THINKING` | `high` | Thinking level (low/medium/high) |
| `GRANJA_REVIEW_REPO_PATH` | `.` | Path to repo for review diffs |
| `GITHUB_TOKEN` | — | GitHub token for PR/merge (required) |

## API Reference

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/health` | Health check |
| `POST` | `/api/projects` | Create project |
| `GET` | `/api/projects` | List projects |
| `GET` | `/api/projects/:id` | Get project |
| `POST` | `/api/epics` | Create epic from PRD |
| `GET` | `/api/epics` | List epics (filter: `?project=&status=`) |
| `GET` | `/api/epics/:id` | Get epic with tasks |
| `DELETE` | `/api/epics/:id` | Cancel/delete epic |
| `GET` | `/api/tasks/:id` | Get task details |
| `POST` | `/api/tasks/:id/complete` | Worker reports success |
| `POST` | `/api/tasks/:id/fail` | Worker reports failure |
| `GET` | `/api/workers` | List active workers |
| `GET` | `/api/workers/:id/logs` | Stream worker logs (SSE) |

## Project Structure

```
granja/
├── cmd/server/              # Entry point
├── internal/
│   ├── api/                 # HTTP handlers + router
│   │   ├── handler/         # Project, Epic, Task, Worker handlers
│   │   └── middleware/      # Logging
│   ├── config/              # Environment config
│   ├── domain/              # Domain models (Epic, Task, Worker)
│   ├── orchestrator/        # Task scheduling + reviewer
│   ├── repository/          # SQLite repositories
│   └── service/             # Parser, Docker, GitHub services
├── docker/
│   ├── worker/              # Worker Dockerfile + entrypoint
│   ├── docker-compose.yml
│   └── build.sh
├── migrations/              # SQLite schema
├── web/                     # Next.js Kanban dashboard
├── skills/granja/           # Local publish skill
├── prd.md                   # Product requirements
└── design.md                # Technical design
```

## Dashboard

Next.js Kanban board showing epics across four columns:

- **Planted** — PRD received, parsing into tasks
- **Growing** — Tasks in progress
- **Ready** — All tasks done, under review
- **Harvested** — Merged to main

```bash
cd web
pnpm install
pnpm dev  # http://localhost:3001
```

Features:
- Real-time updates (5s polling)
- Click epic to expand tasks
- Click in-progress task to view live logs
- Auto-scroll log viewer with ANSI color support

## Key Decisions

1. **Go for backend** — Better concurrency, native Docker SDK
2. **SQLite over Postgres** — Single-user, simpler deployment
3. **Polling over events** — Simpler, 10s latency acceptable
4. **Git worktrees** — Parallel work on same repo without conflicts
5. **Pi as coding agent** — Model-agnostic, works well in containers

## Development

```bash
# Run with live reload
go run ./cmd/server

# Build binary
go build -o server ./cmd/server

# Run tests
go test ./...

# Build worker image
cd docker && ./build.sh
```

## License

MIT
