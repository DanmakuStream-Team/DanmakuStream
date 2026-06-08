<template>
  <div class="player">
    <video
      ref="videoRef"
      :autoplay="autoplay"
      controls
      playsinline
      :poster="mediaUrl(poster)"
      @timeupdate="emitTime"
      @error="handleVideoError"
    />
    <DanmakuLayer :items="danmakus" :current-time="currentTime" />
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
const sourceUrl = computed(() => mediaUrl(props.url))
let hls: Hls | null = null

watch(sourceUrl, setupSource, { immediate: true })

onBeforeUnmount(destroyHls)

async function setupSource(url: string) {
  await nextTick()
  const video = videoRef.value
  if (!video) return

  destroyHls()
  video.removeAttribute('src')
  video.load()
  currentTime.value = 0

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
        if (props.autoplay) video.play().catch(() => {})
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
}

video {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
}
</style>
