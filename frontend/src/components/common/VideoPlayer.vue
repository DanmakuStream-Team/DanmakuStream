<template>
  <div class="video-player">
    <a-spin :loading="playerLoading" class="player-spin">
      <div v-if="playerError" class="player-error">
        <div class="error-title">视频加载失败</div>
        <div class="error-desc">{{ playerError }}</div>
        <a-button type="primary" @click="retry">
          重新加载
        </a-button>
      </div>

      <div
        v-show="!playerError"
        ref="playerContainer"
        class="player-container"
      />

      <DanmakuLayer
        v-if="!playerError"
        :danmakus="danmakus"
        :current-time="currentTime"
        :enabled="danmakuEnabled"
      />
    </a-spin>

    <div class="player-toolbar">
      <a-switch v-model="danmakuEnabled">
        <template #checked>弹幕开</template>
        <template #unchecked>弹幕关</template>
      </a-switch>

      <span class="time-text">
        {{ formatDuration(currentTime) }} / {{ formatDuration(duration) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import Player from 'xgplayer'
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

function play() {
  player.value?.play()
}

function pause() {
  player.value?.pause()
}

function seek(time: number) {
  if (!player.value) return

  player.value.currentTime = time
  currentTime.value = time
  emit('seek', time)
}

function getCurrentTime() {
  return player.value?.currentTime || 0
}

function getDuration() {
  return player.value?.duration || duration.value || 0
}

function formatDuration(seconds: number) {
  if (!Number.isFinite(seconds) || seconds < 0) {
    return '00:00'
  }

  const total = Math.floor(seconds)
  const minute = Math.floor(total / 60)
  const second = total % 60

  return `${String(minute).padStart(2, '0')}:${String(second).padStart(2, '0')}`
}

defineExpose({
  play,
  pause,
  seek,
  getCurrentTime,
  getDuration,
})
</script>

<style scoped>
.video-player {
  position: relative;
  width: 100%;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
  aspect-ratio: 16 / 9;
}

.player-spin {
  width: 100%;
  height: 100%;
}

.player-container {
  width: 100%;
  height: 100%;
  background: #000;
}

.player-error {
  width: 100%;
  height: 100%;
  min-height: 360px;
  color: #fff;
  background: #111;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.error-title {
  font-size: 20px;
  font-weight: 600;
}

.error-desc {
  color: #c9cdd4;
  font-size: 14px;
}

.player-toolbar {
  position: absolute;
  left: 12px;
  right: 12px;
  bottom: 12px;
  z-index: 30;
  display: flex;
  align-items: center;
  gap: 12px;
  pointer-events: none;
}

.player-toolbar :deep(.arco-switch) {
  pointer-events: auto;
}

.time-text {
  color: #fff;
  font-size: 13px;
  text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
}
</style>