import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

test.describe('UC11 私信、实时消息、媒体分享（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC11-01 会话、WebSocket 实时消息、已读和媒体分享', async ({ page, request }) => {
    // NOTE: engagement-service 的消息存储 + WebSocket /ws/chat 路由已在网关预留，
    //       但未读数徽章、媒体上传路径、分享视频卡需前端跨服务联调
    //       待联调完毕后移除 skip
    //       参考实现：../e2e/uc11-chat-message.spec.ts
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  })
})
