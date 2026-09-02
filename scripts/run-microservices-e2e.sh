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
export MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-local-dev-root-password}"
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
  if [[ "$result" != "0" ]]; then
    printf '\nMicroservice E2E failed; container state:\n' >&2
    cat "$artifact_dir/compose-ps.txt" >&2 || true
    for service in mysql user-service content-service engagement-service gateway frontend srs; do
      printf '\n===== %s (last 60 lines) =====\n' "$service" >&2
      "${compose[@]}" logs --no-color --tail 60 "$service" >&2 || true
    done
  fi
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

export E2E_MICROSERVICES=1
export E2E_USE_GATEWAY=1
export E2E_API_BASE="$MICRO_E2E_GATEWAY_URL/api/v1"
export E2E_BASE_URL="$MICRO_E2E_FRONTEND_URL"
export COMPOSE_MICRO="${compose[*]}"
export MYSQL_ROOT_CMD="${compose[*]} exec -T mysql mysql -uroot -p${MYSQL_ROOT_PASSWORD}"

log_step() { printf '\n===== [microservices-e2e] %s =====\n' "$*"; }

log_step "播种微服务 E2E 数据"
(
  cd "$repo_root"
  bash scripts/seed-microservices-e2e-data.sh
) || {
  printf '[WARN] 数据播种失败，尝试继续执行 E2E（部分用例可能因缺数据跳过）\n' >&2
}

e2e_npm_script=test:e2e:microservices
if [[ "${MICRO_E2E_FULL_SUITE:-0}" == "1" ]]; then
  log_step "运行 13/13 全量微服务 E2E（FULL SUITE）"
  e2e_npm_script=test:e2e:microservices:full
else
  log_step "运行微服务 E2E 冒烟基线（网关+JWT 三服务连通，2/13）"
fi

(
  cd "$repo_root/frontend"
  npm run "$e2e_npm_script"
)
