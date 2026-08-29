#!/usr/bin/env bash
# UC13 管理/审核域 API 自动化测试（INT-TC13-01～10）
#
# 前置条件：
#   1. 后端已启动（默认 http://localhost:8080）
#   2. MySQL 可达（默认用户态实例 socket；可用 MYSQL_CMD 覆盖，例如 Docker 环境：
#      MYSQL_CMD="mysql -h127.0.0.1 -P3306 -uroot -ppassword danmakustream"）
#   3. 系统内有 curl、python3、mysql 客户端
#
# 用法：
#   tests/api/uc13-admin-test.sh [2>&1 | tee report.txt]
# 任何断言失败时脚本以非零状态退出，可直接接入 CI。

set -u

API_BASE="${API_BASE:-http://localhost:8080/api/v1}"
MYSQL_CMD="${MYSQL_CMD:-mysql -S /home/haoyue/dms-mysql.sock -uroot -ppassword danmakustream}"
VIDEO_DIR="${VIDEO_DIR:-backend/data}"
RUN_TAG="uc13api-$(date +%s)"
MEDIA_FIXTURE_DIR="$VIDEO_DIR/videos/uc13-$RUN_TAG"
PASS=0
FAIL=0

log()  { printf '%s | %s\n' "$(date '+%F %T')" "$*"; }
ok()   { PASS=$((PASS+1)); log "PASS | $*"; }
bad()  { FAIL=$((FAIL+1)); log "FAIL | $*"; }

assert_eq() { # desc expected actual
  if [ "$2" = "$3" ]; then ok "$1 (got=$3)"; else bad "$1: expected=$2 actual=$3"; fi
}
assert_contains() { # desc needle haystack
  case "$3" in *"$2"*) ok "$1 (contains '$2')" ;; *) bad "$1: '$2' not found in: $(printf '%s' "$3" | head -c 200)" ;; esac
}

http_code() { # method url [token] [json]
  local m=$1 u=$2 t=${3:-} d=${4:-}
  local args=(-s -o /tmp/uc13-body -w '%{http_code}' -X "$m")
  [ -n "$t" ] && args+=(-H "Authorization: Bearer $t")
  [ -n "$d" ] && args+=(-H 'Content-Type: application/json' -d "$d")
  curl "${args[@]}" "$u"
}
body() { cat /tmp/uc13-body; }
jget() { body | python3 -c "import json,sys;d=json.load(sys.stdin);print(eval(sys.argv[1]))" "$1" 2>/dev/null; }
db() { $MYSQL_CMD -N -e "$1" 2>/dev/null; }

cleanup() {
  rm -rf -- "$MEDIA_FIXTURE_DIR"
  rm -f /tmp/uc13-body /tmp/uc13-pub
}
trap cleanup EXIT

# ---------- 测试数据准备 ----------
setup() {
	  mkdir -p "$MEDIA_FIXTURE_DIR"
	  printf 'UC13 local media fixture\n' > "$MEDIA_FIXTURE_DIR/fixture.mp4"
  for n in tuser tmod tadmin; do
    curl -s -X POST "$API_BASE/auth/register" -H 'Content-Type: application/json' \
      -d "{\"nickname\":\"$n\",\"password\":\"Test1234!\"}" > /dev/null
  done
  $MYSQL_CMD -e "
    UPDATE users SET role='user'      WHERE nickname='tuser';
    UPDATE users SET role='moderator' WHERE nickname='tmod';
    UPDATE users SET role='admin'     WHERE nickname='tadmin';
    DELETE FROM videos  WHERE title LIKE 'UC13-API-%';
    DELETE FROM danmakus WHERE content LIKE 'UC13-API-%';
    INSERT INTO videos (created_at,updated_at,title,video_url,status,transcode_status,author_id) VALUES
      (NOW(),NOW(),'UC13-API-$RUN_TAG-ready','/media/videos/uc13-$RUN_TAG/fixture.mp4','pending','ready',(SELECT id FROM users WHERE nickname='tuser')),
      (NOW(),NOW(),'UC13-API-$RUN_TAG-transcoding','','pending','processing',                   (SELECT id FROM users WHERE nickname='tuser'));
    INSERT INTO danmakus (created_at,updated_at,video_id,user_id,content,time) VALUES
      (NOW(),NOW(),(SELECT id FROM videos WHERE title='UC13-API-$RUN_TAG-ready'),
       (SELECT id FROM users WHERE nickname='tuser'),'UC13-API-$RUN_TAG-danmaku',5);" 2>/dev/null
  V_READY=$(db "SELECT id FROM videos WHERE title='UC13-API-$RUN_TAG-ready'")
  V_TRANS=$(db "SELECT id FROM videos WHERE title='UC13-API-$RUN_TAG-transcoding'")
  D_ID=$(db "SELECT id FROM danmakus WHERE content='UC13-API-$RUN_TAG-danmaku'")
  U_USER=$(db "SELECT id FROM users WHERE nickname='tuser'")
  TOKEN_USER=$(login tuser); TOKEN_MOD=$(login tmod); TOKEN_ADMIN=$(login tadmin)
}
login() { curl -s -X POST "$API_BASE/auth/login" -H 'Content-Type: application/json' \
  -d "{\"nickname\":\"$1\",\"password\":\"Test1234!\"}" | python3 -c 'import json,sys;print(json.load(sys.stdin)["data"]["token"])' 2>/dev/null; }

