import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

/** Element Plus 下拉选项（渲染在 body 里的 popper，只取当前可见的那份） */
function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

// ---------- E2E-TC13-01 审核员处理内容（UI 登录 → 审核视频 → 屏蔽弹幕 → 刷新持久） ----------
test('E2E-TC13-01 审核员审核视频并屏蔽弹幕，刷新后状态保持', async ({ page }) => {
  await page.goto('/login')
  await page.getByPlaceholder('请输入昵称').fill(USERS.moderator.nickname)
  await page.getByPlaceholder('请输入密码').fill(USERS.moderator.password)
  await page.getByRole('button', { name: '登录' }).click()
  await expect(page).not.toHaveURL(/login/)

  // 视频审核：把 E2E 待审视频改为通过
  await page.goto('/admin/videos')
  const row = page.locator('.row', { hasText: 'E2E-UC13-待审视频' })
  await expect(row).toBeVisible()
  await row.locator('.el-select').click()
  const reviewResponsePromise = page.waitForResponse((response) =>
    /\/api\/v1\/admin\/videos\/\d+\/status$/.test(response.url())
      && response.request().method() === 'PUT',
  )
  await dropdownOption(page, '通过').click()
  const reviewResponse = await reviewResponsePromise
  expect(reviewResponse.ok(), `视频审核接口返回 ${reviewResponse.status()}: ${await reviewResponse.text()}`).toBeTruthy()
  await expect(row.locator('.el-tag')).toContainText('已通过')

  // 弹幕屏蔽
  await page.goto('/admin/danmaku')
  const drow = page.locator('.row', { hasText: 'E2E-UC13-待屏蔽弹幕' })
  await expect(drow).toBeVisible()
  await drow.getByRole('button', { name: '屏蔽', exact: true }).click()
  await expect(drow.locator('.el-tag')).toContainText('已屏蔽')

  // 刷新后弹幕屏蔽状态保持（视频已通过的持久化由 API 断言）
  await page.reload()
  await expect(page.locator('.row', { hasText: 'E2E-UC13-待屏蔽弹幕' }).locator('.el-tag')).toContainText('已屏蔽')
})

// ---------- E2E-TC13-02 管理员修改角色（搜索 → 修改 → 刷新保持 + 权限生效） ----------
test('E2E-TC13-02 管理员修改用户角色，刷新后保持且权限生效', async ({ page, request }) => {
  const admin = await loginViaApi(request, USERS.admin.nickname, USERS.admin.password)
  await openAs(page, admin, '/admin/users')

  const search = page.getByPlaceholder('搜索用户')
  await search.fill(USERS.target.nickname)
  await search.press('Enter')
  const row = page.locator('.user-row', { hasText: USERS.target.nickname })
  await expect(row).toBeVisible()

  await row.locator('.el-select').click()
  await dropdownOption(page, '内容审核员/版主').click()

  // 刷新 + 重新搜索后角色保持
  await page.reload()
  await page.getByPlaceholder('搜索用户').fill(USERS.target.nickname)
  await page.getByPlaceholder('搜索用户').press('Enter')
  const rowAfter = page.locator('.user-row', { hasText: USERS.target.nickname })
  await expect(rowAfter).toBeVisible()
  await expect(rowAfter).toContainText('内容审核员/版主')

  // 权限随之生效：tuser 重新登录后角色为 moderator
  const relogin = await loginViaApi(request, USERS.target.nickname, USERS.target.password)
  expect(relogin.userInfo.role).toBe('moderator')
})

