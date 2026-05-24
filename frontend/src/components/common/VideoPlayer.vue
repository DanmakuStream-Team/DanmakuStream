<template>
  <div class="player">
    <video
      ref="videoRef"
      controls
      playsinline
      :src="mediaUrl(url)"
      :poster="mediaUrl(poster)"
      @timeupdate="emitTime"
      @error="$emit('error', '视频加载失败，请确认资源已转码完成')"
    />
    <DanmakuLayer :items="danmakus" :current-time="currentTime" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { Danmaku } from '@/types'
import { mediaUrl } from '@/utils/format'
import DanmakuLayer from './DanmakuLayer.vue'

defineProps<{ url: string; poster?: string; danmakus: Danmaku[] }>()
const emit = defineEmits<{ timeupdate: [time: number]; error: [message: string] }>()

const videoRef = ref<HTMLVideoElement>()
const currentTime = ref(0)

function emitTime() {
  currentTime.value = videoRef.value?.currentTime || 0
  emit('timeupdate', currentTime.value)
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
