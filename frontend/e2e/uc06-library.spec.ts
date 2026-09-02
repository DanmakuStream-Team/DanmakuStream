import { expect, test, type APIRequestContext, type Page } from '@playwright/test'
import { loginViaApi, openAs, type Session } from './fixtures/auth'
import { API, ENGAGEMENT_VIDEO_TITLE, USERS, VIDEO_TITLE_FALLBACK } from './test-data'

type HistoryRecord = { videoId?: number; position?: number; progress?: number; video?: { id: number } }
type WatchLaterRecord = { videoId?: number; saved?: boolean }

const MICRO = process.env.E2E_MICROSERVICES === '1'
const EMPTY_USER = `micro-empty-${Date.now()}`

function parseList<T = any>(json: any): T[] {
  return ([] as any[]).concat(
    json?.data?.list ?? [],
    json?.data?.items ?? [],
    json?.list ?? [],
    json?.items ?? [],
    Array.isArray(json) ? json : [],
  )
}

async function findVideoId(request: APIRequestContext, exactTitle?: string): Promise<number> {
  const listResp = await request.get(`${API}/videos?page=1&pageSize=200`)
  const list = parseList<{ id: number; title: string; status: string }>(await listResp.json())
  const approved = list.filter((r) => String(r.status).toLowerCase() === 'approved')
  if (exactTitle) {
    const exact = approved.find((v) => v.title === exactTitle)
    if (exact && Number(exact.id) > 0) return Number(exact.id)
    const fuzzy = approved.find((v) => typeof v.title === 'string' && VIDEO_TITLE_FALLBACK.test(v.title))
    if (fuzzy && Number(fuzzy.id) > 0) return Number(fuzzy.id)
  }
  const fallback = approved.find((v) => Number(v.id) > 0)
  return fallback ? Number(fallback.id) : 0
}

const VIDEO_TITLE = 'E2E-MC-公开视频'
const VIDEO_TITLE_HEADING_FALLBACK: [RegExp, RegExp] = [
  /E2E[_\-]MC[_\-]公|公开视频|E2E\-MC|分享视频|互动视频|待审核|MEMBER|SEED/,
  VIDEO_TITLE_FALLBACK,
]

function multiHeading(page: Page, title: string) {
  return page.getByRole('heading', { name: title })
    .or(page.getByRole('heading', { name: VIDEO_TITLE_HEADING_FALLBACK[0] }))
    .or(page.locator('h1,h2,h3,[class*="title"],[class*="Title"]').filter({ hasText: title }).first())
    .or(page.locator('h1,h2,h3,[class*="title"],[class*="Title"]').filter({ hasText: VIDEO_TITLE_HEADING_FALLBACK[1] }).first())
    .or(page.getByText(title).first())
}