// ---------- E2E-TC13-03 管理员维护横幅与公告（创建/编辑/删除，后台结果一致） ----------
test('E2E-TC13-03 管理员新建、修改、删除横幅和公告', async ({ page, request }) => {
  const admin = await loginViaApi(request, USERS.admin.nickname, USERS.admin.password)
  await openAs(page, admin, '/admin/operations')

  // -- 横幅：新建
  await page.getByRole('button', { name: '新增横幅' }).click()
  const dialog = page.locator('.el-dialog:visible')
  await dialog.getByLabel('标题').fill('E2E-UC13-横幅')
  await dialog.getByRole('button', { name: '保存' }).click()
  const bannerRow = page.locator('.banner-row', { hasText: 'E2E-UC13-横幅' })
  await expect(bannerRow).toBeVisible()

  // -- 横幅：编辑
  await bannerRow.getByRole('button', { name: '编辑' }).click()
  await dialog.getByLabel('标题').fill('E2E-UC13-横幅v2')
  await dialog.getByRole('button', { name: '保存' }).click()
  await expect(page.locator('.banner-row', { hasText: 'E2E-UC13-横幅v2' })).toBeVisible()

  // -- 横幅：删除
  await page.locator('.banner-row', { hasText: 'E2E-UC13-横幅v2' }).getByRole('button', { name: '删除' }).click()
  await expect(page.locator('.banner-row', { hasText: 'E2E-UC13-横幅v2' })).toHaveCount(0)

  // -- 公告：新建（带时间）
  await page.getByRole('button', { name: '新增公告' }).click()
  await dialog.getByLabel('公告内容').fill('E2E-UC13-公告')
  await dialog.locator('.el-form-item', { hasText: '开始时间' }).locator('input').fill('2026-08-26 00:00:00')
  await dialog.getByRole('button', { name: '保存' }).click()
  const annRow = page.locator('.announcement-row', { hasText: 'E2E-UC13-公告' })
  await expect(annRow).toBeVisible()

  // -- 公告：编辑
  await annRow.getByRole('button', { name: '编辑' }).click()
  await dialog.getByLabel('公告内容').fill('E2E-UC13-公告v2')
  await dialog.getByRole('button', { name: '保存' }).click()
  await expect(page.locator('.announcement-row', { hasText: 'E2E-UC13-公告v2' })).toBeVisible()

  // -- 公告：删除
  await page.locator('.announcement-row', { hasText: 'E2E-UC13-公告v2' }).getByRole('button', { name: '删除' }).click()
  await expect(page.locator('.announcement-row', { hasText: 'E2E-UC13-公告v2' })).toHaveCount(0)

  // 后台结果一致性：列表接口中两者均不存在
  const banners = await (await request.get(`${API}/admin/banners`, { headers: { Authorization: `Bearer ${admin.token}` } })).json()
  const anns = await (await request.get(`${API}/admin/announcements`, { headers: { Authorization: `Bearer ${admin.token}` } })).json()
  expect(JSON.stringify(banners)).not.toContain('E2E-UC13-')
  expect(JSON.stringify(anns)).not.toContain('E2E-UC13-')
})

// ---------- E2E-TC13-04 管理员查看基础设施（指标、路径、来源可见） ----------
test('E2E-TC13-04 管理员查看基础设施指标', async ({ page, request }) => {
  const admin = await loginViaApi(request, USERS.admin.nickname, USERS.admin.password)
  await openAs(page, admin, '/admin/infrastructure')

  await expect(page.getByText('存储空间')).toBeVisible()
  await expect(page.getByText('带宽与流量')).toBeVisible()
  await expect(page.getByText('CPU 使用率')).toBeVisible()
  await expect(page.getByText('在线与并发')).toBeVisible()
  // 指标值与来源可见：已用/剩余容量、流量统计来源说明
  await expect(page.getByText(/已用\s*[\d.]+\s*[KMG]B/)).toBeVisible()
  await expect(page.getByText(/剩余\s*[\d.]+\s*[KMG]B/)).toBeVisible()
  await expect(page.getByText(/Go 后端 middleware/)).toBeVisible()
})

// ---------- E2E-TC13-05 普通用户越权（页面拦截 + 接口 403） ----------
test('E2E-TC13-05 普通用户无法进入后台页面，接口返回 403', async ({ page, request }) => {
  const plain = await loginViaApi(request, USERS.plain.nickname, USERS.plain.password)

  // 页面层：路由守卫把已登录的非 staff 用户重定向回首页
  await openAs(page, plain, '/admin/videos')
  await expect(page).not.toHaveURL(/\/admin/)
  await expect(page.getByRole('heading', { name: '视频审核' })).toHaveCount(0)

  // 接口层：直接带 token 调后台接口返回 403，且不产生数据变更
  const res = await request.get(`${API}/admin/videos`, { headers: { Authorization: `Bearer ${plain.token}` } })
  expect(res.status()).toBe(403)
})
