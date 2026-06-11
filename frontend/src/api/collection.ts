import request from '@/utils/request'
import type { VideoCollectionInfo } from '@/types'

export const collectionApi = {
  mine() {
    return request.get<VideoCollectionInfo[]>('/users/me/collections')
  },
  create(data: { title: string; description?: string; coverUrl?: string }) {
    return request.post<VideoCollectionInfo>('/collections', data)
  },
  detail(id: number) {
    return request.get<VideoCollectionInfo>(`/collections/${id}`)
  },
  videoCollections(videoId: number) {
    return request.get<VideoCollectionInfo[]>(`/videos/${videoId}/collections`)
  },
  addVideo(collectionId: number, videoId: number, sort = 0) {
    return request.post<{ id: number }>(`/collections/${collectionId}/videos`, { videoId, sort })
  },
  removeVideo(collectionId: number, videoId: number) {
    return request.delete<{ videoId: number }>(`/collections/${collectionId}/videos/${videoId}`)
  },
  addCollaborator(videoId: number, userId: number) {
    return request.post<{ videoId: number; userId: number }>(`/videos/${videoId}/collaborators`, { userId })
  },
  removeCollaborator(videoId: number, userId: number) {
    return request.delete<{ videoId: number; userId: number }>(`/videos/${videoId}/collaborators/${userId}`)
  },
}