# ---------- INT-TC13 ----------
setup
log "data ready: V_READY=$V_READY V_TRANS=$V_TRANS D_ID=$D_ID U_USER=$U_USER"

# INT-01 GET /admin/videos (moderator) => 200 且返回 list 结构
c=$(http_code GET "$API_BASE/admin/videos" "$TOKEN_MOD")
assert_eq "INT-01 moderator GET /admin/videos => 200" 200 "$c"
assert_contains "INT-01 响应包含待审视频" "UC13-API-$RUN_TAG-ready" "$(body)"

# INT-02 审核通过（媒体就绪）=> 200，DB 持久化，公开列表同步
c=$(http_code PUT "$API_BASE/admin/videos/$V_READY/status" "$TOKEN_MOD" '{"status":"approved"}')
assert_eq "INT-02 审核通过 => 200" 200 "$c"
assert_eq "INT-02 DB status=approved" approved "$(db "SELECT status FROM videos WHERE id=$V_READY")"
curl -s "$API_BASE/videos?page=1&pageSize=50" > /tmp/uc13-pub
if python3 -c "
import json,sys
items=json.load(open('/tmp/uc13-pub'))['data'].get('items') or json.load(open('/tmp/uc13-pub'))['data'].get('list') or []
sys.exit(0 if $V_READY in [v.get('id') for v in items] else 1)" 2>/dev/null; then
  ok "INT-02 公开列表包含已审核视频"
else bad "INT-02 公开列表未包含视频 $V_READY"; fi

# INT-03 未转码视频审核 => 403 且状态不变；不存在 => 404
c=$(http_code PUT "$API_BASE/admin/videos/$V_TRANS/status" "$TOKEN_MOD" '{"status":"approved"}')
assert_eq "INT-03a 转码中审核 => 403" 403 "$c"
assert_eq "INT-03a DB status 仍为 pending" pending "$(db "SELECT status FROM videos WHERE id=$V_TRANS")"
c=$(http_code PUT "$API_BASE/admin/videos/999999/status" "$TOKEN_MOD" '{"status":"approved"}')
assert_eq "INT-03b 不存在视频 => 404" 404 "$c"

# INT-04 屏蔽弹幕 (moderator) => 200，DB blocked=1
c=$(http_code PUT "$API_BASE/admin/danmaku/$D_ID/block" "$TOKEN_MOD")
assert_eq "INT-04 屏蔽弹幕 => 200" 200 "$c"
assert_eq "INT-04 DB blocked=1" 1 "$(db "SELECT blocked FROM danmakus WHERE id=$D_ID")"

# INT-05 admin 修改角色 => 200，DB 同步
c=$(http_code PUT "$API_BASE/admin/users/$U_USER/role" "$TOKEN_ADMIN" '{"role":"moderator"}')
assert_eq "INT-05 admin 改角色 => 200" 200 "$c"
assert_eq "INT-05 DB role=moderator" moderator "$(db "SELECT role FROM users WHERE id=$U_USER")"

# INT-06 moderator 改角色 => 403，DB 不变
c=$(http_code PUT "$API_BASE/admin/users/$U_USER/role" "$TOKEN_MOD" '{"role":"admin"}')
assert_eq "INT-06 moderator 改角色 => 403" 403 "$c"
assert_eq "INT-06 DB role 不变" moderator "$(db "SELECT role FROM users WHERE id=$U_USER")"

