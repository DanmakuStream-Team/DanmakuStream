import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

test.describe('UC04 审核员发布与拒绝视频，普通用户不能越权（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC04-01 审核员通过/拒绝，普通用户同时受页面和 API 权限保护', async ({ page, request }) => {
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
})
