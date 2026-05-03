import request from '@/utils/request'
import type { ApiResponse, Comment } from '@/types'

export interface CreateCommentParams {
  videoId: number
  content: string
  parentId?: number
}

export const commentApi = {
  createComment(data: CreateCommentParams) {
    return request.post<ApiResponse<Comment>>('/comments', data)
  },

  getComments(videoId: number) {
    return request.get<ApiResponse<Comment[]>>(`/comments/${videoId}`)
  },
}