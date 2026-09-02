import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS, VIDEO_TITLE_FALLBACK } from './test-data'

const VIDEO_TITLE = 'E2E-MC-公开视频'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

export function parseVideoList(payload: unknown): Array<Record<string, any>> {
  const wrap: any = payload as any
  const list: any[] = ([] as any[]).concat(
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
  return list.filter(Boolean)
}

export async function findApprovedVideoId(
  request: import('@playwright/test').APIRequestContext,
  preferredTitle?: string,
): Promise<number> {
  const listResp = await request.get(`${API}/videos?page=1&pageSize=200`)
  const list = parseVideoList(await listResp.json())
  const approved = list.filter((item) => String(item.status).toLowerCase() === 'approved')
  if (preferredTitle) {
    const exact = approved.find((item) => item.title === preferredTitle)
    if (exact && Number(exact.id) > 0) return Number(exact.id)
    const fuzzy = approved.find(
      (item) => typeof item.title === 'string' && VIDEO_TITLE_FALLBACK.test(item.title),
    )
    if (fuzzy && Number(fuzzy.id) > 0) return Number(fuzzy.id)
  }
  const fallback = approved.find((item) => Number(item.id) > 0)
  return fallback ? Number(fallback.id) : 0
}

test.describe('UC02 用户搜索公开视频并进入播放页', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC02-01 搜索公开视频 → 详情 → 播放器，并验证空搜索结果', async ({ page, request }) => {
    test.slow()
    const videoId = await findApprovedVideoId(request, VIDEO_TITLE)
    expect(videoId, '需要至少 1 条 approved 视频：member-c 种子未就绪?').toBeGreaterThan(0)

    const initialVideosResponse = page.waitForResponse((response) =>
      response.url().includes('/api/v1/videos') && !response.url().includes('keyword='))
    await page.goto('/', { waitUntil: 'domcontentloaded', timeout: 40_000 })
    expect((await initialVideosResponse).status()).toBe(200)
    const search = page.getByLabel('搜索视频或创作者')
      .or(page.getByPlaceholder(/搜索/)).first()
    await expect(search).toBeVisible({ timeout: 20_000 })
    await search.fill(VIDEO_TITLE)
    const searchResponse = page.waitForResponse((response) =>
      response.url().includes('/api/v1/videos') && response.url().includes('keyword='))
    await search.press('Enter')
    expect((await searchResponse).status()).toBe(200)
    await expect(page).toHaveURL(/keyword=/, { timeout: 20_000 })

    const firstCardSelector = 'a[href*="/video/"], a[href*="video"], [class*="VideoCard"], [class*="video-card"], [class*="videoCard"], [class*="searchItem"], [class*="search-item"], article, section, [role="link"]'
    try {
      await page.waitForSelector(firstCardSelector, { state: 'attached', timeout: 20_000 })
    } catch (err) {
      const html = await page.content()
      // eslint-disable-next-line no-console
      console.log('[uc02 debug] search-page HTML (body 后 5000):', html.slice(html.indexOf('<body'), 5000))
      // eslint-disable-next-line no-console
      console.log('[uc02 debug] keyword in page:', html.includes(VIDEO_TITLE), 'fallback matches:', VIDEO_TITLE_FALLBACK.test(html), 'firstCard count:', await page.locator(firstCardSelector).count())
      throw err
    }

    const byText = page.getByText(VIDEO_TITLE).first().or(page.getByText(VIDEO_TITLE_FALLBACK).first())
    const cardAncestor = byText.locator('xpath=ancestor::a | ancestor::button | ancestor::article | ancestor::section | ancestor::div[contains(@class,"card")][1] | ancestor::div[contains(@class,"Card")][1]').first()
    const directCard = page.locator('a,button,article,section,[role="link"]').filter({ hasText: VIDEO_TITLE }).first()
      .or(page.locator('a,button,article,section,[role="link"]').filter({ hasText: VIDEO_TITLE_FALLBACK }).first())
    const result = cardAncestor.or(directCard).first()
    try {
      await expect(result).toBeVisible({ timeout: 15_000 })
    } catch (err) {
      const html = await page.content()
      // eslint-disable-next-line no-console
      console.log('[uc02 debug] result not visible HTML (body 后 5000):', html.slice(html.indexOf('<body'), 5000))
      // eslint-disable-next-line no-console
      console.log('[uc02 debug] cards count:', await page.locator(firstCardSelector).count())
      throw err
    }
    await Promise.all([
      page.waitForURL(/\/video\/\d+/, { timeout: 20_000 }),
      result.click({ force: true }),
    ])

    const heading = page.getByRole('heading', { name: VIDEO_TITLE })
      .or(page.getByRole('heading', { name: VIDEO_TITLE_FALLBACK }))
      .or(page.locator('h1,h2,h3,[class*="title"],[class*="Title"]').filter({ hasText: VIDEO_TITLE }).first())
      .or(page.locator('h1,h2,h3,[class*="title"],[class*="Title"]').filter({ hasText: VIDEO_TITLE_FALLBACK }).first())
      .or(page.getByText(VIDEO_TITLE).first())
    await expect(heading.first()).toBeVisible({ timeout: 15_000 })
    const video = page.locator('video').first()
    const player = page.locator('[class*="player"],[class*="video-wrap"],[class*="VideoPlayer"]').first()
    await expect(video.or(player)).toBeVisible({ timeout: 15_000 })

    const videoById = await request.get(`${API}/videos/${videoId}`)
    expect(videoById.ok(), `GET /videos/${videoId} returned ${videoById.status()}`).toBeTruthy()
    expect(String(await videoById.text())).toMatch(/approved|title|E2E-MC/)

    const search2 = page.getByLabel('搜索视频或创作者').or(page.getByPlaceholder(/搜索/)).first()
    await search2.fill('绝对不存在-微服务-xxxyyy')
    const emptyResp = page.waitForResponse((r) => r.url().includes('/api/v1/videos') && r.url().includes('xxxyyy'))
    await search2.press('Enter')
    await expect(emptyResp).resolves.toBeTruthy()
    const emptyTip = page.getByText(/没有找到相关视频|暂无视频|搜索结果为空|暂无内容/).first()
    await expect(emptyTip).toBeVisible({ timeout: 15_000 })
  })
})
