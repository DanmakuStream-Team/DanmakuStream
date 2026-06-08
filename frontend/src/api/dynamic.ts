import request from '@/utils/request'
import type { DynamicPost, PageResult } from '@/types'

export const dynamicApi = {
  list(params: { page?: number; pageSize?: number; userId?: number } = {}) {
    return request.get<PageResult<DynamicPost>>('/dynamics', { params })
  },
  create(data: { content: string; images?: string }) {
    return request.post<DynamicPost>('/dynamics', data)
  },
  remove(id: number) {
    return request.delete<{ id: number }>(`/dynamics/${id}`)
  },
}
