#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
USERNAME="${USERNAME:-testuser_$(date +%s)}"
PASSWORD="${PASSWORD:-123456}"
NICKNAME="${NICKNAME:-测试用户}"
MYSQL_USER="${MYSQL_USER:-root}"
MYSQL_PASSWORD="${MYSQL_PASSWORD:-password}"
MYSQL_DATABASE="${MYSQL_DATABASE:-danmakustream}"

request() {
  local method="$1"
  local path="$2"
  local data="${3:-}"
  local response_file
  response_file="$(mktemp)"

  if [[ -n "$data" ]]; then
    status="$(curl -sS -o "$response_file" -w "%{http_code}" \
      -X "$method" "$BASE_URL$path" \
      -H "Content-Type: application/json" \
      -d "$data")"
  else
    status="$(curl -sS -o "$response_file" -w "%{http_code}" \
      -X "$method" "$BASE_URL$path")"
  fi

  body="$(cat "$response_file")"
  rm -f "$response_file"

  if [[ "$status" -lt 200 || "$status" -ge 300 ]]; then
    echo "Request failed: $method $path"
    echo "HTTP status: $status"
    echo "$body"
    exit 1
  fi

  printf '%s' "$body"
}

json_get() {
  local key="$1"
  python3 -c '
import json
import sys

data = json.load(sys.stdin)
value = data
for part in sys.argv[1].split("."):
    value = value[part]
print(value)
' "$key"
}

json_assert_has_key() {
  local key="$1"
  python3 -c '
import json
import sys

data = json.load(sys.stdin)
value = data
for part in sys.argv[1].split("."):
    if part not in value:
        raise SystemExit(f"missing key: {sys.argv[1]}")
    value = value[part]
' "$key"
}

mysql_query() {
  docker compose exec -T mysql mysql \
    --default-character-set=utf8mb4 \
    -u"$MYSQL_USER" \
    -p"$MYSQL_PASSWORD" \
    -N -B "$MYSQL_DATABASE" \
    -e "$1"
}

echo "Testing API server: $BASE_URL"
echo

echo "1. POST /api/v1/auth/register"
register_body="{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\",\"nickname\":\"$NICKNAME\"}"
register_resp="$(request POST /api/v1/auth/register "$register_body")"
printf '%s' "$register_resp" | json_assert_has_key token
printf '%s' "$register_resp" | json_assert_has_key userInfo.id
user_id="$(printf '%s' "$register_resp" | json_get userInfo.id)"
echo "   ok: registered user id=$user_id username=$USERNAME"
echo

echo "2. POST /api/v1/auth/login"
login_body="{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}"
login_resp="$(request POST /api/v1/auth/login "$login_body")"
printf '%s' "$login_resp" | json_assert_has_key token
printf '%s' "$login_resp" | json_assert_has_key userInfo.id
token="$(printf '%s' "$login_resp" | json_get token)"
echo "   ok: login returned token length=${#token}"
echo

echo "3. Seed one approved video for detail/list checks"
video_id="$(mysql_query "
INSERT INTO videos
(created_at, updated_at, title, description, cover_url, video_url, duration, view_count, like_count, collect_count, danmaku_count, status, author_id, tags)
VALUES
(NOW(), NOW(), '接口测试视频', '用于测试视频列表和详情接口', 'http://example.com/test-cover.jpg', 'http://example.com/test-video.mp4', 120, 0, 0, 0, 0, 'approved', $user_id, 'test,backend');
SELECT LAST_INSERT_ID();
")"
echo "   ok: seeded approved video id=$video_id"
echo

echo "4. GET /api/v1/videos?page=1&pageSize=10"
videos_resp="$(request GET '/api/v1/videos?page=1&pageSize=10')"
printf '%s' "$videos_resp" | json_assert_has_key list
printf '%s' "$videos_resp" | json_assert_has_key total
printf '%s' "$videos_resp" | json_assert_has_key page
printf '%s' "$videos_resp" | json_assert_has_key pageSize
total="$(printf '%s' "$videos_resp" | json_get total)"
echo "   ok: video list returned total=$total"
echo

echo "5. GET /api/v1/videos/:id"
detail_resp="$(request GET "/api/v1/videos/$video_id")"
printf '%s' "$detail_resp" | json_assert_has_key id
printf '%s' "$detail_resp" | json_assert_has_key title
printf '%s' "$detail_resp" | json_assert_has_key author.id
printf '%s' "$detail_resp" | json_assert_has_key danmakuCount
printf '%s' "$detail_resp" | json_assert_has_key commentCount
printf '%s' "$detail_resp" | json_assert_has_key viewCount
detail_id="$(printf '%s' "$detail_resp" | json_get id)"
view_count="$(printf '%s' "$detail_resp" | json_get viewCount)"

if [[ "$detail_id" != "$video_id" ]]; then
  echo "Expected detail id=$video_id, got id=$detail_id"
  exit 1
fi

if [[ "$view_count" -lt 1 ]]; then
  echo "Expected viewCount >= 1 after detail request, got viewCount=$view_count"
  exit 1
fi

echo "   ok: video detail returned id=$detail_id viewCount=$view_count"
echo

echo "All core API checks passed."