# INT-07 Banner CRUD
c=$(http_code POST "$API_BASE/admin/banners" "$TOKEN_ADMIN" "{\"title\":\"UC13-API-$RUN_TAG\",\"imageUrl\":\"https://x/a.png\",\"link\":\"/\",\"enabled\":true,\"sort\":1}")
assert_eq "INT-07 banner 创建 => 200" 200 "$c"
B_ID=$(db "SELECT id FROM site_banners WHERE title='UC13-API-$RUN_TAG'")
[ -n "$B_ID" ] && ok "INT-07 banner 已入库 (id=$B_ID)" || bad "INT-07 banner 未入库"
c=$(http_code PUT "$API_BASE/admin/banners/$B_ID" "$TOKEN_ADMIN" "{\"title\":\"UC13-API-$RUN_TAG-v2\",\"imageUrl\":\"https://x/b.png\",\"link\":\"/v\",\"enabled\":false,\"sort\":2}")
assert_eq "INT-07 banner 更新 => 200" 200 "$c"
assert_eq "INT-07 更新持久化" "UC13-API-$RUN_TAG-v2" "$(db "SELECT title FROM site_banners WHERE id=$B_ID")"
c=$(http_code DELETE "$API_BASE/admin/banners/$B_ID" "$TOKEN_ADMIN")
assert_eq "INT-07 banner 删除 => 200" 200 "$c"
assert_eq "INT-07 删除后不在列表" 0 "$(db "SELECT COUNT(*) FROM site_banners WHERE title LIKE 'UC13-API-$RUN_TAG%' AND deleted_at IS NULL")"

# INT-08 Announcement CRUD（含非法时间 400）
c=$(http_code POST "$API_BASE/admin/announcements" "$TOKEN_ADMIN" "{\"content\":\"UC13-API-$RUN_TAG\",\"enabled\":true,\"startedAt\":\"2026-08-26T00:00:00Z\",\"endedAt\":\"2026-09-26T00:00:00Z\"}")
assert_eq "INT-08 公告创建 => 200" 200 "$c"
A_ID=$(db "SELECT id FROM site_announcements WHERE content='UC13-API-$RUN_TAG'")
[ -n "$A_ID" ] && ok "INT-08 公告已入库 (id=$A_ID)" || bad "INT-08 公告未入库"
c=$(http_code POST "$API_BASE/admin/announcements" "$TOKEN_ADMIN" "{\"content\":\"x\",\"startedAt\":\"not-a-time\"}")
assert_eq "INT-08 非法时间 => 400" 400 "$c"
c=$(http_code PUT "$API_BASE/admin/announcements/$A_ID" "$TOKEN_ADMIN" "{\"content\":\"UC13-API-$RUN_TAG-v2\",\"enabled\":false}")
assert_eq "INT-08 公告更新 => 200" 200 "$c"
c=$(http_code DELETE "$API_BASE/admin/announcements/$A_ID" "$TOKEN_ADMIN")
assert_eq "INT-08 公告删除 => 200" 200 "$c"
assert_eq "INT-08 删除后不在列表" 0 "$(db "SELECT COUNT(*) FROM site_announcements WHERE content LIKE 'UC13-API-$RUN_TAG%' AND deleted_at IS NULL")"

# INT-09 基础设施指标 => 200 且四大区块齐全
c=$(http_code GET "$API_BASE/admin/infrastructure" "$TOKEN_ADMIN")
assert_eq "INT-09 GET infrastructure => 200" 200 "$c"
for key in storage traffic cpu online; do
  assert_contains "INT-09 包含 $key 指标" "\"$key\"" "$(body)"
done

# INT-10 无 token / 无效 token => 401
c=$(http_code GET "$API_BASE/admin/videos")
assert_eq "INT-10 无 token => 401" 401 "$c"
c=$(http_code GET "$API_BASE/admin/videos" "invalid.token.value")
assert_eq "INT-10 伪造 token => 401" 401 "$c"

# ---------- 汇总 ----------
log "===== RESULT: TOTAL=$((PASS+FAIL)) PASS=$PASS FAIL=$FAIL ====="
[ "$FAIL" -eq 0 ] || exit 1
exit 0