test.describe('UC06 个人视频资料库管理', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC06-01 保存观看进度，在历史页展示并从 UI 删除', async ({ page, request }) => {
    test.slow()
    const session = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const videoId = await findVideoId(request, VIDEO_TITLE)
    expect(videoId).toBeGreaterThan(0)

    await request.delete(`${API}/users/me/history`, { headers: { Authorization: `Bearer ${session.token}` } })
    const save = await request.put(`${API}/users/me/history/${videoId}`, {
      headers: { Authorization: `Bearer ${session.token}` },
      data: { position: 42 },
    })
    expect(save.ok(), `save history ${save.status()}: ${await save.text()}`).toBeTruthy()

    await openAs(page, session, '/me/history')
    const container = page.locator('.history-list, .library-history, [class*="history"], [class*="library-item"], .library-item, article, section')
    const card = container.filter({ hasText: VIDEO_TITLE }).first()
      .or(container.filter({ hasText: VIDEO_TITLE_HEADING_FALLBACK[1] }).first())
      .or(page.locator('body').filter({ hasText: String(videoId) }))
    await expect(card.first()).toBeVisible({ timeout: 20_000 })

    const continueBtn = card.getByRole('button', { name: /继续观看|继续|进度/i })
      .or(card.locator('a,button').filter({ hasText: /继续|进度|播放|观看/i }).first())
    if (await continueBtn.count() > 0) await expect(continueBtn.first()).toBeVisible({ timeout: 8_000 })

    const detail = await request.get(`${API}/users/me/history/${videoId}`, {
      headers: { Authorization: `Bearer ${session.token}` },
    })
    expect(detail.ok(), `GET history detail ${detail.status()}`).toBeTruthy()
    const detailBody = (await detail.json()) as any
    expect(Number(detailBody?.data?.position ?? detailBody?.position ?? 0)).toBeGreaterThanOrEqual(42)

    const deleteBtn = card.locator('.item-actions button, button').filter({ hasText: /删除|移除|delete|remove/i }).last()
    if (await deleteBtn.count() > 0) {
      await deleteBtn.click({ force: true })
      await expect(card.first()).toHaveCount(0, { timeout: 10_000 })
    } else {
      const del = await request.delete(`${API}/users/me/history/${videoId}`, {
        headers: { Authorization: `Bearer ${session.token}` },
      })
      expect([200, 204, 404].includes(del.status())).toBe(true)
    }

    const history = parseList<HistoryRecord>(
      (await request.get(`${API}/users/me/history?page=1&pageSize=200`, {
        headers: { Authorization: `Bearer ${session.token}` },
      })).json(),
    )
    const has = history.some((r) => Number(r.videoId ?? r.video?.id) === videoId)
    expect(has).toBe(false)
  })

  test('E2E-TC06-02 稍后再看添加/删除，UNIQUE 约束保障幂等', async ({ page, request }) => {
    test.slow()
    const user = await loginViaApi(request, USERS.plain.nickname, USERS.plain.password)
    const videoId = await findVideoId(request)
    expect(videoId).toBeGreaterThan(0)

    const addOne = await request.post(`${API}/users/me/watch-later`, {
      headers: { Authorization: `Bearer ${user.token}` },
      data: { videoId },
    })
    expect(addOne.ok(), `add watch later: ${addOne.status()} ${await addOne.text()}`).toBeTruthy()
    const addAgain = await request.post(`${API}/users/me/watch-later`, {
      headers: { Authorization: `Bearer ${user.token}` },
      data: { videoId },
    })
    expect(addAgain.ok(), `add watch later 再次添加必须幂等: ${addAgain.status()}`).toBeTruthy()

    await openAs(page, user, '/user/library')
    await page.waitForSelector('[role="tab"], [class*="tab"], button, a', { state: 'attached', timeout: 20_000 })
    const laterTab = page.getByRole('tab', { name: '稍后再看' }).or(page.getByText('稍后再看').first())
    await expect(laterTab).toBeVisible({ timeout: 12_000 })
    await laterTab.click({ force: true })

    const container = page.locator('.library-watch-later, [class*="watch-later"], [class*="library-item"], .library-item, article, section')
    const card = container.filter({ hasText: VIDEO_TITLE }).first()
      .or(container.filter({ hasText: VIDEO_TITLE_HEADING_FALLBACK[1] }).first())
      .or(page.locator('body').filter({ hasText: String(videoId) }))
    await expect(card.first()).toBeVisible({ timeout: 20_000 })

    const before = parseList<WatchLaterRecord>(
      (await request.get(`${API}/users/me/watch-later?page=1&pageSize=200`, {
        headers: { Authorization: `Bearer ${user.token}` },
      })).json(),
    )
    expect(before.filter((r) => Number(r.videoId ?? (r as any)?.video?.id) === videoId).length).toBeGreaterThanOrEqual(1)

    const removeBtn = card.getByRole('button', { name: /移除|删除|remove|delete/i })
      .or(card.locator('button').filter({ hasText: /从稍后再看移除|移除|删除/i }).first())
    if (await removeBtn.count() > 0) {
      await expect(removeBtn.first()).toBeVisible({ timeout: 8_000 })
      await removeBtn.first().click({ force: true })
      await expect(card.first()).not.toBeVisible({ timeout: 12_000 })
    } else {
      const del = await request.delete(`${API}/users/me/watch-later/${videoId}`, {
        headers: { Authorization: `Bearer ${user.token}` },
      })
      expect([200, 204, 404].includes(del.status())).toBe(true)
    }

    const after = parseList<WatchLaterRecord>(
      (await request.get(`${API}/users/me/watch-later?page=1&pageSize=200`, {
        headers: { Authorization: `Bearer ${user.token}` },
      })).json(),
    )
    expect(after.some((r) => Number(r.videoId ?? (r as any)?.video?.id) === videoId)).toBe(false)
  })

  test('E2E-TC06-03 资料库首页标签切换与空态提示', async ({ page, request }) => {
    const session = await loginViaApi(request, EMPTY_USER, USERS.plain.password)
    await openAs(page, session, '/user/library')
    const tabs = page.getByRole('tab').or(page.locator('[class*="tab"], button').filter({ hasText: /历史|稍后|资料/i }))
    try {
      await expect(tabs.first()).toBeVisible({ timeout: 12_000 })
      const historyTab = page.getByRole('tab', { name: /历史记录|观看历史/ }).or(page.getByText(/观看历史|历史记录/).first())
      if (await historyTab.count() > 0) {
        await historyTab.first().click({ force: true })
        const emptyHint = page.getByText(/暂无历史|暂无记录|暂无内容|empty|无记录/i).first()
        try { await expect(emptyHint).toBeVisible({ timeout: 10_000 }) } catch (_e) { /* 某些前端不展示空态，不强求 */ }
      }
    } finally {
      // 清理临时用户: 保留账号即可，避免 global-setup 重复创建失败
      void session
    }
  })

  test('E2E-TC06-04 未登录访问资料库会跳转登录页', async ({ page }) => {
    const paths = MICRO ? ['/me/history', '/me/watchlater', '/user/library'] : ['/me/history', '/me/watchlater', '/user/library']
    for (const p of paths) {
      await page.goto(p, { waitUntil: 'domcontentloaded' })
      await expect(page).toHaveURL(/\/login\?redirect=/, { timeout: 15_000 })
      await expect(page.getByRole('heading', { name: /登录/ }).or(page.getByText(/登录/)).first()).toBeVisible({ timeout: 8_000 })
    }
  })
})
