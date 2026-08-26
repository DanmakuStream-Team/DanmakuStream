import { expect, type Page } from '@playwright/test'
import { API } from '../test-data'

export interface Session {
  token: string
  userInfo: { id: number; nickname: string; role: string; [k: string]: unknown }
}

/** 通过 API 登录换取 token（比 UI 登录快且稳定；UI 登录在 E2E-TC13-01 中单独覆盖一次） */
export async function loginViaApi(request: import('@playwright/test').APIRequestContext, nickname: string, password: string): Promise<Session> {
  const res = await request.post(`${API}/auth/login`, { data: { nickname, password } })
  expect(res.ok(), `login ${nickname} should succeed`).toBeTruthy()
  const body = await res.json()
  return { token: body.data.token, userInfo: body.data.userInfo }
}

/** 把会话注入 localStorage（auth store 从 localStorage.token / userInfo 初始化）后打开页面 */
export async function openAs(page: Page, session: Session, path: string) {
  await page.addInitScript(
    ({ token, userInfo }) => {
      localStorage.setItem('token', token)
      localStorage.setItem('userInfo', JSON.stringify(userInfo))
    },
    { token: session.token, userInfo: session.userInfo },
  )
  await page.goto(path)
}
