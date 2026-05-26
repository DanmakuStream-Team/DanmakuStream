import request from '@/utils/request'
import type { UserInfo } from '@/types'

export const userApi = {
  profile(id: number) {
    return request.get<UserInfo>(`/users/${id}`)
  },
  follow(id: number) {
    return request.post<{ followed: boolean }>(`/users/${id}/follow`)
  },
}
