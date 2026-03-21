#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

cd "$SCRIPT_DIR/worker"
docker build -t granja-worker:latest -f Dockerfile .

echo "Built image: granja-worker:latest"
