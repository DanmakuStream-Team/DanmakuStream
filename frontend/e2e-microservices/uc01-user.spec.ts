import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, PASSWORD, USERS } from './test-data'

test.describe('UC01 用户注册、登录与资料维护（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test('E2E-TC01-01 注册成功并保存登录态', async ({ page, request }) => {
    const nickname = `micro-uc01-register-${Date.now()}`
    await page.goto('/register')
    await page.getByPlaceholder('请输入昵称').fill(nickname)
    await page.getByPlaceholder('请输入密码').fill(PASSWORD)
    await page.getByRole('button', { name: '注册', exact: true }).click()
    await page.waitForURL(/\/$/, { timeout: 30_000, waitUntil: 'domcontentloaded' })
    await page.waitForLoadState('domcontentloaded')
    const token = await page.evaluate(() => localStorage.getItem('token'))
    expect(token).toBeTruthy()
    const me = await request.get(`${API}/auth/me`, { headers: { Authorization: `Bearer ${token}` } })
    expect(me.ok()).toBeTruthy()
    expect((await me.json()).data.nickname).toBe(nickname)
  })

  test('E2E-TC01-02 重复昵称显示错误且不建立会话', async ({ page, request }) => {
    const nickname = `micro-uc01-duplicate-${Date.now()}`
    await request.post(`${API}/auth/register`, { data: { nickname, password: PASSWORD } })
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
    const tokenBefore = await page.evaluate(() => localStorage.getItem('token'))
    expect(tokenBefore).toBeTruthy()
    await page.reload()
    await expect.poll(() => page.evaluate(() => localStorage.getItem('token'))).toBe(tokenBefore)
    const me = await request.get(`${API}/auth/me`, { headers: { Authorization: `Bearer ${tokenBefore}` } })
    expect(me.ok()).toBeTruthy()
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
    test.setTimeout(60_000)
    const session = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const bio = `Micro UC01 E2E bio ${Date.now()}`
    await openAs(page, session, `/user/${session.userInfo.id}`)
    await page.locator('.profile-page').waitFor({ state: 'visible', timeout: 20_000 })
    const bioDisplay = page.locator('.bio-display').first()
    await expect(bioDisplay).toBeVisible({ timeout: 10_000 })
    await bioDisplay.click({ force: true })
    const bioEditor = page.locator('.bio-editor').first()
    await expect(bioEditor.getByPlaceholder('写一段个人简介')).toBeVisible({ timeout: 10_000 })
    await bioEditor.getByPlaceholder('写一段个人简介').fill(bio)
    await page.locator('.bio-actions').getByRole('button', { name: '保存', exact: true }).click()
    await expect(page.getByText('简介已更新').last()).toBeVisible({ timeout: 10_000 })
    await expect(bioDisplay).toContainText(bio, { timeout: 10_000 })
    const profile = await request.get(`${API}/users/${session.userInfo.id}`, {
      headers: { Authorization: `Bearer ${session.token}` },
    })
    expect(profile.ok()).toBeTruthy()
    expect((await profile.json()).data.bio).toBe(bio)
  })
})
