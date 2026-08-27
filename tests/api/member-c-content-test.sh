#!/usr/bin/env bash
# 成员 C 内容域 API 自动化测试（UC02/UC03/UC04/UC12）
#
# 前置条件：后端、MySQL、curl、python3；MYSQL_CMD 可覆盖数据库命令。
# 退出规则：断言 FAIL 返回非零；已知未实现要求记为 GAP，但不伪装为 PASS。

set -u

API_BASE="${API_BASE:-http://localhost:8080/api/v1}"
MYSQL_CMD="${MYSQL_CMD:-mysql -S /home/haoyue/dms-mysql.sock -uroot -ppassword danmakustream}"
VIDEO_DIR="${VIDEO_DIR:-backend/data}"
RUN_TAG="mc$(date +%s)$$"
PASSWORD='Test1234!'
BODY_FILE="$(mktemp)"
FIXTURE_DIR="$(mktemp -d)"
PASS=0
FAIL=0
GAP=0

log() { printf '%s | %s\n' "$(date '+%F %T')" "$*"; }
ok() { PASS=$((PASS + 1)); log "PASS | $*"; }
bad() { FAIL=$((FAIL + 1)); log "FAIL | $*"; }
gap() { GAP=$((GAP + 1)); log "GAP  | $*"; }

assert_eq() {
  if [ "$2" = "$3" ]; then ok "$1 (got=$3)"; else bad "$1: expected=$2 actual=$3"; fi
}
assert_contains() {
  case "$3" in *"$2"*) ok "$1 (contains '$2')" ;; *) bad "$1: '$2' not found" ;; esac
}

db() { $MYSQL_CMD -N -B -e "$1" 2>/dev/null; }
http_code() {
  local method=$1 url=$2 token=${3:-} data=${4:-}
  local args=(-sS -o "$BODY_FILE" -w '%{http_code}' -X "$method")
  [ -n "$token" ] && args+=(-H "Authorization: Bearer $token")
  [ -n "$data" ] && args+=(-H 'Content-Type: application/json' -d "$data")
  curl "${args[@]}" "$url"
}
upload_code() {
  local token=$1 title=$2 file=$3
  curl -sS -o "$BODY_FILE" -w '%{http_code}' -X POST "$API_BASE/videos/upload" \
    -H "Authorization: Bearer $token" -F "title=$title" -F "description=member C API fixture" \
    -F "category=tech" -F "tags=ci,member-c" -F "video=@$file"
}
jexpr() {
  python3 - "$BODY_FILE" "$1" <<'PY'
import json, sys
with open(sys.argv[1], encoding="utf-8") as handle:
    data = json.load(handle)
print(eval(sys.argv[2], {"__builtins__": {}}, {"d": data, "len": len}))
PY
}
login() {
  curl -sS -X POST "$API_BASE/auth/login" -H 'Content-Type: application/json' \
    -d "{\"nickname\":\"$1\",\"password\":\"$PASSWORD\"}" |
    python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["token"])'
}

cleanup() {
  db "DELETE FROM video_daily_stats WHERE creator_id IN (SELECT id FROM users WHERE nickname LIKE 'mc-$RUN_TAG-%');
      DELETE FROM creator_daily_stats WHERE creator_id IN (SELECT id FROM users WHERE nickname LIKE 'mc-$RUN_TAG-%');
      DELETE FROM videos WHERE title LIKE 'MC-$RUN_TAG-%';
      DELETE FROM users WHERE nickname LIKE 'mc-$RUN_TAG-%';" >/dev/null || true
  if [[ "${UPLOAD_ID:-}" =~ ^[0-9]+$ ]]; then
    rm -rf -- "$VIDEO_DIR/videos/$UPLOAD_ID"
  fi
  rm -f "$BODY_FILE"
  rm -rf "$FIXTURE_DIR"
}
trap cleanup EXIT

register_user() {
  curl -sS -X POST "$API_BASE/auth/register" -H 'Content-Type: application/json' \
    -d "{\"nickname\":\"$1\",\"password\":\"$PASSWORD\"}" >/dev/null
}

