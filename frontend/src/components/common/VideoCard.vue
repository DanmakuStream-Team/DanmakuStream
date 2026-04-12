<template>
  <a-card hoverable class="video-card" @click="$emit('click')">
    <template #cover>
      <div class="cover-wrapper">
        <img :src="video.coverUrl" :alt="video.title" class="cover-img" />
        <span class="duration">{{ formatDuration(video.duration) }}</span>
      </div>
    </template>
    <a-card-meta :title="video.title">
      <template #description>
        <div class="meta">
          <span class="author">
            <a-avatar :size="18" :image-url="video.author.avatar" />
            {{ video.author.nickname }}
          </span>
          <span class="views">{{ formatCount(video.viewCount) }} 次播放</span>
        </div>
      </template>
    </a-card-meta>
  </a-card>
</template>

<script setup lang="ts">
import type { VideoInfo } from '@/types'

defineProps<{ video: VideoInfo }>()
defineEmits(['click'])

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

function formatCount(count: number): string {
  if (count >= 10000) return `${(count / 10000).toFixed(1)}万`
  return String(count)
}
</script>

<style scoped>
.video-card { cursor: pointer; }
.cover-wrapper { position: relative; }
.cover-img { width: 100%; aspect-ratio: 16/9; object-fit: cover; }
.duration {
  position: absolute;
  bottom: 6px;
  right: 6px;
  background: rgba(0,0,0,.7);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}
.meta { display: flex; justify-content: space-between; align-items: center; margin-top: 4px; }
.author { display: flex; align-items: center; gap: 4px; color: #86909c; font-size: 13px; }
.views { color: #86909c; font-size: 12px; }
</style>
