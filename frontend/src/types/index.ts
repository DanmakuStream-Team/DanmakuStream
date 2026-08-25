export type UserRole = 'user' | 'creator' | 'moderator' | 'admin'
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
  special?: boolean
  groupId?: number | null
  blocked?: boolean
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
  category: string
  author: UserInfo
  tags: string | string[]
  createdAt: string
  commentCount?: number
}

export interface VideoCollectionInfo {
  id: number
  title: string
  description: string
  coverUrl: string
  owner: UserInfo
  videos?: VideoInfo[]
  createdAt: string
}

export interface Danmaku {
  id: number
  videoId: number
  userId: number
  content: string
  time: number
  color: string
  fontSize: 'small' | 'medium' | 'large'
  type: 'scroll' | 'top' | 'bottom' | 'advanced'
  blocked?: boolean
  createdAt?: string
	 author?: UserInfo
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

export interface LiveRoom {
  id: number
  title: string
  coverUrl: string
  streamKey?: string
  publishUrl?: string
  playUrl: string
  streamUrl: string
  status: 'idle' | 'live' | 'ended'
  viewerCount: number
	 viewerPeak: number
	 likeCount: number
  giftValue: number
	 heat: number
  chatMode: 'everyone' | 'followers' | 'members'
  slowModeSeconds: number
  pinnedMessage: string
  ownerId: number
  owner?: UserInfo
  startedAt?: string
  endedAt?: string
  createdAt: string
}

export interface VideoChapter {
  time: number
  label: string
}

export interface LiveGiftDefinition {
	key: 'flower' | 'star' | 'rocket'
	name: string
	value: number
}

export interface LiveSupportRankItem {
	userId: number
	user?: UserInfo
	value: number
	giftCount: number
}

export interface LiveInteraction {
	likeCount: number
	giftValue: number
	heat: number
	gifts: LiveGiftDefinition[]
	supportRank: LiveSupportRankItem[]
	superChats: LiveMonitorSuperChat[]
}

export interface LiveGiftEvent {
	id?: number
	user?: UserInfo
	gift: LiveGiftDefinition
	count: number
	value: number
	giftValue: number
	heat: number
	supportRank: LiveSupportRankItem[]
	createdAt?: string
	message?: string
	displaySeconds?: number
}

export interface LiveMonitorSuperChat {
  id: number
  user?: UserInfo
  gift: LiveGiftDefinition
  count: number
  value: number
  createdAt: string
  message: string
  displaySeconds: number
}

export interface LiveMonitorSnapshot {
  messages: Danmaku[]
  superChats: LiveMonitorSuperChat[]
}

export interface LiveChatSettings {
  chatMode: LiveRoom['chatMode']
  slowModeSeconds: number
  pinnedMessage: string
}

export interface LiveReplay {
  id: number
  roomId: number
  title: string
  coverUrl: string
  replayUrl: string
  status: 'processing' | 'ready' | 'unavailable'
  duration: number
  viewerPeak: number
  ownerId: number
  owner?: UserInfo
  startedAt: string
  endedAt: string
  createdAt: string
}

export interface DynamicPost {
  id: number
  userId: number
  content: string
  images: string
  author?: UserInfo
  createdAt: string
}

export interface LiveSchedule {
  id: number
  title: string
  coverUrl: string
  scheduledAt: string
  status: 'pending' | 'canceled' | 'live'
  reminderCount: number
  reserved: boolean
  ownerId: number
  owner?: UserInfo
  createdAt: string
}

export interface NotificationInfo {
  id: number
  type: string
  title: string
  content: string
  link: string
  read: boolean
  actor?: UserInfo
  createdAt: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

export interface CreatorAnalyticsPoint {
  date: string
  views: number
  collects: number
  growthSpeed: number
  streams: number
}

export interface CreatorAnalytics {
  days: number
  selectedVideoId: number
  summary: {
    totalViews: number
    totalCollects: number
    totalStreams: number
    rangeViews: number
    rangeCollects: number
    rangeStreams: number
    averageDailyViews: number
  }
  points: CreatorAnalyticsPoint[]
  topVideos: Array<{
    id: number
    title: string
    coverUrl: string
    status: VideoStatus
    viewCount: number
    likeCount: number
    collectCount: number
  }>
}

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}
