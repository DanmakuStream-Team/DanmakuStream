import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

const runTag = Date.now()

test('E2E-TC01 注册、重新登录并维护个人资料', async ({ page, request }) => {
  const nickname = `e2e-uc01-${runTag}`
  const password = 'Test1234!'

  await page.goto('/register')
  await page.getByPlaceholder('请输入昵称').fill(nickname)
  await page.getByPlaceholder('请输入密码').fill(password)
  await page.getByRole('button', { name: '注册', exact: true }).click()
  await expect(page.getByText('注册成功')).toBeVisible()
  await expect(page).toHaveURL('/')

  const userInfo = await page.evaluate(() => JSON.parse(localStorage.getItem('userInfo') || '{}')) as { id: number }
  expect(userInfo.id).toBeGreaterThan(0)
  await page.evaluate(() => localStorage.clear())
  await page.goto('/login')
  await page.getByPlaceholder('请输入昵称').fill(nickname)
  await page.getByPlaceholder('请输入密码').fill(password)
  await page.getByRole('button', { name: '登录', exact: true }).click()
  await expect(page.getByText('登录成功')).toBeVisible()

  await page.goto(`/user/${userInfo.id}`)
  await page.locator('.bio-display').click()
  const editor = page.locator('.bio-editor')
  await editor.getByPlaceholder('写一段个人简介').fill(`UC01 profile ${runTag}`)
  await editor.getByRole('button', { name: '保存', exact: true }).click()
  await expect(page.getByText(`UC01 profile ${runTag}`)).toBeVisible()
  await page.reload()
  await expect(page.getByText(`UC01 profile ${runTag}`)).toBeVisible()

  const profile = await (await request.get(`${API}/users/${userInfo.id}`)).json()
  expect(profile.data.bio).toBe(`UC01 profile ${runTag}`)
})

test('E2E-TC07 关注、拉黑与解除拉黑形成完整关系流程', async ({ page, request }) => {
  const viewer = await loginViaApi(request, USERS.domainViewer.nickname, USERS.domainViewer.password)
  const creator = await loginViaApi(request, USERS.domainCreator.nickname, USERS.domainCreator.password)

  await openAs(page, viewer, `/user/${creator.userInfo.id}`)
  await page.getByRole('button', { name: '关注', exact: true }).click()
  await expect(page.getByRole('button', { name: '已关注', exact: true })).toBeVisible()

  await page.getByRole('button', { name: '拉黑', exact: true }).click()
  const popover = page.locator('.el-popconfirm:visible')
  await popover.getByRole('button', { name: '确定', exact: true }).click()
  await expect(page.getByRole('button', { name: '已拉黑', exact: true })).toBeVisible()
  const blocked = await (await request.get(`${API}/users/blocked`, { headers: { Authorization: `Bearer ${viewer.token}` } })).json()
  expect(blocked.data.list.some((item: { id: number }) => item.id === creator.userInfo.id)).toBeTruthy()

  await page.getByRole('button', { name: '解除拉黑', exact: true }).click()
  await page.locator('.el-popconfirm:visible').getByRole('button', { name: '确定', exact: true }).click()
  await expect(page.getByRole('button', { name: '关注', exact: true })).toBeVisible()
})

test('E2E-TC08 订阅创作者、演示支付并展示特别关注', async ({ page, request }) => {
  const viewer = await loginViaApi(request, USERS.domainViewer.nickname, USERS.domainViewer.password)
  const creator = await loginViaApi(request, USERS.domainCreator.nickname, USERS.domainCreator.password)

  await openAs(page, viewer, `/user/${creator.userInfo.id}`)
  await page.locator('.membership-action').click()
  const dialog = page.locator('.el-dialog:visible')
  await expect(dialog.getByText('这是项目演示支付，不会产生真实扣款。')).toBeVisible()
  await dialog.getByRole('button', { name: '确认订阅', exact: true }).click()
  await expect(page.getByText('付费特别关注已开通')).toBeVisible()
  await expect(page.locator('.membership-action')).toContainText('特别关注')

  const status = await (await request.get(`${API}/subscriptions/creators/${creator.userInfo.id}/status`, {
    headers: { Authorization: `Bearer ${viewer.token}` },
  })).json()
  expect(status.data.active).toBe(true)
  expect(status.data.subscription.status).toBe('active')
})

test('E2E-TC11 页面发送私信、接收方读取并与 API 状态一致', async ({ page, request }) => {
  const viewer = await loginViaApi(request, USERS.domainViewer.nickname, USERS.domainViewer.password)
  const other = await loginViaApi(request, USERS.domainOther.nickname, USERS.domainOther.password)
  const content = `E2E-UC11-${runTag}`

  await openAs(page, viewer, `/messages/${other.userInfo.id}`)
  await page.getByPlaceholder('输入消息，Enter 发送，Shift + Enter 换行').fill(content)
  await page.getByRole('button', { name: '发送', exact: true }).click()
  await expect(page.locator('.message-row.mine').getByText(content, { exact: true })).toBeVisible()

  const unreadBefore = await (await request.get(`${API}/messages/unread`, {
    headers: { Authorization: `Bearer ${other.token}` },
  })).json()
  expect(unreadBefore.data.count).toBeGreaterThan(0)

  await openAs(page, other, `/messages/${viewer.userInfo.id}`)
  await expect(page.locator('.message-row:not(.mine)').getByText(content, { exact: true })).toBeVisible()
  const unreadAfter = await (await request.get(`${API}/messages/unread`, {
    headers: { Authorization: `Bearer ${other.token}` },
  })).json()
  expect(unreadAfter.data.count).toBe(0)
})
