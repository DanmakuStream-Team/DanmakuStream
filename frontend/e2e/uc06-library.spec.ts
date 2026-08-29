import { expect, test, type APIRequestContext } from '@playwright/test'
import { loginViaApi, openAs, type Session } from './fixtures/auth'
import { API, ENGAGEMENT_VIDEO_TITLE, USERS } from './test-data'

type LibraryRecord = { video: { id: number; title: string }; position: number; progress: number }

async function findApprovedVideo(request: APIRequestContext) {
  const response = await request.get(`${API}/videos?page=1&pageSize=100&keyword=${encodeURIComponent(ENGAGEMENT_VIDEO_TITLE)}`)
  expect(response.ok(), `GET /videos returned ${response.status()}`).toBeTruthy()
  const list = (await response.json()).data?.list ?? []
  const video = list.find((item: { status: string }) => item.status === 'approved')
  expect(video, 'global setup 应准备一条 UC06 可用的 approved 视频').toBeTruthy()
  return video as { id: number; title: string }
}

function authHeaders(session: Session) {
  return { Authorization: `Bearer ${session.token}` }
}

test.describe('UC06 个人视频资料库管理', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC06-01 保存观看进度，在历史页展示并从 UI 删除', async ({ page, request }) => {
    const session = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const video = await findApprovedVideo(request)
    const headers = authHeaders(session)
    await request.delete(`${API}/users/me/history`, { headers })

    const save = await request.put(`${API}/users/me/history/${video.id}`, {
      headers,
      data: { position: 30 },
    })
    expect(save.ok(), `save history returned ${save.status()}`).toBeTruthy()

    await openAs(page, session, '/me/history')
    await expect(page.getByRole('heading', { name: '历史记录', exact: true })).toBeVisible()
    const item = page.locator('.library-item').filter({ hasText: video.title })
    await expect(item).toBeVisible()
    await expect(item.getByRole('button', { name: '继续观看', exact: true })).toBeVisible()

    const detail = await request.get(`${API}/users/me/history/${video.id}`, { headers })
    expect(detail.ok(), `get history returned ${detail.status()}`).toBeTruthy()
    expect((await detail.json()).data.position).toBe(30)

    await item.locator('.item-actions button').last().click()
    await expect(item).toHaveCount(0)
    const list = await request.get(`${API}/users/me/history?page=1&pageSize=100`, { headers })
    expect(((await list.json()).data.list as LibraryRecord[]).some((record) => record.video.id === video.id)).toBeFalsy()
  })

  test('E2E-TC06-02 从 UI 清空全部观看历史', async ({ page, request }) => {
    const session = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const video = await findApprovedVideo(request)
    const headers = authHeaders(session)
    await request.put(`${API}/users/me/history/${video.id}`, { headers, data: { position: 12 } })

    await openAs(page, session, '/me/history')
    await expect(page.locator('.library-item').filter({ hasText: video.title })).toBeVisible()
    await page.getByRole('button', { name: '清空', exact: true }).click()
    await page.getByRole('button', { name: '清空', exact: true }).last().click()

    await expect(page.getByText('暂无历史记录')).toBeVisible()
    const list = await request.get(`${API}/users/me/history?page=1&pageSize=100`, { headers })
    expect((await list.json()).data.list).toHaveLength(0)
  })

  test('E2E-TC06-03 稍后再看跨页面展示并从 UI 清空', async ({ page, request }) => {
    const session = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const video = await findApprovedVideo(request)
    const headers = authHeaders(session)
    await request.delete(`${API}/users/me/watch-later`, { headers })

    const add = await request.post(`${API}/users/me/watch-later/${video.id}`, { headers })
    expect(add.ok(), `add watch later returned ${add.status()}`).toBeTruthy()
    expect((await add.json()).data.saved).toBe(true)

    await openAs(page, session, '/me/watchlater')
    await expect(page.getByRole('heading', { name: '稍后再看', exact: true })).toBeVisible()
    await expect(page.locator('.library-item').filter({ hasText: video.title })).toBeVisible()

    const status = await request.get(`${API}/users/me/watch-later/${video.id}/status`, { headers })
    expect(status.ok(), `watch-later status returned ${status.status()}`).toBeTruthy()
    expect((await status.json()).data.saved).toBe(true)

    await page.getByRole('button', { name: '清空', exact: true }).click()
    await page.getByRole('button', { name: '清空', exact: true }).last().click()
    await expect(page.getByText('暂未添加稍后再看的视频')).toBeVisible()
    const list = await request.get(`${API}/users/me/watch-later?page=1&pageSize=100`, { headers })
    expect((await list.json()).data.list).toHaveLength(0)
  })

  test('E2E-TC06-04 未登录访问资料库会跳转登录页', async ({ page }) => {
    await page.goto('/me/history')
    await expect(page).toHaveURL(/\/login\?redirect=(?:%2F|\/)me(?:%2F|\/)history/)
    await expect(page.getByRole('heading', { name: '登录账号' })).toBeVisible()
  })
})
