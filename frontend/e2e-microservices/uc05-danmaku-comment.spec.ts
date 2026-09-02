import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, ENGAGEMENT_VIDEO_TITLE, USERS } from './test-data'

const runTag = Date.now()

test.describe('UC05 弹幕、评论、点赞、收藏（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC05-01 发送弹幕和评论、点赞收藏，刷新后数据保持', async ({ page, request }) => {
    // NOTE: engagement-service 的弹幕/评论/点赞/收藏表结构与单体一致，
    //       但跨服务 join 视频标题取路由需补全；待联调后移除 skip
    //       参考实现：../e2e/uc05-danmaku-comment.spec.ts
    const viewer = await loginViaApi(request, USERS.viewer.nickname, USERS.viewer.password)
    const listResponse = await request.get(`${API}/videos?page=1&pageSize=20&keyword=${encodeURIComponent(ENGAGEMENT_VIDEO_TITLE)}`)
    expect(listResponse.ok()).toBeTruthy()
  })
})
