import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

const runTag = Date.now()

test.describe('UC09 直播预约与提醒（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC09-01 创建预约、另一用户预约与取消，刷新后状态保持', async ({ page, request }) => {
    // NOTE: engagement-service 的 live-schedules 表 + 提醒推送需
    //       与 content-service 作者信息跨服务 join；待联调后移除 skip
    //       参考实现：../e2e/uc09-live-schedule.spec.ts
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  })
})
