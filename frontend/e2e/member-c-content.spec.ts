import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

test.describe.serial('成员 C 内容域', () => {
  // E2E-TC02-01 搜索 → 详情 → 播放器，并验证空搜索结果。
  test('UC02 用户搜索公开视频并进入播放页', async ({ page }) => {
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

    await search.fill('E2E-MC-绝对不存在')
    await search.press('Enter')
    await expect(page.getByText('没有找到相关视频')).toBeVisible()
  })

  // E2E-TC03-01 前端取消请求后不显示成功，并可再次完成真实投稿。
  test('UC03 创作者取消上传后重新投稿并看到待审状态', async ({ page, request }) => {
    const creator = await loginViaApi(request, USERS.memberCCreator.nickname, USERS.memberCCreator.password)
    await openAs(page, creator, '/creator/upload')
    const file = '/tmp/danmakustream-member-c-fixtures/member-c.mp4'
    await page.locator('input[type=file]').first().setInputFiles(file)
    await page.getByPlaceholder('请输入视频标题').fill('E2E-MC-取消上传')

    await page.route('**/api/v1/videos/upload', async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 3000))
      await route.abort('aborted').catch(() => undefined)
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

    const failedTitle = `E2E-MC-转码失败-${Date.now()}`
    const failedUpload = await request.post(`${API}/videos/upload`, {
      headers: { Authorization: `Bearer ${creator.token}` },
      multipart: {
        title: failedTitle,
        video: { name: 'invalid.mp4', mimeType: 'video/mp4', buffer: Buffer.from('not a media file') },
      },
    })
    expect(failedUpload.status()).toBe(200)
    await expect.poll(async () => {
      const response = await request.get(`${API}/users/me/videos?page=1&pageSize=100`, {
        headers: { Authorization: `Bearer ${creator.token}` },
      })
      const payload = await response.json()
      return payload.data.list.find((item: { title: string }) => item.title === failedTitle)?.transcodeStatus
    }, { timeout: 15_000 }).toBe('failed')
    await page.reload()
    const failedRow = page.locator('.el-table__row', { hasText: failedTitle })
    await expect(failedRow).toContainText('转码失败')
    await expect(failedRow).toContainText('请检查文件格式后重新上传')
  })

  // E2E-TC04-01 审核员通过/拒绝，普通用户同时受页面和 API 权限保护。
  test('UC04 审核员发布与拒绝视频，普通用户不能越权', async ({ page, request }) => {
    const moderator = await loginViaApi(request, USERS.memberCModerator.nickname, USERS.memberCModerator.password)
    await openAs(page, moderator, '/admin/videos')

    const approveRow = page.locator('.row', { hasText: 'E2E-MC-待审核通过' })
    await expect(approveRow).toBeVisible()
    await approveRow.locator('.el-select').click()
    await dropdownOption(page, '通过').click()
    await expect(approveRow.locator('.el-tag')).toContainText('已通过')
    await expect(approveRow.locator('.el-select__wrapper')).toHaveClass(/is-disabled/)

    const approvedID = Number(await approveRow.getAttribute('data-video-id'))
    expect(approvedID).toBeGreaterThan(0)
    const repeated = await request.put(`${API}/admin/videos/${approvedID}/status`, {
      headers: { Authorization: `Bearer ${moderator.token}` },
      data: { status: 'rejected' },
    })
    expect(repeated.status()).toBe(409)

    const rejectRow = page.locator('.row', { hasText: 'E2E-MC-待审核拒绝' })
    const rejectID = Number(await rejectRow.getAttribute('data-video-id'))
    expect(rejectID).toBeGreaterThan(0)
    const rejected = await request.put(`${API}/admin/videos/${rejectID}/status`, {
      headers: { Authorization: `Bearer ${moderator.token}` },
      data: { status: 'rejected' },
    })
    expect(rejected.status()).toBe(200)
    await page.reload()
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
    const singleWorkPayload = await analyticsResponse.json()
    expect(singleWorkPayload.data.topVideos).toHaveLength(1)
    expect(singleWorkPayload.data.topVideos[0].id).toBe(singleWorkPayload.data.selectedVideoId)
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
