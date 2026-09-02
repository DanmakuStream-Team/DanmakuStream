import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

test.describe('UC03 创作者取消上传后重新投稿（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC03-01 前端取消请求后不显示成功，可再次完成真实投稿', async ({ page, request }) => {
    // NOTE: 微服务版需完善上传路径路由 + content-service 真实文件落盘
    //       目前 content-service 的 multipart 上传需对接媒体卷共享；待运维补全后移除 skip
    //       参考实现：../e2e/uc03-video-upload.spec.ts
    const creator = await loginViaApi(request, USERS.memberCCreator.nickname, USERS.memberCCreator.password)
    await openAs(page, creator, '/creator/upload')
  })
})
