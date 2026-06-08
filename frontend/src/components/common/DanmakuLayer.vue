<template>
  <canvas ref="canvasRef" class="danmaku-layer" />
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount, computed } from 'vue'
import type { Danmaku } from '@/types'

const props = defineProps<{ items: Danmaku[]; currentTime: number }>()

const canvasRef = ref<HTMLCanvasElement>()
let ctx: CanvasRenderingContext2D | null = null
let rafId: number | null = null
let activeDanmakus: ActiveDanmaku[] = []
let lastTime = 0

interface ActiveDanmaku extends Danmaku {
  x: number
  y: number
  speed: number
  duration: number
  width: number
  fontSizePx: number
  startTime?: number
}

const PRESET_COLORS = [
  '#FFFFFF', '#FF5555', '#55FF55', '#5555FF', '#FFFF55',
  '#FF55FF', '#55FFFF', '#FF8C00', '#FF69B4', '#00CED1',
  '#FFD700', '#7FFFD4', '#FF6347', '#98FB98', '#DDA0DD'
]

const FONT_SIZE_MAP = { small: 12, medium: 16, large: 20 }
const ROW_HEIGHT = 32
const TOP_PADDING = 8
const MAX_ROWS = 12
const MAX_ACTIVE = 40

function getFontSize(fontSize?: string): number {
  return FONT_SIZE_MAP[fontSize as keyof typeof FONT_SIZE_MAP] || 16
}

function getColor(item: Danmaku): string {
  if (item.color) return item.color
  return PRESET_COLORS[item.id % PRESET_COLORS.length]
}

function getSpeed(content: string, fontSizePx: number): number {
  const baseSpeed = 250 + fontSizePx * 2
  return Math.min(500, Math.max(150, baseSpeed + content.length * 3))
}

function getDuration(item: Danmaku, width: number, speed: number): number {
  if (!canvasRef.value) return 6
  const canvasWidth = canvasRef.value.clientWidth
  return (canvasWidth + width) / speed
}

function updateDimensions() {
  if (!canvasRef.value) return
  const rect = canvasRef.value.getBoundingClientRect()
  canvasRef.value.width = rect.width
  canvasRef.value.height = rect.height
  ctx = canvasRef.value.getContext('2d')
}

function draw() {
  if (!ctx || !canvasRef.value) return
  ctx.clearRect(0, 0, canvasRef.value.width, canvasRef.value.height)

  for (const d of activeDanmakus) {
    ctx.font = `${d.fontSizePx}px "PingFang SC", "Microsoft YaHei", sans-serif`
    ctx.fillStyle = d.color
    ctx.shadowBlur = 0

    ctx.lineWidth = 2
    ctx.strokeStyle = 'black'
    ctx.shadowColor = 'black'
    ctx.shadowBlur = 0
    ctx.textBaseline = 'middle'

    ctx.strokeText(d.content, d.x, d.y)
    ctx.fillText(d.content, d.x, d.y)
  }
}

