import request from '@/utils/request'
import type { Comment } from '@/types'

export const commentApi = {
  list(videoId: number) {
    return request.get<Comment[]>(`/comments/${videoId}`)
  },
  create(data: { videoId: number; content: string; parentId?: number }) {
    return request.post<Comment>('/comments', data)
  },
  remove(id: number) {
    return request.delete<{ id: number }>(`/comments/${id}`)
  },
  like(id: number) {
    return request.post<{ liked: boolean; likeCount: number }>(`/comments/${id}/like`)
  },
}
