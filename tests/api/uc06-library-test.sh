#!/usr/bin/env bash
# UC06 个人视频资料库 API 自动化测试（INT-TC06-01～12）
#
# 前置条件：后端、MySQL、curl、python3；MYSQL_CMD 可覆盖数据库命令。
# 示例：MYSQL_CMD="mysql -h127.0.0.1 -P3306 -uroot -ppassword danmakustream" \
#       tests/api/uc06-library-test.sh

set -u

API_BASE="${API_BASE:-http://localhost:8080/api/v1}"
MYSQL_CMD="${MYSQL_CMD:-mysql -S /home/haoyue/dms-mysql.sock -uroot -ppassword danmakustream}"
RUN_TAG="uc06api-$(date +%s)-$$"
PASSWORD='Test1234!'
BODY_FILE="$(mktemp)"
PASS=0
FAIL=0

log() { printf '%s | %s\n' "$(date '+%F %T')" "$*"; }
ok() { PASS=$((PASS + 1)); log "PASS | $*"; }
bad() { FAIL=$((FAIL + 1)); log "FAIL | $*"; }

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
jexpr() {
  python3 - "$BODY_FILE" "$1" <<'PY'
import json, sys
with open(sys.argv[1], encoding="utf-8") as handle:
    data = json.load(handle)
print(eval(sys.argv[2], {"__builtins__": {}}, {"d": data, "len": len, "str": str}))
PY
}
register_user() {
  curl -sS -X POST "$API_BASE/auth/register" -H 'Content-Type: application/json' \
    -d "{\"nickname\":\"$1\",\"password\":\"$PASSWORD\"}" >/dev/null
}
login() {
  curl -sS -X POST "$API_BASE/auth/login" -H 'Content-Type: application/json' \
    -d "{\"nickname\":\"$1\",\"password\":\"$PASSWORD\"}" |
    python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["token"])'
}

cleanup() {
  db "SET FOREIGN_KEY_CHECKS=0;
      DELETE FROM watch_histories WHERE user_id IN (SELECT id FROM users WHERE nickname LIKE '$RUN_TAG-%');
      DELETE FROM watch_laters WHERE user_id IN (SELECT id FROM users WHERE nickname LIKE '$RUN_TAG-%');
      DELETE FROM videos WHERE title LIKE '$RUN_TAG-%';
      DELETE FROM users WHERE nickname LIKE '$RUN_TAG-%';
      SET FOREIGN_KEY_CHECKS=1;" >/dev/null || true
  rm -f "$BODY_FILE"
}
trap cleanup EXIT

setup() {
  USER_A="$RUN_TAG-a"
  USER_B="$RUN_TAG-b"
  register_user "$USER_A"
  register_user "$USER_B"
  db "INSERT INTO videos
        (created_at,updated_at,title,description,video_url,status,transcode_status,duration,author_id)
      VALUES
        (NOW(),NOW(),'$RUN_TAG-approved','UC06 approved fixture','/media/videos/$RUN_TAG.mp4','approved','ready',120,
          (SELECT id FROM users WHERE nickname='$USER_A')),
        (NOW(),NOW(),'$RUN_TAG-pending','UC06 pending fixture','/media/videos/$RUN_TAG-pending.mp4','pending','ready',60,
          (SELECT id FROM users WHERE nickname='$USER_A'));" >/dev/null
  APPROVED_ID=$(db "SELECT id FROM videos WHERE title='$RUN_TAG-approved'")
  PENDING_ID=$(db "SELECT id FROM videos WHERE title='$RUN_TAG-pending'")
  TOKEN_A=$(login "$USER_A")
  TOKEN_B=$(login "$USER_B")
  log "data ready: approved=$APPROVED_ID pending=$PENDING_ID"
}

setup

# INT-TC06-01：鉴权与参数校验
c=$(http_code GET "$API_BASE/users/me/history")
assert_eq "INT-TC06-01 历史列表无 token => 401" 401 "$c"
c=$(http_code GET "$API_BASE/users/me/watch-later")
assert_eq "INT-TC06-01 稍后再看无 token => 401" 401 "$c"
c=$(http_code PUT "$API_BASE/users/me/history/not-a-number" "$TOKEN_A" '{"position":10}')
assert_eq "INT-TC06-02 非法视频 ID => 400" 400 "$c"
c=$(http_code PUT "$API_BASE/users/me/history/$APPROVED_ID" "$TOKEN_A" '{"position":-1}')
assert_eq "INT-TC06-02 负播放位置 => 400" 400 "$c"

# INT-TC06-03～07：观看历史保存、覆盖、隔离、删除与清空
c=$(http_code PUT "$API_BASE/users/me/history/$APPROVED_ID" "$TOKEN_A" '{"position":45}')
assert_eq "INT-TC06-03 保存观看位置 => 200" 200 "$c"
assert_eq "INT-TC06-03 DB 位置持久化" 45 "$(db "SELECT position FROM watch_histories WHERE user_id=(SELECT id FROM users WHERE nickname='$USER_A') AND video_id=$APPROVED_ID")"

