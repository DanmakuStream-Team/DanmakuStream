export type UserRole = 'user' | 'creator' | 'admin'
export type VideoStatus = 'pending' | 'approved' | 'rejected'

export interface UserInfo {
  id: number
  username: string
  nickname: string
  avatar: string
  bio: string
  role: UserRole
  followCount: number
  fanCount: number
  followed?: boolean
  videoCount?: number
  createdAt?: string
}

export interface VideoInfo {
  id: number
  title: string
  description: string
  coverUrl: string
  videoUrl: string
  duration: number
  viewCount: number
  likeCount: number
  collectCount: number
  danmakuCount: number
  status: VideoStatus
  author: UserInfo
  tags: string | string[]
  createdAt: string
  commentCount?: number
}

export interface Danmaku {
  id: number
  videoId: number
  userId: number
  content: string
  time: number
  color: string
  fontSize: 'small' | 'medium' | 'large'
  type: 'scroll' | 'top' | 'bottom'
  blocked?: boolean
  createdAt?: string
}

export interface Comment {
  id: number
  videoId: number
  userId: number
  content: string
  likeCount: number
  liked?: boolean
  author: UserInfo
  replies: Comment[]
  createdAt: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}
