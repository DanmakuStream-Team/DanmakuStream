import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, PASSWORD, USERS } from './test-data'

test.describe('UC01 用户注册、登录与资料维护', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC01-01 注册成功并保存登录态', async ({ page, request }) => {
    const nickname = `e2e-uc01-register-${Date.now()}`

    await page.goto('/register')
    await page.getByPlaceholder('请输入昵称').fill(nickname)
    await page.getByPlaceholder('请输入密码').fill(PASSWORD)
    await page.getByRole('button', { name: '注册', exact: true }).click()

    await page.waitForURL(/\/$/, { timeout: 30_000, waitUntil: 'domcontentloaded' })
    await page.waitForLoadState('domcontentloaded')
    await expect.poll(() => page.evaluate(() => localStorage.getItem('token'))).toBeTruthy()
    const session = await page.evaluate(() => ({
      token: localStorage.getItem('token'),
      userInfo: JSON.parse(localStorage.getItem('userInfo') || 'null'),
    }))
    expect(session.token).toBeTruthy()
    expect(session.userInfo?.nickname).toBe(nickname)

    const me = await request.get(`${API}/auth/me`, {
      headers: { Authorization: `Bearer ${session.token}` },
    })
    expect(me.ok(), `GET /auth/me returned ${me.status()}`).toBeTruthy()
    expect((await me.json()).data.nickname).toBe(nickname)
  })

  test('E2E-TC01-02 重复昵称显示错误且不建立会话', async ({ page, request }) => {
    const nickname = `e2e-uc01-duplicate-${Date.now()}`
    const prepared = await request.post(`${API}/auth/register`, {
      data: { nickname, password: PASSWORD },
    })
    expect(prepared.ok(), `prepare register returned ${prepared.status()}`).toBeTruthy()

    await page.goto('/register')
    await page.getByPlaceholder('请输入昵称').fill(nickname)
    await page.getByPlaceholder('请输入密码').fill(PASSWORD)
    await page.getByRole('button', { name: '注册', exact: true }).click()

    await expect(page.getByText('昵称已存在').last()).toBeVisible()
    await expect(page).toHaveURL(/\/register$/)
    expect(await page.evaluate(() => localStorage.getItem('token'))).toBeNull()
  })

  test('E2E-TC01-03 登录成功，刷新后会话仍有效', async ({ page, request }) => {
    const user = USERS.owner
    await page.goto('/login')
    await page.getByPlaceholder('请输入昵称').fill(user.nickname)
    await page.getByPlaceholder('请输入密码').fill(user.password)
    await page.getByRole('button', { name: '登录', exact: true }).click()

    await page.waitForURL(/\/$/, { timeout: 30_000, waitUntil: 'domcontentloaded' })
    const tokenBeforeReload = await page.evaluate(() => localStorage.getItem('token'))
    expect(tokenBeforeReload).toBeTruthy()
    await page.reload()
    await expect.poll(() => page.evaluate(() => localStorage.getItem('token'))).toBe(tokenBeforeReload)

    const me = await request.get(`${API}/auth/me`, {
      headers: { Authorization: `Bearer ${tokenBeforeReload}` },
    })
    expect(me.ok(), `GET /auth/me returned ${me.status()}`).toBeTruthy()
    expect((await me.json()).data.nickname).toBe(user.nickname)
  })

  test('E2E-TC01-04 错误密码显示错误且停留在登录页', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder('请输入昵称').fill(USERS.owner.nickname)
    await page.getByPlaceholder('请输入密码').fill('definitely-wrong')
    await page.getByRole('button', { name: '登录', exact: true }).click()

    await expect(page.getByText('昵称或密码错误').last()).toBeVisible()
    await expect(page).toHaveURL(/\/login$/)
    expect(await page.evaluate(() => localStorage.getItem('token'))).toBeNull()
  })

  test('E2E-TC01-05 编辑个人简介并持久化', async ({ page, request }) => {
    const session = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const bio = `UC01 E2E bio ${Date.now()}`

    await openAs(page, session, `/user/${session.userInfo.id}`)
    await page.getByText('编辑简介', { exact: true }).click()
    await page.getByPlaceholder('写一段个人简介').fill(bio)
    await page.locator('.bio-actions').getByRole('button', { name: '保存', exact: true }).click()

    await expect(page.getByText('简介已更新').last()).toBeVisible()
    await expect(page.locator('.bio-display')).toContainText(bio)
    const profile = await request.get(`${API}/users/${session.userInfo.id}`, {
      headers: { Authorization: `Bearer ${session.token}` },
    })
    expect(profile.ok(), `GET /users/:id returned ${profile.status()}`).toBeTruthy()
    expect((await profile.json()).data.bio).toBe(bio)
  })
})
