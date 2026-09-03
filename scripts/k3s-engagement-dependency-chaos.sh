#!/usr/bin/env bash
set -Eeuo pipefail

# Destructive-by-design acceptance test for a dedicated test/maintenance window.
# It scales content-service to zero, verifies engagement-service fails clearly
# without restarting, then restores the exact original replica count.

namespace=${MICROSERVICE_NAMESPACE:-danmakustream-microservices}
dependency=${CHAOS_DEPENDENCY:-content-service}
gateway_url=${CHAOS_GATEWAY_URL:-http://127.0.0.1:30888}
request_limit_seconds=${CHAOS_REQUEST_LIMIT_SECONDS:-2.0}
evidence_root=${CHAOS_EVIDENCE_DIR:-artifacts/engagement-chaos}
run_id=$(date -u +%Y%m%dT%H%M%SZ)
evidence_dir="$evidence_root/$run_id"

if [[ ${CONFIRM_CHAOS:-} != engagement-dependency-outage ]]; then
	echo "refusing to stop a dependency: set CONFIRM_CHAOS=engagement-dependency-outage" >&2
	exit 2
fi
if [[ $dependency != content-service ]]; then
	echo "only content-service is supported by this repeatable HTTP acceptance test" >&2
	exit 2
fi
if [[ -z ${CHAOS_JWT:-} ]]; then
	echo "CHAOS_JWT is required; use a dedicated test account token" >&2
	exit 2
fi
if [[ ! ${CHAOS_VIDEO_ID:-} =~ ^[1-9][0-9]*$ ]]; then
	echo "CHAOS_VIDEO_ID must be a playable test video ID" >&2
	exit 2
fi
command -v curl >/dev/null || {
	echo "curl is required" >&2
	exit 2
}

mkdir -p "$evidence_dir"

k() {
	sudo k3s kubectl -n "$namespace" "$@"
}

replicas_before=''
restore_needed=false

collect_snapshot() {
	local label=$1
	{
		echo "== deployments =="
		k get deployments -o wide
		echo "== pods =="
		k get pods -o wide
		echo "== endpoints =="
		k get endpoints
		echo "== recent events =="
		k get events --sort-by=.lastTimestamp | tail -80
	} >"$evidence_dir/$label-cluster.txt" 2>&1 || true
	k logs deployment/engagement-service --all-containers=true --tail=250 \
		>"$evidence_dir/$label-engagement.log" 2>&1 || true
}

restore_dependency() {
	if [[ $restore_needed != true ]]; then
		return 0
	fi
	echo "restoring $dependency to $replicas_before replica(s)"
	k scale "deployment/$dependency" --replicas="$replicas_before"
	k rollout status "deployment/$dependency" --timeout=240s
	restore_needed=false
}

cleanup() {
	local rc=$?
	trap - EXIT INT TERM
	set +e
	if ! restore_dependency; then
		echo "failed to restore $dependency; manual recovery is required" >&2
		rc=1
	fi
	collect_snapshot final
	echo "evidence: $evidence_dir"
	exit "$rc"
}
trap cleanup EXIT
trap 'exit 130' INT TERM

pod_uid() {
	local app=$1
	k get pods -l "app=$app" -o jsonpath='{.items[0].metadata.uid}'
}

restart_count() {
	local app=$1
	k get pods -l "app=$app" -o jsonpath='{range .items[*].status.containerStatuses[*]}{.restartCount}{"\n"}{end}' |
		awk '{sum += $1} END {print sum + 0}'
}

assert_ready() {
	local deployment=$1
	local desired ready
	desired=$(k get "deployment/$deployment" -o jsonpath='{.spec.replicas}')
	ready=$(k get "deployment/$deployment" -o jsonpath='{.status.readyReplicas}')
	ready=${ready:-0}
	if ((ready < desired)); then
		echo "$deployment is not ready: desired=$desired ready=$ready" >&2
		return 1
	fi
}

assert_probe() {
	local deployment=$1
	local path=$2
	k exec "deployment/$deployment" -- wget -qO- -T 3 "http://127.0.0.1:8080$path" >/dev/null
}

call_video_like() {
	local label=$1
	local expected_status=$2
	local body="$evidence_dir/$label-body.json"
	local metrics status elapsed
	metrics=$(curl --silent --show-error --max-time 3 \
		--output "$body" --write-out '%{http_code} %{time_total}' \
		--request POST \
		--header "Authorization: Bearer $CHAOS_JWT" \
		"$gateway_url/api/v1/videos/$CHAOS_VIDEO_ID/like")
	read -r status elapsed <<<"$metrics"
	printf '%s status=%s elapsed=%ss\n' "$label" "$status" "$elapsed" |
		tee -a "$evidence_dir/requests.txt"
	if [[ $status != "$expected_status" ]]; then
		echo "$label returned HTTP $status, expected $expected_status; body: $(cat "$body")" >&2
		return 1
	fi
	if ! awk -v elapsed="$elapsed" -v limit="$request_limit_seconds" \
		'BEGIN { exit !(elapsed + 0 <= limit + 0) }'; then
		echo "$label took ${elapsed}s, above ${request_limit_seconds}s" >&2
		return 1
	fi
}

for deployment in mysql user-service content-service engagement-service gateway; do
	k get "deployment/$deployment" >/dev/null
	assert_ready "$deployment"
done

replicas_before=$(k get "deployment/$dependency" -o jsonpath='{.spec.replicas}')
if [[ ! $replicas_before =~ ^[1-9][0-9]*$ ]]; then
	echo "$dependency must have at least one replica before the experiment" >&2
	exit 1
fi

engagement_uid_before=$(pod_uid engagement-service)
engagement_restarts_before=$(restart_count engagement-service)
collect_snapshot before

# Two toggles prove the fixture is valid and restore its original like state.
call_video_like baseline-toggle-1 200
call_video_like baseline-toggle-2 200

restore_needed=true
k scale "deployment/$dependency" --replicas=0
for _ in $(seq 1 60); do
	if [[ -z $(k get pods -l "app=$dependency" -o name) ]]; then
		break
	fi
	sleep 2
done
if [[ -n $(k get pods -l "app=$dependency" -o name) ]]; then
	echo "$dependency pods did not terminate" >&2
	exit 1
fi

collect_snapshot outage
call_video_like dependency-outage 503

# Dependency failure must not poison engagement's own process/database probes.
assert_probe engagement-service /api/v1/livez
assert_probe engagement-service /api/v1/health
assert_probe user-service /api/v1/livez
k exec deployment/gateway -- wget -qO- -T 3 http://127.0.0.1/gateway/health >/dev/null

engagement_uid_after=$(pod_uid engagement-service)
engagement_restarts_after=$(restart_count engagement-service)
if [[ $engagement_uid_after != "$engagement_uid_before" ]]; then
	echo "engagement-service Pod was replaced during dependency outage" >&2
	exit 1
fi
if [[ $engagement_restarts_after != "$engagement_restarts_before" ]]; then
	echo "engagement-service restart count changed: $engagement_restarts_before -> $engagement_restarts_after" >&2
	exit 1
fi
for deployment in mysql user-service engagement-service gateway; do
	assert_ready "$deployment"
done

restore_dependency
assert_ready "$dependency"
call_video_like recovery-toggle-1 200
call_video_like recovery-toggle-2 200
collect_snapshot recovered

cat >"$evidence_dir/summary.txt" <<EOF
result=PASS
namespace=$namespace
dependency=$dependency
dependency_replicas_before=$replicas_before
engagement_pod_uid=$engagement_uid_before
engagement_restarts_before=$engagement_restarts_before
engagement_restarts_after=$engagement_restarts_after
failure_http_status=503
request_limit_seconds=$request_limit_seconds
EOF

echo "PASS: dependency failed clearly, engagement stayed healthy, and recovery succeeded"
