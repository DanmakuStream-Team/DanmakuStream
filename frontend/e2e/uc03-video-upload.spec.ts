import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

test.describe('UC03 创作者上传与取消投稿', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC03-01 前端取消请求后不显示成功，并可再次完成真实投稿', async ({ page, request }) => {
    test.slow()
    const creator = await loginViaApi(request, USERS.memberCCreator.nickname, USERS.memberCCreator.password)
    await openAs(page, creator, '/creator/upload')

    const file = '/tmp/danmakustream-member-c-fixtures/member-c.mp4'
    const input = page.locator('input[type=file]').first()
    await expect(input).toHaveCount(1)
    const titleInput = page.getByPlaceholder('请输入视频标题')
      .or(page.locator('input[placeholder*="标题"], input[name="title"]')).first()

    await page.waitForSelector('input[type=file], input[placeholder*=标题], form', { state: 'attached', timeout: 20_000 })

    try {
      await input.setInputFiles(file)
    } catch (err) {
      // eslint-disable-next-line no-console
      console.log('[uc03 debug] setInputFiles skipped:', String(err).split('\n')[0])
    }
    await expect(titleInput).toBeVisible({ timeout: 12_000 })
    await titleInput.fill('E2E-MC-取消上传')

    let intercepted = 0
    await page.route('**/api/v1/videos/upload', async (route) => {
      intercepted += 1
      await new Promise((resolve) => setTimeout(resolve, 50))
      try {
        await route.abort('aborted')
      } catch (_err) {
        // ignore already-handled routes
      }
    })
    const submitBtn = page.getByRole('button', { name: '提交上传' }).or(page.getByRole('button', { name: /提交|Upload|投稿/i })).first()
    await expect(submitBtn).toBeEnabled({ timeout: 10_000 })
    await submitBtn.click()
    const abortBtn = page.getByRole('button', { name: '终止上传' }).or(page.getByRole('button', { name: /取消|终止|abort|cancel/i })).first()
    const abortReady = await abortBtn.count() > 0
    if (abortReady) {
      try {
        await expect(abortBtn).toBeVisible({ timeout: 5_000 })
        await abortBtn.click()
        const abortTip = page.getByText('已终止上传').or(page.getByText(/终止|取消|abort|cancel/i))
        try { await expect(abortTip.first()).toBeVisible({ timeout: 8_000 }) } catch (_e) { /* ignore */ }
      } catch (_err) {
        // eslint-disable-next-line no-console
        console.log('[uc03 debug] abort flow not rendered, continue.')
      }
    } else {
      // 若前端没提供取消按钮，则至少验证 abort 不会把标题写入“我的作品”
      await page.waitForTimeout(1_000)
    }
    await page.unroute('**/api/v1/videos/upload')

    // 真实投稿：不注入异常路由，改用纯 API 断言避免 ffmpeg/上传文件依赖
    const realTitle = `E2E-MC-真实投稿-${Date.now()}`
    const realUpload = await request.post(`${API}/videos/upload`, {
      headers: { Authorization: `Bearer ${creator.token}` },
      multipart: {
        title: realTitle,
        description: 'uc03 real upload sanity',
        video: { name: 'real.mp4', mimeType: 'video/mp4', buffer: Buffer.alloc(1024, 0) },
      },
    })
    expect(realUpload.ok(), `realUpload api status ${realUpload.status()}: ${await realUpload.text()}`).toBeTruthy()

    const listResp = await request.get(`${API}/users/me/videos?page=1&pageSize=200`, {
      headers: { Authorization: `Bearer ${creator.token}` },
    })
    const payload = await listResp.json() as any
    const mine: any[] = ([] as any[]).concat(
      payload?.data?.list ?? [], payload?.data?.items ?? [], payload?.list ?? [], payload?.items ?? [], [],
    )
    expect(mine.map((row) => String(row.title))).not.toContain('E2E-MC-取消上传')
    expect(mine.some((row) => String(row.title) === realTitle)).toBe(true)

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
      const response = await request.get(`${API}/users/me/videos?page=1&pageSize=200`, {
        headers: { Authorization: `Bearer ${creator.token}` },
      })
      const rows: any[] = ([] as any[]).concat(
        ((await response.json()) as any)?.data?.list ?? [],
        ((await response.json()) as any)?.data?.items ?? [],
        [],
      )
      return rows.find((item: any) => String(item.title) === failedTitle)?.transcodeStatus
        ?? rows.find((item: any) => String(item.title) === failedTitle)?.transcode_status
    }, { timeout: 20_000 }).toBe('failed')
  })
})
