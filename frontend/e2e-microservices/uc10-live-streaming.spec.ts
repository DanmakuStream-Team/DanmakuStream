import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

const runTag = Date.now()

test.describe('UC10 开播、观众互动、结束直播（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC10-01 创建直播、观众点赞赠礼、主播结束，页面与 API 一致', async ({ page, request }) => {
    // NOTE: 微服务版 SRS 推流、WebSocket live room、礼物/点赞回调依赖 engagement-service
    //       与 SRS 的 webhook 回调；待 webhook 打通 + 流文件准备完毕后移除 skip
    //       参考实现：../e2e/uc10-live-streaming.spec.ts
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  })
})
