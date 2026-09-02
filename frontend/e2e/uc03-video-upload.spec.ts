import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

test.describe('UC03 创作者取消上传后重新投稿并看到待审状态', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC03-01 前端取消请求后不显示成功，并可再次完成真实投稿', async ({ page, request }) => {
    // NOTE: 骨架占位 — 完整实现详见 member-c-content.spec.ts 第 38~86 行
    // 领域负责人（成员C）：请在此补全断言后移除 test.skip
    // 步骤：登录创作者 → 选择文件并填标题 → 拦截上传请求模拟慢请求
    //       → 点击终止上传 → 验证终止提示 → 重新投稿 → 验证待审核状态
    //       → 验证被取消的投稿不入库 → 验证转码失败场景
    const creator = await loginViaApi(request, USERS.memberCCreator.nickname, USERS.memberCCreator.password)
    await openAs(page, creator, '/creator/upload')
  })
})
