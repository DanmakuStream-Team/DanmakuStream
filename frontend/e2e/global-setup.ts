import { execSync } from 'node:child_process'
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { request as playwrightRequest } from '@playwright/test'
import { API, ENGAGEMENT_VIDEO_TITLE, USERS } from './test-data'

type ApiContext = Awaited<ReturnType<typeof playwrightRequest.newContext>>

function runSql(mysqlCommand: string, sql: string) {
  const directory = mkdtempSync(join(tmpdir(), 'danmakustream-e2e-'))
  const sqlFile = join(directory, 'setup.sql')
  try {
    writeFileSync(sqlFile, `SET NAMES utf8mb4;\n${sql}`, 'utf8')
    execSync(`${mysqlCommand} < "${sqlFile}"`, { stdio: 'pipe' })
  } finally {
    rmSync(directory, { recursive: true, force: true })
  }
}

async function ensureUser(api: ApiContext, nickname: string, password: string) {
  await api.post(`${API}/auth/register`, { data: { nickname, password } })
  const response = await api.post(`${API}/auth/login`, { data: { nickname, password } })
  if (!response.ok()) throw new Error(`cannot prepare ${nickname}: ${await response.text()}`)
  return (await response.json()).data as { token: string; userInfo: { id: number } }
}

async function prepareEngagementData(api: ApiContext) {
  const owner = await ensureUser(api, USERS.owner.nickname, USERS.owner.password)
  await ensureUser(api, USERS.viewer.nickname, USERS.viewer.password)
  const headers = { Authorization: `Bearer ${owner.token}` }

  const roomsResponse = await api.get(`${API}/live?page=1&pageSize=100`)
  if (roomsResponse.ok()) {
    const rooms = (await roomsResponse.json()).data?.list || []
    for (const room of rooms) {
      if (room.ownerId === owner.userInfo.id) await api.put(`${API}/live/${room.id}/end`, { headers })
    }
  }

  const schedulesResponse = await api.get(`${API}/live-schedules?status=pending&page=1&pageSize=100`, { headers })
  if (schedulesResponse.ok()) {
    const schedules = (await schedulesResponse.json()).data?.list || []
    for (const schedule of schedules) {
      if (schedule.ownerId === owner.userInfo.id) await api.delete(`${API}/live-schedules/${schedule.id}`, { headers })
    }
  }

  const mysqlCommand = process.env.MYSQL_CMD
  if (mysqlCommand) {
    runSql(mysqlCommand, `
      UPDATE users SET role='creator' WHERE nickname='${USERS.owner.nickname}';
      SET FOREIGN_KEY_CHECKS=0;
      DELETE FROM comment_likes WHERE comment_id IN (
        SELECT id FROM comments WHERE video_id IN (SELECT id FROM videos WHERE title='${ENGAGEMENT_VIDEO_TITLE}')
      );
      DELETE FROM comments WHERE video_id IN (SELECT id FROM videos WHERE title='${ENGAGEMENT_VIDEO_TITLE}');
      DELETE FROM danmakus WHERE video_id IN (SELECT id FROM videos WHERE title='${ENGAGEMENT_VIDEO_TITLE}') AND scene='video';
      DELETE FROM likes WHERE video_id IN (SELECT id FROM videos WHERE title='${ENGAGEMENT_VIDEO_TITLE}');
      DELETE FROM collects WHERE video_id IN (SELECT id FROM videos WHERE title='${ENGAGEMENT_VIDEO_TITLE}');
      DELETE FROM videos WHERE title='${ENGAGEMENT_VIDEO_TITLE}';
      SET FOREIGN_KEY_CHECKS=1;
      INSERT INTO videos (created_at,updated_at,title,description,video_url,status,author_id)
      VALUES (NOW(),NOW(),'${ENGAGEMENT_VIDEO_TITLE}','UC05 自动化测试数据','/data/videos/e2e-uc05.mp4','approved',
        (SELECT id FROM users WHERE nickname='${USERS.owner.nickname}'));
    `)
  }
}

async function prepareUC13Data(api: ApiContext) {
  for (const user of [USERS.target, USERS.moderator, USERS.admin, USERS.plain]) {
    await ensureUser(api, user.nickname, user.password)
  }

  const mysqlCommand = process.env.MYSQL_CMD
    ?? (process.platform === 'linux' ? 'mysql -S /home/haoyue/dms-mysql.sock -uroot -ppassword danmakustream' : '')
  if (!mysqlCommand) throw new Error('UC13 E2E setup requires MYSQL_CMD on this platform')

  runSql(mysqlCommand, `
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
       (SELECT id FROM users WHERE nickname='${USERS.target.nickname}'),'E2E-UC13-待屏蔽弹幕',5);
  `)
}

export default async function globalSetup() {
  const api = await playwrightRequest.newContext()
  try {
    await prepareEngagementData(api)
    if (process.env.E2E_SKIP_UC13_SETUP !== '1') await prepareUC13Data(api)
  } finally {
    await api.dispose()
  }
}
