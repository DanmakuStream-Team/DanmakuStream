#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
CONFIG_FILE="${CONFIG_FILE:-etc/config.yaml}"
GOCACHE="${GOCACHE:-/tmp/go-build}"

export GOCACHE

echo "Backend directory: $BACKEND_DIR"
echo "Config file: $CONFIG_FILE"
echo "Go cache: $GOCACHE"
echo

cd "$BACKEND_DIR"

echo "1. Checking backend build with go test ./..."
go test ./...
echo

echo "2. Starting backend server"
echo "   URL: http://localhost:8080"
echo

exec go run api/main.go -f "$CONFIG_FILE"
