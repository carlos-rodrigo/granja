#!/usr/bin/env bash
set -Eeuo pipefail

# Required env vars
: "${TASK_ID:?TASK_ID required}"
: "${REPO_URL:?REPO_URL required}"
: "${BRANCH:?BRANCH required}"
: "${GRANJA_API:?GRANJA_API required}"
: "${TASK_PROMPT:?TASK_PROMPT required}"
: "${TASK_TITLE:?TASK_TITLE required}"

post_fail() {
  local message="${1:-Worker failed}"
  curl -fsS -X POST "$GRANJA_API/api/tasks/$TASK_ID/fail" \
    -H "Content-Type: application/json" \
    -d "{\"error\":\"$message\"}" || true
}

post_complete() {
  curl -fsS -X POST "$GRANJA_API/api/tasks/$TASK_ID/complete" \
    -H "Content-Type: application/json" \
    -d '{"success":true}'
}

on_error() {
  local exit_code="$1"
  post_fail "entrypoint error (exit $exit_code)"
  exit "$exit_code"
}
trap 'on_error $?' ERR

# 1) Configure git user
git config --global user.email "granja@localhost"
git config --global user.name "Granja Worker"

# 2) Clone repo (or use existing)
REPO_DIR="/workspace/repo"
if [[ ! -d "$REPO_DIR/.git" ]]; then
  rm -rf "$REPO_DIR"
  git clone "$REPO_URL" "$REPO_DIR"
fi
cd "$REPO_DIR"

# 3) Checkout/create epic branch
CURRENT_ORIGIN_URL="$(git remote get-url origin 2>/dev/null || true)"
if [[ "$CURRENT_ORIGIN_URL" != "$REPO_URL" ]]; then
  git remote remove origin 2>/dev/null || true
  git remote add origin "$REPO_URL"
fi

git fetch origin
if git show-ref --verify --quiet "refs/remotes/origin/$BRANCH"; then
  git checkout -B "$BRANCH" "origin/$BRANCH"
elif git show-ref --verify --quiet "refs/remotes/origin/main"; then
  git checkout -B "$BRANCH" "origin/main"
else
  git checkout -B "$BRANCH"
fi

# 4) Run Pi agent in non-interactive mode
if ! pi --print -p "$TASK_PROMPT"; then
  post_fail "Pi exited with non-zero"
  exit 1
fi

# 5) On success: commit, push, callback complete
git add -A
if ! git diff --cached --quiet; then
  git commit -m "feat: $TASK_TITLE"
fi
git push -u origin "$BRANCH"
post_complete

echo "Task $TASK_ID completed successfully"
