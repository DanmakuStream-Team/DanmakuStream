import request from '@/utils/request'
import type { Danmaku, PageResult } from '@/types'

export const danmakuApi = {
  list(videoId: number) {
    return request.get<Danmaku[]>(`/danmaku/${videoId}`)
  },
  send(data: Pick<Danmaku, 'videoId' | 'content' | 'time' | 'color' | 'fontSize' | 'type'>) {
    return request.post<Danmaku>('/danmaku', data)
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

  constructor(
    private options: {
      roomId: number
      token: string
      onMessage: (danmaku: Danmaku) => void
      onViewerCount: (count: number) => void
    }
  ) {}

  connect() {
    if (!this.options.token) return
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
    this.ws = new WebSocket(`${protocol}://${location.host}/ws/live/${this.options.roomId}?token=${this.options.token}`)
    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data)
      if (data.type === 'danmaku') this.options.onMessage(data.payload)
      if (data.type === 'viewer_count') this.options.onViewerCount(data.payload)
    }
    this.ws.onclose = () => {
      this.reconnectTimer = setTimeout(() => this.connect(), 3000)
    }
  }

  send(content: string, color = '#FFFFFF') {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'danmaku', content, color }))
    }
  }

  disconnect() {
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    this.ws?.close()
    this.ws = null
  }
}
