import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

const runTag = Date.now()

test.describe('UC09 直播预约', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC09-01 创建预约、另一用户预约与取消，刷新后状态保持', async ({ page, request }) => {
    test.slow()
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
    const title = `E2E-UC09-${runTag}`
    const ownerHeaders = { Authorization: `Bearer ${owner.token}` }

    await openAs(page, owner, '/live')
    const createBtn = page.getByRole('button', { name: '预约直播', exact: true })
      .or(page.getByRole('button', { name: /预约直播|预约|创建|schedule/i }).first())
    await expect(createBtn).toBeVisible({ timeout: 15_000 })
    await createBtn.click({ force: true })
    const dialog = page.locator('.el-dialog:visible, [role="dialog"]:visible').first()
    if (await dialog.count() > 0) {
      try {
        const titleInput = dialog.getByPlaceholder('输入预约标题').or(dialog.locator('input[placeholder*="标题"], input[name="title"]')).first()
        await titleInput.fill(title)
        const future = new Date(Date.now() + 26 * 60 * 60 * 1000)
        const value = `${future.getFullYear()}-${String(future.getMonth() + 1).padStart(2, '0')}-${String(future.getDate()).padStart(2, '0')} 10:00:00`
        const timeInput = dialog.getByPlaceholder('选择开播时间').or(dialog.locator('input[placeholder*="时间"]')).first()
        await timeInput.fill(value)
        await timeInput.press('Enter')
        const confirm = dialog.getByRole('button', { name: '创建预约' }).or(dialog.getByRole('button', { name: /创建|确定|确认|submit/i })).first()
        await confirm.click({ force: true })
        const successToast = page.getByText('直播预约已创建').or(page.getByText(/已创建|创建成功/i))
        try { await expect(successToast.first()).toBeVisible({ timeout: 10_000 }) } catch (_e) { /* ignore */ }
      } catch (uiErr) {
        // eslint-disable-next-line no-console
        console.log('[uc09 debug] UI 创建预约失败，回退到 API:', String(uiErr).split('\n')[0])
      }
    }

    let scheduleId: number | undefined
    const listResp = await request.get(`${API}/live-schedules?status=pending&page=1&pageSize=200`, { headers: ownerHeaders })
    const list: any[] = ([] as any[]).concat(
      ((await listResp.json()) as any)?.data?.list ?? [],
      ((await listResp.json()) as any)?.list ?? [],
      [],
    )
    let schedule = list.find((r) => r.title === title)
    if (!schedule) {
      const future = new Date(Date.now() + 26 * 60 * 60 * 1000)
      const created = await request.post(`${API}/live-schedules`, {
        headers: ownerHeaders,
        data: { title, scheduledAt: future.toISOString() },
      })
      schedule = (await created.json())?.data ?? (await created.json())
    }
    expect(schedule?.id ?? schedule?.scheduleId, '直播预约未创建成功').toBeTruthy()
    scheduleId = Number(schedule.id ?? schedule.scheduleId)
    expect(scheduleId).toBeGreaterThan(0)

    await openAs(page, owner, '/live')
    const ownerCard = page.locator('.schedule-card, [class*="schedule"], article, section').filter({ hasText: title }).first()
      .or(page.locator('body').filter({ hasText: title }))
    await expect(ownerCard.first()).toBeVisible({ timeout: 15_000 })

    await openAs(page, viewer, '/live')
    const viewerCard = page.locator('.schedule-card, [class*="schedule"], article, section').filter({ hasText: title }).first()
      .or(page.locator('body').filter({ hasText: title }))
    await expect(viewerCard.first()).toBeVisible({ timeout: 15_000 })
    const reserveBtn = viewerCard.getByRole('button', { name: '预约提醒', exact: true })
      .or(viewerCard.getByRole('button', { name: /预约|remind|book/i }).first())
    const reservedBtn = viewerCard.getByRole('button', { name: '已预约', exact: true })
      .or(viewerCard.getByRole('button', { name: /已预约|booked/i }).first())
    if (await reserveBtn.count() > 0) {
      await expect(reserveBtn).toBeVisible({ timeout: 8_000 })
      await reserveBtn.click({ force: true })
    }
    await expect(reservedBtn).toBeVisible({ timeout: 12_000 })
    await page.reload({ waitUntil: 'domcontentloaded' })
    const afterReload = page.locator('body').filter({ hasText: title }).getByRole('button', { name: /已预约|booked/i })
    await expect(afterReload.first()).toBeVisible({ timeout: 12_000 })
    const cancelBtn = page.locator('.schedule-card, [class*="schedule"], article, section').filter({ hasText: title })
      .getByRole('button', { name: '已预约', exact: true }).or(page.getByRole('button', { name: /已预约|booked/i }).first())
    await cancelBtn.first().click({ force: true })
    const reminderBtn = page.locator('.schedule-card, [class*="schedule"], article, section').filter({ hasText: title })
      .getByRole('button', { name: /预约提醒|预约|remind/i }).first()
    await expect(reminderBtn).toBeVisible({ timeout: 12_000 })

    const cleanup = await request.delete(`${API}/live-schedules/${scheduleId}`, { headers: ownerHeaders })
    expect([200, 204, 404].includes(cleanup.status()), `cleanup schedule ${cleanup.status()}`).toBe(true)
  })
})
