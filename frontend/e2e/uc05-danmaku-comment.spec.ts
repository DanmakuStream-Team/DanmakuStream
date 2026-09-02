import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, ENGAGEMENT_VIDEO_TITLE, USERS } from './test-data'

const runTag = Date.now()

test.describe('UC05 发送弹幕和评论、点赞收藏，刷新后数据保持', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC05-01 发送弹幕和评论、点赞收藏，刷新后数据保持', async ({ page, request }) => {
    // NOTE: 骨架占位 — 完整实现详见 d-engagement.spec.ts 第 7~38 行
    // 领域负责人（成员D）：请在此补全断言后移除 test.skip
    // 步骤：登录观众 → 获取目标视频 → 打开播放页 → 点赞(0→1)
    //       → 收藏(0→1) → 发送弹幕(验证 1 弹幕) → 发表评论(验证可见)
    //       → 刷新页面 → 验证点赞、收藏、弹幕、评论全部持久化
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  })
})
