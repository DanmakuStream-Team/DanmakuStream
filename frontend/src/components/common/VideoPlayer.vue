<template>
  <div
    ref="playerRef"
    class="player"
    :class="{ 'controls-visible': controlsVisible || isPaused, 'is-live': live }"
    tabindex="0"
    @mousemove="showControls"
    @mouseleave="scheduleControlsHide"
    @keydown="handleShortcut"
  >
    <video
      ref="videoRef"
      :autoplay="autoplay"
      playsinline
      :poster="mediaUrl(poster)"
      @click="togglePlayback"
      @dblclick="toggleFullscreen"
      @timeupdate="emitTime"
      @durationchange="syncDuration"
      @loadedmetadata="syncDuration"
      @volumechange="syncVolume"
      @ratechange="syncPlaybackRate"
      @play="handlePlay"
      @playing="buffering = false"
      @waiting="buffering = true"
      @canplay="buffering = false"
      @pause="handlePause"
      @ended="handleEnded"
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
    <div v-if="buffering" class="buffering" aria-label="正在加载" />
    <button
      v-else-if="isPaused"
      class="center-play"
      type="button"
      aria-label="播放"
      @click="togglePlayback"
    >
      <el-icon><VideoPlay /></el-icon>
    </button>

    <div class="control-layer" @click.stop>
      <div v-if="!live" class="progress-wrap">
        <input
          class="progress"
          type="range"
          min="0"
          :max="Math.max(duration, 0)"
          step="0.1"
          :value="currentTime"
          :style="{ '--played': playedPercent + '%' }"
          aria-label="播放进度"
          @input="seekFromInput"
        >
        <button
          v-for="chapter in visibleChapters"
          :key="`${chapter.time}-${chapter.label}`"
          class="chapter-marker"
          type="button"
          :title="`${formatPlayerTime(chapter.time)} ${chapter.label}`"
          :style="{ left: `${chapter.time / duration * 100}%` }"
          @click="seek(chapter.time)"
        />
      </div>

      <div class="control-row">
        <div class="control-group">
          <button class="icon-control" type="button" :title="isPaused ? '播放 (K)' : '暂停 (K)'" @click="togglePlayback">
            <el-icon><VideoPlay v-if="isPaused" /><VideoPause v-else /></el-icon>
          </button>
          <button class="icon-control" type="button" :title="muted ? '取消静音 (M)' : '静音 (M)'" @click="toggleMute">
            <el-icon><Mute v-if="muted || volume === 0" /><Microphone v-else /></el-icon>
          </button>
          <input
            class="volume"
            type="range"
            min="0"
            max="1"
            step="0.05"
            :value="muted ? 0 : volume"
            aria-label="音量"
            @input="setVolumeFromInput"
          >
          <span class="time-label">
            <template v-if="live"><i />直播</template>
            <template v-else>{{ formatPlayerTime(currentTime) }} / {{ formatPlayerTime(duration) }}</template>
          </span>
          <span v-if="currentChapter" class="chapter-label">{{ currentChapter.label }}</span>
        </div>

        <div class="control-group control-right">
          <el-select
            v-if="!live"
            v-model="playbackRate"
            class="speed-select"
            size="small"
            title="播放速度"
            @change="setPlaybackRate"
          >
            <el-option v-for="rate in playbackRates" :key="rate" :label="rate === 1 ? '倍速' : `${rate}x`" :value="rate" />
          </el-select>
          <el-select
            v-if="qualityOptions.length > 1"
            v-model="selectedQuality"
            class="quality-select"
            size="small"
            title="清晰度"
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
          <button class="icon-control" type="button" title="全屏 (F)" @click="toggleFullscreen">
            <el-icon><FullScreen /></el-icon>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Hls from 'hls.js'
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { FullScreen, Microphone, Mute, VideoPause, VideoPlay } from '@element-plus/icons-vue'
import type { Danmaku, VideoChapter } from '@/types'
import { mediaUrl } from '@/utils/format'
import DanmakuLayer from './DanmakuLayer.vue'

