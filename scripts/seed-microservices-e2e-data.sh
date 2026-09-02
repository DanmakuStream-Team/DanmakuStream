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

log() { printf '[micro-seed] %s\n' "$*"; }

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

mysql_root "
DELETE FROM content_db.videos WHERE title LIKE '${MARK}-%' OR title IN ('E2E-MC-公开视频','E2E-MC-待审核通过','E2E-MC-待审核拒绝','E2E-MEMBER-B-分享视频','E2E-UC05-互动测试视频');
"

SEED_VIDEO_FILE="$repo_root/tests/fixtures/seed-mini.mp4"
if [ ! -f "$SEED_VIDEO_FILE" ]; then
  log "  准备媒体占位：创建 10KB 空文件用于挂载内容"
  mkdir -p "$(dirname "$SEED_VIDEO_FILE")"
  printf 'ftypmp42' > "$SEED_VIDEO_FILE" 2>/dev/null || true
  head -c 10240 /dev/zero 2>/dev/null >> "$SEED_VIDEO_FILE" || head -c 10240 /dev/urandom >> "$SEED_VIDEO_FILE" 2>/dev/null || true
fi
log "  占位视频源文件：$SEED_VIDEO_FILE （大小=$(du -b "$SEED_VIDEO_FILE" 2>/dev/null | awk '{print $1}' || echo unknown) 字节）"

CREATOR_ID=$(mysql_root "SELECT id FROM user_db.users WHERE nickname='e2e-mc-creator' LIMIT 1;" | head -n1 || echo "")
log "  creator(e2e-mc-creator) id=$CREATOR_ID"

mysql_insert_video() {
  local title=$1 status=$2 owner_id=$3 description="$MARK 演示视频"
  if [ -z "$owner_id" ]; then
    log "  [$title] 跳过（owner_id 为空）"
    echo ""
    return 0
  fi
  mysql_root "INSERT INTO content_db.videos (user_id,title,description,video_url,cover_url,status,view_count,like_count,danmaku_count,created_at,updated_at) VALUES ($owner_id,'$title','$description','/data/videos/seed.mp4','','$status',0,0,0,NOW(),NOW());" 2>/dev/null || true
  local id
  id=$(mysql_root "SELECT id FROM content_db.videos WHERE title='$title' AND user_id=$owner_id ORDER BY id DESC LIMIT 1;" 2>/dev/null | tr -d '\r\n' || echo "")
  if [ -z "$id" ]; then
    log "  [$title] INSERT 失败，请检查 content_db.videos 是否就绪"
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

if [ -n "$OWNER_TOKEN" ] && [ -n "${VID_UC05:-}" ]; then
  log "互动域测试数据：为 UC05 视频预埋 0 弹幕/评论/点赞/收藏（E2E脚本自行验证从0起步）"
fi

log "微服务 E2E 数据播种完成。账号密码统一：$PASSWORD"
