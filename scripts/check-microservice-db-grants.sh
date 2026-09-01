#!/usr/bin/env bash
set -euo pipefail

files=(
  scripts/microservices-db-init.sql
  deploy/k8s/microservices/mysql.yaml
)

for file in "${files[@]}"; do
  if grep -Eq 'GRANT[[:space:]]+ALL[[:space:]]+PRIVILEGES' "$file"; then
    echo "forbidden broad grant found in $file" >&2
    exit 1
  fi

  grep -Fq "ON user_db.* TO 'user_app'@'%'" "$file"
  grep -Fq "ON content_db.* TO 'content_app'@'%'" "$file"
  grep -Fq "ON engagement_db.* TO 'engagement_app'@'%'" "$file"

  if grep -Eq "ON (content_db|engagement_db)\.\* TO 'user_app'" "$file" ||
     grep -Eq "ON (user_db|engagement_db)\.\* TO 'content_app'" "$file" ||
     grep -Eq "ON (user_db|content_db)\.\* TO 'engagement_app'" "$file"; then
    echo "cross-schema grant found in $file" >&2
    exit 1
  fi
done

echo "microservice database grants are schema-scoped"
