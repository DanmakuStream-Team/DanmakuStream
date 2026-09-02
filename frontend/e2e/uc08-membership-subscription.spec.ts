import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

test.describe('UC08 配置套餐、页面订阅、幂等支付与特别关注', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC08-01 配置套餐、页面订阅、幂等支付与特别关注', async ({ page, request }) => {
    // NOTE: 骨架占位 — 完整实现详见 member-b-workflows.spec.ts 第 52~85 行
    // 领域负责人（成员B）：请在此补全断言后移除 test.skip
    // 步骤：主播设置会员套餐(6元/月) → 粉丝打开主播个人页 → 点击订阅按钮
    //       → 弹出付费特别关注对话框 → 确认订阅 → 提示已开通
    //       → API 查询订阅状态 active=true → 查询订单列表
    //       → 对同一订单重复支付(幂等) → 订阅列表仍为1条
    //       → 打开关注管理页 → 按钮显示已付费特别关注
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  })
})
