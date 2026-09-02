#!/usr/bin/env bash
set -Eeuo pipefail

namespace=${MICROSERVICE_NAMESPACE:-danmakustream-microservices}
action=${1:?usage: k3s-hpa-chaos.sh hpa SERVICE DURATION_SECONDS CONCURRENCY | chaos}

k() {
  sudo k3s kubectl -n "$namespace" "$@"
}

require_service() {
  case "$1" in
    user-service|content-service|engagement-service) ;;
    *) echo "unsupported service: $1" >&2; exit 2 ;;
  esac
}

number_in_range() {
  local value=$1 min=$2 max=$3 label=$4
  if [[ ! "$value" =~ ^[0-9]+$ ]] || (( value < min || value > max )); then
    echo "$label must be an integer in [$min,$max]" >&2
    exit 2
  fi
}

sample_hpa() {
  local service=$1 now iso current desired cpu target
  now=$(date +%s)
  iso=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
  current=$(k get hpa "$service" -o jsonpath='{.status.currentReplicas}' 2>/dev/null || true)
  desired=$(k get hpa "$service" -o jsonpath='{.status.desiredReplicas}' 2>/dev/null || true)
  cpu=$(k get hpa "$service" -o jsonpath='{.status.currentMetrics[0].resource.current.averageUtilization}' 2>/dev/null || true)
  target=$(k get hpa "$service" -o jsonpath='{.spec.metrics[0].resource.target.averageUtilization}' 2>/dev/null || true)
  current=${current:-0}
  desired=${desired:-0}
  cpu=${cpu:-0}
  target=${target:-0}
  echo "HPA_CSV,$now,$iso,$service,$current,$desired,$cpu,$target"
  echo "HPA_SAMPLE time=$iso service=$service current=$current desired=$desired cpu=${cpu}% target=${target}%"
  HPA_CURRENT=$current
  HPA_DESIRED=$desired
}

run_hpa() {
  local service=${2:?service is required}
  local duration=${3:-180}
  local concurrency=${4:-80}
  require_service "$service"
  number_in_range "$duration" 60 600 duration
  number_in_range "$concurrency" 1 500 concurrency

  k get deployment "$service" >/dev/null
  k get hpa "$service" >/dev/null
  if ! k top pods -l "app=$service" >/dev/null 2>&1; then
    echo "resource metrics are unavailable; verify the k3s metrics-server before running HPA" >&2
    exit 1
  fi

  local min_replicas job_name deadline max_observed=0 scaled_down=0
  min_replicas=$(k get hpa "$service" -o jsonpath='{.spec.minReplicas}')
  job_name="hpa-load-${service%-service}-$(date +%s)"
  cleanup_hpa_job() {
    set +e
    k delete job "$job_name" --ignore-not-found --wait=false >/dev/null 2>&1
  }
  trap cleanup_hpa_job EXIT

  echo "HPA_EXERCISE service=$service duration=${duration}s concurrency=$concurrency min=$min_replicas"
  sample_hpa "$service"
  sed \
    -e "s/__JOB_NAME__/$job_name/g" \
    -e "s/__SERVICE__/$service/g" \
    -e "s/__DURATION__/$duration/g" \
    -e "s/__CONCURRENCY__/$concurrency/g" <<'YAML' | k apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: __JOB_NAME__
spec:
  backoffLimit: 0
  template:
    metadata:
      labels:
        app: hpa-load-generator
    spec:
      restartPolicy: Never
      containers:
        - name: load
          image: curlimages/curl:8.10.1
          command: ["/bin/sh", "-c"]
          args:
            - |
              set -eu
              results=/tmp/results.csv
              : > "$results"
              end=$(( $(date +%s) + __DURATION__ ))
              worker=0
              while [ "$worker" -lt __CONCURRENCY__ ]; do
                (
                  while [ "$(date +%s)" -lt "$end" ]; do
                    curl -sS --no-keepalive --max-time 3 -o /dev/null \
                      -w '%{time_total},%{http_code}\n' \
                      http://__SERVICE__:8080/api/v1/livez >> "$results" \
                      || printf '3.000000,000\n' >> "$results"
                  done
                ) &
                worker=$((worker + 1))
              done
              wait
              count=$(wc -l < "$results")
              errors=$(awk -F, '$2 !~ /^2/ {n++} END {print n+0}' "$results")
              average=$(awk -F, '{sum+=$1} END {if (NR) printf "%.6f", sum/NR; else print "0"}' "$results")
              p95_index=$(( (count * 95 + 99) / 100 ))
              p95=$(cut -d, -f1 "$results" | sort -n | sed -n "${p95_index}p")
              error_rate=$(awk -v errors="$errors" -v count="$count" 'BEGIN {if (count) printf "%.4f", errors*100/count; else print "100"}')
              echo "LOAD_RESULT requests=$count errors=$errors errorRate=${error_rate}% average=${average}s p95=${p95:-0}s"
YAML

  deadline=$(( $(date +%s) + duration + 180 ))
  while (( $(date +%s) < deadline )); do
    sample_hpa "$service"
    if (( HPA_CURRENT > max_observed )); then max_observed=$HPA_CURRENT; fi
    if (( HPA_DESIRED > max_observed )); then max_observed=$HPA_DESIRED; fi
    if [[ "$(k get job "$job_name" -o jsonpath='{.status.succeeded}' 2>/dev/null || true)" == "1" ]]; then
      break
    fi
    if [[ -n "$(k get job "$job_name" -o jsonpath='{.status.failed}' 2>/dev/null || true)" ]]; then
      echo "load generator job failed" >&2
      break
    fi
    sleep 10
  done

  echo "== load generator summary =="
  k logs "job/$job_name" --tail=120 || true
  echo "== waiting for scale down =="
  deadline=$(( $(date +%s) + 480 ))
  while (( $(date +%s) < deadline )); do
    sample_hpa "$service"
    if (( HPA_CURRENT > max_observed )); then max_observed=$HPA_CURRENT; fi
    if (( HPA_DESIRED > max_observed )); then max_observed=$HPA_DESIRED; fi
    if (( HPA_CURRENT <= min_replicas && HPA_DESIRED <= min_replicas )); then
      scaled_down=1
      break
    fi
    sleep 10
  done

  echo "HPA_RESULT service=$service min=$min_replicas maxObserved=$max_observed final=$HPA_CURRENT"
  if (( max_observed <= min_replicas )); then
    echo "HPA did not scale up; increase load or inspect resource metrics" >&2
    return 1
  fi
  if (( scaled_down != 1 )); then
    echo "HPA did not return to minReplicas within 480 seconds" >&2
    return 1
  fi
  echo "HPA scale-up and scale-down verified"
}

