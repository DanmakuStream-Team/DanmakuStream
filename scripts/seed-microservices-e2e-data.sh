#!/usr/bin/env bash
# 微服务专用 E2E 数据播种（幂等，可重复执行）
# 通过网关 + 真实注册接口创建 13UC 所需账号与视频数据
# 依赖：docker compose 微服务栈已启动（gateway, mysql, 三服务健康）
set -Eeuo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
project_name="${MICRO_E2E_PROJECT_NAME:-danmakustream-e2e}"
compose_file="$repo_root/docker-compose.microservices.yml"
compose=(docker compose --project-name "$project_name" -f "$compose_file")
gateway_url="${MICRO_E2E_GATEWAY_URL:-http://127.0.0.1:18888}"
PASSWORD='Test1234!'
MARK='MICRO-E2E-SEED'

log() { printf '[micro-seed] %s\n' "$*" >&2; }

wait_for_gateway() {
  for i in $(seq 1 60); do
    if curl --fail --silent --show-error "$gateway_url/gateway/health" >/dev/null 2>&1; then
      log "gateway 健康"
      return 0
    fi
    [ "$i" = 60 ] && { log "gateway 未就绪，放弃"; return 1; }
    sleep 2
  done
}
wait_for_gateway

wait_for_content_service() {
  for i in $(seq 1 60); do
    if curl --fail --silent --show-error "$gateway_url/api/v1/videos?pageSize=1" >/dev/null 2>&1; then
      log "content-service 健康（videos 列表可调用）"
      return 0
    fi
    [ "$i" = 60 ] && { log "content-service 未就绪（videos 列表始终不可达），放弃"; return 1; }
    sleep 2
  done
}
wait_for_content_service

register() {
  local nick=$1
  curl --silent --show-error --output /dev/null -w '%{http_code}' \
    --header 'Content-Type: application/json' \
    --data "{\"nickname\":\"$nick\",\"password\":\"$PASSWORD\"}" \
    "$gateway_url/api/v1/auth/register" || echo "000"
}

log "创建 13UC 所需 E2E 账号（若已存在则跳过）"
declare -A codes
for nick in \
  e2e-d-owner e2e-d-viewer tuser tmod tadmin e2eplain \
  e2e-mc-creator e2e-mc-moderator e2e-mc-plain; do
  codes[$nick]=$(register "$nick")
  log "  $nick -> HTTP ${codes[$nick]}"
done

mysql_root() {
  "${compose[@]}" exec -T mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD:-local-dev-root-password}" -N -B -e "$1" 2>/dev/null || true
}

mysql_exec() {
  local sql="$1" label="${2:-unnamed-sql}"
  local stdout_tmp stderr_tmp rc
  stdout_tmp=$(mktemp)
  stderr_tmp=$(mktemp)
  set +e
  "${compose[@]}" exec -T mysql mysql -uroot -p"${MYSQL_ROOT_PASSWORD:-local-dev-root-password}" --default-character-set=utf8mb4 -N -B -e "SET NAMES utf8mb4; SET CHARACTER SET utf8mb4; $sql" \
    >"$stdout_tmp" 2>"$stderr_tmp"
  rc=$?
  set -e
  local stdout_text stderr_text
  stdout_text=$(tr -d '\r' < "$stdout_tmp" || true)
  stderr_text=$(tr -d '\r' < "$stderr_tmp" || true)
  rm -f "$stdout_tmp" "$stderr_tmp"

  if [ "$rc" -ne 0 ] || [ -n "$stderr_text" ]; then
    log "  MySQL [$label] rc=$rc"
    if [ -n "$stdout_text" ]; then
      local IFS=$'\n'
      local line
      for line in $stdout_text; do
        [ -n "$line" ] && log "    [stdout] $line"
      done
    fi
    if [ -n "$stderr_text" ]; then
      local IFS=$'\n'
      local line
      for line in $stderr_text; do
        [ -n "$line" ] && log "    [stderr] $line"
      done
    fi
    log "  MySQL [$label] sql>>> SET NAMES utf8mb4; SET CHARACTER SET utf8mb4; $sql"
  fi

  if [ -n "$stdout_text" ]; then
    printf '%s\n' "$stdout_text"
  fi
  return $rc
}

