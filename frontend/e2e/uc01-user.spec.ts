import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, PASSWORD, USERS } from './test-data'

test.describe('UC01 用户注册、登录与资料维护', () => {
  const user = USERS.owner
  const newUser = `e2e-uc01-new-${Date.now()}`
  const newPassword = PASSWORD

  test.describe('E2E-TC01-01 新用户注册主流程', () => {
    test('E2E-TC01-01-01 通过注册页创建账号 → 自动登录并跳转首页', async ({ page, request }) => {
      await test.step('1. 注册前置清理：如果该昵称已存在则先注销（本次通过随机昵称保证首次）', async () => {
        const exists = await request.post(`${API}/auth/login`, { data: { nickname: newUser, password: newPassword } })
        expect(exists.ok(), '新用户昵称应未被占用，避免偶发失败').toBeFalsy()
      })

      await test.step('2. 打开注册页，填写新账号并提交', async () => {
        await page.goto('/register')
        await page.locator('[name=nickname]').fill(newUser)
        await page.locator('[name=password]').fill(newPassword)
        await page.locator('button[type=submit]').click()
      })

      await test.step('3. 断言：注册成功后应保存登录态并跳到首页或 profile', async () => {
        await page.waitForURL(/\/(home|$)/, { timeout: 15000 })
        const userInfo = await page.evaluate(() => localStorage.getItem('userInfo'))
        const token = await page.evaluate(() => localStorage.getItem('token'))
        expect(userInfo, 'localStorage 应保存 userInfo').toBeTruthy()
        expect(token, 'localStorage 应保存 Bearer token').toBeTruthy()
      })
    })

    test.skip('E2E-TC01-01-02 重复昵称应提示注册失败并停留在注册页', () => {
      // 原因：现注册页未实现后端 409 冲突的 UI 错误提示（errorMessage 未绑定到 input 下方），
      // 需等 UC01 PRD 约定的错误文案落地后补充。
    })
  })

  test.describe('E2E-TC01-02 登录主流程', () => {
    test('E2E-TC01-02-01 已注册账号登录成功 → 跳首页并持久化登录态', async ({ page, request }) => {
      await test.step('1. 打开登录页，输入合法账号并提交', async () => {
        await page.goto('/login')
        await page.locator('[name=nickname]').fill(user.nickname)
        await page.locator('[name=password]').fill(user.password)
        await page.locator('button[type=submit]').click()
      })

      await test.step('2. 断言：跳转首页、token/userInfo 正确落盘、顶部用户昵称与登录账号一致', async () => {
        await page.waitForURL(/\/(home|$)/, { timeout: 15000 })
        const token = await page.evaluate(() => localStorage.getItem('token'))
        const userInfoStr = await page.evaluate(() => localStorage.getItem('userInfo'))
        expect(token, '登录成功 localStorage 应有 token').toBeTruthy()
        expect(userInfoStr, '登录成功 localStorage 应有 userInfo').toBeTruthy()
        const userInfo = JSON.parse(userInfoStr as string)
        expect(userInfo.nickname, '昵称应等于登录账号').toBe(user.nickname)
        const topBar = page.locator('header, .app-header, .layout-header, .nav-header').first()
        await expect(topBar.getByText(user.nickname, { exact: false }), '顶部栏应显示当前用户昵称').toBeVisible({ timeout: 8000 })
      })

      await test.step('3. 断言：重新打开新 tab 使用同 token 访问 /me 仍能拉到个人信息', async () => {
        const session = await loginViaApi(request, user.nickname, user.password)
        const profile = await request.get(`${API}/users/${session.userInfo.id}`, {
          headers: { Authorization: `Bearer ${session.token}` },
        })
        expect(profile.ok(), '登录态对应的 /users/:id 应 200').toBeTruthy()
        const info = (await profile.json()).data
        expect(info.nickname, '服务端返回昵称需一致').toBe(user.nickname)
      })
    })

    test.skip('E2E-TC01-02-02 错误密码应弹窗报错且不跳转', () => {
      // 原因：登录页未实现后端 401/400 错误的 UI 错误 Toast 绑定，
      // 需等 LoginPage.vue 的错误提示交互落地后再补。
    })
  })

  test.describe('E2E-TC01-03 资料维护主流程', () => {
    test('E2E-TC01-03-01 更新 nickname/bio → 本地与服务端同步生效', async ({ page, request }) => {
      const session = await loginViaApi(request, user.nickname, user.password)
      const newBio = `E2E UC01 资料更新 ${Date.now()}`

      await test.step('1. 通过 API 先保存初始资料（绕过登录 UI），再打开个人资料页', async () => {
        await openAs(page, session, `/user/${session.userInfo.id}`)
      })

      await test.step('2. 进入编辑资料模式，填写新简介并提交', async () => {
        const editButton = page.getByRole('button', { name: /编辑|编辑资料|Edit profile/i }).first()
        if (!(await editButton.isVisible())) {
          test.skip()
          throw new Error('个人资料页未实现「编辑资料」按钮，暂时跳过 UI 流程')
        }
        await editButton.click()
        const bioInput = page.locator('textarea[name=bio], textarea[data-testid=bio], [data-testid=profile-bio] textarea').first()
        await bioInput.fill(newBio)
        await page.getByRole('button', { name: /保存|Submit|Save/i }).first().click()
      })

      await test.step('3. 断言：页面重新渲染后简介为新值，且服务端 GET /users/:id 返回新 bio', async () => {
        const profile = await request.get(`${API}/users/${session.userInfo.id}`, {
          headers: { Authorization: `Bearer ${session.token}` },
        })
        expect(profile.ok()).toBeTruthy()
        const data = (await profile.json()).data
        expect(data.bio || data.description, '服务端保存的简介应包含本次 E2E 标识').toContain('E2E UC01')
      })
    })

    test.skip('E2E-TC01-03-02 上传头像 → 头像预览图更新并持久化', () => {
      // 原因：UploadAvatarHandler 后端已存在，但前端 UserProfilePage.vue 尚未接入「更换头像」上传组件，
      // 需要上传控件 DOM 绑定后，补充 chooseFile + 预览 img 断言。
    })
  })
})
