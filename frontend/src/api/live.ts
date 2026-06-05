import request from '@/utils/request'
import type { LiveRoom, PageResult } from '@/types'

export const liveApi = {
  list(params: { page?: number; pageSize?: number } = {}) {
    return request.get<PageResult<LiveRoom>>('/live', { params })
  },
  detail(id: number) {
    return request.get<LiveRoom>(`/live/${id}`)
  },
  create(data: { title: string; coverUrl?: string }) {
    return request.post<LiveRoom>('/live', data)
  },
  end(id: number) {
    return request.put<LiveRoom>(`/live/${id}/end`)
  },
}
