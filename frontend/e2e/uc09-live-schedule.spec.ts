import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

const runTag = Date.now()

test.describe('UC09 创建预约、另一用户预约与取消，刷新后状态保持', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC09-01 创建预约、另一用户预约与取消，刷新后状态保持', async ({ page, request }) => {
    // NOTE: 骨架占位 — 完整实现详见 d-engagement.spec.ts 第 40~70 行
    // 领域负责人（成员D）：请在此补全断言后移除 test.skip
    // 步骤：主播登录 → 打开直播页 → 点击预约直播 → 填标题和开播时间
    //       → 创建预约 → 验证提示成功和卡片显示
    //       → 切换至观众账号 → 打开直播页 → 点击预约提醒 → 验证已预约
    //       → 刷新页面 → 验证已预约状态持久化 → 取消预约 → 变回预约提醒
    //       → API 查询确认存在 → 清理删除
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  })
})
