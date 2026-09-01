import { expect, request as playwrightRequest, test } from '@playwright/test'

const gatewayURL = process.env.MICRO_E2E_GATEWAY_URL ?? 'http://127.0.0.1:18888'
const password = 'micro-e2e-password'

test.describe('微服务公开入口', () => {
  test('网关可用、服务目录完整且内部接口不暴露', async () => {
    const gateway = await playwrightRequest.newContext({ baseURL: gatewayURL })
    try {
      const health = await gateway.get('/gateway/health')
      expect(health.ok(), await health.text()).toBeTruthy()
      expect((await health.json()).data.service).toBe('api-gateway')

      const catalog = await gateway.get('/api/v1/platform/services')
      expect(catalog.ok(), await catalog.text()).toBeTruthy()
      const names = (await catalog.json()).data.services.map((item: { name: string }) => item.name)
      expect(names).toEqual(['user-service', 'content-service', 'engagement-service'])

      const internal = await gateway.get('/internal/v1/users/1')
      expect(internal.status()).toBe(404)
    } finally {
      await gateway.dispose()
    }
  })

  test('前端经网关注册，并用同一 JWT 访问三个独立服务', async ({ page, request }) => {
    const nickname = `micro-e2e-${Date.now()}`

    await page.goto('/register')
    await page.getByPlaceholder('请输入昵称').fill(nickname)
    await page.getByPlaceholder('请输入密码').fill(password)
    await page.getByRole('button', { name: '注册', exact: true }).click()
    await page.waitForURL(/\/$/, { waitUntil: 'domcontentloaded' })

    const token = await page.evaluate(() => localStorage.getItem('token'))
    expect(token).toBeTruthy()
    if (!token) throw new Error('registration did not persist a JWT')
    const headers = { Authorization: `Bearer ${token}` }

    // user-service: authentication and persistence
    const me = await request.get('/api/v1/auth/me', { headers })
    expect(me.ok(), await me.text()).toBeTruthy()
    expect((await me.json()).data.nickname).toBe(nickname)

    // content-service: public content route and authenticated creator route
    const videos = await request.get('/api/v1/videos')
    expect(videos.ok(), await videos.text()).toBeTruthy()
    const mine = await request.get('/api/v1/users/me/videos', { headers })
    expect(mine.ok(), await mine.text()).toBeTruthy()

    // engagement-service: public live route and authenticated personal library route
    const live = await request.get('/api/v1/live')
    expect(live.ok(), await live.text()).toBeTruthy()
    const history = await request.get('/api/v1/users/me/history', { headers })
    expect(history.ok(), await history.text()).toBeTruthy()
    expect((await history.json()).data.total).toBe(0)
  })
})
