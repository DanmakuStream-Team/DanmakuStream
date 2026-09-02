import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

test.describe('UC11 会话、WebSocket 实时消息、已读和媒体分享', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC11-01 会话、WebSocket 实时消息、已读和媒体分享', async ({ page, request }) => {
    // NOTE: 骨架占位 — 完整实现详见 member-b-workflows.spec.ts 第 87~187 行
    // 领域负责人（成员B）：请在此补全断言后移除 test.skip
    // 步骤：双方登录 → 观众先发3条未读消息 → 主播首页 badge 显示未读3
    //       → 安装 WS 跟踪脚本 → 打开私信页 → 验证聊天头部和消息
    //       → 验证未读计数 API 变为 0
    //       → 主动关闭聊天 WS → 验证离线图标
    //       → 离线期间发历史消息 → 重连后在线图标 → 刷新可见
    //       → 发实时文本 → 页面立即可见
    //       → 页面输入框发回复 → 可见
    //       → clientMessageId 幂等重试(两次同ID返回同一消息ID)
    //       → 发送图片附件 → message-image 可见
    //       → 发送视频附件 → message-video 可见
    //       → 分享视频卡 → 视频标题消息卡片可见
    //       → 最终未读=0，历史API查询包含回复和 video_share
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  })
})
