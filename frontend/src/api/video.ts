import request from '@/utils/request'
import type { PageResult, VideoInfo, VideoStatus } from '@/types'

export const videoApi = {
  list(params: { page: number; pageSize: number; keyword?: string; tag?: string }) {
    return request.get<PageResult<VideoInfo>>('/videos', { params })
  },
  detail(id: number) {
    return request.get<VideoInfo>(`/videos/${id}`)
  },
  upload(formData: FormData, onProgress?: (percent: number) => void) {
    return request.post<VideoInfo>('/videos/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 0,
      onUploadProgress: (event) => {
        if (event.total && onProgress) {
          onProgress(Math.round((event.loaded / event.total) * 100))
        }
      },
    })
  },
  update(id: number, data: { title?: string; description?: string; tags?: string }) {
    return request.put<{ id: number }>(`/videos/${id}`, data)
  },
  updateCover(id: number, formData: FormData) {
    return request.post<{ coverUrl: string }>(`/videos/${id}/cover`, formData)
  },
  remove(id: number) {
    return request.delete<{ id: number }>(`/videos/${id}`)
  },
  like(id: number) {
    return request.post<{ liked: boolean }>(`/videos/${id}/like`)
  },
  collect(id: number) {
    return request.post<{ collected: boolean }>(`/videos/${id}/collect`)
  },
  myVideos(params: { page: number; pageSize: number; status?: VideoStatus | '' }) {
    return request.get<PageResult<VideoInfo>>('/users/me/videos', { params })
  },
  userVideos(userId: number, params: { page: number; pageSize: number }) {
    return request.get<PageResult<VideoInfo>>(`/users/${userId}/videos`, { params })
  },
  adminList(params: { page: number; pageSize: number; status?: VideoStatus | ''; keyword?: string }) {
    return request.get<PageResult<VideoInfo>>('/admin/videos', { params })
  },
  adminUpdateStatus(id: number, status: VideoStatus) {
    return request.put<{ id: number; status: VideoStatus }>(`/admin/videos/${id}/status`, { status })
  },
}
