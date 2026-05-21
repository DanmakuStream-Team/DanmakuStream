import request from '@/utils/request'
import type { UserInfo } from '@/types'

export interface AuthResponse {
  token: string
  userInfo: UserInfo
}

export const authApi = {
  login(data: { nickname: string; password: string }) {
    return request.post<AuthResponse>('/auth/login', data)
  },
  register(data: { password: string; nickname: string }) {
    return request.post<AuthResponse>('/auth/register', data)
  },
  getUserInfo() {
    return request.get<UserInfo>('/auth/me')
  },
  logout() {
    return request.post<void>('/auth/logout')
  },
}
