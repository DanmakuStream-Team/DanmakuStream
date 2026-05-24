<template>
  <article class="video-card" @click="$emit('open')">
    <div class="cover">
      <img v-if="video.coverUrl" :src="mediaUrl(video.coverUrl)" :alt="video.title" />
      <div v-else class="cover-fallback">Danmaku</div>
      <span class="duration">{{ formatDuration(video.duration) }}</span>
    </div>
    <div class="body">
      <h3>{{ video.title }}</h3>
      <div class="meta">
        <span>{{ formatCount(video.viewCount) }} 播放</span>
        <span>{{ formatCount(video.danmakuCount) }} 弹幕</span>
      </div>
      <div class="author">
        <el-avatar :size="24" :src="mediaUrl(video.author?.avatar)">
          {{ video.author?.nickname?.slice(0, 1) || 'U' }}
        </el-avatar>
        <span>{{ video.author?.nickname || '匿名用户' }}</span>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import type { VideoInfo } from '@/types'
import { formatCount, formatDuration, mediaUrl } from '@/utils/format'

defineProps<{ video: VideoInfo }>()
defineEmits<{ open: [] }>()
</script>

<style scoped>
.video-card {
  overflow: hidden;
  border: 1px solid rgba(20, 32, 51, 0.08);
  border-radius: 10px;
  background: #fff;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}

.video-card:hover {
  transform: translateY(-3px);
  border-color: rgba(22, 93, 255, 0.22);
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.1);
}

.cover {
  position: relative;
  overflow: hidden;
  aspect-ratio: 16 / 9;
  background: #e8edf5;
}

.cover img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
}

.cover-fallback {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  color: #165dff;
  font-weight: 800;
}

.duration {
  position: absolute;
  right: 8px;
  bottom: 8px;
  padding: 3px 7px;
  border-radius: 6px;
  background: rgba(10, 16, 28, 0.78);
  color: #fff;
  font-size: 12px;
}

.body {
  display: grid;
  gap: 10px;
  padding: 14px;
}

h3 {
  min-height: 44px;
  margin: 0;
  color: #142033;
  font-size: 15px;
  line-height: 1.45;
}

.meta {
  display: flex;
  gap: 10px;
  color: #667085;
  font-size: 12px;
}

.author {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #475467;
  font-size: 13px;
}
</style>
