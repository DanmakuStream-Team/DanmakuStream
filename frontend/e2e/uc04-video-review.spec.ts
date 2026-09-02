import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

test.describe('UC04 审核员发布与拒绝视频，普通用户不能越权', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC04-01 审核员通过/拒绝，普通用户同时受页面和 API 权限保护', async ({ page, request }) => {
    // NOTE: 骨架占位 — 完整实现详见 member-c-content.spec.ts 第 89~126 行
    // 领域负责人（成员C）：请在此补全断言后移除 test.skip
    // 步骤：登录审核员 → 通过待审视频 → 验证已通过状态且状态不可再改
    //       → API 重复修改状态返回 409 → 拒绝另一视频 → 验证已拒绝
    //       → 普通用户访问 /admin/videos 被重定向 → API 访问返回 403
    const moderator = await loginViaApi(request, USERS.memberCModerator.nickname, USERS.memberCModerator.password)
    await openAs(page, moderator, '/admin/videos')
  })
})
