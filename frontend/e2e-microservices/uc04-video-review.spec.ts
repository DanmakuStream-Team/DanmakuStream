import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

test.describe('UC04 审核员发布与拒绝视频，普通用户不能越权（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC04-01 审核员通过/拒绝，普通用户同时受页面和 API 权限保护', async ({ page, request }) => {
    test.slow()
    const moderator = await loginViaApi(request, USERS.memberCModerator.nickname, USERS.memberCModerator.password)

    const listResp = await request.get(`${API}/admin/videos?page=1&pageSize=50`, {
      headers: { Authorization: `Bearer ${moderator.token}` },
    })
    expect(listResp.ok()).toBeTruthy()
    const listJson = await listResp.json().catch(() => ({}))
    const list = ([] as Array<any>).concat(
      (listJson as any)?.data?.list ?? [],
      (listJson as any)?.data?.items ?? [],
      (listJson as any)?.data?.records ?? [],
      (listJson as any)?.list ?? [],
      (listJson as any)?.items ?? [],
      Array.isArray(listJson) ? listJson : [],
    )
    function findByTitle(t: string) { return list.find((v) => v.title === t) }
    const pendingApproved = findByTitle('E2E-MC-待审核通过')
    const pendingRejected = findByTitle('E2E-MC-待审核拒绝')
    const fallbackPending = list.find((v) => v.status === 'pending')

    const approvedID: number = Number(pendingApproved?.id ?? fallbackPending?.id ?? 0)
    const rejectedID: number = Number(pendingRejected?.id ?? list.filter((v) => v.status === 'pending').slice(-1)[0]?.id ?? 0)

    if (approvedID <= 0 || rejectedID <= 0) {
      // eslint-disable-next-line no-console
      console.log('[uc04 debug] admin/videos list payload:', JSON.stringify(listJson).slice(0, 1500))
      // eslint-disable-next-line no-console
      console.log('[uc04 debug] parsed list items:', list.map((v) => ({ id: v.id, status: v.status, title: v.title })).slice(0, 20))
    }
    expect(approvedID).toBeGreaterThan(0)
    expect(rejectedID).toBeGreaterThan(0)
    expect(approvedID).not.toBe(rejectedID)

    const repeated = await request.put(`${API}/admin/videos/${approvedID}/status`, {
      headers: { Authorization: `Bearer ${moderator.token}` },
      data: { status: 'approved' },
    })
    expect([200, 201, 409]).toContain(repeated.status())

    const rejected = await request.put(`${API}/admin/videos/${rejectedID}/status`, {
      headers: { Authorization: `Bearer ${moderator.token}` },
      data: { status: 'rejected' },
    })
    expect(rejected.status()).toBe(200)

    const detailAfterA = await request.get(`${API}/admin/videos?page=1&pageSize=50&status=approved`, {
      headers: { Authorization: `Bearer ${moderator.token}` },
    })
    expect(detailAfterA.ok()).toBeTruthy()
    const afterA = await detailAfterA.json().catch(() => ({}))
    const afterList = ([] as Array<any>).concat(
      (afterA as any)?.data?.list ?? [],
      (afterA as any)?.data?.items ?? [],
      (afterA as any)?.list ?? [],
      Array.isArray(afterA) ? afterA : [],
    )
    expect(afterList.some((v) => Number(v.id) === approvedID && v.status === 'approved')).toBe(true)

    await openAs(page, moderator, '/admin/videos')
    try {
      const approveRow = page.locator('tr,[class*="row"],[class*="table-row"],article,li').filter({ hasText: 'E2E-MC-待审核通过' }).first()
        .or(page.locator('tr,[class*="row"],[class*="table-row"],article,li').filter({ hasText: String(approvedID) }).first())
      await expect(approveRow).toBeVisible({ timeout: 15_000 })
    } catch (err) {
      const html = await page.content()
      // eslint-disable-next-line no-console
      console.log('[uc04 debug] admin/videos page HTML snapshot (前2500字):', html.slice(0, 2500))
      // eslint-disable-next-line no-console
      console.log('[uc04 debug] approvedID in html:', html.includes(String(approvedID)), 'rejectedID in html:', html.includes(String(rejectedID)))
      throw err
    }

    const plain = await loginViaApi(request, USERS.memberCPlain.nickname, USERS.memberCPlain.password)
    await openAs(page, plain, '/admin/videos')
    await expect(page).not.toHaveURL(/\/admin/, { timeout: 15_000 })
    const response = await request.get(`${API}/admin/videos`, {
      headers: { Authorization: `Bearer ${plain.token}` },
    })
    expect(response.status()).toBe(403)
  })
})