const props = withDefaults(defineProps<{
  url: string
  poster?: string
  danmakus: Danmaku[]
  autoplay?: boolean
  live?: boolean
  danmakuVisible?: boolean
  danmakuOpacity?: number
  chapters?: VideoChapter[]
}>(), {
  autoplay: false,
  live: false,
  danmakuVisible: true,
  danmakuOpacity: 85,
  chapters: () => [],
})

const emit = defineEmits<{
  timeupdate: [time: number]
  play: []
  pause: []
  ended: []
  error: [message: string]
}>()

const playerRef = ref<HTMLElement>()
const videoRef = ref<HTMLVideoElement>()
const currentTime = ref(0)
const duration = ref(0)
const isPaused = ref(true)
const buffering = ref(false)
const controlsVisible = ref(true)
const volume = ref(readNumberSetting('danmaku:player-volume', 1))
const muted = ref(localStorage.getItem('danmaku:player-muted') === 'true')
const playbackRate = ref(readNumberSetting('danmaku:player-rate', 1))
const playbackRates = [0.5, 0.75, 1, 1.25, 1.5, 2]
const selectedQuality = ref(-1)
const qualityOptions = ref<{ label: string; value: number }[]>([])
const sourceUrl = computed(() => mediaUrl(props.url))
const playedPercent = computed(() => duration.value > 0 ? Math.min(100, currentTime.value / duration.value * 100) : 0)
const visibleChapters = computed(() => props.chapters.filter((chapter) => chapter.time >= 0 && chapter.time < duration.value))
const currentChapter = computed(() => {
  for (let index = visibleChapters.value.length - 1; index >= 0; index -= 1) {
    if (currentTime.value >= visibleChapters.value[index].time) return visibleChapters.value[index]
  }
  return undefined
})
let hls: Hls | null = null
let controlsTimer: number | undefined
let recoveringMediaError = false

watch(sourceUrl, setupSource, { immediate: true })

onBeforeUnmount(() => {
  destroyHls()
  if (controlsTimer) clearTimeout(controlsTimer)
})

async function setupSource(url: string) {
  await nextTick()
  const video = videoRef.value
  if (!video) return

  destroyHls()
  qualityOptions.value = []
  selectedQuality.value = -1
  video.removeAttribute('src')
  video.load()
  video.volume = volume.value
  video.muted = muted.value
  video.playbackRate = props.live ? 1 : playbackRate.value
  currentTime.value = 0
  duration.value = 0
  isPaused.value = true
  buffering.value = Boolean(url)
  recoveringMediaError = false

  if (!url) return

  if (isHlsSource(url)) {
    if (video.canPlayType('application/vnd.apple.mpegurl')) {
      video.src = url
      video.load()
      return
    }

    if (Hls.isSupported()) {
      hls = new Hls({
        enableWorker: true,
        lowLatencyMode: props.live,
        backBufferLength: props.live ? 30 : 90,
      })
      hls.loadSource(url)
      hls.attachMedia(video)
      hls.on(Hls.Events.MANIFEST_PARSED, () => {
        qualityOptions.value = hls?.levels.map((level, index) => ({
          label: level.height ? `${level.height}P` : `线路 ${index + 1}`,
          value: index,
        })) || []
        if (props.autoplay) void video.play().catch(() => undefined)
      })
      hls.on(Hls.Events.ERROR, (_event, data) => {
        if (!data.fatal || !hls) return
        if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
          hls.startLoad()
          return
        }
        if (data.type === Hls.ErrorTypes.MEDIA_ERROR && !recoveringMediaError) {
          recoveringMediaError = true
          hls.recoverMediaError()
          window.setTimeout(() => { recoveringMediaError = false }, 1500)
          return
        }
        emit('error', '视频加载失败，请确认资源可用')
      })
      return
    }

    emit('error', '当前浏览器不支持 HLS 视频播放')
    return
  }

  video.src = url
  video.load()
  if (props.autoplay) void video.play().catch(() => undefined)
}

function isHlsSource(url: string) {
  return /\.m3u8($|\?)/i.test(url)
}

function destroyHls() {
  hls?.destroy()
  hls = null
}

