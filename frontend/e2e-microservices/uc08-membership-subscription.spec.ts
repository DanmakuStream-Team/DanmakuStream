import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

test.describe('UC08 付费特别关注与订阅（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC08-01 配置套餐、页面订阅、幂等支付', async ({ page, request }) => {
    // NOTE: 会员/订单/支付目前在 user-service 内，支付回调与 Stripe Demo 模式
    //       需在微服务 compose 中增加 mock 支付回调服务；待补全后移除 skip
    //       参考实现：../e2e/uc08-membership-subscription.spec.ts
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  })
})
