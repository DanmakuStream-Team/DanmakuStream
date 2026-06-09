<template>
  <canvas ref="canvasRef" class="danmaku-layer" />
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { Danmaku } from '@/types'

const props = defineProps<{ items: Danmaku[]; currentTime: number; paused?: boolean }>()

const canvasRef = ref<HTMLCanvasElement>()
let ctx: CanvasRenderingContext2D | null = null
let animationFrame: number | null = null
let resizeObserverInstance: ResizeObserver | null = null
let activeDanmakus: ActiveDanmaku[] = []
let displayedIds = new Set<number>()
let lastFrameAt = 0
let lastMediaTime = 0

interface AdvancedSpec {
  text: string
  x: number
  y: number
  targetX: number
  targetY: number
  duration: number
  fontSizePx: number
  color: string
  alpha: number
}

interface ActiveDanmaku extends Danmaku {
  x: number
  y: number
  fromX: number
  fromY: number
  targetX: number
  targetY: number
  speed: number
  duration: number
  width: number
  fontSizePx: number
  startedAt: number
  displayType: string
  displayText: string
  displayColor: string
  alpha: number
}

const PRESET_COLORS = [
  '#FFFFFF', '#FF5555', '#55FF55', '#5555FF', '#FFFF55',
  '#FF55FF', '#55FFFF', '#FF8C00', '#FF69B4', '#00CED1',
  '#FFD700', '#7FFFD4', '#FF6347', '#98FB98', '#DDA0DD',
]

const FONT_SIZE_MAP = { small: 12, medium: 16, large: 20 }
const ROW_HEIGHT = 32
const TOP_PADDING = 8
const MAX_ACTIVE = 80

function getFontSize(fontSize?: string): number {
  return FONT_SIZE_MAP[fontSize as keyof typeof FONT_SIZE_MAP] || 16
}

function getColor(item: Danmaku): string {
  if (item.color) return item.color
  return PRESET_COLORS[item.id % PRESET_COLORS.length]
}

function getSpeed(content: string, fontSizePx: number): number {
  const baseSpeed = 120 + fontSizePx * 2
  return Math.min(280, Math.max(120, baseSpeed + content.length * 2))
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value))
}

function updateDimensions() {
  if (!canvasRef.value) return
  const rect = canvasRef.value.getBoundingClientRect()
  const ratio = window.devicePixelRatio || 1
  canvasRef.value.width = Math.max(1, Math.floor(rect.width * ratio))
  canvasRef.value.height = Math.max(1, Math.floor(rect.height * ratio))
  ctx = canvasRef.value.getContext('2d')
  if (ctx) ctx.scale(ratio, ratio)
}

function getCanvasWidth() {
  return canvasRef.value?.clientWidth || 0
}

function getCanvasHeight() {
  return canvasRef.value?.clientHeight || 0
}

function getRows() {
  return Math.max(3, Math.floor((getCanvasHeight() - TOP_PADDING * 2) / ROW_HEIGHT))
}

function measureText(content: string, fontSizePx: number) {
  if (!ctx) return content.length * fontSizePx
  ctx.font = `${fontSizePx}px "PingFang SC", "Microsoft YaHei", sans-serif`
  return ctx.measureText(content).width
}

function pickRow(type: string) {
  const rows = getRows()
  if (type === 'top') return Math.min(1, rows - 1)
  if (type === 'bottom') return Math.max(rows - 2, 0)
  const occupied = new Set(
    activeDanmakus
      .filter((item) => item.displayType === 'scroll')
      .map((item) => Math.floor((item.y - TOP_PADDING) / ROW_HEIGHT)),
  )
  for (let row = 0; row < rows; row++) {
    if (!occupied.has(row)) return row
  }
  return activeDanmakus.length % rows
}

