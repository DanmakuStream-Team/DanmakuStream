import request from '@/utils/request'
import type { Comment } from '@/types'

export interface CreateCommentParams {
  videoId: number
  content: string
  parentId?: number
}

export const commentApi = {
  createComment(data: CreateCommentParams) {
    return request.post<Comment>('/comments', data)
  },

  getComments(videoId: number) {
    return request.get<Comment[]>(`/comments/${videoId}`)
  },
}
