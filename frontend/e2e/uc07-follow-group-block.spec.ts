import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

test.describe('UC07 关注、分组与黑名单状态在页面可验证', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC07-01 关注、分组与黑名单状态在页面可验证', async ({ page, request }) => {
    // NOTE: 骨架占位 — 完整实现详见 member-b-workflows.spec.ts 第 6~50 行
    // 领域负责人（成员B）：请在此补全断言后移除 test.skip
    // 步骤：登录双方用户 → 关注主播 → API 返回 followed=true
    //       → 创建关注分组 → 将主播加入分组
    //       → 打开关注管理页 → 验证分组名和关注用户可见
    //       → 主播发布动态 → 粉丝收到关注通知
    //       → 拉黑主播 → API 返回 blocked=true → 刷新后黑名单页可见
    //       → 解除拉黑 → API 返回 blocked=false
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  })
})
