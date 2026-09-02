import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

const runTag = Date.now()
const VIDEO_TITLE = 'E2E-MC-公开视频'

function extractVideoList(payload: any): Array<{ id?: number; title?: string; status?: string }> {
  const candidates: unknown[] = []
  if (Array.isArray(payload)) candidates.push(payload)
  if (payload && typeof payload === 'object') {
    const d = (payload as any).data
    if (Array.isArray(d)) candidates.push(d)
    if (d && typeof d === 'object') {
      if (Array.isArray((d as any).list)) candidates.push((d as any).list)
      if (Array.isArray((d as any).items)) candidates.push((d as any).items)
      if (Array.isArray((d as any).records)) candidates.push((d as any).records)
    }
    if (Array.isArray((payload as any).list)) candidates.push((payload as any).list)
    if (Array.isArray((payload as any).items)) candidates.push((payload as any).items)
  }
  for (const c of candidates) {
    if (Array.isArray(c) && c.length > 0) return c as Array<{ id?: number; title?: string }>
  }
  return []
}

async function findVideoId(request: import('@playwright/test').APIRequestContext): Promise<number> {
  const byKeyword = await request.get(`${API}/videos?page=1&pageSize=100&keyword=${encodeURIComponent(VIDEO_TITLE)}`)
  const byKeywordBody = await byKeyword.json().catch(() => ({}))
  const list1 = extractVideoList(byKeywordBody)
  const exactTitle = list1.find((v) => v.title === VIDEO_TITLE && (!v.status || v.status === 'approved'))
  if (exactTitle?.id) return Number(exactTitle.id)

  const all = await request.get(`${API}/videos?page=1&pageSize=100`)
  const allBody = await all.json().catch(() => ({}))
  const list2 = extractVideoList(allBody)
  const byTitle2 = list2.find((v) => v.title === VIDEO_TITLE && (!v.status || v.status === 'approved'))
  if (byTitle2?.id) return Number(byTitle2.id)

  const fallback = list2.find((v) => v.id && (!v.status || v.status === 'approved'))
  if (fallback?.id) return Number(fallback.id)
  return 0
}

test.describe('UC06 个人资料库（观看历史/进度/稍后再看）（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC06-01 播放页触发历史记录与进度写入并持久化', async ({ page, request }) => {
    const user = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
    const videoId = await findVideoId(request)
    expect(videoId).toBeGreaterThan(0)

    await openAs(page, user, `/video/${videoId}`)
    await expect(page.getByRole('heading', { name: VIDEO_TITLE })).toBeVisible()
    await page.evaluate(() => {
      window.dispatchEvent(new CustomEvent('video-watch-progress', { detail: { position: 42, duration: 120 } }))
    })
    await page.waitForTimeout(500)
    await page.reload()
    await openAs(page, user, '/user/library')
    const historyCard = page.locator('.library-history .video-card', { hasText: VIDEO_TITLE }).first()
    await expect(historyCard).toBeVisible({ timeout: 10_000 })
    await expect(historyCard.locator('.progress-bar')).toBeVisible()

    const historyApi = await request.get(`${API}/users/me/history?page=1&pageSize=10`, {
      headers: { Authorization: `Bearer ${user.token}` },
    })
    expect(historyApi.ok()).toBeTruthy()
    const historyList = (await historyApi.json()).data.list as Array<{ videoId: number; progress: number }>
    expect(historyList.some((h) => h.videoId === videoId && h.progress >= 42)).toBe(true)
  })

  test('E2E-TC06-02 稍后再看添加/删除，UNIQUE 约束保障幂等', async ({ page, request }) => {
    const user = await loginViaApi(request, USERS.plain.nickname, USERS.plain.password)
    const videoId = await findVideoId(request)
    expect(videoId).toBeGreaterThan(0)

    const addOne = await request.post(`${API}/users/me/watch-later`, {
      headers: { Authorization: `Bearer ${user.token}` },
      data: { videoId },
    })
    expect(addOne.ok()).toBeTruthy()
    const addAgain = await request.post(`${API}/users/me/watch-later`, {
      headers: { Authorization: `Bearer ${user.token}` },
      data: { videoId },
    })
    expect(addAgain.ok()).toBeTruthy()

    await openAs(page, user, '/user/library')
    await page.getByRole('tab', { name: '稍后再看' }).click()
    const card = page.locator('.library-watch-later .video-card', { hasText: VIDEO_TITLE }).first()
    await expect(card).toBeVisible()
    const countBefore = (await (await request.get(`${API}/users/me/watch-later?page=1&pageSize=100`, {
      headers: { Authorization: `Bearer ${user.token}` },
    })).json()).data.list as Array<{ videoId: number }>
    expect(countBefore.filter((c) => c.videoId === videoId)).toHaveLength(1)

    await card.getByRole('button', { name: '从稍后再看移除' }).click()
    await expect(card).not.toBeVisible()
    const countAfter = (await (await request.get(`${API}/users/me/watch-later?page=1&pageSize=100`, {
      headers: { Authorization: `Bearer ${user.token}` },
    })).json()).data.list as Array<{ videoId: number }>
    expect(countAfter.filter((c) => c.videoId === videoId)).toHaveLength(0)
  })

  test('E2E-TC06-03 资料库首页标签切换与空态提示', async ({ page, request }) => {
    const session = await loginViaApi(request, `micro-empty-${runTag}`, USERS.plain.password)
    await openAs(page, session, '/user/library')
    await expect(page.getByRole('heading', { name: '个人资料库' })).toBeVisible()
    await page.getByRole('tab', { name: '稍后再看' }).click()
    await expect(page.locator('.library-watch-later').getByText('稍后再看还是空的')).toBeVisible()
  })

  test('E2E-TC06-04 所有资料库接口位于 auth 中间件之后（鉴权保护）', async ({ request }) => {
    const h1 = await request.get(`${API}/users/me/history`)
    const h2 = await request.get(`${API}/users/me/watch-later`)
    const h3 = await request.post(`${API}/users/me/watch-later`, { data: { videoId: 1 } })
    const h4 = await request.delete(`${API}/users/me/watch-later/1`)
    expect(h1.status()).toBe(401)
    expect(h2.status()).toBe(401)
    expect(h3.status()).toBe(401)
    expect(h4.status()).toBe(401)
  })
})
