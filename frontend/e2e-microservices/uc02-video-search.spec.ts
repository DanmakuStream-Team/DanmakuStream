import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

test.describe('UC02 用户搜索公开视频并进入播放页（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC02-01 搜索公开视频 → 详情 → 播放器，并验证空搜索结果', async ({ page }) => {
    test.slow()
    const initialVideosResponse = page.waitForResponse((response) =>
      response.url().includes('/api/v1/videos') && !response.url().includes('keyword='))
    await page.goto('/', { waitUntil: 'domcontentloaded', timeout: 40_000 })
    expect((await initialVideosResponse).status()).toBe(200)
    const search = page.getByLabel('搜索视频或创作者').or(page.getByPlaceholder(/搜索/)).first()
    await expect(search).toBeVisible({ timeout: 20_000 })
    await search.fill('E2E-MC-公开视频')
    const searchResponse = page.waitForResponse((response) =>
      response.url().includes('/api/v1/videos') && response.url().includes('keyword='))
    await search.press('Enter')
    expect((await searchResponse).status()).toBe(200)
    await expect(page).toHaveURL(/keyword=/, { timeout: 20_000 })

    const exactTitleCard = page.getByRole('link', { name: /E2E-MC-公开视频/ }).first()
    const titleTextLocator = page.locator('a,button,article,[role="link"],[class*="video-card"],[class*="search-item"],[class*="video-item"],[class*="videoCard"],[class*="searchItem"]').filter({ hasText: 'E2E-MC-公开视频' }).first()
    const result = exactTitleCard.or(titleTextLocator).first()
    try {
      await expect(result).toBeVisible({ timeout: 15_000 })
    } catch (err) {
      const html = await page.content()
      // eslint-disable-next-line no-console
      console.log('[uc02 debug] search-result-page HTML snapshot (前2000字):', html.slice(0, 2000))
      // eslint-disable-next-line no-console
      console.log('[uc02 debug] search-result-page search-query in html:', html.includes('E2E-MC-公开视频'))
      throw err
    }
    await Promise.all([
      page.waitForURL(/\/video\/\d+/, { timeout: 20_000 }),
      result.click(),
    ])
    await expect(page.getByRole('heading', { name: 'E2E-MC-公开视频' })).toBeVisible({ timeout: 15_000 })
    const video = page.locator('video').first()
    const player = page.locator('[class*="player"],[class*="video-wrap"],[class*="VideoPlayer"]').first()
    await expect(video.or(player)).toBeVisible({ timeout: 15_000 })

    const search2 = page.getByLabel('搜索视频或创作者').or(page.getByPlaceholder(/搜索/)).first()
    await search2.fill('绝对不存在-微服务-xxxyyy')
    await search2.press('Enter')
    await expect(page.getByText(/没有找到相关视频|暂无视频|搜索结果为空/).first()).toBeVisible({ timeout: 15_000 })
  })
})
