<template>
  <div class="player">
    <video
      ref="videoRef"
      :autoplay="autoplay"
      controls
      playsinline
      :poster="mediaUrl(poster)"
      @timeupdate="emitTime"
      @play="isPaused = false"
      @pause="isPaused = true"
      @ended="isPaused = true"
      @error="handleVideoError"
    />
    <DanmakuLayer
      :items="danmakus"
      :current-time="currentTime"
      :paused="isPaused"
      :visible="danmakuVisible"
      :opacity="danmakuOpacity / 100"
    />

    <div class="player-brand">DanmakuStream</div>
    <div class="player-controls">
      <el-select
        v-if="qualityOptions.length > 1"
        v-model="selectedQuality"
        class="quality-select"
        size="small"
        @change="switchQuality"
      >
        <el-option label="自动" :value="-1" />
        <el-option
          v-for="option in qualityOptions"
          :key="option.value"
          :label="option.label"
          :value="option.value"
        />
      </el-select>
      <el-switch
        v-model="danmakuVisibleModel"
        size="small"
        inline-prompt
        active-text="弹"
        inactive-text="关"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import Hls from 'hls.js'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import type { Danmaku } from '@/types'
import { mediaUrl } from '@/utils/format'
import DanmakuLayer from './DanmakuLayer.vue'

const props = withDefaults(defineProps<{
  url: string
  poster?: string
  danmakus: Danmaku[]
  autoplay?: boolean
  danmakuVisible?: boolean
  danmakuOpacity?: number
}>(), {
  danmakuVisible: true,
  danmakuOpacity: 85,
})

const emit = defineEmits<{
  timeupdate: [time: number]
  error: [message: string]
  'update:danmakuVisible': [visible: boolean]
}>()

const videoRef = ref<HTMLVideoElement>()
const currentTime = ref(0)
const isPaused = ref(true)
const selectedQuality = ref(-1)
const qualityOptions = ref<{ label: string; value: number }[]>([])
const sourceUrl = computed(() => mediaUrl(props.url))
const danmakuVisibleModel = computed({
  get: () => props.danmakuVisible,
  set: (value: boolean) => emit('update:danmakuVisible', value),
})
let hls: Hls | null = null

watch(sourceUrl, setupSource, { immediate: true })

onBeforeUnmount(destroyHls)

async function setupSource(url: string) {
  await nextTick()
  const video = videoRef.value
  if (!video) return

  destroyHls()
  qualityOptions.value = []
  selectedQuality.value = -1
  video.removeAttribute('src')
  video.load()
  currentTime.value = 0
  isPaused.value = true

  if (!url) return

  if (isHlsSource(url)) {
    if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = url
      video.load()
      return
    }

    if (Hls.isSupported()) {
      hls = new Hls({ enableWorker: true })
      hls.loadSource(url)
      hls.attachMedia(video)
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        qualityOptions.value = hls?.levels.map((level, index) => ({
          label: `${level.height || '未知'}P`,
          value: index,
        })) || []
      })
      hls.on(Hls.Events.ERROR, (_event, data) => {
        if (data.fatal) emit('error', '视频加载失败，请确认资源已转码完成')
      })
      return
    }

    emit('error', '当前浏览器不支持 HLS 视频播放')
    return
  }

  video.src = url
  video.load()
}

function isHlsSource(url: string) {
  return /\.m3u8($|\?)/i.test(url)
}

function destroyHls() {
  if (hls) {
    hls.destroy()
    hls = null
  }
}

function emitTime() {
  currentTime.value = videoRef.value?.currentTime || 0
  emit('timeupdate', currentTime.value)
}

function handleVideoError() {
  emit('error', '视频加载失败，请确认资源已转码完成')
}

function switchQuality(value: number) {
  if (!hls) return
  hls.currentLevel = value
}

defineExpose({
  play: () => videoRef.value?.play(),
  pause: () => videoRef.value?.pause(),
  seek: (time: number) => {
    if (videoRef.value) videoRef.value.currentTime = time
  },
  getCurrentTime: () => videoRef.value?.currentTime || 0,
  getDuration: () => videoRef.value?.duration || 0,
})
</script>

<style scoped>
.player {
  position: relative;
  overflow: hidden;
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 10px;
  background: #05070d;
  box-shadow: 0 16px 40px rgb(15 23 42 / 16%);
}

video {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
  background: #000;
}

.player-brand {
  position: absolute;
  top: 16px;
  right: 18px;
  z-index: 3;
  color: rgb(255 255 255 / 72%);
  font-size: 13px;
  font-weight: 700;
  pointer-events: none;
}

.player-controls {
  position: absolute;
  right: 14px;
  bottom: 48px;
  z-index: 5;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 10px;
  border-radius: 8px;
  background: rgb(0 0 0 / 42%);
  backdrop-filter: blur(8px);
  opacity: 0;
  transform: translateY(6px);
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.player:hover .player-controls,
.player:focus-within .player-controls {
  opacity: 1;
  transform: translateY(0);
}

.quality-select {
  width: 92px;
}
</style>
