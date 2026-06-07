import request from '@/utils/request'
import type { NotificationInfo, PageResult } from '@/types'

export const notificationApi = {
  list(params: { page?: number; pageSize?: number; read?: boolean } = {}) {
    return request.get<PageResult<NotificationInfo> & { unreadCount: number }>('/notifications', { params })
  },
  read(id: number) {
    return request.put<{ id: number; read: boolean }>(`/notifications/${id}/read`)
  },
  readAll() {
    return request.put<{ read: boolean }>('/notifications')
  },
}
