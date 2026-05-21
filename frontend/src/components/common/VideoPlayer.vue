<template>
  <div class="relative w-full bg-black rounded-xl overflow-hidden aspect-video">
    <div v-if="playerError" class="absolute inset-0 z-30 flex flex-col items-center justify-center gap-3 text-white bg-neutral-900">
      <div class="text-xl font-semibold">视频加载失败</div>
      <div class="text-sm text-gray-400">{{ playerError }}</div>
      <el-button type="primary" @click="retry">重新加载</el-button>
    </div>

    <div
      v-loading="playerLoading && !playerError"
      element-loading-background="rgba(0,0,0,0.4)"
      class="w-full h-full"
    >
      <div
        v-show="!playerError"
        ref="playerContainer"
        class="w-full h-full bg-black"
      />

      <DanmakuLayer
        v-if="!playerError"
        :danmakus="danmakus"
        :current-time="currentTime"
        :enabled="danmakuEnabled"
      />
    </div>

    <div class="absolute left-3 right-3 bottom-3 z-30 flex items-center gap-3 pointer-events-none">
      <el-switch
        v-model="danmakuEnabled"
        active-text="弹幕开"
        inactive-text="弹幕关"
        class="pointer-events-auto"
      />
      <span class="text-white text-xs drop-shadow">
        {{ formatDuration(currentTime) }} / {{ formatDuration(duration) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import Player from 'xgplayer'
import HlsPlugin from 'xgplayer-hls'
import 'xgplayer/dist/index.min.css'
import DanmakuLayer from '@/components/common/DanmakuLayer.vue'
import type { Danmaku } from '@/types'

const props = withDefaults(
  defineProps<{
    url: string
    poster?: string
    danmakus?: Danmaku[]
    autoplay?: boolean
  }>(),
  {
    poster: '',
    danmakus: () => [],
    autoplay: false,
  },
)

const emit = defineEmits<{
  ready: []
  error: [message: string]
  timeupdate: [time: number]
  play: []
  pause: []
  seek: [time: number]
}>()

const playerContainer = ref<HTMLElement>()
const player = ref<Player | null>(null)

const playerLoading = ref(false)
const playerError = ref('')
const currentTime = ref(0)
const duration = ref(0)
const danmakuEnabled = ref(true)

onMounted(async () => {
  await nextTick()
  initPlayer()
})

onUnmounted(() => {
  destroyPlayer()
})

watch(
  () => props.url,
  async () => {
    await nextTick()
    initPlayer()
  },
)

function initPlayer() {
  destroyPlayer()

  if (!props.url) {
    playerError.value = '视频地址为空'
    return
  }

  if (!playerContainer.value) {
    playerError.value = '播放器容器初始化失败'
    return
  }

  playerLoading.value = true
  playerError.value = ''

  try {
    const isHls = /\.m3u8(\?|$)/i.test(props.url)

    player.value = new Player({
      el: playerContainer.value,
      url: props.url,
      poster: props.poster,
      width: '100%',
      height: '100%',
      autoplay: props.autoplay,
      fluid: true,
      playbackRate: [0.5, 0.75, 1, 1.25, 1.5, 2],
      defaultPlaybackRate: 1,
      playsinline: true,
      lang: 'zh-cn',
      plugins: isHls ? [HlsPlugin] : [],
    })

    player.value.once('ready', () => {
      playerLoading.value = false
      duration.value = player.value?.duration || 0
      emit('ready')
    })

    player.value.on('error', () => {
      playerLoading.value = false
      playerError.value = '播放器加载出错，请检查视频地址或网络连接'
      emit('error', playerError.value)
    })

    player.value.on('canplay', () => {
      playerLoading.value = false
      playerError.value = ''
      duration.value = player.value?.duration || 0
    })

    player.value.on('waiting', () => {
      playerLoading.value = true
    })

    player.value.on('playing', () => {
      playerLoading.value = false
      emit('play')
    })

    player.value.on('pause', () => {
      emit('pause')
    })

    player.value.on('timeupdate', () => {
      currentTime.value = player.value?.currentTime || 0
      duration.value = player.value?.duration || duration.value
      emit('timeupdate', currentTime.value)
    })

    player.value.on('seeked', () => {
      currentTime.value = player.value?.currentTime || 0
      emit('seek', currentTime.value)
    })
  } catch {
    playerLoading.value = false
    playerError.value = '播放器初始化失败'
    emit('error', playerError.value)
  }
}

function destroyPlayer() {
  if (player.value) {
    player.value.destroy()
    player.value = null
  }
}

function retry() {
  initPlayer()
}

function play() { player.value?.play() }
function pause() { player.value?.pause() }
function seek(time: number) {
  if (!player.value) return
  player.value.currentTime = time
  currentTime.value = time
  emit('seek', time)
}
function getCurrentTime() { return player.value?.currentTime || 0 }
function getDuration() { return player.value?.duration || duration.value || 0 }

function formatDuration(seconds: number) {
  if (!Number.isFinite(seconds) || seconds < 0) return '00:00'
  const total = Math.floor(seconds)
  const minute = Math.floor(total / 60)
  const second = total % 60
  return `${String(minute).padStart(2, '0')}:${String(second).padStart(2, '0')}`
}

defineExpose({ play, pause, seek, getCurrentTime, getDuration })
</script>
