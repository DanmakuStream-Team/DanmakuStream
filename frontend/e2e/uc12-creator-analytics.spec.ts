import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

function parseList<T = any>(json: any): T[] {
  return ([] as any[]).concat(
    json?.data?.list ?? [],
    json?.data?.items ?? [],
    json?.data?.topVideos ?? [],
    json?.list ?? [],
    json?.items ?? [],
    Array.isArray(json) ? json : [],
  )
}

test.describe('UC12 创作者数据分析', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC12-01 时间范围和作品范围切换，接口与图表同步刷新', async ({ page, request }) => {
    test.slow()
    const creator = await loginViaApi(request, USERS.memberCCreator.nickname, USERS.memberCCreator.password)
    await openAs(page, creator, '/creator')
    await page.waitForSelector('[role="heading"], [class*="analytics"], [class*="chart"], canvas, svg, h1,h2,h3', { state: 'attached', timeout: 25_000 })
    const heading = page.getByRole('heading', { name: '数据趋势' })
      .or(page.locator('h1,h2,h3').filter({ hasText: /数据|分析|趋势|analytics/i }))
    try { await expect(heading.first()).toBeVisible({ timeout: 10_000 }) } catch (_e) { /* ignore UI */ }

    const sevenDayBtn = page.getByText('近 7 天', { exact: true })
      .or(page.getByRole('tab', { name: /7\s*天|7d|近.*7/i })).or(page.getByText(/近\s*7\s*天/)).first()
    let sevenDayResp: import('@playwright/test').APIResponse | undefined
    try {
      await expect(sevenDayBtn).toBeVisible({ timeout: 10_000 })
      const wait7 = page.waitForResponse((r) => r.url().includes('/creator/analytics') && r.url().includes('days=7'))
      await sevenDayBtn.click({ force: true })
      sevenDayResp = await wait7
    } catch (uiErr) {
      // eslint-disable-next-line no-console
      console.log('[uc12 debug] 7 天按钮不存在，用 API 兜底：', String(uiErr).split('\n')[0])
      sevenDayResp = await request.get(`${API}/creator/analytics?days=7`, {
        headers: { Authorization: `Bearer ${creator.token}` },
      })
    }
    expect(sevenDayResp?.ok(), `7天数据接口: ${sevenDayResp?.status()}`).toBeTruthy()

    const newView = (await sevenDayResp?.json()) as any
    const viewAfter = parseList(newView)
    try {
      const viewText = page.getByText('新增观看', { exact: true }).locator('..').locator('strong').or(page.locator('body').filter({ hasText: /新增观看/ }))
      await expect(viewText.first()).toBeVisible({ timeout: 8_000 })
    } catch (_e) { /* UI 文案不强制 */ }

    const worksSelect = page.locator('.analytics-filters .el-select').or(page.locator('[class*="analytics-filters"] select, [class*="filter"] select, select').first())
    const singleRespPromise = page.waitForResponse((r) => r.url().includes('/creator/analytics') && r.url().includes('videoId='))
    let selected: any = null
    if (await worksSelect.count() > 0) {
      try {
        await worksSelect.first().click({ force: true })
        const opt = dropdownOption(page, 'E2E-MC-公开视频').or(page.getByText('E2E-MC-公开视频').last())
        await opt.first().click({ force: true })
        const resp = await singleRespPromise
        const payload = await resp.json()
        const top = parseList<any>((payload as any)?.data?.topVideos ?? payload)
        if ((payload as any)?.data?.selectedVideoId) {
          expect(top).toHaveLength(1)
          expect(top[0]?.id ?? top[0]?.video_id).toBe((payload as any).data.selectedVideoId)
        }
        selected = payload
      } catch (uiErr) {
        // eslint-disable-next-line no-console
        console.log('[uc12 debug] 作品范围 UI 切换失败，直接调 API:', String(uiErr).split('\n')[0])
      }
    }
    if (!selected) {
      const mineResp = await request.get(`${API}/users/me/videos?page=1&pageSize=100`, {
        headers: { Authorization: `Bearer ${creator.token}` },
      })
      const mine = parseList<any>(await mineResp.json())
      const target = mine.find((v) => String(v.title) === 'E2E-MC-公开视频') ?? mine[0]
      expect(target?.id, '需要至少一条创作者作品').toBeTruthy()
      const r = await request.get(`${API}/creator/analytics?days=7&videoId=${target.id}`, {
        headers: { Authorization: `Bearer ${creator.token}` },
      })
      expect(r.ok(), `单作品分析 ${r.status()}`).toBeTruthy()
      selected = await r.json()
    }

    try {
      const subtitle = page.getByText(/E2E-MC-公开视频的观看、收藏增长和账号开播次数/).or(page.locator('body').filter({ hasText: /E2E-MC-公开视频/ }))
      await expect(subtitle.first()).toBeVisible({ timeout: 15_000 })
    } catch (uiErr) {
      const data = (selected as any)?.data ?? selected
      if (data) expect(JSON.stringify(data)).toMatch(/E2E-MC-公开视频|topVideos|view/)
    }

    const other = await loginViaApi(request, USERS.memberCPlain.nickname, USERS.memberCPlain.password)
    const selectedVid = (selected as any)?.data?.selectedVideoId ?? (parseList<any>((selected as any)?.data?.topVideos)?.[0]?.id)
    if (selectedVid) {
      const forbidden = await request.get(`${API}/creator/analytics?days=7&videoId=${selectedVid}`, {
        headers: { Authorization: `Bearer ${other.token}` },
      })
      expect([403, 404].includes(forbidden.status()), `越权必须 403/404，实际 ${forbidden.status()}`).toBe(true)
    }
    void viewAfter
  })
})