function togglePlayback() {
  const video = videoRef.value
  if (!video) return
  if (video.paused) void video.play().catch(() => undefined)
  else video.pause()
}

function handlePlay() {
  isPaused.value = false
  emit('play')
  scheduleControlsHide()
}

function handlePause() {
  isPaused.value = true
  controlsVisible.value = true
  emit('pause')
}

function handleEnded() {
  isPaused.value = true
  controlsVisible.value = true
  emit('ended')
}

function emitTime() {
  currentTime.value = videoRef.value?.currentTime || 0
  emit('timeupdate', currentTime.value)
}

function syncDuration() {
  const value = videoRef.value?.duration || 0
  duration.value = Number.isFinite(value) ? value : 0
}

function seekFromInput(event: Event) {
  seek(Number((event.target as HTMLInputElement).value))
}

function seek(time: number) {
  const video = videoRef.value
  if (!video || props.live || !Number.isFinite(time) || time < 0) return
  const applySeek = () => {
    const limit = Number.isFinite(video.duration) ? video.duration : time
    video.currentTime = Math.min(time, limit)
    currentTime.value = video.currentTime
  }
  if (video.readyState >= 1) applySeek()
  else video.addEventListener('loadedmetadata', applySeek, { once: true })
}

function toggleMute() {
  const video = videoRef.value
  if (!video) return
  video.muted = !video.muted
}

function setVolumeFromInput(event: Event) {
  const video = videoRef.value
  if (!video) return
  const value = Number((event.target as HTMLInputElement).value)
  video.volume = Math.min(1, Math.max(0, value))
  video.muted = value === 0
}

function syncVolume() {
  const video = videoRef.value
  if (!video) return
  volume.value = video.volume
  muted.value = video.muted
  localStorage.setItem('danmaku:player-volume', String(video.volume))
  localStorage.setItem('danmaku:player-muted', String(video.muted))
}

function setPlaybackRate(value: number) {
  const video = videoRef.value
  if (!video || props.live) return
  video.playbackRate = value
}

function syncPlaybackRate() {
  const video = videoRef.value
  if (!video || props.live) return
  playbackRate.value = video.playbackRate
  localStorage.setItem('danmaku:player-rate', String(video.playbackRate))
}

function switchQuality(value: number) {
  if (hls) hls.currentLevel = value
}

async function toggleFullscreen() {
  const player = playerRef.value
  if (!player) return
  if (document.fullscreenElement) await document.exitFullscreen()
  else await player.requestFullscreen()
}

function handleShortcut(event: KeyboardEvent) {
  const target = event.target as HTMLElement
  if (target.matches('input, textarea, select') || target.closest('.el-select')) return
  const key = event.key.toLowerCase()
  if (key === ' ' || key === 'k') {
    event.preventDefault()
    togglePlayback()
  } else if (key === 'm') {
    toggleMute()
  } else if (key === 'f') {
    void toggleFullscreen()
  } else if (!props.live && (key === 'arrowleft' || key === 'arrowright')) {
    event.preventDefault()
    seek(currentTime.value + (key === 'arrowleft' ? -5 : 5))
  }
  showControls()
}

function showControls() {
  controlsVisible.value = true
  scheduleControlsHide()
}

function scheduleControlsHide() {
  if (controlsTimer) clearTimeout(controlsTimer)
  if (isPaused.value) return
  controlsTimer = window.setTimeout(() => { controlsVisible.value = false }, 2200)
}

function handleVideoError() {
  buffering.value = false
  emit('error', '视频加载失败，请确认资源可用')
}

function formatPlayerTime(value: number) {
  if (!Number.isFinite(value) || value < 0) return '00:00'
  const seconds = Math.floor(value % 60).toString().padStart(2, '0')
  const minutes = Math.floor(value / 60) % 60
  const hours = Math.floor(value / 3600)
  return hours > 0
    ? `${hours}:${minutes.toString().padStart(2, '0')}:${seconds}`
    : `${minutes.toString().padStart(2, '0')}:${seconds}`
}

