import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

test.describe('UC02 用户搜索公开视频并进入播放页（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC02-01 搜索公开视频 → 详情 → 播放器，并验证空搜索结果', async ({ page }) => {
    const initialVideosResponse = page.waitForResponse((response) =>
      response.url().includes('/api/v1/videos') && !response.url().includes('keyword='))
    await page.goto('/', { waitUntil: 'domcontentloaded', timeout: 30_000 })
    expect((await initialVideosResponse).status()).toBe(200)
    const search = page.getByLabel('搜索视频或创作者')
    await search.fill('E2E-MC-公开视频')
    const searchResponse = page.waitForResponse((response) =>
      response.url().includes('/api/v1/videos') && response.url().includes('keyword='))
    await search.press('Enter')
    expect((await searchResponse).status()).toBe(200)
    await expect(page).toHaveURL(/keyword=/)
    const result = page.locator('.search-item', { hasText: 'E2E-MC-公开视频' })
    await expect(result).toBeVisible()
    await Promise.all([
      page.waitForURL(/\/video\/\d+/, { timeout: 15_000 }),
      result.click(),
    ])
    await expect(page.getByRole('heading', { name: 'E2E-MC-公开视频' })).toBeVisible()
    await expect(page.locator('video')).toBeVisible()

    await search.fill('绝对不存在-微服务-xxxyyy')
    await search.press('Enter')
    await expect(page.getByText('没有找到相关视频')).toBeVisible()
  })
})
