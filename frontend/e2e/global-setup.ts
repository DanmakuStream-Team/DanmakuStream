import { execSync } from 'node:child_process'
import { API, USERS } from './test-data'

/**
 * E2E 测试数据初始化：
 * 1. 注册四个测试账号（已存在则忽略）并固化角色；
 * 2. 重置 tuser 角色（E2E-TC13-02 会修改它）；
 * 3. 造一条“待审 + 媒体就绪”视频和一条未屏蔽弹幕（E2E-TC13-01 消费）。
 * MySQL 连接可用 MYSQL_CMD 环境变量覆盖（默认本地用户态实例）。
 */
const MYSQL = process.env.MYSQL_CMD ?? 'mysql -S /home/haoyue/dms-mysql.sock -uroot -ppassword danmakustream'
const BACKEND = process.env.API_BASE ?? 'http://localhost:8080'

async function register(nickname: string, password: string) {
  const res = await fetch(`${BACKEND}${API}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ nickname, password }),
  })
  // 已存在（400）视为成功
  if (!res.ok && res.status !== 400) throw new Error(`register ${nickname} failed: HTTP ${res.status}`)
}

export default async function globalSetup() {
  for (const u of Object.values(USERS)) await register(u.nickname, u.password)

  execSync(
    `${MYSQL} -e "
    UPDATE users SET role='user'      WHERE nickname='${USERS.target.nickname}';
    UPDATE users SET role='moderator' WHERE nickname='${USERS.moderator.nickname}';
    UPDATE users SET role='admin'     WHERE nickname='${USERS.admin.nickname}';
    UPDATE users SET role='user'      WHERE nickname='${USERS.plain.nickname}';
    DELETE FROM videos   WHERE title LIKE 'E2E-UC13-%';
    DELETE FROM danmakus WHERE content LIKE 'E2E-UC13-%';
    DELETE FROM site_banners        WHERE title LIKE 'E2E-UC13-%';
    DELETE FROM site_announcements  WHERE content LIKE 'E2E-UC13-%';
    INSERT INTO videos (created_at,updated_at,title,video_url,status,author_id) VALUES
      (NOW(),NOW(),'E2E-UC13-待审视频','/data/videos/e2e.mp4','pending',(SELECT id FROM users WHERE nickname='${USERS.target.nickname}'));
    INSERT INTO danmakus (created_at,updated_at,video_id,user_id,content,time) VALUES
      (NOW(),NOW(),(SELECT id FROM videos WHERE title='E2E-UC13-待审视频'),
       (SELECT id FROM users WHERE nickname='${USERS.target.nickname}'),'E2E-UC13-待屏蔽弹幕',5);"`,
    { stdio: 'pipe' },
  )
  console.log('[globalSetup] UC13 E2E 数据就绪')
}
