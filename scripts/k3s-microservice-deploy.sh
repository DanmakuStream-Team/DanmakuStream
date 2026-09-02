#!/usr/bin/env bash
set -Eeuo pipefail

namespace=${MICROSERVICE_NAMESPACE:-danmakustream-microservices}
action=${1:?usage: k3s-microservice-deploy.sh precheck|deploy|rollback|evidence [COMMIT_SHA]}
target_sha=${2:-}
rollout_timeout=${ROLLOUT_TIMEOUT:-240s}
deployments=(user-service content-service engagement-service gateway)
services=(user-service content-service engagement-service)
hpas=(user-service-hpa content-service-hpa engagement-service-hpa gateway-hpa)

k() {
  sudo k3s kubectl -n "$namespace" "$@"
}

require_target_sha() {
  if [[ ! "$target_sha" =~ ^[0-9a-f]{40}$ ]]; then
    echo "target commit must be a full 40-character lowercase hexadecimal SHA" >&2
    exit 2
  fi
}

wait_for_endpoint() {
  local service=$1
  local path=$2
  local expected=${3:-}
  local body=''
  for _ in $(seq 1 24); do
    if body=$(k exec "deployment/$service" -- \
      wget -qO- -T 5 "http://127.0.0.1:8080$path" 2>/dev/null); then
      if [[ -z "$expected" || "$body" == *"$expected"* ]]; then
        printf '%s %s => %s\n' "$service" "$path" "$body"
        return 0
      fi
    fi
    sleep 5
  done
  echo "$service $path did not become valid; last response: $body" >&2
  return 1
}

precheck() {
  sudo k3s kubectl wait --for=condition=Ready node --all --timeout=60s
  if ! sudo k3s kubectl get --raw '/apis/metrics.k8s.io/v1beta1/nodes' >/dev/null; then
    echo "metrics.k8s.io is unavailable; HPA requires metrics-server" >&2
    return 1
  fi
  k get secret micro-secrets >/dev/null
  local keys=(
    MYSQL_ROOT_PASSWORD USER_DB_PASSWORD CONTENT_DB_PASSWORD ENGAGEMENT_DB_PASSWORD
    USER_DATABASE_DSN CONTENT_DATABASE_DSN ENGAGEMENT_DATABASE_DSN
    JWT_SECRET INTERNAL_API_TOKEN
  )
  local key value
  for key in "${keys[@]}"; do
    value=$(k get secret micro-secrets -o "jsonpath={.data.${key}}")
    if [[ -z "$value" ]]; then
      echo "micro-secrets is missing required key: $key" >&2
      return 1
    fi
  done
  echo "precheck OK: node Ready, metrics-server available, and micro-secrets contains all required keys"
}

deploy() {
  require_target_sha
  k rollout status deployment/mysql --timeout="$rollout_timeout"
  local deployment image
  for deployment in "${deployments[@]}"; do
    k rollout status "deployment/$deployment" --timeout="$rollout_timeout"
  done
  for deployment in "${services[@]}"; do
    image=$(k get "deployment/$deployment" -o jsonpath='{.spec.template.spec.containers[0].image}')
    if [[ "$image" != *":$target_sha" ]]; then
      echo "$deployment uses unexpected image: $image" >&2
      return 1
    fi
    echo "$deployment image verified: $image"
    wait_for_endpoint "$deployment" /api/v1/livez
    wait_for_endpoint "$deployment" /api/v1/health
    wait_for_endpoint "$deployment" /api/v1/version "$target_sha"
  done
  local hpa
  for hpa in "${hpas[@]}"; do
    k get "horizontalpodautoscaler/$hpa" >/dev/null
  done
  echo "HPA objects verified: ${hpas[*]}"
  echo "microservice deployment verified at commit $target_sha"
}

rollback() {
  local deployment rc=0
  for deployment in "${deployments[@]}"; do
    echo "rolling back $deployment"
    k rollout undo "deployment/$deployment" || rc=1
  done
  for deployment in "${deployments[@]}"; do
    k rollout status "deployment/$deployment" --timeout="$rollout_timeout" || rc=1
  done
  return "$rc"
}

evidence() {
  set +e
  echo "== workloads and services =="
  k get deployments,pods,services,pvc,horizontalpodautoscalers -o wide
  echo "== resource usage =="
  k top pods --containers
  echo "== deployed images =="
  k get deployments -o custom-columns='NAME:.metadata.name,IMAGE:.spec.template.spec.containers[*].image,READY:.status.readyReplicas,AVAILABLE:.status.availableReplicas'
  echo "== probes and versions =="
  local service path
  for service in "${services[@]}"; do
    for path in /api/v1/livez /api/v1/health /api/v1/version; do
      printf '%s %s => ' "$service" "$path"
      k exec "deployment/$service" -- wget -qO- -T 5 "http://127.0.0.1:8080$path"
      echo
    done
  done
  echo "== recent events =="
  k get events --sort-by=.lastTimestamp | tail -50
  for service in "${services[@]}" gateway; do
    echo "== $service logs (last 150 lines) =="
    k logs "deployment/$service" --all-containers=true --tail=150
  done
  return 0
}

case "$action" in
  precheck) precheck ;;
  deploy) deploy ;;
  rollback) rollback ;;
  evidence) evidence ;;
  *)
    echo "unknown action: $action" >&2
    exit 2
    ;;
esac
