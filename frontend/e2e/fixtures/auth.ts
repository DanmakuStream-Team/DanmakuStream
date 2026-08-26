import { expect, type APIRequestContext, type Page } from '@playwright/test'
import { API } from '../test-data'

export interface Session {
  token: string
  userInfo: { id: number; username: string; nickname: string; role: string }
}

export async function loginViaApi(request: APIRequestContext, nickname: string, password: string): Promise<Session> {
  const response = await request.post(`${API}/auth/login`, { data: { nickname, password } })
  expect(response.ok(), `login failed for ${nickname}: ${await response.text()}`).toBeTruthy()
  const body = await response.json()
  return body.data as Session
}

export async function openAs(page: Page, session: Session, path: string) {
  await page.addInitScript(({ token, userInfo }) => {
    localStorage.setItem('token', token)
    localStorage.setItem('userInfo', JSON.stringify(userInfo))
  }, session)
  await page.goto(path)
}
