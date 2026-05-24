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
  register(data: { nickname: string; password: string }) {
    return request.post<AuthResponse>('/auth/register', data)
  },
  me() {
    return request.get<UserInfo>('/auth/me')
  },
  updateMe(data: { nickname?: string; bio?: string }) {
    return request.put<void>('/users/me', data)
  },
  uploadAvatar(formData: FormData) {
    return request.post<{ avatar: string }>('/users/me/avatar', formData)
  },
}
