# Granja 🌾

AI agent orchestration system for automated software development. Submit a PRD, get working code.

## How It Works

```
PRD + Design → Parser (Pi) → Tasks → Workers (Pi in Docker) → Review (Pi) → PR → Merge
```

1. **Publish** a PRD and Technical Design
2. **Parser** (Pi agent) breaks it into executable tasks
3. **Orchestrator** spawns Docker workers for each task
4. **Workers** (Pi agent) implement tasks, commit, push
5. **Reviewer** (Pi agent) validates against PRD
6. **Merger** creates PR, monitors CI, auto-merges

## Quick Start

```bash
# 1. Build worker image
cd docker && ./build.sh

# 2. Start server
export GITHUB_TOKEN="ghp_..."  # For PR/merge flow
./server

# 3. Create a project
curl -X POST http://localhost:3000/api/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"my-project","repo_url":"git@github.com:user/repo.git"}'

# 4. Publish a PRD
curl -X POST http://localhost:3000/api/epics \
  -H "Content-Type: application/json" \
  -d '{"project_id":"proj_xxx","prd":"# Feature\n\n## User Stories\n...","design":"..."}'
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Granja Server                          │
│  ┌──────────┐  ┌──────────────┐  ┌─────────┐  ┌──────────┐ │
│  │ HTTP API │  │ Orchestrator │  │ Parser  │  │ Reviewer │ │
│  │  (chi)   │  │  (polling)   │  │  (Pi)   │  │   (Pi)   │ │
│  └──────────┘  └──────────────┘  └─────────┘  └──────────┘ │
│                       │                                     │
│               ┌───────┴───────┐                            │
│               │    SQLite     │                            │
│               └───────────────┘                            │
└─────────────────────────────────────────────────────────────┘
                        │ spawns
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                   Docker Workers                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ Worker 1    │  │ Worker 2    │  │ Worker 3    │        │
│  │ (Pi agent)  │  │ (Pi agent)  │  │ (Pi agent)  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## Epic Lifecycle

```
planted → growing → ready → harvested
   │         │        │
   │         │        └─ Review passed, PR merged
   │         └─ Tasks in progress
   └─ PRD received, parsing tasks

blocked ← (any stage can fail)
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
| `GRANJA_PI_THINKING` | `high` | Thinking level for Pi |
| `GITHUB_TOKEN` | (required for merge) | GitHub token for PR/merge |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/projects` | Create project |
| `GET` | `/api/projects` | List projects |
| `POST` | `/api/epics` | Create epic from PRD |
| `GET` | `/api/epics` | List epics |
| `GET` | `/api/epics/:id` | Get epic with tasks |
| `GET` | `/api/workers` | List active workers |
| `GET` | `/api/workers/:id/logs` | Stream worker logs (SSE) |
| `GET` | `/api/health` | Health check |

## Project Structure

```
granja/
├── cmd/server/          # Entry point
├── internal/
│   ├── api/             # HTTP handlers
│   ├── config/          # Configuration
│   ├── domain/          # Domain models
│   ├── orchestrator/    # Task scheduling, review
│   ├── repository/      # SQLite repositories
│   └── service/         # Business logic
├── docker/
│   ├── worker/          # Worker Dockerfile
│   └── docker-compose.yml
├── migrations/          # SQL migrations
├── web/                 # Next.js dashboard
└── skills/granja/       # OpenClaw skill
```

## Dashboard

The Next.js dashboard provides a Kanban view of epics and tasks.

```bash
cd web
pnpm install
pnpm dev  # http://localhost:3001
```

## Worker Image

Workers run Pi coding agent in isolated Docker containers:

```bash
cd docker
./build.sh  # Creates granja-worker:latest
```

Each worker:
- Clones the project repo
- Checks out the epic branch
- Runs Pi with the task prompt
- Commits and pushes changes
- Reports completion to orchestrator

## Local Skill

Publish PRDs from your local machine using the granja skill:

```bash
# From your project directory with .features/my-feature/prd.md
granja publish --project my-project --server http://localhost:3000
```

## Development

```bash
# Run server with hot reload
go run ./cmd/server

# Build
go build -o server ./cmd/server

# Run tests
go test ./...
```

## License

MIT
