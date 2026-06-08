import request from '@/utils/request'
import type { LiveRoom, LiveSchedule, PageResult } from '@/types'

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
  schedules(params: { page?: number; pageSize?: number; status?: LiveSchedule['status'] } = {}) {
    return request.get<PageResult<LiveSchedule>>('/live-schedules', { params })
  },
  createSchedule(data: { title: string; coverUrl?: string; scheduledAt: string }) {
    return request.post<LiveSchedule>('/live-schedules', data)
  },
  cancelSchedule(id: number) {
    return request.delete<{ id: number; status: LiveSchedule['status'] }>(`/live-schedules/${id}`)
  },
  reserveSchedule(id: number) {
    return request.post<{ reserved: boolean; reminderCount: number }>(`/live-schedules/${id}/reserve`)
  },
}
