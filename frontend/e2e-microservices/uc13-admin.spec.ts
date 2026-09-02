import { expect, test } from '@playwright/test'
import { loginViaApi, openAs } from './fixtures/auth'
import { API, USERS } from './test-data'

const runTag = Date.now()

test.describe('UC13 管理员用户与内容管理（微服务版）', () => {
  test.describe.configure({ mode: 'serial' })

  test.skip('E2E-TC13-01 管理员禁用/启用用户，普通用户不能访问管理页', async ({ page, request }) => {
    // NOTE: admin/v1/users 在 user-service 内部接口，网关需显式路由
    //       同时前端管理页需和三服务的禁用状态联动；待打通后移除 skip
    const admin = await loginViaApi(request, USERS.admin.nickname, USERS.admin.password)
    await openAs(page, admin, '/admin/users')
  })

  test.skip('E2E-TC13-02 管理员下架违规视频，搜索和播放页不再可见', async ({ page, request }) => {
    // NOTE: content-service 的管理员下架 + 搜索过滤 + 播放页 404 联动
    //       待打通 content-service admin 路由后移除 skip
    const admin = await loginViaApi(request, USERS.admin.nickname, USERS.admin.password)
    await openAs(page, admin, '/admin/videos')
  })
})
