#!/usr/bin/env bash
set -euo pipefail

commit_sha=${1:?usage: render-microservice-manifests.sh COMMIT_SHA [GHCR_OWNER]}
ghcr_owner=${2:-danmakustream-team}
ghcr_owner=${ghcr_owner,,}

if [[ ! "$commit_sha" =~ ^[0-9a-f]{40}$ ]]; then
  echo "COMMIT_SHA must be a full 40-character lowercase hexadecimal SHA" >&2
  exit 2
fi
if [[ ! "$ghcr_owner" =~ ^[a-z0-9][a-z0-9-]*$ ]]; then
  echo "GHCR owner contains unsupported characters" >&2
  exit 2
fi

repo_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
manifest_dir="$repo_root/deploy/k8s/microservices"
placeholder=0000000000000000000000000000000000000000

emit_file() {
  local path=$1
  sed \
    -e "s|ghcr.io/danmakustream-team/user-service:${placeholder}|ghcr.io/${ghcr_owner}/user-service:${commit_sha}|g" \
    -e "s|ghcr.io/danmakustream-team/content-service:${placeholder}|ghcr.io/${ghcr_owner}/content-service:${commit_sha}|g" \
    -e "s|ghcr.io/danmakustream-team/engagement-service:${placeholder}|ghcr.io/${ghcr_owner}/engagement-service:${commit_sha}|g" \
    -e "s|danmakustream.io/release-commit: \"${placeholder}\"|danmakustream.io/release-commit: \"${commit_sha}\"|g" \
    "$path"
}

manifests=(
  namespace.yaml
  mysql.yaml
  user-service.yaml
  content-service.yaml
  engagement-service.yaml
  gateway.yaml
  hpa.yaml
)

for index in "${!manifests[@]}"; do
  emit_file "$manifest_dir/${manifests[$index]}"
  if (( index + 1 < ${#manifests[@]} )); then
    printf '\n---\n'
  fi
done