function readNumberSetting(key: string, fallback: number) {
  const value = Number(localStorage.getItem(key))
  return Number.isFinite(value) && value > 0 ? value : fallback
}

defineExpose({
  play: () => videoRef.value?.play(),
  pause: () => videoRef.value?.pause(),
  seek,
  getCurrentTime: () => videoRef.value?.currentTime || 0,
  getDuration: () => videoRef.value?.duration || 0,
})
</script>

<style scoped>
.player { position: relative; overflow: hidden; width: 100%; aspect-ratio: 16 / 9; border-radius: 8px; outline: none; background: #05070d; box-shadow: 0 12px 32px rgb(15 23 42 / 14%); }
video { display: block; width: 100%; height: 100%; object-fit: contain; background: #000; cursor: pointer; }
.player-brand { position: absolute; top: 16px; right: 18px; z-index: 3; color: rgb(255 255 255 / 72%); font-size: 13px; font-weight: 700; pointer-events: none; }
.buffering { position: absolute; top: 50%; left: 50%; z-index: 6; width: 42px; height: 42px; border: 3px solid rgb(255 255 255 / 30%); border-top-color: #fff; border-radius: 50%; animation: spin 0.8s linear infinite; pointer-events: none; }
.center-play { position: absolute; top: 50%; left: 50%; z-index: 6; display: grid; width: 64px; height: 64px; padding: 0; place-items: center; border: 0; border-radius: 50%; background: rgb(0 0 0 / 56%); color: #fff; cursor: pointer; transform: translate(-50%, -50%); }
.center-play .el-icon { font-size: 32px; }
.control-layer { position: absolute; right: 0; bottom: 0; left: 0; z-index: 8; padding: 38px 14px 10px; background: linear-gradient(transparent, rgb(0 0 0 / 78%)); opacity: 0; pointer-events: none; transform: translateY(8px); transition: opacity 0.18s ease, transform 0.18s ease; }
.player.controls-visible .control-layer, .player:focus-within .control-layer { opacity: 1; pointer-events: auto; transform: translateY(0); }
.progress-wrap { position: relative; height: 13px; margin: 0 0 5px; }
.progress { position: absolute; inset: 0; width: 100%; height: 4px; margin: auto 0; accent-color: #00aeec; cursor: pointer; }
.chapter-marker { position: absolute; top: 3px; z-index: 2; width: 3px; height: 7px; padding: 0; border: 0; background: rgb(255 255 255 / 88%); cursor: pointer; transform: translateX(-1px); }
.control-row, .control-group { display: flex; align-items: center; }
.control-row { min-height: 32px; justify-content: space-between; gap: 12px; }
.control-group { min-width: 0; gap: 6px; }
.control-right { justify-content: flex-end; }
.icon-control { display: grid; width: 32px; height: 32px; flex: 0 0 32px; padding: 0; place-items: center; border: 0; background: transparent; color: #fff; cursor: pointer; }
.icon-control .el-icon { font-size: 20px; }
.volume { width: 74px; accent-color: #fff; cursor: pointer; }
.time-label { display: inline-flex; align-items: center; gap: 6px; color: #fff; font-size: 12px; white-space: nowrap; }
.chapter-label { overflow: hidden; max-width: 220px; color: rgb(255 255 255 / 82%); font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.time-label i { width: 7px; height: 7px; border-radius: 50%; background: #f04438; }
.speed-select { width: 74px; }
.quality-select { width: 86px; }
.speed-select :deep(.el-select__wrapper), .quality-select :deep(.el-select__wrapper) { min-height: 28px; border: 0; background: transparent; box-shadow: none; }
.speed-select :deep(.el-select__selected-item), .quality-select :deep(.el-select__selected-item), .speed-select :deep(.el-select__caret), .quality-select :deep(.el-select__caret) { color: #fff; }
@keyframes spin { to { transform: rotate(360deg); } }
@media (max-width: 640px) { .volume, .chapter-label { display: none; } .time-label { font-size: 11px; } .control-layer { padding-right: 8px; padding-left: 8px; } .speed-select { width: 64px; } .quality-select { width: 76px; } }
</style>