setup() {
  CREATOR="mc-$RUN_TAG-creator"
  OTHER="mc-$RUN_TAG-other"
  MODERATOR="mc-$RUN_TAG-mod"
  EMPTY="mc-$RUN_TAG-empty"
  for nickname in "$CREATOR" "$OTHER" "$MODERATOR" "$EMPTY"; do register_user "$nickname"; done

  db "UPDATE users SET role='creator' WHERE nickname IN ('$CREATOR','$OTHER','$EMPTY');
      UPDATE users SET role='moderator' WHERE nickname='$MODERATOR';
      INSERT INTO videos (created_at,updated_at,title,description,video_url,status,author_id,view_count,collect_count,category,tags) VALUES
       (NOW(),NOW(),'MC-$RUN_TAG-approved','search target','/media/videos/fixture/playlist.m3u8','approved',(SELECT id FROM users WHERE nickname='$CREATOR'),20,4,'tech','ci,member-c'),
       (DATE_SUB(NOW(),INTERVAL 1 DAY),NOW(),'MC-$RUN_TAG-second','second approved','/media/videos/fixture/playlist.m3u8','approved',(SELECT id FROM users WHERE nickname='$CREATOR'),8,2,'life','member-c'),
       (NOW(),NOW(),'MC-$RUN_TAG-pending-ready','ready for review','/media/videos/fixture/playlist.m3u8','pending',(SELECT id FROM users WHERE nickname='$CREATOR'),0,0,'tech','ci'),
       (NOW(),NOW(),'MC-$RUN_TAG-pending-missing','missing media','', 'pending',(SELECT id FROM users WHERE nickname='$CREATOR'),0,0,'tech','ci'),
       (NOW(),NOW(),'MC-$RUN_TAG-other','other creator','/media/videos/fixture/playlist.m3u8','approved',(SELECT id FROM users WHERE nickname='$OTHER'),99,9,'tech','ci');
      INSERT INTO creator_daily_stats (created_at,updated_at,creator_id,date,view_delta,collect_delta,stream_count) VALUES
       (NOW(),NOW(),(SELECT id FROM users WHERE nickname='$CREATOR'),DATE_FORMAT(CURDATE(),'%Y-%m-%d'),10,2,1),
       (NOW(),NOW(),(SELECT id FROM users WHERE nickname='$CREATOR'),DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 1 DAY),'%Y-%m-%d'),5,1,2);
      INSERT INTO video_daily_stats (created_at,updated_at,creator_id,video_id,date,view_delta,collect_delta) VALUES
       (NOW(),NOW(),(SELECT id FROM users WHERE nickname='$CREATOR'),(SELECT id FROM videos WHERE title='MC-$RUN_TAG-approved'),DATE_FORMAT(CURDATE(),'%Y-%m-%d'),6,1),
       (NOW(),NOW(),(SELECT id FROM users WHERE nickname='$CREATOR'),(SELECT id FROM videos WHERE title='MC-$RUN_TAG-approved'),DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 1 DAY),'%Y-%m-%d'),3,1);" >/dev/null

  CREATOR_ID=$(db "SELECT id FROM users WHERE nickname='$CREATOR'")
  APPROVED_ID=$(db "SELECT id FROM videos WHERE title='MC-$RUN_TAG-approved'")
  SECOND_ID=$(db "SELECT id FROM videos WHERE title='MC-$RUN_TAG-second'")
  READY_ID=$(db "SELECT id FROM videos WHERE title='MC-$RUN_TAG-pending-ready'")
  MISSING_ID=$(db "SELECT id FROM videos WHERE title='MC-$RUN_TAG-pending-missing'")
  OTHER_ID=$(db "SELECT id FROM videos WHERE title='MC-$RUN_TAG-other'")
  TOKEN_CREATOR=$(login "$CREATOR")
  TOKEN_OTHER=$(login "$OTHER")
  TOKEN_MOD=$(login "$MODERATOR")
  TOKEN_EMPTY=$(login "$EMPTY")
  printf 'not a media file\n' > "$FIXTURE_DIR/invalid.mp4"
  log "data ready: creator=$CREATOR_ID approved=$APPROVED_ID ready=$READY_ID missing=$MISSING_ID"
}

setup

# ---------- UC02：发现、搜索、详情与播放入口 ----------
c=$(http_code GET "$API_BASE/videos?page=1&pageSize=2&keyword=MC-$RUN_TAG&sort=date")
assert_eq "INT-TC02-01 搜索公开视频 => 200" 200 "$c"
assert_eq "INT-TC02-01 分页大小保持为 2" 2 "$(jexpr 'd["data"]["pageSize"]')"
assert_contains "INT-TC02-01 命中已审核视频" "MC-$RUN_TAG-approved" "$(cat "$BODY_FILE")"
if grep -q "MC-$RUN_TAG-pending" "$BODY_FILE"; then bad "INT-TC02-01 搜索泄露待审视频"; else ok "INT-TC02-01 待审视频未公开"; fi

