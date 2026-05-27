import request from '@/utils/request'
import type { UserInfo } from '@/types'

export interface FolloweeInfo {
  id: number
  nickname: string
  avatar: string
  role: string
}

export const userApi = {
  profile(id: number) {
    return request.get<UserInfo>(`/users/${id}`)
  },
  follow(id: number) {
    return request.post<{ followed: boolean }>(`/users/${id}/follow`)
  },
  following() {
    return request.get<{ list: FolloweeInfo[] }>('/users/following')
  },
}
