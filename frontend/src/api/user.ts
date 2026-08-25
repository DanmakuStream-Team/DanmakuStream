import request from '@/utils/request'
import type { UserInfo } from '@/types'

export interface FolloweeInfo {
  id: number
  nickname: string
  avatar: string
  role: string
  special: boolean
  groupId: number | null
  groupName: string
}

export interface FollowGroupInfo {
  id: number
  name: string
  count: number
}

export interface BlockedUserInfo {
  id: number
  nickname: string
  avatar: string
  role: string
  blockedAt: string
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
  followGroups() {
    return request.get<{ list: FollowGroupInfo[] }>('/users/follow-groups')
  },
  createFollowGroup(name: string) {
    return request.post<FollowGroupInfo>('/users/follow-groups', { name })
  },
  updateFollowGroup(id: number, name: string) {
    return request.put<{ id: number; name: string }>(`/users/follow-groups/${id}`, { name })
  },
  deleteFollowGroup(id: number) {
    return request.delete(`/users/follow-groups/${id}`)
  },
  updateFollowSettings(id: number, data: { special?: boolean; groupId?: number }) {
    return request.put(`/users/${id}/follow-settings`, data)
  },
  blocked() {
    return request.get<{ list: BlockedUserInfo[] }>('/users/blocked')
  },
  block(id: number) {
    return request.post<{ blocked: boolean }>(`/users/${id}/block`)
  },
  search(params: { q: string; page?: number; pageSize?: number }) {
    return request.get<{ list: UserInfo[]; total: number; page: number; pageSize: number }>('/search/users', { params })
  },
}
