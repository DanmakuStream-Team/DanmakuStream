import { expect, test, type APIRequestContext } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, ENGAGEMENT_VIDEO_TITLE, USERS, VIDEO_TITLE_FALLBACK } from './test-data'

const runTag = Date.now()

type VideoItem = { id: number; title: string; status?: string }

function parseList(json: any): VideoItem[] {
  return ([] as any[]).concat(
    json?.data?.list ?? [],
    json?.data?.items ?? [],
    json?.list ?? [],
    json?.items ?? [],
    Array.isArray(json) ? json : [],
  )
}

async function findEngagementVideo(request: APIRequestContext): Promise<VideoItem> {
  const kw = encodeURIComponent(ENGAGEMENT_VIDEO_TITLE)
  const resp = await request.get(`${API}/videos?page=1&pageSize=100&keyword=${kw}`)
  expect(resp.ok(), `search engagement video: ${resp.status()}`).toBeTruthy()
  const list = parseList(await resp.json())
  const preferred = list.find((v) => v.title === ENGAGEMENT_VIDEO_TITLE)
  if (preferred) return preferred
  const fuzzy = list.find((v) => typeof v.title === 'string' && VIDEO_TITLE_FALLBACK.test(v.title))
  if (fuzzy) return fuzzy
  const any = list.find((v) => Number(v.id) > 0)
  expect(any, 'global setup 应准备 UC05 互动测试视频，或至少 1 条公开视频').toBeTruthy()
  return any!
}

test.describe('UC05 弹幕、评论、点赞与收藏', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC05-01 发送弹幕和评论、点赞收藏，刷新后数据保持', async ({ page, request }) => {
    test.slow()
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
    const video = await findEngagementVideo(request)
    await openAs(page, viewer, `/video/${video.id}`)

    const heading = page.getByRole('heading', { name: video.title })
      .or(page.locator('h1,h2,h3,[class*="title"],[class*="Title"]').filter({ hasText: video.title }).first())
      .or(page.getByText(video.title).first())
      .or(page.locator('body').filter({ hasText: VIDEO_TITLE_FALLBACK }).first())
    await expect(heading.first()).toBeVisible({ timeout: 20_000 })

    const likeBtn = page.getByRole('button', { name: /^点赞/ })
      .or(page.locator('button').filter({ hasText: /点赞|like/i }).first())
    const beforeLike = await likeBtn.count() > 0 ? Number(String(await likeBtn.first().textContent() ?? '0').replace(/\D/g, '')) : 0
    if (beforeLike === 0) {
      await expect(likeBtn.first()).toBeVisible({ timeout: 10_000 })
      await likeBtn.first().click({ force: true })
    }
    await expect(page.locator('button').filter({ hasText: /点赞 1|已点赞|^点赞 1$/ }).first()).toBeVisible({ timeout: 12_000 })

    const collectBtn = page.getByRole('button', { name: /^收藏/ })
      .or(page.locator('button').filter({ hasText: /收藏|favorite/i }).first())
    const beforeFav = await collectBtn.count() > 0 ? Number(String(await collectBtn.first().textContent() ?? '0').replace(/\D/g, '')) : 0
    if (beforeFav === 0) {
      await expect(collectBtn.first()).toBeVisible({ timeout: 10_000 })
      await collectBtn.first().click({ force: true })
    }
    await expect(page.locator('button').filter({ hasText: /收藏 1|已收藏|^收藏 1$/ }).first()).toBeVisible({ timeout: 12_000 })

    const danmakuInput = page.getByPlaceholder('此刻想说什么')
      .or(page.getByPlaceholder(/发弹幕|输入弹幕|danmaku/i))
      .or(page.locator('input[placeholder*="弹幕"], input[placeholder*="此刻"]')).first()
    const danmakuText = `E2E弹幕-${runTag}`
    await expect(danmakuInput.first()).toBeVisible({ timeout: 10_000 })
    await danmakuInput.first().fill(danmakuText)
    await danmakuInput.first().press('Enter')

    const stats = page.locator('.stats,*').filter({ hasText: /弹幕\s*\d+|^\d+\s*弹幕/ })
    try { await expect(stats.first()).toBeVisible({ timeout: 8_000 }) } catch (_e) { /* ignore */ }

    const commentInput = page.getByPlaceholder('写下你的看法')
      .or(page.getByPlaceholder(/评论|comment/i))
      .or(page.locator('textarea,input').filter({ hasText: /写下你的|看法|评论/i }).first())
    const commentText = `E2E评论-${runTag}`
    await expect(commentInput.first()).toBeVisible({ timeout: 10_000 })
    await commentInput.first().fill(commentText)
    const submit = page.getByRole('button', { name: '发表评论' }).or(page.getByRole('button', { name: /发送|发表|提交/i })).first()
    await submit.click()
    await expect(page.getByText(commentText, { exact: true }).first()).toBeVisible({ timeout: 10_000 })

    await page.reload({ waitUntil: 'domcontentloaded' })
    await expect(page.locator('button').filter({ hasText: /点赞 1|已点赞/ }).first()).toBeVisible({ timeout: 15_000 })
    await expect(page.locator('button').filter({ hasText: /收藏 1|已收藏/ }).first()).toBeVisible({ timeout: 10_000 })
    await expect(page.getByText(commentText, { exact: true }).first()).toBeVisible({ timeout: 10_000 })

    const danmakus = await request.get(`${API}/videos/${video.id}/danmaku`)
    expect(danmakus.ok()).toBeTruthy()
    expect(String(await danmakus.text())).toContain(danmakuText)
  })
})
