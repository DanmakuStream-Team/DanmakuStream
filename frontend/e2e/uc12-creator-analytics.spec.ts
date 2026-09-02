import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

test.describe('UC12 创作者切换 7 天和单作品分析范围', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC12-01 时间范围和作品范围切换，接口与图表同步刷新', async ({ page, request }) => {
    // NOTE: 骨架占位 — 完整实现详见 member-c-content.spec.ts 第 129~160 行
    // 领域负责人（成员C）：请在此补全断言后移除 test.skip
    // 步骤：创作者登录 → 打开创作中心 → 数据趋势标题可见
    //       → 点击近 7 天 → 等待 days=7 接口 → 新增观看显示 15或16
    //       → 作品下拉选择目标视频 → 等待带 videoId= 的 analytics 接口
    //       → 验证 topVideos 仅1条且 id === selectedVideoId
    //       → 验证"X的观看、收藏增长和账号开播次数"说明文字
    //       → 其他账号 API 访问该创作者 analytics 返回 404
    const creator = await loginViaApi(request, USERS.memberCCreator.nickname, USERS.memberCCreator.password)
  })
})