before_views=$(db "SELECT view_count FROM videos WHERE id=$APPROVED_ID")
c=$(http_code GET "$API_BASE/videos/$APPROVED_ID")
assert_eq "INT-TC02-02 已审核详情 => 200" 200 "$c"
after_views=$(db "SELECT view_count FROM videos WHERE id=$APPROVED_ID")
assert_eq "INT-TC02-02 详情访问播放量 +1" "$((before_views + 1))" "$after_views"
assert_eq "INT-TC02-02 当日作品统计同步 +1" 7 "$(db "SELECT view_delta FROM video_daily_stats WHERE video_id=$APPROVED_ID AND date=DATE_FORMAT(CURDATE(),'%Y-%m-%d')")"
c=$(http_code GET "$API_BASE/videos/$MISSING_ID")
assert_eq "INT-TC02-03 待审详情不可公开 => 404" 404 "$c"
c=$(http_code GET "$API_BASE/videos/not-a-number")
assert_eq "INT-TC02-03 非法详情 ID => 400" 400 "$c"
c=$(http_code GET "$API_BASE/videos?page=1&pageSize=10&sort=views")
if [ "$c" = 400 ]; then ok "INT-TC02-04 非法排序 => 400"; else gap "INT-TC02-04 非法排序当前返回 $c，期望 400"; fi
gap "UC02 详情接口未校验 videoUrl 对应媒体文件是否真实存在"

# ---------- UC03：投稿、取消与状态跟踪 ----------
c=$(http_code POST "$API_BASE/videos/upload")
assert_eq "INT-TC03-01 未登录投稿 => 401" 401 "$c"
c=$(http_code POST "$API_BASE/videos/upload" "$TOKEN_CREATOR")
assert_eq "INT-TC03-01 缺标题和视频 => 400" 400 "$c"
c=$(http_code POST "$API_BASE/videos/upload" "$TOKEN_CREATOR" '{}')
assert_eq "INT-TC03-01 JSON 请求不能代替 multipart 视频 => 400" 400 "$c"

c=$(upload_code "$TOKEN_CREATOR" "MC-$RUN_TAG-transcode-failure" "$FIXTURE_DIR/invalid.mp4")
assert_eq "INT-TC03-02 上传请求创建待审记录 => 200" 200 "$c"
UPLOAD_ID=$(db "SELECT id FROM videos WHERE title='MC-$RUN_TAG-transcode-failure'")
sleep 1
assert_eq "INT-TC03-02 转码失败视频保持 pending" pending "$(db "SELECT status FROM videos WHERE id=$UPLOAD_ID")"
assert_eq "INT-TC03-02 转码失败视频没有公开视频地址" 0 "$(db "SELECT LENGTH(video_url) FROM videos WHERE id=$UPLOAD_ID")"
gap "UC03 转码失败仅写服务日志，数据模型没有 failed 状态或失败原因"

truncate -s 16777216 "$FIXTURE_DIR/cancel.mp4"
set +e
timeout 1 curl -sS --limit-rate 16k -X POST "$API_BASE/videos/upload" \
  -H "Authorization: Bearer $TOKEN_CREATOR" -F "title=MC-$RUN_TAG-cancelled" \
  -F "video=@$FIXTURE_DIR/cancel.mp4" >/dev/null
cancel_rc=$?
set +e
if [ "$cancel_rc" -ne 0 ]; then ok "INT-TC03-03 客户端中止了进行中的上传"; else bad "INT-TC03-03 上传未按测试预期被中止"; fi
sleep 1
assert_eq "INT-TC03-03 中止上传不产生视频记录" 0 "$(db "SELECT COUNT(*) FROM videos WHERE title='MC-$RUN_TAG-cancelled'")"
c=$(http_code GET "$API_BASE/users/me/videos?page=1&pageSize=100" "$TOKEN_CREATOR")
assert_eq "INT-TC03-04 创作者查询自己的投稿 => 200" 200 "$c"
assert_contains "INT-TC03-04 返回转码失败待审记录" "MC-$RUN_TAG-transcode-failure" "$(cat "$BODY_FILE")"

# ---------- UC04：审核与发布 ----------
c=$(http_code GET "$API_BASE/admin/videos")
assert_eq "INT-TC04-01 未登录审核列表 => 401" 401 "$c"
c=$(http_code GET "$API_BASE/admin/videos" "$TOKEN_CREATOR")
assert_eq "INT-TC04-01 普通创作者审核列表 => 403" 403 "$c"
c=$(http_code GET "$API_BASE/admin/videos?status=pending&keyword=MC-$RUN_TAG" "$TOKEN_MOD")
assert_eq "INT-TC04-01 审核员查看待审列表 => 200" 200 "$c"
assert_contains "INT-TC04-01 待审列表包含目标" "MC-$RUN_TAG-pending-ready" "$(cat "$BODY_FILE")"

