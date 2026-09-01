import { execSync } from 'node:child_process'
import { copyFileSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path, { join } from 'node:path'
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
    const videoDir = process.env.VIDEO_DIR ?? path.resolve('../backend/data')
    const mediaFixture = path.join(videoDir, 'videos', 'e2e-uc05.mp4')
    mkdirSync(path.dirname(mediaFixture), { recursive: true })
    writeFileSync(mediaFixture, 'UC05 E2E media fixture', 'utf8')

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
      VALUES (NOW(),NOW(),'${ENGAGEMENT_VIDEO_TITLE}','UC05 自动化测试数据','/media/videos/e2e-uc05.mp4','approved',
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

  const videoDir = process.env.VIDEO_DIR ?? path.resolve('../backend/data')
  const mediaFixture = path.join(videoDir, 'videos', 'e2e-uc13.mp4')
  mkdirSync(path.dirname(mediaFixture), { recursive: true })
  writeFileSync(mediaFixture, 'UC13 E2E media fixture', 'utf8')

  runSql(mysqlCommand, `
    UPDATE users SET role='user'      WHERE nickname='${USERS.target.nickname}';
    UPDATE users SET role='moderator' WHERE nickname='${USERS.moderator.nickname}';
    UPDATE users SET role='admin'     WHERE nickname='${USERS.admin.nickname}';
    UPDATE users SET role='user'      WHERE nickname='${USERS.plain.nickname}';
    DELETE FROM videos   WHERE title LIKE 'E2E-UC13-%';
    DELETE FROM danmakus WHERE content LIKE 'E2E-UC13-%';
    DELETE FROM site_banners        WHERE title LIKE 'E2E-UC13-%';
    DELETE FROM site_announcements  WHERE content LIKE 'E2E-UC13-%';
    INSERT INTO videos (created_at,updated_at,title,video_url,status,transcode_status,author_id) VALUES
      (NOW(),NOW(),'E2E-UC13-待审视频','/media/videos/e2e-uc13.mp4','pending','ready',
       (SELECT id FROM users WHERE nickname='${USERS.target.nickname}'));
    INSERT INTO danmakus (created_at,updated_at,video_id,user_id,content,time) VALUES
      (NOW(),NOW(),(SELECT id FROM videos WHERE title='E2E-UC13-待审视频'),
       (SELECT id FROM users WHERE nickname='${USERS.target.nickname}'),'E2E-UC13-待屏蔽弹幕',5);
  `)
}

async function prepareMemberCData(api: ApiContext) {
  for (const user of [USERS.memberCCreator, USERS.memberCModerator, USERS.memberCPlain]) {
    await ensureUser(api, user.nickname, user.password)
  }

  const mysqlCommand = process.env.MYSQL_CMD
    ?? (process.platform === 'linux' ? 'mysql -S /home/haoyue/dms-mysql.sock -uroot -ppassword danmakustream' : '')
  if (!mysqlCommand) throw new Error('member C E2E setup requires MYSQL_CMD')

  const fixtureDir = '/tmp/danmakustream-member-c-fixtures'
  const backendMediaDir = path.resolve('../backend/data/videos/e2e-member-c')
  mkdirSync(fixtureDir, { recursive: true })
  mkdirSync(backendMediaDir, { recursive: true })
  const fixture = path.join(fixtureDir, 'member-c.mp4')
  const mediaFixture = path.join(backendMediaDir, 'fixture.mp4')
  const ffmpeg = process.env.FFMPEG_BIN ?? 'ffmpeg'
  execSync(
    `${ffmpeg} -hide_banner -loglevel error -y -f lavfi -i color=c=blue:s=320x180:d=1 -f lavfi -i anullsrc=r=44100:cl=stereo -shortest -c:v libx264 -pix_fmt yuv420p -c:a aac "${fixture}"`,
    { stdio: 'pipe' },
  )
  copyFileSync(fixture, mediaFixture)

  runSql(mysqlCommand, `
    UPDATE users SET role='creator'   WHERE nickname='${USERS.memberCCreator.nickname}';
    UPDATE users SET role='moderator' WHERE nickname='${USERS.memberCModerator.nickname}';
    UPDATE users SET role='user'      WHERE nickname='${USERS.memberCPlain.nickname}';
    DELETE FROM video_daily_stats WHERE creator_id=(SELECT id FROM users WHERE nickname='${USERS.memberCCreator.nickname}');
    DELETE FROM creator_daily_stats WHERE creator_id=(SELECT id FROM users WHERE nickname='${USERS.memberCCreator.nickname}');
    DELETE FROM videos WHERE title LIKE 'E2E-MC-%';
    INSERT INTO videos (created_at,updated_at,title,description,video_url,status,transcode_status,author_id,view_count,collect_count,category,tags) VALUES
      (NOW(),NOW(),'E2E-MC-公开视频','成员 C 搜索播放用例','/media/videos/e2e-member-c/fixture.mp4','approved','ready',(SELECT id FROM users WHERE nickname='${USERS.memberCCreator.nickname}'),20,4,'tech','e2e,member-c'),
      (NOW(),NOW(),'E2E-MC-待审核通过','成员 C 审核通过用例','/media/videos/e2e-member-c/fixture.mp4','pending','ready',(SELECT id FROM users WHERE nickname='${USERS.memberCCreator.nickname}'),3,1,'tech','e2e'),
      (NOW(),NOW(),'E2E-MC-待审核拒绝','成员 C 审核拒绝用例','/media/videos/e2e-member-c/fixture.mp4','pending','ready',(SELECT id FROM users WHERE nickname='${USERS.memberCCreator.nickname}'),2,0,'life','e2e');
    INSERT INTO creator_daily_stats (created_at,updated_at,creator_id,date,view_delta,collect_delta,stream_count) VALUES
      (NOW(),NOW(),(SELECT id FROM users WHERE nickname='${USERS.memberCCreator.nickname}'),DATE_FORMAT(CURDATE(),'%Y-%m-%d'),10,2,1),
      (NOW(),NOW(),(SELECT id FROM users WHERE nickname='${USERS.memberCCreator.nickname}'),DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 1 DAY),'%Y-%m-%d'),5,1,0);
    INSERT INTO video_daily_stats (created_at,updated_at,creator_id,video_id,date,view_delta,collect_delta) VALUES
      (NOW(),NOW(),(SELECT id FROM users WHERE nickname='${USERS.memberCCreator.nickname}'),(SELECT id FROM videos WHERE title='E2E-MC-公开视频'),DATE_FORMAT(CURDATE(),'%Y-%m-%d'),6,1),
      (NOW(),NOW(),(SELECT id FROM users WHERE nickname='${USERS.memberCCreator.nickname}'),(SELECT id FROM videos WHERE title='E2E-MC-公开视频'),DATE_FORMAT(DATE_SUB(CURDATE(),INTERVAL 1 DAY),'%Y-%m-%d'),3,1);
  `)
}

async function prepareMemberBData(api: ApiContext) {
  const owner = await ensureUser(api, USERS.owner.nickname, USERS.owner.password)
  const viewer = await ensureUser(api, USERS.viewer.nickname, USERS.viewer.password)
  const mysqlCommand = process.env.MYSQL_CMD
  if (!mysqlCommand) throw new Error('member B E2E setup requires MYSQL_CMD')

  runSql(mysqlCommand, `
    SET FOREIGN_KEY_CHECKS=0;
    UPDATE users SET role='creator' WHERE id=${owner.userInfo.id};
    UPDATE users SET role='user' WHERE id=${viewer.userInfo.id};
    DELETE FROM chat_messages WHERE sender_id IN (${owner.userInfo.id},${viewer.userInfo.id}) OR receiver_id IN (${owner.userInfo.id},${viewer.userInfo.id});
    DELETE FROM notifications WHERE user_id IN (${owner.userInfo.id},${viewer.userInfo.id}) OR actor_id IN (${owner.userInfo.id},${viewer.userInfo.id});
    DELETE FROM subscription_orders WHERE subscriber_id=${viewer.userInfo.id} OR creator_id=${owner.userInfo.id};
    DELETE FROM creator_subscriptions WHERE subscriber_id=${viewer.userInfo.id} OR creator_id=${owner.userInfo.id};
    DELETE FROM creator_membership_plans WHERE creator_id=${owner.userInfo.id};
    DELETE FROM follows WHERE follower_id IN (${owner.userInfo.id},${viewer.userInfo.id}) OR followee_id IN (${owner.userInfo.id},${viewer.userInfo.id});
    DELETE FROM follow_groups WHERE owner_id=${viewer.userInfo.id};
    DELETE FROM user_blocks WHERE blocker_id IN (${owner.userInfo.id},${viewer.userInfo.id}) OR blocked_id IN (${owner.userInfo.id},${viewer.userInfo.id});
    DELETE FROM dynamic_posts WHERE user_id IN (${owner.userInfo.id},${viewer.userInfo.id});
    DELETE FROM videos WHERE title LIKE 'E2E-MEMBER-B-%';
    INSERT INTO videos (created_at,updated_at,title,description,video_url,status,transcode_status,author_id)
    VALUES (NOW(),NOW(),'E2E-MEMBER-B-分享视频','UC11 视频分享夹具','/media/videos/e2e-member-b.mp4','approved','ready',${owner.userInfo.id});
    SET FOREIGN_KEY_CHECKS=1;
  `)
}

export default async function globalSetup() {
  const api = await playwrightRequest.newContext()
  try {
    // 统一 E2E：三域数据全部就绪（各域数据按前缀隔离、可重复执行）
    await prepareEngagementData(api)
    await prepareUC13Data(api)
    await prepareMemberCData(api)
    await prepareMemberBData(api)
  } finally {
    await api.dispose()
  }
}
