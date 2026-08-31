import request from '@/utils/request'
import type { UserInfo } from '@/types'

export type ChatMessageType = 'text' | 'image' | 'video' | 'video_share'

export interface SharedChatVideo {
  id: number
  title: string
  coverUrl: string
  duration: number
  author: UserInfo
}

export interface MessageSendPayload {
  receiverId: number
  clientMessageId?: string
  type: ChatMessageType
  content?: string
  mediaUrl?: string
  mediaName?: string
  videoId?: number
}

export interface UploadedMessageMedia {
  url: string
  mediaType: 'image' | 'video'
  name: string
  contentType: string
  size: number
}

export interface ChatMessageInfo {
  id: number
  clientMessageId?: string
  senderId: number
  receiverId: number
  type: ChatMessageType
  content: string
  mediaUrl: string
  mediaName: string
  video?: SharedChatVideo
  read: boolean
  sender: UserInfo
  receiver: UserInfo
  createdAt: string
}

export interface ConversationInfo {
  user: UserInfo
  lastMessage: ChatMessageInfo
  unreadCount: number
}

export const messageApi = {
  conversations() {
    return request.get<{ list: ConversationInfo[] }>('/messages/conversations')
  },
  history(userId: number, params: { page?: number; pageSize?: number } = {}) {
    return request.get<{ list: ChatMessageInfo[]; total: number; page: number; pageSize: number }>(`/messages/${userId}`, { params })
  },
  send(payload: MessageSendPayload) {
    return request.post<ChatMessageInfo>('/messages', payload)
  },
  uploadMedia(file: File, onProgress?: (percent: number) => void) {
    const formData = new FormData()
    formData.append('file', file)
    return request.post<UploadedMessageMedia>('/messages/media', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 0,
      onUploadProgress: event => {
        if (event.total && onProgress) onProgress(Math.round(event.loaded / event.total * 100))
      },
    })
  },
  read(userId: number) {
    return request.put(`/messages/${userId}/read`)
  },
  unread() {
    return request.get<{ count: number }>('/messages/unread')
  },
}

export class ChatWebSocket {
  private socket: WebSocket | null = null
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private shouldReconnect = true

  constructor(private options: {
    token: string
    onMessage: (message: ChatMessageInfo) => void
    onError: (message: string) => void
    onStateChange?: (connected: boolean) => void
  }) {}

  connect() {
    if (!this.options.token) return
    this.shouldReconnect = true
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
    this.socket = new WebSocket(`${protocol}://${location.host}/ws/chat?token=${encodeURIComponent(this.options.token)}`)
    this.socket.onopen = () => this.options.onStateChange?.(true)
    this.socket.onmessage = (event) => {
      const data = JSON.parse(event.data)
      if (data.type === 'message') this.options.onMessage(data.payload)
      if (data.type === 'error') this.options.onError(String(data.payload || '消息发送失败'))
    }
    this.socket.onclose = () => {
      this.options.onStateChange?.(false)
      if (this.shouldReconnect) this.reconnectTimer = setTimeout(() => this.connect(), 3000)
    }
  }

  send(message: MessageSendPayload) {
    if (this.socket?.readyState !== WebSocket.OPEN) return false
    this.socket.send(JSON.stringify({ type: 'message', message }))
    return true
  }

  disconnect() {
    this.shouldReconnect = false
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    this.socket?.close()
    this.socket = null
  }
}
