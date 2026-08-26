import { request as playwrightRequest } from '@playwright/test'
import { API, USERS } from './test-data'

async function ensureUser(api: Awaited<ReturnType<typeof playwrightRequest.newContext>>, nickname: string, password: string) {
  await api.post(`${API}/auth/register`, { data: { nickname, password } })
  const response = await api.post(`${API}/auth/login`, { data: { nickname, password } })
  if (!response.ok()) throw new Error(`cannot prepare ${nickname}: ${await response.text()}`)
  return (await response.json()).data as { token: string; userInfo: { id: number } }
}

export default async function globalSetup() {
  const api = await playwrightRequest.newContext()
  const owner = await ensureUser(api, USERS.owner.nickname, USERS.owner.password)
  await ensureUser(api, USERS.viewer.nickname, USERS.viewer.password)
  const headers = { Authorization: `Bearer ${owner.token}` }

  const roomsResponse = await api.get(`${API}/live?page=1&pageSize=100`)
  if (roomsResponse.ok()) {
    const rooms = (await roomsResponse.json()).data?.list || []
    for (const room of rooms) {
      if (room.ownerId === owner.userInfo.id) await api.put(`${API}/live/${room.id}/end`, { headers })
    }
  }

  const schedulesResponse = await api.get(`${API}/live-schedules?status=pending&page=1&pageSize=100`, { headers })
  if (schedulesResponse.ok()) {
    const schedules = (await schedulesResponse.json()).data?.list || []
    for (const schedule of schedules) {
      if (schedule.ownerId === owner.userInfo.id) await api.delete(`${API}/live-schedules/${schedule.id}`, { headers })
    }
  }
  await api.dispose()
}
