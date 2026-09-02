import type { APIRequestContext, Page } from '@playwright/test'
import { API } from '../test-data'

export interface Session {
  token: string
  userInfo: { id: number; nickname: string; role?: string }
}

export async function loginViaApi(request: APIRequestContext, nickname: string, password: string): Promise<Session> {
  const resp = await request.post(`${API}/auth/login`, { data: { nickname, password } })
  if (!resp.ok()) {
    const register = await request.post(`${API}/auth/register`, { data: { nickname, password } })
    if (!register.ok()) {
      throw new Error(`loginViaApi failed: login=${resp.status()} register=${register.status()} body=${await register.text()}`)
    }
    const d = (await register.json()).data
    return { token: d.token, userInfo: d.userInfo ?? d }
  }
  const d = (await resp.json()).data
  return { token: d.token, userInfo: d.userInfo ?? d }
}

export async function openAs(page: Page, session: Session, path: string): Promise<void> {
  await page.addInitScript((payload) => {
    localStorage.setItem('token', payload.token)
    localStorage.setItem('userInfo', JSON.stringify(payload.userInfo))
  }, session)
  await page.goto(path, { waitUntil: 'domcontentloaded' })
}