c=$(http_code GET "$API_BASE/users/me/history/$APPROVED_ID" "$TOKEN_A")
assert_eq "INT-TC06-04 查询单条历史 => 200" 200 "$c"
assert_eq "INT-TC06-04 返回 position=45" 45 "$(jexpr 'd["data"]["position"]')"
assert_eq "INT-TC06-04 进度按 45/120 计算" 37 "$(jexpr 'd["data"]["progress"]')"

c=$(http_code PUT "$API_BASE/users/me/history/$APPROVED_ID" "$TOKEN_A" '{"position":999}')
assert_eq "INT-TC06-05 超出时长仍保存成功" 200 "$c"
assert_eq "INT-TC06-05 播放位置钳制到 duration" 120 "$(jexpr 'd["data"]["position"]')"

c=$(http_code PUT "$API_BASE/users/me/history/$PENDING_ID" "$TOKEN_A" '{"position":10}')
assert_eq "INT-TC06-05 未审核视频不能写历史 => 404" 404 "$c"
c=$(http_code GET "$API_BASE/users/me/history?page=1&pageSize=100" "$TOKEN_B")
assert_eq "INT-TC06-06 另一用户历史列表 => 200" 200 "$c"
assert_eq "INT-TC06-06 用户间历史隔离" 0 "$(jexpr 'len(d["data"]["list"])')"

c=$(http_code DELETE "$API_BASE/users/me/history/$APPROVED_ID" "$TOKEN_A")
assert_eq "INT-TC06-07 删除单条历史 => 200" 200 "$c"
assert_eq "INT-TC06-07 删除后 DB 无记录" 0 "$(db "SELECT COUNT(*) FROM watch_histories WHERE user_id=(SELECT id FROM users WHERE nickname='$USER_A')")"
http_code PUT "$API_BASE/users/me/history/$APPROVED_ID" "$TOKEN_A" '{"position":20}' >/dev/null
c=$(http_code DELETE "$API_BASE/users/me/history" "$TOKEN_A")
assert_eq "INT-TC06-07 清空历史 => 200" 200 "$c"
assert_eq "INT-TC06-07 清空后列表为空" true "$(jexpr 'str(d["data"]["cleared"]).lower()')"

# INT-TC06-08～12：稍后再看状态、切换、隔离、删除与清空
c=$(http_code GET "$API_BASE/users/me/watch-later/$APPROVED_ID/status" "$TOKEN_A")
assert_eq "INT-TC06-08 初始稍后再看状态 => 200" 200 "$c"
assert_eq "INT-TC06-08 初始 saved=false" false "$(jexpr 'str(d["data"]["saved"]).lower()')"

c=$(http_code POST "$API_BASE/users/me/watch-later/$APPROVED_ID" "$TOKEN_A")
assert_eq "INT-TC06-09 添加稍后再看 => 200" 200 "$c"
assert_eq "INT-TC06-09 saved=true" true "$(jexpr 'str(d["data"]["saved"]).lower()')"
c=$(http_code GET "$API_BASE/users/me/watch-later?page=1&pageSize=100" "$TOKEN_A")
assert_eq "INT-TC06-10 稍后再看列表 => 200" 200 "$c"
assert_contains "INT-TC06-10 列表包含目标视频" "$RUN_TAG-approved" "$(cat "$BODY_FILE")"

c=$(http_code GET "$API_BASE/users/me/watch-later?page=1&pageSize=100" "$TOKEN_B")
assert_eq "INT-TC06-10 另一用户稍后再看 => 200" 200 "$c"
assert_eq "INT-TC06-10 用户间稍后再看隔离" 0 "$(jexpr 'len(d["data"]["list"])')"
c=$(http_code POST "$API_BASE/users/me/watch-later/$PENDING_ID" "$TOKEN_A")
assert_eq "INT-TC06-11 未审核视频不能加入稍后再看 => 404" 404 "$c"

c=$(http_code DELETE "$API_BASE/users/me/watch-later/$APPROVED_ID" "$TOKEN_A")
assert_eq "INT-TC06-12 删除稍后再看 => 200" 200 "$c"
http_code POST "$API_BASE/users/me/watch-later/$APPROVED_ID" "$TOKEN_A" >/dev/null
c=$(http_code DELETE "$API_BASE/users/me/watch-later" "$TOKEN_A")
assert_eq "INT-TC06-12 清空稍后再看 => 200" 200 "$c"
assert_eq "INT-TC06-12 清空后 DB 无记录" 0 "$(db "SELECT COUNT(*) FROM watch_laters WHERE user_id=(SELECT id FROM users WHERE nickname='$USER_A')")"

log "===== RESULT: TOTAL=$((PASS + FAIL)) PASS=$PASS FAIL=$FAIL ====="
[ "$FAIL" -eq 0 ] || exit 1
