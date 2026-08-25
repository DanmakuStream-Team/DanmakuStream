import type { VideoChapter } from '@/types'

const TIMESTAMP_LINE = /^\s*((?:\d{1,2}:)?\d{1,2}:\d{2})\s+(.+?)\s*$/

export function parseVideoChapters(description = ''): VideoChapter[] {
  const chapters = description.split(/\r?\n/).flatMap((line) => {
    const match = line.match(TIMESTAMP_LINE)
    if (!match) return []
    const parts = match[1].split(':').map(Number)
    const time = parts.length === 3
      ? parts[0] * 3600 + parts[1] * 60 + parts[2]
      : parts[0] * 60 + parts[1]
    return Number.isFinite(time) ? [{ time, label: match[2] }] : []
  })

  chapters.sort((a, b) => a.time - b.time)
  if (chapters.length < 2 || chapters[0]?.time !== 0) return []
  return chapters.filter((chapter, index) => index === 0 || chapter.time > chapters[index - 1].time)
}
