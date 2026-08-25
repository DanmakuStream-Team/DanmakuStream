import request from '@/utils/request'
import type { Danmaku, LiveGiftEvent, PageResult } from '@/types'

export const danmakuApi = {
  list(videoId: number) {
    return request.get<Danmaku[]>(`/danmaku/${videoId}`)
  },
  send(data: Pick<Danmaku, 'videoId' | 'content' | 'time' | 'color' | 'fontSize' | 'type'>) {
    return request.post<Danmaku>('/danmaku', data)
  },
  uploadAdvanced(videoId: number, file: File) {
    const form = new FormData()
    form.append('videoId', String(videoId))
    form.append('file', file)
    return request.post<{ list: Danmaku[]; count: number }>('/danmaku/advanced/upload', form)
  },
  adminList(params: { page: number; pageSize: number; videoId?: number; keyword?: string; blocked?: boolean }) {
    return request.get<PageResult<Danmaku>>('/admin/danmaku', { params })
  },
  block(id: number) {
    return request.put<void>(`/admin/danmaku/${id}/block`)
  },
}

export class DanmakuWebSocket {
  private ws: WebSocket | null = null
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private shouldReconnect = true

  constructor(
    private options: {
      roomId: number
      token: string
      onMessage: (danmaku: Danmaku) => void
      onViewerCount: (count: number) => void
	  onConnectionChange?: (connected: boolean) => void
	  onLike?: (payload: { userId: number; liked: boolean; likeCount: number; heat: number }) => void
	  onGift?: (payload: LiveGiftEvent) => void
	  onChatError?: (payload: { message: string; retryAfter: number }) => void
	  monitor?: boolean
    }
  ) {}

  connect() {
    this.shouldReconnect = true
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
	const query = new URLSearchParams()
	if (this.options.token) query.set('token', this.options.token)
	if (this.options.monitor) query.set('monitor', '1')
	const queryString = query.size ? `?${query.toString()}` : ''
    this.ws = new WebSocket(`${protocol}://${location.host}/ws/live/${this.options.roomId}${queryString}`)
	this.ws.onopen = () => this.options.onConnectionChange?.(true)
    this.ws.onmessage = (event) => {
      let data: any
      try {
        data = JSON.parse(event.data)
      } catch {
        return
      }
      if (data.type === 'danmaku') {
        const p = data.payload
        this.options.onMessage({ ...p, type: p.danmakuType || 'scroll' })
      }
      if (data.type === 'viewer_count') this.options.onViewerCount(data.payload)
	  if (data.type === 'live_like') this.options.onLike?.(data.payload)
	  if (data.type === 'live_gift') this.options.onGift?.(data.payload)
	  if (data.type === 'chat_error') this.options.onChatError?.(data.payload)
    }
    this.ws.onclose = () => {
	  this.options.onConnectionChange?.(false)
      if (!this.shouldReconnect) return
      this.reconnectTimer = setTimeout(() => this.connect(), 3000)
    }
  }

  send(content: string, color = '#FFFFFF', fontSize = 'medium', danmakuType = 'scroll') {
    if (this.ws?.readyState !== WebSocket.OPEN) return false
    this.ws.send(JSON.stringify({ type: 'danmaku', content, color, fontSize, danmakuType }))
    return true
  }

  disconnect() {
    this.shouldReconnect = false
	this.options.onConnectionChange?.(false)
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    this.ws?.close()
    this.ws = null
  }
}