restore_content_hpa() {
  k apply -f - <<'YAML'
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: content-service
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: content-service
  minReplicas: 1
  maxReplicas: 4
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 0
      selectPolicy: Max
      policies:
        - {type: Pods, value: 2, periodSeconds: 30}
        - {type: Percent, value: 100, periodSeconds: 30}
    scaleDown:
      stabilizationWindowSeconds: 60
      selectPolicy: Max
      policies:
        - {type: Pods, value: 1, periodSeconds: 60}
        - {type: Percent, value: 50, periodSeconds: 60}
  metrics:
    - type: Resource
      resource:
        name: cpu
        target: {type: Utilization, averageUtilization: 60}
YAML
}

pod_restart_snapshot() {
  local service=$1
  k get pods -l "app=$service" \
    -o jsonpath='{range .items[*]}{range .status.containerStatuses[*]}{.restartCount}{"\n"}{end}{end}' \
    | awk '{sum+=$1} END {print sum+0}'
}

base64url() {
  openssl base64 -A | tr '+/' '-_' | tr -d '='
}

run_chaos() {
  if [[ ${CONFIRM_CHAOS:-} != "yes" ]]; then
    echo "set CONFIRM_CHAOS=yes to acknowledge the temporary content-service outage" >&2
    exit 2
  fi
  local target=content-service probe="chaos-probe-$(date +%s)"
  local original_replicas had_hpa=0 before_user before_engagement
  original_replicas=$(k get deployment "$target" -o jsonpath='{.spec.replicas}')
  if k get hpa "$target" >/dev/null 2>&1; then had_hpa=1; fi
  before_user=$(pod_restart_snapshot user-service)
  before_engagement=$(pod_restart_snapshot engagement-service)

  cleanup_chaos() {
    local exit_code=$?
    set +e
    echo "CHAOS_RESTORE content-service replicas=$original_replicas hpa=$had_hpa"
    k scale deployment "$target" --replicas="$original_replicas"
    k rollout status deployment/"$target" --timeout=240s
    if (( had_hpa == 1 )); then restore_content_hpa; fi
    k delete pod "$probe" --ignore-not-found --wait=false >/dev/null 2>&1
    echo "CHAOS_RESTORE completed"
    return "$exit_code"
  }
  trap cleanup_chaos EXIT

  k run "$probe" --image=curlimages/curl:8.10.1 --restart=Never --command -- sleep 900
  k wait --for=condition=Ready "pod/$probe" --timeout=180s

  local jwt_secret header payload signature token now
  jwt_secret=$(k get secret micro-secrets -o jsonpath='{.data.JWT_SECRET}' | base64 -d)
  now=$(date +%s)
  header=$(printf '%s' '{"alg":"HS256","typ":"JWT"}' | base64url)
  payload=$(printf '{"userId":1,"username":"chaos-probe","role":"user","iat":%d,"exp":%d}' "$now" "$((now + 600))" | base64url)
  signature=$(printf '%s' "$header.$payload" | openssl dgst -sha256 -hmac "$jwt_secret" -binary | base64url)
  token="$header.$payload.$signature"
  unset jwt_secret signature

  echo "CHAOS_BASELINE target=$target replicas=$original_replicas"
  k exec "$probe" -- curl -fsS --max-time 3 http://content-service:8080/api/v1/livez >/dev/null
  k exec "$probe" -- curl -fsS --max-time 3 http://engagement-service:8080/api/v1/health >/dev/null

  if (( had_hpa == 1 )); then k delete hpa "$target" --wait=true; fi
  k scale deployment "$target" --replicas=0
  if [[ -n "$(k get pods -l "app=$target" -o name)" ]]; then
    k wait --for=delete pod -l "app=$target" --timeout=180s
  fi
  echo "CHAOS_INJECTED target=$target endpoints=$(k get endpoints "$target" -o jsonpath='{.subsets}' 2>/dev/null || true)"

  local result http_code elapsed body after_user after_engagement
  result=$(k exec "$probe" -- curl -sS --max-time 3 \
    -o /tmp/chaos-response.json -w '%{http_code},%{time_total}' \
    -H "Authorization: Bearer $token" -H 'Content-Type: application/json' \
    -H 'X-Request-ID: chaos-content-outage' \
    -X POST http://engagement-service:8080/api/v1/danmaku \
    --data '{"videoId":1,"content":"chaos-probe","time":0}')
  http_code=${result%%,*}
  elapsed=${result#*,}
  body=$(k exec "$probe" -- cat /tmp/chaos-response.json)
  echo "CHAOS_DEPENDENCY_RESPONSE status=$http_code elapsed=${elapsed}s body=$body"

  k exec "$probe" -- curl -fsS --max-time 3 http://engagement-service:8080/api/v1/livez >/dev/null
  k exec "$probe" -- curl -fsS --max-time 3 http://engagement-service:8080/api/v1/health >/dev/null
  k exec "$probe" -- curl -fsS --max-time 3 http://user-service:8080/api/v1/health >/dev/null
  after_user=$(pod_restart_snapshot user-service)
  after_engagement=$(pod_restart_snapshot engagement-service)
  echo "CHAOS_RESTARTS user-before=$before_user user-after=$after_user engagement-before=$before_engagement engagement-after=$after_engagement"
  echo "== engagement logs during dependency outage =="
  k logs deployment/engagement-service --tail=120 || true

  if [[ "$http_code" != "503" && "$http_code" != "504" ]]; then
    echo "expected an explicit 503/504 dependency failure, got $http_code" >&2
    return 1
  fi
  if ! awk -v elapsed="$elapsed" 'BEGIN { exit !(elapsed <= 2.5) }'; then
    echo "dependency failure exceeded 2.5 seconds: ${elapsed}s" >&2
    return 1
  fi
  if [[ "$before_user" != "$after_user" || "$before_engagement" != "$after_engagement" ]]; then
    echo "an unaffected service restarted during the exercise" >&2
    return 1
  fi
  echo "CHAOS_RESULT explicitFailure=$http_code elapsed=${elapsed}s unaffectedServicesHealthy=true restartsUnchanged=true"
}

case "$action" in
  hpa) run_hpa "$@" ;;
  chaos) run_chaos ;;
  *) echo "unknown action: $action" >&2; exit 2 ;;
esac
