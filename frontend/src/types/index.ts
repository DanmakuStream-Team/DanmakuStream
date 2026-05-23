// 用户相关
export interface UserInfo {
  id: number
  username: string
  nickname: string
  avatar: string
  role: 'user' | 'creator' | 'admin'
  bio: string
  followCount: number
  fanCount: number
  followed?: boolean
  videoCount?: number
  createdAt: string
}

// 视频相关
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
  status: 'pending' | 'approved' | 'rejected'
  author: UserInfo
  tags: string[]
  createdAt: string
}

// 弹幕相关
export interface Danmaku {
  id: number
  videoId: number
  userId: number
  content: string
  time: number       // 视频时间点（秒）
  color: string
  fontSize: 'small' | 'medium' | 'large'
  type: 'scroll' | 'top' | 'bottom'
  createdAt: string
}

// 直播间相关
export interface LiveRoom {
  id: number
  title: string
  coverUrl: string
  streamKey: string
  status: 'idle' | 'live' | 'ended'
  viewerCount: number
  author: UserInfo
  createdAt: string
}

// 评论相关
export interface Comment {
  id: number
  videoId: number
  userId: number
  content: string
  likeCount: number
  author: UserInfo
  replies: Comment[]
  createdAt: string
}

// 通用分页响应
export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

// API 响应
export interface ApiResponse<T = void> {
  code: number
  message: string
  data: T
}
