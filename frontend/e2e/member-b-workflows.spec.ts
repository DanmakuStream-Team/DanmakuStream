/**
 * ============================================================
 *  LEGACY FILE — 内容已迁移为独立骨架文件（Day4 任务）
 * ============================================================
 *  本文件内的 UC 用例已按 UC 编号拆分到以下独立骨架文件，
 *  各自带 test.skip 占位并标注了对应领域负责人。
 *  待领域负责人补全断言后，可删除本文件。
 *
 *   UC07  →  uc07-follow-group-block.spec.ts        (成员B)
 *   UC08  →  uc08-membership-subscription.spec.ts   (成员B)
 *   UC11  →  uc11-chat-message.spec.ts              (成员B)
 * ============================================================
 */
import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

test.describe.serial('UC07 / UC08 / UC11 完整业务流程 [LEGACY — 见上方注释]', () => {
  test('E2E-TC07 关注、分组与黑名单状态在页面可验证', async ({ page, request }) => {
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
    const headers = { Authorization: `Bearer ${viewer.token}` }

    const follow = await request.post(`${API}/users/${owner.userInfo.id}/follow`, { headers })
    expect(follow.ok(), await follow.text()).toBeTruthy()
    expect((await follow.json()).data.followed).toBe(true)

    const groupName = `E2E核心${Date.now() % 100000}`
    const createGroup = await request.post(`${API}/users/follow-groups`, { headers, data: { name: groupName } })
    expect(createGroup.ok(), await createGroup.text()).toBeTruthy()
    const groupID = (await createGroup.json()).data.id as number
    const bind = await request.put(`${API}/users/${owner.userInfo.id}/follow-settings`, {
      headers,
      data: { groupId: groupID },
    })
    expect(bind.ok(), await bind.text()).toBeTruthy()

    await openAs(page, viewer, '/subscriptions')
    await expect(page.getByRole('heading', { name: '关注管理' })).toBeVisible()
    await expect(page.getByText(groupName, { exact: true }).first()).toBeVisible()
    await expect(page.getByText(USERS.owner.nickname, { exact: true }).last()).toBeVisible()

    const dynamicText = `UC07 关注通知 ${Date.now()}`
    const dynamic = await request.post(`${API}/dynamics`, {
      headers: { Authorization: `Bearer ${owner.token}` },
      data: { content: dynamicText },
    })
    expect(dynamic.ok(), await dynamic.text()).toBeTruthy()
    await page.getByRole('button', { name: '通知' }).click()
    await expect(page.getByText('你关注的用户发布了新动态', { exact: true })).toBeVisible()
    await expect(page.getByText(dynamicText, { exact: true })).toBeVisible()

    const block = await request.post(`${API}/users/${owner.userInfo.id}/block`, { headers })
    expect(block.ok(), await block.text()).toBeTruthy()
    expect((await block.json()).data.blocked).toBe(true)
    await page.reload()
    await page.getByRole('button', { name: /黑名单/ }).click()
    await expect(page.getByText(USERS.owner.nickname, { exact: true }).last()).toBeVisible()

    const unblock = await request.post(`${API}/users/${owner.userInfo.id}/block`, { headers })
    expect(unblock.ok(), await unblock.text()).toBeTruthy()
    expect((await unblock.json()).data.blocked).toBe(false)
  })

  test('E2E-TC08 配置套餐、页面订阅、幂等支付与特别关注', async ({ page, request }) => {
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
    const ownerHeaders = { Authorization: `Bearer ${owner.token}` }
    const viewerHeaders = { Authorization: `Bearer ${viewer.token}` }

    const updatePlan = await request.put(`${API}/creator/membership-plan`, {
      headers: ownerHeaders,
      data: { priceCents: 600, benefits: 'E2E 专属权益', enabled: true },
    })
    expect(updatePlan.ok(), await updatePlan.text()).toBeTruthy()

    await openAs(page, viewer, `/user/${owner.userInfo.id}`)
    const membershipButton = page.getByRole('button', { name: /¥6\/月|特别关注|确认续费/ }).first()
    await expect(membershipButton).toBeVisible()
    await membershipButton.click()
    await expect(page.getByRole('dialog', { name: '付费特别关注' })).toBeVisible()
    await page.getByRole('button', { name: '确认订阅' }).click()
    await expect(page.getByText('付费特别关注已开通')).toBeVisible()

    const status = await request.get(`${API}/subscriptions/creators/${owner.userInfo.id}/status`, { headers: viewerHeaders })
    expect(status.ok(), await status.text()).toBeTruthy()
    expect((await status.json()).data.active).toBe(true)

    const orders = await request.get(`${API}/subscriptions/orders`, { headers: viewerHeaders })
    const order = (await orders.json()).data.list[0]
    const payAgain = await request.post(`${API}/subscriptions/orders/${order.orderNo}/demo-pay`, { headers: viewerHeaders })
    expect(payAgain.ok(), await payAgain.text()).toBeTruthy()
    const subscriptions = await request.get(`${API}/subscriptions`, { headers: viewerHeaders })
    expect((await subscriptions.json()).data.list).toHaveLength(1)

    await openAs(page, viewer, '/subscriptions')
    await expect(page.getByRole('button', { name: '已付费特别关注' })).toBeVisible()
  })

  test('E2E-TC11 会话、WebSocket 实时消息、已读和媒体分享', async ({ page, request }) => {
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
    const ownerHeaders = { Authorization: `Bearer ${owner.token}` }
    const viewerHeaders = { Authorization: `Bearer ${viewer.token}` }

    for (let index = 1; index <= 3; index += 1) {
      const unreadMessage = await request.post(`${API}/messages`, {
        headers: viewerHeaders,
        data: { receiverId: owner.userInfo.id, clientMessageId: `uc11-unread-${index}-${Date.now()}`, type: 'text', content: `UC11 未读消息 ${index}` },
      })
      expect(unreadMessage.ok(), await unreadMessage.text()).toBeTruthy()
    }
    await openAs(page, owner, '/')
    await expect(page.getByTitle('私信').locator('xpath=ancestor::*[contains(@class,"el-badge")]').getByText('3', { exact: true })).toBeVisible()

    await page.addInitScript(() => {
      const NativeWebSocket = window.WebSocket
      const trackedWindow = window as Window & { __e2eChatSockets?: WebSocket[] }
      trackedWindow.__e2eChatSockets = []
      class TrackedWebSocket extends NativeWebSocket {
        constructor(url: string | URL, protocols?: string | string[]) {
          super(url, protocols)
          trackedWindow.__e2eChatSockets?.push(this)
        }
      }
      window.WebSocket = TrackedWebSocket
    })
    await openAs(page, owner, `/messages/${viewer.userInfo.id}`)
    await expect(page.locator('.chat-head strong')).toHaveText(USERS.viewer.nickname)
    await expect(page.locator('.message-scroll p').filter({ hasText: 'UC11 未读消息 3' })).toBeVisible()
    await expect.poll(async () => {
      const openedUnread = await request.get(`${API}/messages/unread`, { headers: ownerHeaders })
      return (await openedUnread.json()).data.count
    }).toBe(0)

    await page.evaluate(() => {
      const trackedWindow = window as Window & { __e2eChatSockets?: WebSocket[] }
      trackedWindow.__e2eChatSockets?.find(socket => socket.url.includes('/ws/chat'))?.close()
    })
    await expect(page.locator('.panel-head i')).not.toHaveClass(/online/)
    const offlineText = `UC11 离线历史 ${Date.now()}`
    const offlineSend = await request.post(`${API}/messages`, {
      headers: viewerHeaders,
      data: { receiverId: owner.userInfo.id, clientMessageId: `offline-${Date.now()}`, type: 'text', content: offlineText },
    })
    expect(offlineSend.ok(), await offlineSend.text()).toBeTruthy()
    await expect(page.locator('.panel-head i')).toHaveClass(/online/, { timeout: 8_000 })
    await page.reload()
    await expect(page.locator('.message-scroll p').filter({ hasText: offlineText })).toBeVisible()

    const realtimeText = `UC11 实时消息 ${Date.now()}`
    const send = await request.post(`${API}/messages`, {
      headers: viewerHeaders,
      data: { receiverId: owner.userInfo.id, type: 'text', content: realtimeText },
    })
    expect(send.ok(), await send.text()).toBeTruthy()
    await expect(page.locator('.message-scroll p').filter({ hasText: realtimeText })).toBeVisible()

    const replyText = `UC11 页面回复 ${Date.now()}`
    await page.getByPlaceholder('输入消息，Enter 发送，Shift + Enter 换行').fill(replyText)
    await page.getByRole('button', { name: '发送', exact: true }).click()
    await expect(page.locator('.message-scroll p').filter({ hasText: replyText })).toBeVisible()

    const retryID = `uc11-idempotent-${Date.now()}`
    const retryPayload = { receiverId: owner.userInfo.id, clientMessageId: retryID, type: 'text', content: 'UC11 幂等重试' }
    const retryOne = await request.post(`${API}/messages`, { headers: viewerHeaders, data: retryPayload })
    const retryTwo = await request.post(`${API}/messages`, { headers: viewerHeaders, data: retryPayload })
    expect((await retryOne.json()).data.id).toBe((await retryTwo.json()).data.id)

    await page.locator('.composer-tools input[accept*="image/png"]').setInputFiles({
      name: 'uc11-picture.png',
      mimeType: 'image/png',
      buffer: Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=', 'base64'),
    })
    await expect(page.locator('.message-scroll .message-image')).toBeVisible()

    await page.locator('.composer-tools input[accept*="video/mp4"]').setInputFiles({
      name: 'uc11-clip.mp4',
      mimeType: 'video/mp4',
      buffer: Buffer.from([0, 0, 0, 24, 0x66, 0x74, 0x79, 0x70, 0x69, 0x73, 0x6f, 0x6d, 0, 0, 0, 0]),
    })
    await expect(page.locator('.message-scroll .message-video')).toContainText('uc11-clip.mp4')

    const videos = await request.get(`${API}/videos?keyword=E2E-MEMBER-B-分享视频&page=1&pageSize=20`)
    const video = (await videos.json()).data.list.find((item: { title: string }) => item.title === 'E2E-MEMBER-B-分享视频')
    expect(video?.id).toBeTruthy()
    const share = await request.post(`${API}/messages`, {
      headers: viewerHeaders,
      data: { receiverId: owner.userInfo.id, type: 'video_share', videoId: video.id },
    })
    expect(share.ok(), await share.text()).toBeTruthy()
    await expect(page.locator('.message-scroll').getByText('E2E-MEMBER-B-分享视频', { exact: true })).toBeVisible()

    const unread = await request.get(`${API}/messages/unread`, { headers: ownerHeaders })
    expect((await unread.json()).data.count).toBe(0)
    const history = await request.get(`${API}/messages/${owner.userInfo.id}`, { headers: viewerHeaders })
    const historyList = (await history.json()).data.list as { content: string; type: string }[]
    expect(historyList.some((item) => item.content === replyText)).toBeTruthy()
    expect(historyList.some((item) => item.type === 'video_share')).toBeTruthy()
  })
})
