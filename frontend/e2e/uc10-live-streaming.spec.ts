import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

const runTag = Date.now()

test.describe('UC10 创建直播、观众点赞赠礼、主播结束，页面与 API 一致', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC10-01 创建直播、观众点赞赠礼、主播结束，页面与 API 一致', async ({ page, request }) => {
    // NOTE: 骨架占位 — 完整实现详见 d-engagement.spec.ts 第 72~135 行
    // 领域负责人（成员D）：请在此补全断言后移除 test.skip
    // 步骤：安装 WebSocket 跟踪脚本 → 主播登录 → 开始直播 → 填标题
    //       → 生成串流密钥 → 进入直播间 → 验证标题可见
    //       → 观众进入 → 验证直播中 + 已连接 + 等待直播流
    //       → 主动断开 WS → 验证未连接 → 重连后已连接 → 1人观看
    //       → 点赞按钮 → 已点赞 → 赠送礼物 → 确认赠送 → 已送出
    //       → API 查询互动数据(likeCount=1, giftValue>0)
    //       → 主播结束直播 → 返回 /live → 详情 API 404
    const owner = await loginViaApi(request, USERS.owner.nickname, USERS.owner.password)
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
  })
})