log "固化账号角色（e2e-mc-moderator=moderator, tadmin=admin, 其余=user）"
mysql_root "
UPDATE user_db.users SET role='moderator' WHERE nickname IN ('tmod','e2e-mc-moderator');
UPDATE user_db.users SET role='admin'     WHERE nickname='tadmin';
UPDATE user_db.users SET role='user'      WHERE nickname NOT IN ('tmod','e2e-mc-moderator','tadmin');
"

login_and_get_token() {
  local nick=$1
  local resp
  resp=$(curl --silent --show-error --header 'Content-Type: application/json' \
    --data "{\"nickname\":\"$nick\",\"password\":\"$PASSWORD\"}" \
    "$gateway_url/api/v1/auth/login")
  echo "$resp" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])" 2>/dev/null || echo ""
}

CREATOR_TOKEN=$(login_and_get_token e2e-mc-creator)
OWNER_TOKEN=$(login_and_get_token e2e-d-owner)

log "内容域测试数据：清理并插入视频（直接 MySQL INSERT 绕过上传接口校验 40010）"

mysql_exec "SELECT SCHEMA_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME IN ('content_db','user_db','engagement_db') ORDER BY SCHEMA_NAME;" "检查3库是否存在"
mysql_exec "SHOW CREATE TABLE content_db.videos\G" "DESCRIBE content_db.videos（排查字段名/类型/NOT NULL/DEFAULT）"

mysql_exec "
DELETE FROM content_db.videos WHERE title LIKE '${MARK}-%' OR title IN ('E2E-MC-公开视频','E2E-MC-待审核通过','E2E-MC-待审核拒绝','E2E-MEMBER-B-分享视频','E2E-UC05-互动测试视频');
" "清理旧播种视频"

SEED_VIDEO_FILE="$repo_root/tests/fixtures/seed-mini.mp4"
if [ ! -f "$SEED_VIDEO_FILE" ]; then
  log "  准备媒体占位：创建 10KB 空文件用于挂载内容"
  mkdir -p "$(dirname "$SEED_VIDEO_FILE")"
  printf 'ftypmp42' > "$SEED_VIDEO_FILE" 2>/dev/null || true
  head -c 10240 /dev/zero 2>/dev/null >> "$SEED_VIDEO_FILE" || head -c 10240 /dev/urandom >> "$SEED_VIDEO_FILE" 2>/dev/null || true
fi
log "  占位视频源文件：$SEED_VIDEO_FILE （大小=$(du -b "$SEED_VIDEO_FILE" 2>/dev/null | awk '{print $1}' || echo unknown) 字节）"

CREATOR_ID_RAW=$(mysql_exec "SELECT id, nickname, role FROM user_db.users WHERE nickname='e2e-mc-creator' LIMIT 1;" "查 creator(e2e-mc-creator) 完整行" || true)
CREATOR_ID=$(printf '%s\n' "$CREATOR_ID_RAW" | grep -Eo '^[0-9]+' | head -n1)
CREATOR_ID="${CREATOR_ID:-}"
log "  creator(e2e-mc-creator) id=$CREATOR_ID"

MODERATOR_RAW=$(mysql_exec "SELECT id, nickname, role FROM user_db.users WHERE nickname='e2e-mc-moderator' LIMIT 1;" "查 moderator(e2e-mc-moderator)" || true)
MODERATOR_ID=$(printf '%s\n' "$MODERATOR_RAW" | grep -Eo '^[0-9]+' | head -n1)
MODERATOR_ID="${MODERATOR_ID:-}"
[ -n "$MODERATOR_ID" ] && log "  moderator(e2e-mc-moderator) id=$MODERATOR_ID"

