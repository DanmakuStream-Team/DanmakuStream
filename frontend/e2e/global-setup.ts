import { execFileSync, execSync } from 'node:child_process'
import { request as playwrightRequest } from '@playwright/test'
import { API, USERS } from './test-data'

type ApiContext = Awaited<ReturnType<typeof playwrightRequest.newContext>>

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
}

async function prepareUC13Data(api: ApiContext) {
  for (const user of [USERS.target, USERS.moderator, USERS.admin, USERS.plain]) {
    await ensureUser(api, user.nickname, user.password)
  }

  const mysqlCommand = process.env.MYSQL_CMD
    ?? (process.platform === 'linux' ? 'mysql -S /home/haoyue/dms-mysql.sock -uroot -ppassword danmakustream' : '')
  if (!mysqlCommand) throw new Error('UC13 E2E setup requires MYSQL_CMD on this platform')

  execSync(
    `${mysqlCommand} -e "
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
}

async function prepareUserDomainData(api: ApiContext) {
  for (const user of [USERS.domainViewer, USERS.domainCreator, USERS.domainOther]) {
    await ensureUser(api, user.nickname, user.password)
  }

  const mysqlCommand = process.env.MYSQL_CMD
    ?? (process.platform === 'linux' ? 'mysql -S /home/haoyue/dms-mysql.sock -uroot -ppassword danmakustream' : '')
  if (!mysqlCommand && !process.env.MYSQL_TEST_CLIENT) {
    throw new Error('UC01/07/08/11 E2E setup requires MYSQL_CMD or MYSQL_TEST_CLIENT on this platform')
  }

  const sql = `
    SET @viewer=(SELECT id FROM users WHERE nickname='${USERS.domainViewer.nickname}');
    SET @creator=(SELECT id FROM users WHERE nickname='${USERS.domainCreator.nickname}');
    SET @other=(SELECT id FROM users WHERE nickname='${USERS.domainOther.nickname}');
    UPDATE users SET role='user' WHERE id IN (@viewer,@other);
    UPDATE users SET role='creator' WHERE id=@creator;
    DELETE FROM chat_messages WHERE sender_id IN (@viewer,@creator,@other) OR receiver_id IN (@viewer,@creator,@other);
    DELETE FROM subscription_orders WHERE subscriber_id IN (@viewer,@other) OR creator_id=@creator;
    DELETE FROM creator_subscriptions WHERE subscriber_id IN (@viewer,@other) OR creator_id=@creator;
    DELETE FROM creator_membership_plans WHERE creator_id=@creator;
    DELETE FROM notifications WHERE user_id IN (@viewer,@creator,@other) OR actor_id IN (@viewer,@creator,@other);
    DELETE FROM user_blocks WHERE blocker_id IN (@viewer,@creator,@other) OR blocked_id IN (@viewer,@creator,@other);
    DELETE FROM follows WHERE follower_id IN (@viewer,@creator,@other) OR followee_id IN (@viewer,@creator,@other);
    DELETE FROM follow_groups WHERE owner_id IN (@viewer,@creator,@other);
    INSERT INTO creator_membership_plans
      (created_at,updated_at,creator_id,price_cents,benefits,enabled)
      VALUES (NOW(),NOW(),@creator,600,'E2E 会员权益',TRUE);`
  if (process.env.MYSQL_TEST_CLIENT) {
    execFileSync(process.env.MYSQL_TEST_CLIENT, ['-e', sql], { stdio: 'pipe' })
  } else {
    execSync(`${mysqlCommand} -e "${sql}"`, { stdio: 'pipe' })
  }
}

export default async function globalSetup() {
  const api = await playwrightRequest.newContext()
  try {
    await prepareEngagementData(api)
    if (process.env.E2E_RUN_USER_DOMAIN === '1') await prepareUserDomainData(api)
    if (process.env.E2E_SKIP_UC13_SETUP !== '1') await prepareUC13Data(api)
  } finally {
    await api.dispose()
  }
}
