import request from '@/utils/request'
import type { UserInfo } from '@/types'

export const userApi = {
  getProfile(id: number) {
    return request.get<UserInfo>(`/users/${id}`)
  },
}