sql_escape() {
  local s="${1:-}"
  printf '%s' "$s" | sed "s/'/\\\\'/g"
}

mysql_insert_video() {
  local title="$1"
  local status="$2"
  local owner_id="$3"
  local description="${MARK:-MICRO-E2E-SEED} 演示视频"
  if [ -z "$owner_id" ]; then
    log "  [$title] 跳过（owner_id 为空）"
    echo ""
    return 0
  fi
  if ! [[ "$owner_id" =~ ^[0-9]+$ ]] || [ "$owner_id" -le 0 ]; then
    log "  [$title] 跳过（owner_id=$owner_id 不是正整数）"
    echo ""
    return 0
  fi
  local esc_title esc_desc
  esc_title=$(sql_escape "$title")
  esc_desc=$(sql_escape "$description")
  log "  [$title] 开始 INSERT：author_id=$owner_id status=$status description='$description'"
  local insert_sql_full insert_sql_min lookup_sql
  insert_sql_full="INSERT INTO content_db.videos (author_id,title,description,video_url,cover_url,status,transcode_status,transcode_error,duration,view_count,like_count,collect_count,danmaku_count,tags,category,created_at,updated_at,deleted_at) VALUES ($owner_id,'$esc_title','$esc_desc','/data/videos/seed.mp4','','$status','ready','',0,0,0,0,0,'','',NOW(3),NOW(3),NULL);"
  insert_sql_min="INSERT INTO content_db.videos (author_id,title,status,video_url,created_at,updated_at) VALUES ($owner_id,'$esc_title','$status','/data/videos/seed.mp4',NOW(3),NOW(3));"
  lookup_sql="SELECT id FROM content_db.videos WHERE title='$esc_title' AND author_id=$owner_id AND deleted_at IS NULL ORDER BY id DESC LIMIT 1;"
  local used_sql=""
  set +e
  mysql_exec "$insert_sql_full" "video-insert-full[$title]"
  local rc_full=$?
  set -e
  used_sql="$insert_sql_full"
  local id
  id=$(mysql_exec "$lookup_sql" "lookup[$title]" 2>/dev/null | tr -d '\r\n' || echo "")
  if [ -z "$id" ]; then
    log "  [$title] 全字段 INSERT 未生效（rc=$rc_full），回退到最小 NOT NULL 字段集（走 DEFAULT）"
    set +e
    mysql_exec "$insert_sql_min" "video-insert-min[$title]"
    local rc_min=$?
    set -e
    used_sql="$insert_sql_min"
    id=$(mysql_exec "$lookup_sql" "lookup2[$title]" 2>/dev/null | tr -d '\r\n' || echo "")
    if [ -z "$id" ]; then
      log "  [$title] 两版 INSERT 都失败（full rc=$rc_full / min rc=$rc_min）"
      log "  [$title] 最终 used_sql>>> $used_sql"
    else
      log "  [$title] 回退最小字段集成功，id=$id"
    fi
  else
    log "  [$title] 全字段 INSERT 成功，id=$id"
  fi
  echo "$id"
}

VID_APPROVED=$(mysql_insert_video "E2E-MC-公开视频"      "approved" "$CREATOR_ID") || true
VID_PENDING1=$(mysql_insert_video "E2E-MC-待审核通过"     "pending"  "$CREATOR_ID") || true
VID_PENDING2=$(mysql_insert_video "E2E-MC-待审核拒绝"     "pending"  "$CREATOR_ID") || true
VID_SHARED_B=$(mysql_insert_video "E2E-MEMBER-B-分享视频" "approved" "$CREATOR_ID") || true
VID_UC05=$(mysql_insert_video     "E2E-UC05-互动测试视频" "approved" "$CREATOR_ID") || true
log "  E2E-MC-公开视频      id=$VID_APPROVED"
log "  E2E-MC-待审核通过     id=$VID_PENDING1"
log "  E2E-MC-待审核拒绝     id=$VID_PENDING2"
log "  E2E-MEMBER-B-分享视频 id=$VID_SHARED_B"
log "  E2E-UC05-互动测试视频 id=$VID_UC05"
set +e
mysql_exec "SELECT id, author_id, status, title, video_url, HEX(title) AS hex_title FROM content_db.videos WHERE author_id=$CREATOR_ID AND deleted_at IS NULL ORDER BY id DESC LIMIT 6;" "播种后 content_db 回查（6条，含 HEX(title) 验证 UTF-8 未乱码）" >/dev/null
set -e

