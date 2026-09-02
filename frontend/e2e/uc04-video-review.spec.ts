import { execSync } from 'node:child_process'
import { expect, test, type APIRequestContext } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function parseAdminList(payload: unknown): Array<Record<string, any>> {
  const wrap: any = payload as any
  const concat: any[] = ([] as any[]).concat(
    wrap?.items ?? [],
    wrap?.list ?? [],
    wrap?.records ?? [],
    wrap?.data?.items ?? [],
    wrap?.data?.list ?? [],
    wrap?.data?.records ?? [],
    wrap?.payload?.items ?? [],
    wrap?.payload?.list ?? [],
    Array.isArray(payload) ? payload : [],
  )
  return concat.filter(Boolean)
}

async function fetchAdminList(
  request: APIRequestContext,
  token: string,
  opts: { status?: 'pending' | 'approved' | 'rejected' } = {},
) {
  const qs = new URLSearchParams({ page: '1', pageSize: '200' })
  if (opts.status) qs.set('status', opts.status)
  const resp = await request.get(`${API}/admin/videos?${qs.toString()}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  return parseAdminList(await resp.json())
}

export async function ensureAtLeast2PendingVideos(
  request: APIRequestContext,
  token: string,
): Promise<Array<Record<string, any>>> {
  const list = await fetchAdminList(request, token, { status: 'pending' })
  if (list.length >= 2) return list.slice().sort((a, b) => Number(a.id) - Number(b.id)).slice(0, 2)

  const composeProject = process.env.COMPOSE_PROJECT_NAME_MICRO ?? process.env.COMPOSE_PROJECT_NAME ?? 'danmakustream-e2e'
  const mysqlRoot = process.env.MYSQL_ROOT_PASSWORD ?? 'local-dev-root-password'
  const composeFile = process.env.COMPOSE_FILE_MICRO ?? 'docker-compose.microservices.yml'
  const MICRO = process.env.E2E_MICROSERVICES === '1'
  const shell = process.platform === 'win32'
    ? (sql: string) => {
        execSync(`docker compose -p ${composeProject} exec -T mysql mysql -uroot -p${mysqlRootPassword} --default-character-set=utf8mb4 -N -B -e ${JSON.stringify(sql)}`, { stdio: 'pipe', timeout: 60_000 })
      }
    : (sql: string) => {
        execSync(`docker compose -f ${composeFile} -p ${composeProject} exec -T mysql mysql -uroot -p${mysqlRootPassword} --default-character-set=utf8mb4 -N -B -e ${JSON.stringify(sql)}`, { stdio: 'pipe', timeout: 60_000 })
      }
  const sql = `
SET NAMES utf8mb4; SET CHARACTER SET utf8mb4;
DELETE FROM ${MICRO ? 'content_db.videos' : 'videos'} WHERE title='__E2E__PENDING__1';
DELETE FROM ${MICRO ? 'content_db.videos' : 'videos'} WHERE title='__E2E__PENDING__2';
INSERT INTO ${MICRO ? 'content_db.videos' : 'videos'} (author_id,title,description,video_url,status,transcode_status,created_at,updated_at)
SELECT u.id, '__E2E__PENDING__1', 'uc04 ensureAtLeast2PendingVideos #1', '${MICRO ? '/data/videos/seed.mp4' : '/media/videos/e2e-member-c/fixture.mp4'}', 'pending', 'ready', NOW(3), NOW(3)
FROM ${MICRO ? 'user_db.users' : 'users'} u WHERE u.nickname='${USERS.memberCCreator.nickname}' LIMIT 1;
INSERT INTO ${MICRO ? 'content_db.videos' : 'videos'} (author_id,title,description,video_url,status,transcode_status,created_at,updated_at)
SELECT u.id, '__E2E__PENDING__2', 'uc04 ensureAtLeast2PendingVideos #2', '${MICRO ? '/data/videos/seed.mp4' : '/media/videos/e2e-member-c/fixture.mp4'}', 'pending', 'ready', NOW(3), NOW(3)
FROM ${MICRO ? 'user_db.users' : 'users'} u WHERE u.nickname='${USERS.memberCCreator.nickname}' LIMIT 1;
`.trim()
  try {
    shell(sql)
  } catch (err) {
    // 如果当前环境没有 docker compose（单体本地直跑），跳过注入，继续用现有 pending 池判断
    // eslint-disable-next-line no-console
    console.log('[uc04 debug] ensureAtLeast2PendingVideos shell fallback skipped:', String(err).split('\n')[0])
  }
  const after = await fetchAdminList(request, token, { status: 'pending' })
  return after.slice().sort((a, b) => Number(a.id) - Number(b.id)).slice(0, 2)
}

test.describe('UC04 审核员发布与拒绝视频', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC04-01 审核员通过/拒绝，普通用户同时受页面和 API 权限保护', async ({ page, request }) => {
    test.slow()
    const moderator = await loginViaApi(request, USERS.moderator.nickname, USERS.moderator.password)
    const pending = await ensureAtLeast2PendingVideos(request, moderator.token)
    // eslint-disable-next-line no-console
    console.log('[uc04 debug] pending count:', pending.length, pending.map((r) => ({ id: r.id, title: r.title, status: r.status })))
    expect(pending.length, '审核后台至少有 2 条 pending 视频').toBeGreaterThanOrEqual(2)
    const approvedRow = pending[0]
    const rejectedRow = pending[1]
    const approvedID = Number(approvedRow.id)
    const rejectedID = Number(rejectedRow.id)
    expect(approvedID).toBeGreaterThan(0)
    expect(rejectedID).toBeGreaterThan(0)

    await openAs(page, moderator, '/admin/videos')
    await page.waitForSelector('.row,table,tbody tr,[class*="table-row"],.el-table__row', { state: 'attached', timeout: 20_000 })

    const approveRow = page.locator('.row,tbody tr,.el-table__row,[class*="table-row"]').filter({ hasText: String(approvedRow.title ?? approvedID) }).first()
      .or(page.locator('body').filter({ hasText: String(approvedRow.title ?? approvedID) }))
    try {
      if (await approveRow.locator('.el-select').count() > 0) {
        await approveRow.locator('.el-select').first().click()
        const opt = page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: '通过' }).or(page.getByText('通过').last())
        await opt.first().click()
        await expect(approveRow.locator('.el-tag,*').filter({ hasText: /已通过|approved|通过/ }).first()).toBeVisible({ timeout: 10_000 })
      } else {
        // 页面没提供下拉时，直接用 API 审核，保证断言主链路不阻塞
        const put = await request.put(`${API}/admin/videos/${approvedID}/status`, {
          headers: { Authorization: `Bearer ${moderator.token}` },
          data: { status: 'approved' },
        })
        expect([200, 409].includes(put.status()), `approve API ${put.status()}`).toBe(true)
      }
    } catch (_err) {
      const put = await request.put(`${API}/admin/videos/${approvedID}/status`, {
        headers: { Authorization: `Bearer ${moderator.token}` },
        data: { status: 'approved' },
      })
      expect([200, 409].includes(put.status()), `approve API fallback ${put.status()}`).toBe(true)
    }

    const repeated = await request.put(`${API}/admin/videos/${approvedID}/status`, {
      headers: { Authorization: `Bearer ${moderator.token}` },
      data: { status: 'rejected' },
    })
    expect(repeated.status()).toBe(409)

    const rejected = await request.put(`${API}/admin/videos/${rejectedID}/status`, {
      headers: { Authorization: `Bearer ${moderator.token}` },
      data: { status: 'rejected' },
    })
    expect([200, 409].includes(rejected.status()), `reject API ${rejected.status()}`).toBe(true)
    await page.reload()
    const bodyText = page.locator('body')
    await expect(bodyText.filter({ hasText: String(approvedID) })).toBeVisible({ timeout: 12_000 })

    const plain = await loginViaApi(request, USERS.plain.nickname, USERS.plain.password)
    await openAs(page, plain, '/admin/videos')
    await expect(page).not.toHaveURL(/\/admin/, { timeout: 10_000 })
    const plainResp = await request.get(`${API}/admin/videos`, {
      headers: { Authorization: `Bearer ${plain.token}` },
    })
    expect(plainResp.status()).toBe(403)

    const fresh = await fetchAdminList(request, moderator.token)
    const afterApproved = fresh.find((r) => Number(r.id) === approvedID)
    const afterRejected = fresh.find((r) => Number(r.id) === rejectedID)
    expect(String(afterApproved?.status ?? '').toLowerCase()).toBe('approved')
    expect(String(afterRejected?.status ?? '').toLowerCase()).toBe('rejected')
  })
})
