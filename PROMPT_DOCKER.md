# Task: Create Docker Worker Image

Read `design.md` for the Docker worker specs. Create in `docker/worker/` directory.

## Requirements:

### 1. Dockerfile (`docker/worker/Dockerfile`)
- Base: ubuntu:22.04
- Install: git, curl, ca-certificates, nodejs (for pi)
- Install Pi coding agent: `npm install -g @anthropics/claude-code` or similar
- Create /workspace volume
- Copy entrypoint script

### 2. Entrypoint Script (`docker/worker/entrypoint.sh`)
Required env vars:
- TASK_ID
- REPO_URL
- BRANCH
- GRANJA_API (base URL for callbacks)
- TASK_PROMPT
- TASK_TITLE

Flow:
1. Configure git user
2. Clone repo (or use existing)
3. Checkout/create epic branch
4. Run Pi agent with task prompt (non-interactive mode)
5. On success: commit, push, POST /api/tasks/:id/complete
6. On failure: POST /api/tasks/:id/fail with error

### 3. Docker Compose (`docker/docker-compose.yml`)
- Worker image build config
- Volume mounts for Pi config
- Network config

### 4. Build script (`docker/build.sh`)
- Build the worker image
- Tag as granja-worker:latest

Commit your changes when done.
