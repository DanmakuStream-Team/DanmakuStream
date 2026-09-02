#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
hpa_file="$repo_root/deploy/k8s/microservices/autoscaling.yaml"
targets=(user-service content-service engagement-service gateway)

[[ -f "$hpa_file" ]] || {
  echo "missing HPA manifest: $hpa_file" >&2
  exit 1
}

hpa_count=$(grep -c '^kind: HorizontalPodAutoscaler$' "$hpa_file")
[[ "$hpa_count" -eq "${#targets[@]}" ]] || {
  echo "expected ${#targets[@]} HPAs, found $hpa_count" >&2
  exit 1
}

for target in "${targets[@]}"; do
  grep -q "^[[:space:]]*name: ${target}-hpa[[:space:]]*$" "$hpa_file" || {
    echo "missing ${target}-hpa" >&2
    exit 1
  }
  grep -q "^[[:space:]]*name: ${target}[[:space:]]*$" "$hpa_file" || {
    echo "${target}-hpa does not target deployment/$target" >&2
    exit 1
  }

  manifest="$repo_root/deploy/k8s/microservices/${target}.yaml"
  grep -q '^[[:space:]]*resources:[[:space:]]*$' "$manifest" || {
    echo "deployment/$target is missing container resources" >&2
    exit 1
  }
  grep -q '^[[:space:]]*cpu:[[:space:]]*' "$manifest" || {
    echo "deployment/$target is missing a CPU request" >&2
    exit 1
  }
  if grep -q '^[[:space:]]*replicas:[[:space:]]*' "$manifest"; then
    echo "deployment/$target declares replicas and would fight its HPA" >&2
    exit 1
  fi
done

[[ $(grep -c '^  minReplicas: 1$' "$hpa_file") -eq "${#targets[@]}" ]]
[[ $(grep -c '^  maxReplicas: 5$' "$hpa_file") -eq "${#targets[@]}" ]]
[[ $(grep -c '^          averageUtilization: 60$' "$hpa_file") -eq "${#targets[@]}" ]]

if grep -q '^    name: mysql$' "$hpa_file"; then
  echo "mysql must not be targeted by a HorizontalPodAutoscaler" >&2
  exit 1
fi

echo "HPA validation OK: ${targets[*]}"