function updateActiveDanmakus(nowSeconds: number) {
  if (!canvasRef.value) return
  const canvasWidth = canvasRef.value.width
  if (canvasWidth === 0) return

  activeDanmakus = activeDanmakus.filter(d => {
    if (d.type === 'scroll') {
      return d.x + d.width > 0
    } else {
      const elapsed = nowSeconds - (d.startTime || 0)
      return elapsed < d.duration
    }
  })

  const newItems = props.items.filter(item => {
    const elapsed = nowSeconds - item.time
    return elapsed >= -0.2 && elapsed < (item as any)._duration
  }).slice(0, MAX_ACTIVE - activeDanmakus.length)

  const rowsOccupiedUntil = new Array(MAX_ROWS).fill(0)

  for (const d of activeDanmakus) {
    const rowIndex = Math.floor((d.y - TOP_PADDING) / ROW_HEIGHT)
    if (rowIndex >= 0 && rowIndex < MAX_ROWS) {
      rowsOccupiedUntil[rowIndex] = Math.max(rowsOccupiedUntil[rowIndex], nowSeconds + d.duration)
    }
  }

  for (const item of newItems) {
    const fontSizePx = getFontSize(item.fontSize)
    const color = getColor(item)
    const content = item.content
    const type = item.type || 'scroll'

    if (type === 'scroll') {
      if (!ctx) continue
      ctx.font = `${fontSizePx}px "PingFang SC", "Microsoft YaHei", sans-serif`
      const width = ctx.measureText(content).width

      const speed = getSpeed(content, fontSizePx)
      const duration = getDuration(item, width, speed)

      let bestRow = -1
      let minEndTime = Infinity
      for (let i = 0; i < MAX_ROWS; i++) {
        if (rowsOccupiedUntil[i] <= nowSeconds) {
          bestRow = i
          break
        }
        if (rowsOccupiedUntil[i] < minEndTime) {
          minEndTime = rowsOccupiedUntil[i]
          bestRow = i
        }
      }
      if (bestRow === -1) bestRow = 0

      const y = TOP_PADDING + bestRow * ROW_HEIGHT + ROW_HEIGHT / 2
      const startX = canvasWidth
      const endX = -width
      const travelDistance = canvasWidth + width
      const speedPxPerSec = travelDistance / duration
      const elapsed = nowSeconds - item.time
      let currentX = startX - speedPxPerSec * elapsed
      if (currentX > canvasWidth) currentX = canvasWidth
      if (currentX < endX) currentX = endX

      activeDanmakus.push({
        ...item,
        x: currentX,
        y,
        speed: speedPxPerSec,
        duration,
        width,
        fontSizePx,
        color,
        startTime: nowSeconds,
      })

      rowsOccupiedUntil[bestRow] = nowSeconds + duration
    }
    
    else if (type === 'top' || type === 'bottom') {
      let row: number
      if (type === 'top') {
        row = Math.min(2, MAX_ROWS - 1)
      } else {
        row = Math.max(MAX_ROWS - 3, 0)
      }
      const y = TOP_PADDING + row * ROW_HEIGHT + ROW_HEIGHT / 2
      const duration = 4

      if (!ctx) continue
      ctx.font = `${fontSizePx}px "PingFang SC", "Microsoft YaHei", sans-serif`
      const width = ctx.measureText(content).width
      const x = (canvasRef.value.width - width) / 2

      activeDanmakus.push({
        ...item,
        x,
        y,
        speed: 0,
        duration,
        width,
        fontSizePx,
        color,
        startTime: nowSeconds,
      })
    }

    if (!ctx) continue
    ctx.font = `${fontSizePx}px "PingFang SC", "Microsoft YaHei", sans-serif`
    const width = ctx.measureText(content).width

    const speed = getSpeed(content, fontSizePx)
    const duration = getDuration(item, width, speed)

    let bestRow = -1
    let minEndTime = Infinity
    for (let i = 0; i < MAX_ROWS; i++) {
      if (rowsOccupiedUntil[i] <= nowSeconds) {
        bestRow = i
        break
      }
      if (rowsOccupiedUntil[i] < minEndTime) {
        minEndTime = rowsOccupiedUntil[i]
        bestRow = i
      }
    }
    if (bestRow === -1) bestRow = 0

    const y = TOP_PADDING + bestRow * ROW_HEIGHT + ROW_HEIGHT / 2
    const startX = canvasWidth
    const endX = -width
    const travelDistance = canvasWidth + width
    const speedPxPerSec = travelDistance / duration
    const elapsed = nowSeconds - item.time
    let currentX = startX - speedPxPerSec * elapsed
    if (currentX > canvasWidth) currentX = canvasWidth
    if (currentX < endX) currentX = endX

    activeDanmakus.push({
      ...item,
      x: currentX,
      y,
      speed: speedPxPerSec,
      duration,
      width,
      fontSizePx,
      color,
    })

    rowsOccupiedUntil[bestRow] = nowSeconds + duration

    if (activeDanmakus.length > MAX_ACTIVE) activeDanmakus.shift()
  }

  const deltaSeconds = nowSeconds - lastTime
  if (deltaSeconds > 0.03) {
    for (const d of activeDanmakus) {
      d.x -= d.speed * deltaSeconds
    }
    lastTime = nowSeconds
  }

  draw()
}

let animationFrame: number | null = null
function animate() {
  const nowSeconds = performance.now() / 1000
  updateActiveDanmakus(nowSeconds)
  animationFrame = requestAnimationFrame(animate)
}

function resizeObserver() {
  if (!canvasRef.value) return
  updateDimensions()
  activeDanmakus = []
  lastTime = performance.now() / 1000
}

let resizeObserverInstance: ResizeObserver | null = null

onMounted(() => {
  updateDimensions()
  resizeObserverInstance = new ResizeObserver(resizeObserver)
  if (canvasRef.value) resizeObserverInstance.observe(canvasRef.value)
  lastTime = performance.now() / 1000
  animationFrame = requestAnimationFrame(animate)
})

onBeforeUnmount(() => {
  if (animationFrame) cancelAnimationFrame(animationFrame)
  if (resizeObserverInstance) resizeObserverInstance.disconnect()
})

watch(() => props.items, (items) => {
  if (!ctx || !canvasRef.value) return
  const canvasWidth = canvasRef.value.width
  for (const item of items) {
    const fontSizePx = getFontSize(item.fontSize)
    ctx.font = `${fontSizePx}px "PingFang SC", "Microsoft YaHei", sans-serif`
    const width = ctx.measureText(item.content).width
    const speed = getSpeed(item.content, fontSizePx)
    const duration = getDuration(item, width, speed)
    ;(item as any)._duration = duration
  }
}, { immediate: true, deep: false })
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