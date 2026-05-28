<template>
  <div ref="layerRef" class="danmaku-layer">
    <span
      v-for="item in visibleItems"
      :key="item.id"
      :class="['danmaku', `danmaku-${item._fontSize}`, `danmaku-${item._type}`]"
      :style="getStyle(item)"
    >
      {{ item.content }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Danmaku } from '@/types'

const props = defineProps<{ items: Danmaku[]; currentTime: number }>()

const layerRef = ref<HTMLElement>()

const PRESET_COLORS = [
  '#FFFFFF', '#FF5555', '#55FF55', '#5555FF', '#FFFF55',
  '#FF55FF', '#55FFFF', '#FF8C00', '#FF69B4', '#00CED1',
  '#FFD700', '#7FFFD4', '#FF6347', '#98FB98', '#DDA0DD',
]

const ROW_HEIGHT = 32
const TOP_PADDING = 8

interface VisibleDanmaku extends Danmaku {
  _row: number
  _fontSize: string
  _type: string
  _color: string
  _duration: number
}

const visibleItems = computed<VisibleDanmaku[]>(() => {
  if (!layerRef.value) return []

  const containerHeight = layerRef.value.clientHeight
  const totalRows = Math.max(3, Math.floor((containerHeight - TOP_PADDING * 2) / ROW_HEIGHT * 0.5))
  const rowOccupiedUntil = new Array<number>(totalRows).fill(0)
  const result: VisibleDanmaku[] = []

  const candidates = props.items
    .filter((item) => {
      const elapsed = props.currentTime - item.time
      return elapsed > -1 && elapsed < 12
    })
    .sort((a, b) => a.time - b.time)

  for (const item of candidates) {
    const fontSize = item.fontSize || 'medium'
    const type = item.type || 'scroll'
    const duration = type === 'scroll'
      ? Math.max(5, Math.min(12, 6 + item.content.length * 0.03))
      : 5
    // Deterministic color from preset palette, unless user explicitly picked a non-default color
    const color = (!item.color || item.color === '#FFFFFF')
      ? PRESET_COLORS[item.id % PRESET_COLORS.length]
      : item.color

    const elapsed = props.currentTime - item.time
    if (type !== 'scroll' && elapsed > duration) continue

    if (type === 'scroll') {
      const occupyDuration = duration * 0.3
      let bestRow = -1
      let bestTime = Infinity

      for (let row = 0; row < totalRows; row++) {
        if (rowOccupiedUntil[row] <= props.currentTime) {
          bestRow = row
          break
        }
        if (rowOccupiedUntil[row] < bestTime) {
          bestTime = rowOccupiedUntil[row]
          bestRow = row
        }
      }

      if (bestRow >= 0) {
        const startTime = Math.max(props.currentTime, rowOccupiedUntil[bestRow])
        rowOccupiedUntil[bestRow] = startTime + occupyDuration
        result.push({ ...item, _row: bestRow, _fontSize: fontSize, _type: type, _color: color, _duration: duration })
      }
    } else if (type === 'top') {
      const maxTopRow = Math.min(2, totalRows - 1)
      for (let row = 0; row <= maxTopRow; row++) {
        if (rowOccupiedUntil[row] <= props.currentTime) {
          rowOccupiedUntil[row] = props.currentTime + duration
          result.push({ ...item, _row: row, _fontSize: fontSize, _type: type, _color: color, _duration: duration })
          break
        }
      }
    } else if (type === 'bottom') {
      const minBottomRow = Math.max(0, totalRows - 3)
      for (let row = minBottomRow; row < totalRows; row++) {
        if (rowOccupiedUntil[row] <= props.currentTime) {
          rowOccupiedUntil[row] = props.currentTime + duration
          result.push({ ...item, _row: row, _fontSize: fontSize, _type: type, _color: color, _duration: duration })
          break
        }
      }
    }

    if (result.length >= 30) break
  }

  return result
})

function getStyle(item: VisibleDanmaku) {
  const top = TOP_PADDING + item._row * ROW_HEIGHT

  if (item._type !== 'scroll') {
    return {
      top: `${top}px`,
      color: item._color,
      animation: `danmaku-fixed ${item._duration}s linear forwards`,
    }
  }

  // scroll type
  return {
    top: `${top}px`,
    color: item._color,
    '--duration': `${item._duration}s`,
  }
}
</script>

<style scoped>
.danmaku-layer {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
}

.danmaku {
  position: absolute;
  min-width: max-content;
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
  user-select: none;
  /* Bilibili-style text stroke for readability on any background */
  text-shadow:
    1px 0 0 #000,
    -1px 0 0 #000,
    0 1px 0 #000,
    0 -1px 0 #000,
    1px 1px 0 #000,
    -1px -1px 0 #000,
    1px -1px 0 #000,
    -1px 1px 0 #000;
}

/* Font sizes */
.danmaku-small  { font-size: 12px; }
.danmaku-medium { font-size: 16px; }
.danmaku-large  { font-size: 20px; }

/* Scroll type */
.danmaku-scroll {
  left: 100%;
  animation: danmaku-fly var(--duration, 6s) linear forwards;
}

@keyframes danmaku-fly {
  to {
    transform: translateX(calc(-100vw - 100%));
  }
}

/* Fixed top/bottom types */
.danmaku-top,
.danmaku-bottom {
  left: 50%;
  transform: translateX(-50%);
}

@keyframes danmaku-fixed {
  0%   { opacity: 0; }
  10%  { opacity: 1; }
  90%  { opacity: 1; }
  100% { opacity: 0; }
}
</style>
