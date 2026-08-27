#!/usr/bin/env bash
# 播放固定测试数据（幂等，可重复执行）：
#   1. 等待后端健康检查通过（GORM 此时已完成自动建表）
#   2. 通过真实注册接口创建 test_user / test_moderator / test_admin
#   3. 修正角色并插入示例视频/弹幕/横幅/公告
# 依赖：docker compose 已启动（所有命令通过 docker compose exec 执行，宿主机无需 mysql/curl）
set -euo pipefail
cd "$(dirname "$0")/.."

PASSWORD='Test1234!'
MARK='SEED'

log() { printf '[seed] %s\n' "$*"; }

# 1. 等待后端健康
for i in $(seq 1 30); do
  if docker compose exec -T backend wget -qO- http://localhost:8080/api/v1/health 2>/dev/null | grep -q '"status":"ok"'; then
    break
  fi
  [ "$i" = 30 ] && { log "backend 未就绪，放弃"; exit 1; }
  sleep 2
done
log "backend 健康"

# 2. 注册三个固定账号（已存在返回 400，忽略）
for nick in test_user test_moderator test_admin; do
  docker compose exec -T backend wget -qO- \
    --header='Content-Type: application/json' \
    --post-data="{\"nickname\":\"$nick\",\"password\":\"$PASSWORD\"}" \
    http://localhost:8080/api/v1/auth/register > /dev/null 2>&1 || true
done
log "账号注册完成（密码均为 $PASSWORD，仅限本地测试）"

mysql_exec() { docker compose exec -T mysql mysql -uroot -ppassword danmakustream -e "$1" 2>/dev/null; }

# 3. 角色固化 + 内容数据（按 SEED 标记清理重建，保证幂等）
mysql_exec "
UPDATE users SET role='user'      WHERE nickname='test_user';
UPDATE users SET role='moderator' WHERE nickname='test_moderator';
UPDATE users SET role='admin'     WHERE nickname='test_admin';
DELETE FROM danmakus        WHERE content LIKE '${MARK}-%';
DELETE FROM videos          WHERE title LIKE '${MARK}-%';
DELETE FROM site_banners    WHERE title LIKE '${MARK}-%';
DELETE FROM site_announcements WHERE content LIKE '${MARK}-%';
INSERT INTO videos (created_at,updated_at,title,description,video_url,status,author_id) VALUES
 (NOW(),NOW(),'${MARK}-待审核视频','等待审核的演示视频','/data/videos/seed.mp4','pending', (SELECT id FROM users WHERE nickname='test_user')),
 (NOW(),NOW(),'${MARK}-转码中视频','尚未完成转码的演示视频','','pending',             (SELECT id FROM users WHERE nickname='test_user')),
 (NOW(),NOW(),'${MARK}-已发布视频','已通过审核的演示视频','/data/videos/seed.mp4','approved',(SELECT id FROM users WHERE nickname='test_user'));
INSERT INTO danmakus (created_at,updated_at,video_id,user_id,content,time) VALUES
 (NOW(),NOW(),(SELECT id FROM videos WHERE title='${MARK}-已发布视频'),(SELECT id FROM users WHERE nickname='test_user'),'${MARK}-示例弹幕',5);
INSERT INTO site_banners (created_at,updated_at,title,image_url,link,enabled,sort) VALUES
 (NOW(),NOW(),'${MARK}-示例横幅','https://example.com/banner.png','/videos',1,0);
INSERT INTO site_announcements (created_at,updated_at,content,enabled,started_at,ended_at) VALUES
 (NOW(),NOW(),'${MARK}-示例公告：系统测试数据，可通过管理员账号在运营管理中维护。',1,NOW(),DATE_ADD(NOW(), INTERVAL 30 DAY));
"
log "测试数据就绪：账号 test_user/test_moderator/test_admin（密码 $PASSWORD），视频/弹幕/横幅/公告各含 SEED- 前缀示例"
