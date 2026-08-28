import { expect, type APIRequestContext, type Page } from '@playwright/test'
import { API } from '../test-data'

export interface Session {
  token: string
  userInfo: { id: number; username?: string; nickname: string; role: string; [key: string]: unknown }
}

/** 通过 API 登录换取 token，比重复执行 UI 登录更快；UC13 另有独立 UI 登录用例。 */
export async function loginViaApi(request: APIRequestContext, nickname: string, password: string): Promise<Session> {
  const response = await request.post(`${API}/auth/login`, { data: { nickname, password } })
  expect(response.ok(), `login failed for ${nickname}: ${await response.text()}`).toBeTruthy()
  const body = await response.json()
  return { token: body.data.token, userInfo: body.data.userInfo }
}

/** 注入前端认证状态后打开指定页面。 */
export async function openAs(page: Page, session: Session, path: string) {
  await page.addInitScript(
    ({ token, userInfo }) => {
      localStorage.setItem('token', token)
      localStorage.setItem('userInfo', JSON.stringify(userInfo))
    },
    { token: session.token, userInfo: session.userInfo },
  )
  await page.goto(path, { waitUntil: 'domcontentloaded', timeout: 30_000 })
}
