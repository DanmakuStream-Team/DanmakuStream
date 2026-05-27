#!/bin/bash
set -e

BASE_URL="http://localhost:8080/api/v1"
VIDEO_DIR="/home/haoyue/Videos/Hidamari"

# Login to get token
echo "=== Logging in ==="
LOGIN_RESP=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"nickname":"testadmin","password":"123456"}')
TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "Token obtained"

# Upload each video
TOTAL=0
SUCCESS=0
FAILED=0

for video in "$VIDEO_DIR"/*.mp4; do
  TOTAL=$((TOTAL + 1))
  filename=$(basename "$video")
  # Generate a random-ish title
  titles=(
    "精彩瞬间" "每日精选" "热门推荐" "不容错过" "超清画质"
    "治愈日常" "搞笑合集" "创意短片" "经典回顾" "独家放送"
    "萌宠时刻" "美食诱惑" "旅行日记" "音乐现场" "科技前沿"
    "游戏实况" "动漫剪辑" "影视解说" "运动健身" "学习分享"
    "手工制作" "摄影作品" "城市风光" "自然美景" "人文纪实"
    "街拍时尚" "美妆教程" "读书心得" "生活妙招" "数码评测"
  )
  title="${titles[$((RANDOM % 30))]} - 第${TOTAL}期"

  echo -n "[$TOTAL] $filename -> $title ... "

  RESP=$(curl -s -X POST "$BASE_URL/videos/upload" \
    -H "Authorization: Bearer $TOKEN" \
    -F "title=$title" \
    -F "description=测试视频 $TOTAL，来自 $filename" \
    -F "tags=测试,自动上传" \
    -F "video=@$video" \
    --max-time 600 2>&1)

  CODE=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code', -1))" 2>/dev/null || echo "-1")

  if [ "$CODE" = "0" ]; then
    SUCCESS=$((SUCCESS + 1))
    echo "OK"
  else
    FAILED=$((FAILED + 1))
    echo "FAILED: $(echo "$RESP" | head -c 200)"
  fi
done

echo ""
echo "=== Done ==="
echo "Total: $TOTAL, Success: $SUCCESS, Failed: $FAILED"
