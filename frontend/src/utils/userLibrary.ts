import type { VideoInfo } from '@/types'

export type UserLibraryKind = 'history' | 'liked' | 'collections' | 'downloads'

export interface UserLibraryRecord {
  video: VideoInfo
  savedAt: string
  progress?: number
}

const STORAGE_KEYS: Record<UserLibraryKind, string> = {
  history: 'danmaku:user-library:history',
  liked: 'danmaku:user-library:liked',
  collections: 'danmaku:user-library:collections',
  downloads: 'danmaku:user-library:downloads',
}

function readRecords(kind: UserLibraryKind): UserLibraryRecord[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEYS[kind])
    if (!raw) return []
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function writeRecords(kind: UserLibraryKind, records: UserLibraryRecord[]) {
  localStorage.setItem(STORAGE_KEYS[kind], JSON.stringify(records.slice(0, 80)))
}

export function getUserLibraryRecords(kind: UserLibraryKind) {
  return readRecords(kind)
}

export function upsertUserLibraryRecord(kind: UserLibraryKind, video: VideoInfo, progress?: number) {
  const records = readRecords(kind).filter((item) => item.video.id !== video.id)
  records.unshift({
    video,
    savedAt: new Date().toISOString(),
    progress,
  })
  writeRecords(kind, records)
}

export function removeUserLibraryRecord(kind: UserLibraryKind, videoId: number) {
  writeRecords(kind, readRecords(kind).filter((item) => item.video.id !== videoId))
}

export function clearUserLibraryRecords(kind: UserLibraryKind) {
  writeRecords(kind, [])
}
