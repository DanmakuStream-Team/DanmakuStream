import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

const runTag = Date.now()

test('E2E-TC09 创建预约、另一用户预约与取消，刷新后状态保持', async ({ page, request }) => {
  const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
  const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  const title = `E2E-UC09-${runTag}`

  await openAs(page, owner, '/live')
  await page.getByRole('button', { name: '预约直播', exact: true }).click()
  const dialog = page.locator('.el-dialog:visible')
  await dialog.getByPlaceholder('输入预约标题').fill(title)
  const future = new Date(Date.now() + 24 * 60 * 60 * 1000)
  const value = `${future.getFullYear()}-${String(future.getMonth() + 1).padStart(2, '0')}-${String(future.getDate()).padStart(2, '0')} 10:00:00`
  await dialog.getByPlaceholder('选择开播时间').fill(value)
  await dialog.getByPlaceholder('选择开播时间').press('Enter')
  await dialog.getByRole('button', { name: '创建预约', exact: true }).click()
  await expect(page.getByText('直播预约已创建')).toBeVisible()
  await expect(page.locator('.schedule-card', { hasText: title })).toBeVisible()

  await openAs(page, viewer, '/live')
  const viewerCard = page.locator('.schedule-card', { hasText: title })
  await viewerCard.getByRole('button', { name: '预约提醒', exact: true }).click()
  await expect(viewerCard.getByRole('button', { name: '已预约', exact: true })).toBeVisible()
  await page.reload()
  await expect(page.locator('.schedule-card', { hasText: title }).getByRole('button', { name: '已预约', exact: true })).toBeVisible()
  await page.locator('.schedule-card', { hasText: title }).getByRole('button', { name: '已预约', exact: true }).click()
  await expect(page.locator('.schedule-card', { hasText: title }).getByRole('button', { name: '预约提醒', exact: true })).toBeVisible()

  const list = await (await request.get(`${API}/live-schedules?status=pending&page=1&pageSize=100`, { headers: { Authorization: `Bearer ${owner.token}` } })).json()
  const schedule = list.data.list.find((item: { title: string }) => item.title === title)
  expect(schedule).toBeTruthy()
  await request.delete(`${API}/live-schedules/${schedule.id}`, { headers: { Authorization: `Bearer ${owner.token}` } })
})

test('E2E-TC10 创建直播、观众点赞赠礼、主播结束，页面与 API 一致', async ({ page, request }) => {
  const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
  const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  const title = `E2E-UC10-${runTag}`

  await openAs(page, owner, '/live')
  await page.getByRole('button', { name: '开始直播', exact: true }).first().click()
  const dialog = page.locator('.el-dialog:visible')
  await dialog.getByPlaceholder('输入直播标题').fill(title)
  await dialog.getByRole('button', { name: '开始直播', exact: true }).click()
  // OBS is the default mode; successful creation is evidenced by generated stream parameters.
  await expect(dialog.getByText('串流密钥', { exact: true })).toBeVisible()
  await Promise.all([
    page.waitForURL(/\/live\/\d+$/),
    dialog.getByRole('button', { name: '进入直播间', exact: true }).click(),
  ])
  await expect(page.getByRole('heading', { name: title })).toBeVisible()
  const roomID = Number(new URL(page.url()).pathname.match(/\/live\/(\d+)/)?.[1])
  expect(roomID).toBeGreaterThan(0)

  await openAs(page, viewer, `/live/${roomID}`)
  await expect(page.getByText('直播中', { exact: true })).toBeVisible()
  await page.getByRole('button', { name: /^点赞/ }).click()
  await expect(page.getByRole('button', { name: /^已点赞/ })).toBeVisible()
  await page.getByRole('button', { name: '赠送礼物', exact: true }).click()
  const giftDialog = page.locator('.el-dialog:visible')
  await giftDialog.locator('.gift-grid button').first().click()
  await giftDialog.getByRole('button', { name: '确认赠送', exact: true }).click()
  await expect(page.getByText(/已送出/)).toBeVisible()

  const interaction = await (await request.get(`${API}/live/${roomID}/interaction`)).json()
  expect(interaction.data.likeCount).toBe(1)
  expect(interaction.data.giftValue).toBeGreaterThan(0)

  await openAs(page, owner, `/live/${roomID}`)
  await page.getByRole('button', { name: '结束直播', exact: true }).click()
  await expect(page).toHaveURL(/\/live$/)
  const detail = await request.get(`${API}/live/${roomID}`)
  expect(detail.status()).toBe(404)
})
