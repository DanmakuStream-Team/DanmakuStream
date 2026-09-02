import { expect, test, type Locator } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

function dropdownOption(page: import('@playwright/test').Page, text: string): Locator {
  return page.locator('.el-select-dropdown:visible .el-select-dropdown__item', { hasText: text })
}

test.describe('UC12 创作者数据分析（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC12-01 时间范围和作品范围切换，接口与图表同步刷新', async ({ page, request }) => {
    // NOTE: analytics 聚合在 content-service 内，跨服务拉 engagement 观看/收藏数据
    //       需补全跨服务数据查询 + 图表渲染联调；待完成后移除 skip
    //       参考实现：../e2e/uc12-creator-analytics.spec.ts
    const creator = await loginViaApi(request, USERS.memberCCreator.nickname, USERS.memberCCreator.password)
    await openAs(page, creator, '/creator')
  })
})