log "===== [HTTP 自检] content-service 通过网关的 7 个真实接口响应（直接判断接口认不认 DB 里的数据） ====="
do_http_check() {
  local label="$1" url="$2" auth_header="${3:-}"
  local body_file code
  body_file=$(mktemp)
  if [ -n "$auth_header" ]; then
    code=$(curl -sS -o "$body_file" -w '%{http_code}' --max-time 10 "$url" -H "$auth_header" 2>/dev/null || echo 000)
  else
    code=$(curl -sS -o "$body_file" -w '%{http_code}' --max-time 10 "$url" 2>/dev/null || echo 000)
  fi
  local body
  body=$(head -c 800 "$body_file" 2>/dev/null | tr -d '\r' | tr '\n' ' ' || echo "")
  rm -f "$body_file"
  log "  HTTP [$label] HTTP=$code url=$url"
  log "    body=$body"
}
GW="${MICRO_E2E_GATEWAY_URL:-http://127.0.0.1:18888}"
API="$GW/api/v1"

CREATOR_TOKEN_HTTP=""
if [ -n "$CREATOR_TOKEN" ]; then CREATOR_TOKEN_HTTP="Authorization: Bearer $CREATOR_TOKEN"; fi
MODERATOR_TOKEN_HTTP=""
if [ -n "$MODERATOR_TOKEN" ]; then MODERATOR_TOKEN_HTTP="Authorization: Bearer $MODERATOR_TOKEN"; fi

do_http_check "C1 GET /videos (公开 无关键字 approved+ready)" "$API/videos?page=1&pageSize=50"
do_http_check "C2 GET /videos keyword=公开" "$API/videos?page=1&pageSize=50&keyword=$(printf '公开' | python3 -c 'import sys,urllib.parse;print(urllib.parse.quote(sys.stdin.read().strip()))' 2>/dev/null || printf '%%E5%%85%%AC%%E5%%BC%%80')"
if [ -n "$VID_APPROVED" ]; then do_http_check "C3 GET /videos/$VID_APPROVED （单条详情）" "$API/videos/$VID_APPROVED"; fi
if [ -n "$CREATOR_TOKEN_HTTP" ]; then
  do_http_check "C4 GET /creator/mine/videos （作者后台含 pending）" "$API/creator/mine/videos?page=1&pageSize=20" "$CREATOR_TOKEN_HTTP"
fi
if [ -n "$MODERATOR_TOKEN_HTTP" ]; then
  do_http_check "C5 GET /admin/videos （审核后台含全部状态）" "$API/admin/videos?page=1&pageSize=20" "$MODERATOR_TOKEN_HTTP"
  do_http_check "C6 GET /admin/videos status=pending" "$API/admin/videos?page=1&pageSize=20&status=pending" "$MODERATOR_TOKEN_HTTP"
fi
if [ -n "$CREATOR_ID" ]; then
  do_http_check "C7 GET /creator/$CREATOR_ID/videos （他人创作者主页 只看 approved）" "$API/creator/$CREATOR_ID/videos?page=1&pageSize=20"
fi

if [ -n "$OWNER_TOKEN" ] && [ -n "${VID_UC05:-}" ]; then
  log "互动域测试数据：为 UC05 视频预埋 0 弹幕/评论/点赞/收藏（E2E脚本自行验证从0起步）"
fi

log "微服务 E2E 数据播种完成。账号密码统一：$PASSWORD"
