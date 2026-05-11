import request from '@/utils/request'
import type { Danmaku } from '@/types'

export const danmakuApi = {
  getDanmakuList(videoId: number) {
    return request.get<Danmaku[]>(`/danmaku/${videoId}`)
  },
  sendDanmaku(data: Pick<Danmaku, 'videoId' | 'content' | 'time' | 'color' | 'fontSize' | 'type'>) {
    return request.post<Danmaku>('/danmaku', data)
  },
  blockDanmaku(id: number) {
    return request.put<void>(`/danmaku/${id}/block`)
  },
}

// WebSocket 弹幕连接（直播间）
export class DanmakuWebSocket {
  private ws: WebSocket | null = null
  private roomId: number
  private token: string
  private onMessage: (danmaku: Danmaku) => void
  private onViewerCount: (count: number) => void
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null

  constructor(options: {
    roomId: number
    token: string
    onMessage: (danmaku: Danmaku) => void
    onViewerCount: (count: number) => void
  }) {
    this.roomId = options.roomId
    this.token = options.token
    this.onMessage = options.onMessage
    this.onViewerCount = options.onViewerCount
  }

  connect() {
    const url = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws/live/${this.roomId}?token=${this.token}`
    this.ws = new WebSocket(url)

    this.ws.onopen = () => {
      console.log('[WS] Connected to live room', this.roomId)
    }

    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data)
      if (data.type === 'danmaku') {
        this.onMessage(data.payload)
      } else if (data.type === 'viewer_count') {
        this.onViewerCount(data.payload)
      }
    }

    this.ws.onclose = () => {
      console.log('[WS] Disconnected, reconnecting in 3s...')
      this.reconnectTimer = setTimeout(() => this.connect(), 3000)
    }

    this.ws.onerror = (err) => {
      console.error('[WS] Error:', err)
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
