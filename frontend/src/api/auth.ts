import request from '@/utils/request'
import type { ApiResponse, UserInfo } from '@/types'

export const authApi = {
  login(data: { username: string; password: string }) {
    return request.post<ApiResponse<{ token: string }>>('/auth/login', data)
  },
  register(data: { username: string; password: string; nickname: string }) {
    return request.post<ApiResponse>('/auth/register', data)
  },
  getUserInfo() {
    return request.get<ApiResponse<UserInfo>>('/auth/me')
  },
  logout() {
    return request.post<ApiResponse>('/auth/logout')
  },
}
