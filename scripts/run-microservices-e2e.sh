#!/usr/bin/env bash
set -Eeuo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
compose_file="$repo_root/docker-compose.microservices.yml"
project_name="${MICRO_E2E_PROJECT_NAME:-danmakustream-e2e}"
artifact_dir="${MICRO_E2E_ARTIFACT_DIR:-$repo_root/artifacts/microservices-e2e}"

export MICRO_GATEWAY_PORT="${MICRO_GATEWAY_PORT:-18888}"
export MICRO_FRONTEND_PORT="${MICRO_FRONTEND_PORT:-18080}"
export MICRO_RTMP_PORT="${MICRO_RTMP_PORT:-19350}"
export MICRO_HLS_PORT="${MICRO_HLS_PORT:-18081}"
export MICRO_E2E_GATEWAY_URL="http://127.0.0.1:${MICRO_GATEWAY_PORT}"
export MICRO_E2E_FRONTEND_URL="http://127.0.0.1:${MICRO_FRONTEND_PORT}"
export MICRO_E2E_ARTIFACT_DIR="$artifact_dir"
export COMMIT_SHA="${COMMIT_SHA:-$(git -C "$repo_root" rev-parse HEAD 2>/dev/null || printf 'development')}"
export BUILD_TIME="${BUILD_TIME:-$(date -u '+%Y-%m-%dT%H:%M:%SZ')}"

compose=(docker compose --project-name "$project_name" -f "$compose_file")
mkdir -p "$artifact_dir"

collect_evidence() {
  "${compose[@]}" ps --all > "$artifact_dir/compose-ps.txt" 2>&1 || true
  "${compose[@]}" logs --no-color > "$artifact_dir/compose.log" 2>&1 || true
}

cleanup() {
  result=$?
  collect_evidence
  if [[ "${MICRO_E2E_KEEP_STACK:-0}" != "1" ]]; then
    "${compose[@]}" down --volumes --remove-orphans >/dev/null 2>&1 || true
  else
    printf 'MICRO_E2E_KEEP_STACK=1; stack %s remains running.\n' "$project_name"
  fi
  exit "$result"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

wait_for_url() {
  name=$1
  url=$2
  for attempt in $(seq 1 90); do
    if curl --fail --silent --show-error "$url" >/dev/null; then
      printf '%s is ready: %s\n' "$name" "$url"
      return 0
    fi
    if [[ "$attempt" == "90" ]]; then
      printf '%s did not become ready: %s\n' "$name" "$url" >&2
      return 1
    fi
    sleep 2
  done
}

"${compose[@]}" up --detach --build --remove-orphans
wait_for_url gateway "$MICRO_E2E_GATEWAY_URL/gateway/health"
wait_for_url frontend "$MICRO_E2E_FRONTEND_URL/"

(
  cd "$repo_root/frontend"
  npm run test:e2e:microservices
)
