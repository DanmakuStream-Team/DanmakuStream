import request from '@/utils/request'
import type { PageResult, VideoInfo } from '@/types'

export interface ServerLibraryRecord {
  video: VideoInfo
  savedAt: string
  progress: number
  position: number
}

export const libraryApi = {
  history(params: { page?: number; pageSize?: number } = {}) {
    return request.get<PageResult<ServerLibraryRecord>>('/users/me/history', { params })
  },
  historyDetail(videoId: number) {
    return request.get<ServerLibraryRecord>(`/users/me/history/${videoId}`, {
      skipErrorMessage: true,
    })
  },
  saveHistory(videoId: number, position: number) {
    return request.put<{ videoId: number; position: number }>(`/users/me/history/${videoId}`, { position })
  },
  removeHistory(videoId: number) {
    return request.delete<{ videoId: number }>(`/users/me/history/${videoId}`)
  },
  clearHistory() {
    return request.delete<{ cleared: boolean }>('/users/me/history')
  },
  watchLater(params: { page?: number; pageSize?: number } = {}) {
    return request.get<PageResult<ServerLibraryRecord>>('/users/me/watch-later', { params })
  },
  watchLaterStatus(videoId: number) {
    return request.get<{ saved: boolean }>(`/users/me/watch-later/${videoId}/status`)
  },
  toggleWatchLater(videoId: number) {
    return request.post<{ saved: boolean }>(`/users/me/watch-later/${videoId}`)
  },
  removeWatchLater(videoId: number) {
    return request.delete<{ videoId: number }>(`/users/me/watch-later/${videoId}`)
  },
  clearWatchLater() {
    return request.delete<{ cleared: boolean }>('/users/me/watch-later')
  },
}
