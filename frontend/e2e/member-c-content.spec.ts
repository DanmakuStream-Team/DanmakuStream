import path from 'node:path'
import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

test.describe.serial('成员 C 内容域', () => {
  // E2E-TC02-01 搜索 → 详情 → 播放器，并验证空搜索结果。
  test('UC02 用户搜索公开视频并进入播放页', async ({ page }) => {
    await page.goto('/', { waitUntil: 'domcontentloaded', timeout: 30_000 })
    const search = page.getByLabel('搜索视频或创作者')
    await search.fill('E2E-MC-公开视频')
    const searchResponse = page.waitForResponse((response) =>
      response.url().includes('/api/v1/videos') && response.url().includes('keyword='))
    await search.press('Enter')
    expect((await searchResponse).status()).toBe(200)
    await expect(page).toHaveURL(/keyword=/)
    const result = page.locator('.search-item', { hasText: 'E2E-MC-公开视频' })
    await expect(result).toBeVisible()
    await result.getByRole('heading', { name: 'E2E-MC-公开视频' }).click()
    await expect(page).toHaveURL(/\/video\/\d+/)
    await expect(page.getByRole('heading', { name: 'E2E-MC-公开视频' })).toBeVisible()
    await expect(page.locator('video')).toBeVisible()

    await search.fill('E2E-MC-绝对不存在')
    await search.press('Enter')
    await expect(page.getByText('没有找到相关视频')).toBeVisible()
  })

  // E2E-TC03-01 前端取消请求后不显示成功，并可再次完成真实投稿。
  test('UC03 创作者取消上传后重新投稿并看到待审状态', async ({ page, request }) => {
    const creator = await loginViaApi(request, USERS.memberCCreator.nickname, USERS.memberCCreator.password)
    await openAs(page, creator, '/creator/upload')
    const file = path.resolve('.e2e-fixtures/member-c.mp4')
    await page.locator('input[type=file]').first().setInputFiles(file)
    await page.getByPlaceholder('请输入视频标题').fill('E2E-MC-取消上传')

    await page.route('**/api/v1/videos/upload', async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 3000))
      await route.abort('aborted')
    })
    await page.getByRole('button', { name: '提交上传' }).click()
    await page.getByRole('button', { name: '终止上传' }).click()
    await expect(page.getByText('已终止上传')).toBeVisible()
    await page.unroute('**/api/v1/videos/upload')

    await page.getByPlaceholder('请输入视频标题').fill('E2E-MC-真实投稿')
    await page.getByRole('button', { name: '提交上传' }).click()
    await expect(page).toHaveURL(/\/creator$/, { timeout: 30_000 })
    const workRow = page.locator('.el-table__row', { hasText: 'E2E-MC-真实投稿' })
    await expect(workRow).toBeVisible()
    await expect(workRow).toContainText('待审核')

    const canceled = await request.get(`${API}/users/me/videos?page=1&pageSize=100`, {
      headers: { Authorization: `Bearer ${creator.token}` },
    })
    expect(await canceled.text()).not.toContain('E2E-MC-取消上传')
  })

  // E2E-TC04-01 审核员通过/拒绝，普通用户同时受页面和 API 权限保护。
  test('UC04 审核员发布与拒绝视频，普通用户不能越权', async ({ page, request }) => {
    const moderator = await loginViaApi(request, USERS.memberCModerator.nickname, USERS.memberCModerator.password)
    await openAs(page, moderator, '/admin/videos')

    const approveRow = page.locator('.row', { hasText: 'E2E-MC-待审核通过' })
    await expect(approveRow).toBeVisible()
    await approveRow.locator('.el-select').click()
    await page.keyboard.press('ArrowDown')
    await page.keyboard.press('Enter')
    await expect(approveRow.locator('.el-tag')).toContainText('已通过')

    const rejectRow = page.locator('.row', { hasText: 'E2E-MC-待审核拒绝' })
    await rejectRow.locator('.el-select').click()
    await page.keyboard.press('ArrowDown')
    await page.keyboard.press('ArrowDown')
    await page.keyboard.press('Enter')
    await expect(rejectRow.locator('.el-tag')).toContainText('已拒绝')

    const plain = await loginViaApi(request, USERS.memberCPlain.nickname, USERS.memberCPlain.password)
    await openAs(page, plain, '/admin/videos')
    await expect(page).not.toHaveURL(/\/admin/)
    const response = await request.get(`${API}/admin/videos`, {
      headers: { Authorization: `Bearer ${plain.token}` },
    })
    expect(response.status()).toBe(403)
  })

  // E2E-TC12-01 时间范围和作品范围切换，接口与图表同步刷新。
  test('UC12 创作者切换 7 天和单作品分析范围', async ({ page, request }) => {
    const creator = await loginViaApi(request, USERS.memberCCreator.nickname, USERS.memberCCreator.password)
    await openAs(page, creator, '/creator')
    await expect(page.getByRole('heading', { name: '数据趋势' })).toBeVisible()

    const sevenDayResponse = page.waitForResponse((response) =>
      response.url().includes('/creator/analytics') && response.url().includes('days=7'))
    await page.getByText('近 7 天', { exact: true }).click()
    expect((await sevenDayResponse).status()).toBe(200)
    await expect(page.getByText('新增观看', { exact: true }).locator('..').locator('strong')).toContainText(/15|16/)

    const singleWorkResponse = page.waitForResponse((response) =>
      response.url().includes('/creator/analytics') && response.url().includes('videoId='))
    await page.locator('.analytics-filters .el-select').click()
    await dropdownOption(page, 'E2E-MC-公开视频').click()
    const analyticsResponse = await singleWorkResponse
    expect(analyticsResponse.status()).toBe(200)
    await expect(page.getByText('E2E-MC-公开视频的观看、收藏增长和账号开播次数。')).toBeVisible()

    const other = await loginViaApi(request, USERS.memberCPlain.nickname, USERS.memberCPlain.password)
    const ownVideos = await (await request.get(`${API}/users/me/videos?page=1&pageSize=100`, {
      headers: { Authorization: `Bearer ${creator.token}` },
    })).json()
    const selectedID = ownVideos.data.list.find((item: { title: string }) => item.title === 'E2E-MC-公开视频').id
    const forbidden = await request.get(`${API}/creator/analytics?days=7&videoId=${selectedID}`, {
      headers: { Authorization: `Bearer ${other.token}` },
    })
    expect(forbidden.status()).toBe(404)
  })
})
