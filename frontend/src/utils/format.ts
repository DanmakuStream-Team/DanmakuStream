export function mediaUrl(url?: string) {
  if (!url) return ''
  if (/^https?:\/\//.test(url)) return url
  return url
}

export function formatCount(count = 0) {
  if (count >= 10000) return `${(count / 10000).toFixed(1)}万`
  return String(count)
}

export function formatDuration(seconds = 0) {
  if (!seconds) return '00:00'
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}:${String(s).padStart(2, '0')}`
}

export function normalizeTags(tags: string | string[] | undefined) {
  if (!tags) return []
  if (Array.isArray(tags)) return tags
  return tags.split(',').map((tag) => tag.trim()).filter(Boolean)
}

export function formatTime(value?: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
}