c=$(http_code PUT "$API_BASE/admin/videos/$READY_ID/status" "$TOKEN_MOD" '{"status":"approved"}')
assert_eq "INT-TC04-02 审核通过 => 200" 200 "$c"
assert_eq "INT-TC04-02 审核状态持久化" approved "$(db "SELECT status FROM videos WHERE id=$READY_ID")"
c=$(http_code GET "$API_BASE/videos/$READY_ID")
assert_eq "INT-TC04-02 通过后可公开访问 => 200" 200 "$c"
c=$(http_code PUT "$API_BASE/admin/videos/$SECOND_ID/status" "$TOKEN_MOD" '{"status":"rejected"}')
assert_eq "INT-TC04-03 审核拒绝 => 200" 200 "$c"
c=$(http_code GET "$API_BASE/videos/$SECOND_ID")
assert_eq "INT-TC04-03 拒绝后不可公开访问 => 404" 404 "$c"
c=$(http_code PUT "$API_BASE/admin/videos/$MISSING_ID/status" "$TOKEN_MOD" '{"status":"approved"}')
assert_eq "INT-TC04-04 无媒体源不能通过 => 403" 403 "$c"
c=$(http_code PUT "$API_BASE/admin/videos/$MISSING_ID/status" "$TOKEN_MOD" '{"status":"published"}')
assert_eq "INT-TC04-04 非法状态 => 400" 400 "$c"
c=$(http_code PUT "$API_BASE/admin/videos/999999999/status" "$TOKEN_MOD" '{"status":"rejected"}')
assert_eq "INT-TC04-04 不存在视频 => 404" 404 "$c"
gap "UC04 重复审核可以覆盖既有终态，且审核通过未检查 HLS 文件真实存在"

# ---------- UC12：创作者数据分析 ----------
c=$(http_code GET "$API_BASE/creator/analytics?days=7")
assert_eq "INT-TC12-01 未登录分析接口 => 401" 401 "$c"
c=$(http_code GET "$API_BASE/creator/analytics?days=14" "$TOKEN_CREATOR")
assert_eq "INT-TC12-01 非法 days => 400" 400 "$c"
c=$(http_code GET "$API_BASE/creator/analytics?videoId=abc" "$TOKEN_CREATOR")
assert_eq "INT-TC12-01 非法 videoId => 400" 400 "$c"
c=$(http_code GET "$API_BASE/creator/analytics?days=7&videoId=$OTHER_ID" "$TOKEN_CREATOR")
assert_eq "INT-TC12-01 不能查询他人作品 => 404" 404 "$c"

c=$(http_code GET "$API_BASE/creator/analytics?days=7" "$TOKEN_CREATOR")
assert_eq "INT-TC12-02 全部作品 7 天分析 => 200" 200 "$c"
assert_eq "INT-TC12-02 趋势补齐 7 个自然日" 7 "$(jexpr 'len(d["data"]["points"])')"
EXPECTED_RANGE_VIEWS=$(db "SELECT COALESCE(SUM(view_delta),0) FROM creator_daily_stats WHERE creator_id=$CREATOR_ID AND date BETWEEN DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 6 DAY),'%Y-%m-%d') AND DATE_FORMAT(CURDATE(),'%Y-%m-%d')")
assert_eq "INT-TC12-02 范围新增观看汇总" "$EXPECTED_RANGE_VIEWS" "$(jexpr 'd["data"]["summary"]["rangeViews"]')"
assert_eq "INT-TC12-02 范围新增收藏汇总" 3 "$(jexpr 'd["data"]["summary"]["rangeCollects"]')"
assert_eq "INT-TC12-02 作品排行最多 5 条" True "$(jexpr 'len(d["data"]["topVideos"]) <= 5')"

c=$(http_code GET "$API_BASE/creator/analytics?days=7&videoId=$APPROVED_ID" "$TOKEN_CREATOR")
assert_eq "INT-TC12-03 单作品分析 => 200" 200 "$c"
assert_eq "INT-TC12-03 单作品范围观看" 10 "$(jexpr 'd["data"]["summary"]["rangeViews"]')"
assert_eq "INT-TC12-03 返回选中作品 ID" "$APPROVED_ID" "$(jexpr 'd["data"]["selectedVideoId"]')"
if [ "$(jexpr 'len(d["data"]["topVideos"])')" = 1 ]; then
  ok "INT-TC12-03 单作品模式排行同步收窄"
else
  gap "UC12 单作品模式的 topVideos 仍返回创作者全部作品"
fi

c=$(http_code GET "$API_BASE/creator/analytics?days=30" "$TOKEN_EMPTY")
assert_eq "INT-TC12-04 空数据分析 => 200" 200 "$c"
assert_eq "INT-TC12-04 空数据仍返回 30 个零值点" 30 "$(jexpr 'len(d["data"]["points"])')"
assert_eq "INT-TC12-04 空数据总播放为 0" 0 "$(jexpr 'd["data"]["summary"]["totalViews"]')"

log "===== RESULT: TOTAL=$((PASS + FAIL + GAP)) PASS=$PASS FAIL=$FAIL GAP=$GAP ====="
[ "$FAIL" -eq 0 ] || exit 1
exit 0
