import request from '@/utils/request'
import type { Danmaku, LiveChatSettings, LiveGiftEvent, LiveInteraction, LiveMonitorSnapshot, LiveReplay, LiveRoom, LiveSchedule, PageResult } from '@/types'

export const liveApi = {
  list(params: { page?: number; pageSize?: number } = {}) {
    return request.get<PageResult<LiveRoom>>('/live', { params })
  },
  detail(id: number) {
    return request.get<LiveRoom>(`/live/${id}`)
  },
  create(data: { title: string; coverUrl?: string }) {
    return request.post<LiveRoom>('/live', data)
  },
  end(id: number) {
    return request.put<LiveRoom>(`/live/${id}/end`)
  },
	manageDetail(id: number) {
		return request.get<LiveRoom>(`/live/${id}/manage`)
	},
	monitor(id: number) {
		return request.get<LiveMonitorSnapshot>(`/live/${id}/monitor`)
	},
	updateChatSettings(id: number, data: LiveChatSettings) {
		return request.put<LiveRoom>(`/live/${id}/chat-settings`, data)
	},
	interaction(id: number) {
		return request.get<LiveInteraction>(`/live/${id}/interaction`)
	},
	liveDanmaku(id: number) {
		return request.get<Danmaku[]>(`/live/${id}/danmaku`)
	},
	heatRanking() {
		return request.get<Array<{ room: LiveRoom; heat: number }>>('/live/rankings/heat')
	},
	like(id: number) {
		return request.post<{ userId: number; liked: boolean; likeCount: number; heat: number }>(`/live/${id}/like`)
	},
	likeStatus(id: number) {
		return request.get<{ liked: boolean }>(`/live/${id}/like/status`)
	},
	sendGift(id: number, giftKey: string, count: number, message = '') {
		return request.post<LiveGiftEvent>(`/live/${id}/gifts`, { giftKey, count, message })
	},
  replays(params: { page?: number; pageSize?: number } = {}) {
    return request.get<PageResult<LiveReplay>>('/live-replays', { params })
  },
  replayDetail(id: number) {
    return request.get<LiveReplay>(`/live-replays/${id}`)
  },
  replayDanmaku(id: number) {
    return request.get<Danmaku[]>(`/live-replays/${id}/danmaku`)
  },
  schedules(params: { page?: number; pageSize?: number; status?: LiveSchedule['status'] } = {}) {
    return request.get<PageResult<LiveSchedule>>('/live-schedules', { params })
  },
  createSchedule(data: { title: string; coverUrl?: string; scheduledAt: string }) {
    return request.post<LiveSchedule>('/live-schedules', data)
  },
  cancelSchedule(id: number) {
    return request.delete<{ id: number; status: LiveSchedule['status'] }>(`/live-schedules/${id}`)
  },
  reserveSchedule(id: number) {
    return request.post<{ reserved: boolean; reminderCount: number }>(`/live-schedules/${id}/reserve`)
  },
}