function parseAdvancedDanmaku(item: Danmaku): AdvancedSpec | null {
  const raw = item.content.trim()
  if (item.type !== 'advanced' && !raw.startsWith('@adv')) return null

  const body = raw.replace(/^@adv\s*/, '')
  const [paramPart, textPart = ''] = body.split('|')
  const paramsText = paramPart.replace(/^\{/, '').replace(/\}$/, '')
  const params = new Map<string, string>()
  const matcher = /(\w+)=("[^"]*"|'[^']*'|[^\s,;]+)/g
  let match: RegExpExecArray | null

  while ((match = matcher.exec(paramsText))) {
    params.set(match[1], match[2].replace(/^['"]|['"]$/g, ''))
  }

  const readNumber = (keys: string[], fallback: number) => {
    for (const key of keys) {
      const value = Number(params.get(key))
      if (Number.isFinite(value)) return value
    }
    return fallback
  }

  const text = params.get('text') || textPart.trim() || '*'
  const x = clamp(readNumber(['x', 'fromX'], 50), 0, 100)
  const y = clamp(readNumber(['y', 'fromY'], 50), 0, 100)
  const targetX = clamp(readNumber(['tx', 'toX', 'targetX'], x), 0, 100)
  const targetY = clamp(readNumber(['ty', 'toY', 'targetY'], y), 0, 100)
  const duration = Math.max(0.2, Math.min(30, readNumber(['dur', 'duration'], 4)))
  const fontSizePx = Math.max(8, Math.min(96, readNumber(['size', 'fontSize'], getFontSize(item.fontSize))))
  const color = params.get('color') || item.color || '#FFFFFF'
  const alpha = clamp(readNumber(['alpha', 'opacity'], 1), 0, 1)

  return { text, x, y, targetX, targetY, duration, fontSizePx, color, alpha }
}

function addDanmaku(item: Danmaku, mediaTime: number, elapsed = 0) {
  if (!canvasRef.value || !ctx || displayedIds.has(item.id)) return

  const advanced = parseAdvancedDanmaku(item)
  const displayType = advanced ? 'advanced' : item.type || 'scroll'
  const fontSizePx = advanced?.fontSizePx || getFontSize(item.fontSize)
  const displayColor = advanced?.color || getColor(item)
  const displayText = advanced?.text || item.content
  const width = measureText(displayText, fontSizePx)
  const canvasWidth = getCanvasWidth()
  const canvasHeight = getCanvasHeight()
  const row = pickRow(displayType)
  const baseY = TOP_PADDING + row * ROW_HEIGHT + ROW_HEIGHT / 2
  const duration = advanced?.duration || (displayType === 'scroll'
    ? Math.max(5, Math.min(12, (canvasWidth + width) / getSpeed(displayText, fontSizePx)))
    : 4)
  const speed = displayType === 'scroll' ? (canvasWidth + width) / duration : 0
  const x = advanced
    ? canvasWidth * advanced.x / 100
    : displayType === 'scroll'
      ? canvasWidth - speed * Math.max(0, elapsed)
      : (canvasWidth - width) / 2
  const y = advanced ? canvasHeight * advanced.y / 100 : baseY
  const targetX = advanced ? canvasWidth * advanced.targetX / 100 : x
  const targetY = advanced ? canvasHeight * advanced.targetY / 100 : y

  activeDanmakus.push({
    ...item,
    x,
    y,
    fromX: x,
    fromY: y,
    targetX,
    targetY,
    speed,
    duration,
    width,
    fontSizePx,
    startedAt: mediaTime,
    displayType,
    displayText,
    displayColor,
    alpha: advanced?.alpha ?? 1,
  })
  displayedIds.add(item.id)

  if (activeDanmakus.length > MAX_ACTIVE) {
    const removed = activeDanmakus.shift()
    if (removed) displayedIds.delete(removed.id)
  }
}

function enqueueDueDanmakus(mediaTime: number) {
  if (!ctx) return

  const candidates = props.items
    .slice()
    .sort((a, b) => a.time - b.time)
    .filter((item) => !displayedIds.has(item.id))

  for (const item of candidates) {
    const elapsed = mediaTime - item.time

    if (item.time > 0) {
      if (elapsed >= -0.1 && elapsed <= 0.6) addDanmaku(item, mediaTime, elapsed)
      continue
    }

    if (mediaTime <= 1 || props.items.length <= MAX_ACTIVE) {
      addDanmaku(item, mediaTime, 0)
    }
  }
}

function updateActive(deltaSeconds: number, mediaTime: number) {
  activeDanmakus = activeDanmakus.filter((item) => {
    if (item.displayType === 'scroll') {
      item.x -= item.speed * deltaSeconds
      return item.x + item.width > 0
    }

    if (item.displayType === 'advanced') {
      const progress = clamp((mediaTime - item.startedAt) / item.duration, 0, 1)
      const ease = progress < 0.5 ? 2 * progress * progress : 1 - Math.pow(-2 * progress + 2, 2) / 2
      item.x = item.fromX + (item.targetX - item.fromX) * ease
      item.y = item.fromY + (item.targetY - item.fromY) * ease
      return progress < 1
    }

    return mediaTime - item.startedAt < item.duration
  })
}

function draw() {
  if (!ctx || !canvasRef.value) return
  ctx.clearRect(0, 0, getCanvasWidth(), getCanvasHeight())

  for (const item of activeDanmakus) {
    ctx.save()
    ctx.globalAlpha = item.alpha
    ctx.font = `${item.fontSizePx}px "PingFang SC", "Microsoft YaHei", sans-serif`
    ctx.textBaseline = 'middle'
    ctx.fillStyle = item.displayColor
    ctx.fillText(item.displayText, item.x, item.y)
    ctx.restore()
  }
}

function animate(now: number) {
  const deltaSeconds = lastFrameAt ? Math.min((now - lastFrameAt) / 1000, 0.05) : 0
  lastFrameAt = now

  if (!props.paused) {
    enqueueDueDanmakus(props.currentTime)
    updateActive(deltaSeconds, props.currentTime)
  }

  draw()
  animationFrame = requestAnimationFrame(animate)
}

function resetLayer() {
  activeDanmakus = []
  displayedIds = new Set()
  lastFrameAt = 0
  draw()
}

onMounted(() => {
  updateDimensions()
  resizeObserverInstance = new ResizeObserver(() => {
    updateDimensions()
    resetLayer()
  })
  if (canvasRef.value) resizeObserverInstance.observe(canvasRef.value)
  animationFrame = requestAnimationFrame(animate)
})

onBeforeUnmount(() => {
  if (animationFrame) cancelAnimationFrame(animationFrame)
  if (resizeObserverInstance) resizeObserverInstance.disconnect()
})

watch(() => props.currentTime, (time) => {
  if (time + 1 < lastMediaTime || Math.abs(time - lastMediaTime) > 3) {
    resetLayer()
  }
  lastMediaTime = time
})

watch(() => props.items, () => {
  if (!props.paused) enqueueDueDanmakus(props.currentTime)
}, { deep: false })
</script>

<style scoped>
.danmaku-layer {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  display: block;
}
</style>
