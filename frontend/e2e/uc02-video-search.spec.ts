import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

test.describe('UC02 用户搜索公开视频并进入播放页', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC02-01 搜索 → 详情 → 播放器，并验证空搜索结果', async ({ page }) => {
    // NOTE: 骨架占位 — 完整实现详见 member-c-content.spec.ts 第 11~35 行
    // 领域负责人（成员C）：请在此补全断言后移除 test.skip
    // 步骤：打开首页 → 搜索公开视频 → 验证结果 → 点击进入播放页 → 验证播放器可见
    //       → 搜索不存在的关键词 → 验证空提示
    const initialVideosResponse = page.waitForResponse((response) =>
      response.url().includes('/api/v1/videos') && !response.url().includes('keyword='))
    await page.goto('/', { waitUntil: 'domcontentloaded', timeout: 30_000 })
    expect((await initialVideosResponse).status()).toBe(200)
  })
})
