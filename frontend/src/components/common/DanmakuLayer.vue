<template>
  <div class="danmaku-layer">
    <div
      v-for="item in activeDanmakus"
      :key="item.renderId"
      class="danmaku-item"
      :class="[
        `danmaku-${item.type}`,
        `danmaku-size-${item.fontSize}`,
      ]"
      :style="getDanmakuStyle(item)"
    >
      {{ item.content }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Danmaku } from '@/types'

interface RenderDanmaku extends Danmaku {
  renderId: string
  track: number
}

const props = withDefaults(
  defineProps<{
    danmakus: Danmaku[]
    currentTime: number
    enabled?: boolean
    maxTracks?: number
    timeWindow?: number
  }>(),
  {
    enabled: true,
    maxTracks: 8,
    timeWindow: 0.35,
  },
)

const activeDanmakus = ref<RenderDanmaku[]>([])
const renderedKeys = ref<Set<string>>(new Set())
const lastTime = ref(0)
const trackIndex = ref(0)

const sortedDanmakus = computed(() => {
  return [...props.danmakus].sort((a, b) => a.time - b.time)
})

watch(
  () => props.currentTime,
  (newTime, oldTime) => {
    if (!props.enabled) {
      return
    }

    if (!Number.isFinite(newTime)) {
      return
    }

    const isSeek = Math.abs(newTime - oldTime) > 2

    if (isSeek) {
      renderedKeys.value.clear()
      activeDanmakus.value = []
    }

    const from = isSeek ? newTime - props.timeWindow : lastTime.value
    const to = newTime + props.timeWindow

    const matched = sortedDanmakus.value.filter((danmaku) => {
      const key = getDanmakuKey(danmaku)

      return (
        danmaku.time >= from &&
        danmaku.time <= to &&
        !renderedKeys.value.has(key)
      )
    })

    matched.forEach(addDanmaku)

    lastTime.value = newTime
  },
)

watch(
  () => props.enabled,
  (enabled) => {
    if (!enabled) {
      activeDanmakus.value = []
    }
  },
)

function addDanmaku(danmaku: Danmaku) {
  const key = getDanmakuKey(danmaku)
  renderedKeys.value.add(key)

  const renderItem: RenderDanmaku = {
    ...danmaku,
    renderId: `${key}-${Date.now()}-${Math.random()}`,
    track: getNextTrack(),
  }

  activeDanmakus.value.push(renderItem)

  const duration = renderItem.type === 'scroll' ? 8000 : 3500

  window.setTimeout(() => {
    activeDanmakus.value = activeDanmakus.value.filter(
      item => item.renderId !== renderItem.renderId,
    )
  }, duration)
}

function getNextTrack() {
  const track = trackIndex.value % props.maxTracks
  trackIndex.value += 1
  return track
}

function getDanmakuKey(danmaku: Danmaku) {
  return `${danmaku.id}-${danmaku.time}-${danmaku.content}`
}

function getDanmakuStyle(item: RenderDanmaku) {
  const top = `${8 + item.track * 10}%`

  if (item.type === 'top') {
    return {
      top,
      color: item.color,
    }
  }

  if (item.type === 'bottom') {
    return {
      bottom: `${8 + item.track * 10}%`,
      color: item.color,
    }
  }

  return {
    top,
    color: item.color,
  }
}
</script>

<style scoped>
.danmaku-layer {
  position: absolute;
  inset: 0;
  z-index: 20;
  pointer-events: none;
  overflow: hidden;
}

.danmaku-item {
  position: absolute;
  max-width: 80%;
  white-space: nowrap;
  font-weight: 600;
  text-shadow:
    1px 1px 2px rgba(0, 0, 0, 0.9),
    -1px -1px 2px rgba(0, 0, 0, 0.9);
  will-change: transform, opacity;
}

.danmaku-scroll {
  left: 100%;
  animation: danmaku-scroll 8s linear forwards;
}

.danmaku-top {
  left: 50%;
  transform: translateX(-50%);
  animation: danmaku-fixed 3.5s linear forwards;
}

.danmaku-bottom {
  left: 50%;
  transform: translateX(-50%);
  animation: danmaku-fixed 3.5s linear forwards;
}

.danmaku-size-small {
  font-size: 16px;
}

.danmaku-size-medium {
  font-size: 22px;
}

.danmaku-size-large {
  font-size: 28px;
}

@keyframes danmaku-scroll {
  from {
    transform: translateX(0);
  }

  to {
    transform: translateX(calc(-100vw - 100%));
  }
}

@keyframes danmaku-fixed {
  0% {
    opacity: 0;
  }

  8% {
    opacity: 1;
  }

  92% {
    opacity: 1;
  }

  100% {
    opacity: 0;
  }
}
</style>