<template>
  <article class="video-card" @click="$emit('open')">
    <div class="cover">
      <img v-if="video.coverUrl" :src="mediaUrl(video.coverUrl)" :alt="video.title" />
      <div v-else class="cover-fallback">
        <span>Danmaku</span>
      </div>
      <div class="cover-mask">
        <span><el-icon><VideoPlay /></el-icon>{{ formatCount(video.viewCount) }}</span>
        <span><el-icon><ChatDotRound /></el-icon>{{ formatCount(video.danmakuCount) }}</span>
        <strong>{{ formatDuration(video.duration) }}</strong>
      </div>
    </div>
    <div class="body">
      <h3>{{ video.title }}</h3>
      <button class="author" type="button" :disabled="!video.author?.id" @click.stop="openAuthor">
        <el-icon><User /></el-icon>
        <span>{{ video.author?.nickname || '匿名用户' }}</span>
        <em>{{ formatTime(video.createdAt) }}</em>
      </button>
    </div>
  </article>
</template>

<script setup lang="ts">
import { ChatDotRound, User, VideoPlay } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import type { VideoInfo } from '@/types'
import { formatCount, formatDuration, formatTime, mediaUrl } from '@/utils/format'

const props = defineProps<{ video: VideoInfo }>()
defineEmits<{ open: [] }>()
const router = useRouter()

function openAuthor() {
  if (!props.video.author?.id) return
  router.push(`/user/${props.video.author.id}`)
}
</script>

<style scoped>
.video-card {
  min-width: 0;
  cursor: pointer;
}

.cover {
  position: relative;
  overflow: hidden;
  aspect-ratio: 16 / 9;
  border-radius: 8px;
  background: #f1f2f3;
}

.cover img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
  transition: transform 0.25s ease;
}

.video-card:hover .cover img {
  transform: scale(1.04);
}

.cover-fallback {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  background:
    linear-gradient(135deg, rgba(0, 174, 236, 0.16), rgba(251, 114, 153, 0.16)),
    #f6f7f8;
  color: #00aeec;
  font-size: 16px;
  font-weight: 900;
}

.cover-mask {
  position: absolute;
  inset: auto 0 0;
  display: grid;
  grid-template-columns: auto auto 1fr;
  align-items: center;
  gap: 8px;
  padding: 22px 8px 7px;
  background: linear-gradient(180deg, transparent, rgba(0, 0, 0, 0.72));
  color: #fff;
  font-size: 11px;
}

.cover-mask span {
  display: inline-flex;
  align-items: center;
  gap: 3px;
}

.cover-mask strong {
  justify-self: end;
  font-weight: 600;
}

.body {
  padding: 9px 2px 0;
}

h3 {
  display: block;
  min-height: 22px;
  margin: 0;
  overflow: hidden;
  color: #18191c;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-card:hover h3 {
  color: #00aeec;
}

.author {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-top: 7px;
  padding: 0;
  border: 0;
  background: transparent;
  color: #9499a0;
  cursor: pointer;
  font-size: 12px;
}

.author:hover {
  color: #00aeec;
}

.author:disabled {
  cursor: default;
}

.author:disabled:hover {
  color: #9499a0;
}

.author span {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.author em {
  flex-shrink: 0;
  font-style: normal;
}
</style>
