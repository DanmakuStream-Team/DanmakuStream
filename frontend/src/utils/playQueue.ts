import type { VideoInfo } from '@/types'

const QUEUE_KEY = 'danmaku:play-queue'
const AUTOPLAY_KEY = 'danmaku:autoplay'
const MAX_QUEUE_SIZE = 30

export function loadPlayQueue(): VideoInfo[] {
  try {
    const value = JSON.parse(sessionStorage.getItem(QUEUE_KEY) || '[]')
    return Array.isArray(value) ? value : []
  } catch {
    return []
  }
}

export function savePlayQueue(items: VideoInfo[]) {
  sessionStorage.setItem(QUEUE_KEY, JSON.stringify(items.slice(0, MAX_QUEUE_SIZE)))
}

export function readAutoplay() {
  return localStorage.getItem(AUTOPLAY_KEY) !== 'false'
}

export function saveAutoplay(enabled: boolean) {
  localStorage.setItem(AUTOPLAY_KEY, String(enabled))
}
