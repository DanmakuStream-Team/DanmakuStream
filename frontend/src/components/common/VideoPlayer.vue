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
    <DanmakuLayer :items="danmakus" :current-time="currentTime" :paused="isPaused" />
    <div class="player-toolbar">
      <span class="player-brand">DanmakuStream</span>
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
    </div>
  </div>
</template>

<script setup lang="ts">
import Hls from 'hls.js'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import type { Danmaku } from '@/types'
import { mediaUrl } from '@/utils/format'
import DanmakuLayer from './DanmakuLayer.vue'

const props = defineProps<{ url: string; poster?: string; danmakus: Danmaku[]; autoplay?: boolean }>()
const emit = defineEmits<{ timeupdate: [time: number]; error: [message: string] }>()

const videoRef = ref<HTMLVideoElement>()
const currentTime = ref(0)
const isPaused = ref(true)
const selectedQuality = ref(-1)
const qualityOptions = ref<{ label: string; value: number }[]>([])
const sourceUrl = computed(() => mediaUrl(props.url))
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
      hls = new Hls()
      hls.on(Hls.Events.ERROR, (_, data) => {
        if (data.fatal) {
          emit('error', '视频加载失败，请确认资源已转码完成')
          destroyHls()
        }
      })
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        qualityOptions.value = hls?.levels.map((level, index) => ({
          value: index,
          label: level.height ? `${level.height}P` : `${Math.round(level.bitrate / 1000)}Kbps`,
        })) || []
        if (props.autoplay) video.play().catch(() => {})
      })
      hls.on(Hls.Events.LEVEL_SWITCHED, (_, data) => {
        selectedQuality.value = hls?.autoLevelEnabled ? -1 : data.level
      })
      hls.loadSource(url)
      hls.attachMedia(video)
      return
    }

    emit('error', '当前浏览器不支持 HLS 视频播放')
    return
  }

  video.src = url
  video.load()
  if (props.autoplay) video.play().catch(() => {})
}

function destroyHls() {
  if (!hls) return
  hls.destroy()
  hls = null
}

function isHlsSource(url: string) {
  return /\.m3u8($|\?)/i.test(url)
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
  border-radius: 10px;
  background: #0b1020;
  aspect-ratio: 16 / 9;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.18);
}

video {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
}

.player-toolbar {
  position: absolute;
  top: 12px;
  right: 12px;
  left: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.18s ease;
}

.player:hover .player-toolbar {
  opacity: 1;
}

.player-brand {
  padding: 5px 9px;
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.42);
  color: rgba(255, 255, 255, 0.88);
  font-size: 12px;
  line-height: 1;
}

.quality-select {
  width: 96px;
  pointer-events: auto;
}
</style>
