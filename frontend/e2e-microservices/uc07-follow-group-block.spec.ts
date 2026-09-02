import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

test.describe('UC07 关注、分组与黑名单（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC07-01 关注、分组、动态通知、黑名单，状态持久化', async ({ page, request }) => {
    // NOTE: user-service 已提供 follow / follow-groups / block API，
    //       微服务前端的订阅页/通知中心需和 content-service 动态模块联调
    //       待通知推送跨服务联调通过后移除 skip
    //       参考实现：../e2e/uc07-follow-group-block.spec.ts
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  })
})
