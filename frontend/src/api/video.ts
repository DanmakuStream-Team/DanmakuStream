import request from '@/utils/request'
import type { VideoInfo, PageResult } from '@/types'

export const videoApi = {
  getVideoList(params: { page: number; pageSize: number; keyword?: string; tag?: string }) {
    return request.get<PageResult<VideoInfo>>('/videos', { params })
  },
  getVideoDetail(id: number) {
    return request.get<VideoInfo>(`/videos/${id}`)
  },
  uploadVideo(formData: FormData, onProgress?: (percent: number) => void) {
    return request.post<{ videoId: number }>('/videos/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (e) => {
        if (onProgress && e.total) onProgress(Math.round((e.loaded / e.total) * 100))
      },
    })
  },
  updateVideoMeta(id: number, data: Partial<Pick<VideoInfo, 'title' | 'description' | 'tags'>>) {
    return request.put<void>(`/videos/${id}`, data)
  },
  likeVideo(id: number) {
    return request.post<void>(`/videos/${id}/like`)
  },
  collectVideo(id: number) {
    return request.post<void>(`/videos/${id}/collect`)
  },
  deleteVideo(id: number) {
    return request.delete<void>(`/videos/${id}`)
  },
}
