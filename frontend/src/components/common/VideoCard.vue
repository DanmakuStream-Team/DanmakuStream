<template>
  <article class="video-card" @click="$emit('open')">
    <div class="cover">
      <img
        v-if="video.coverUrl"
        :src="mediaUrl(video.coverUrl)"
        :alt="video.title"
        loading="lazy"
        decoding="async"
      />
      <div v-else class="cover-fallback">
        <span>Danmaku</span>
      </div>
      <el-tooltip content="稍后再看" placement="left">
        <button class="watch-later" type="button" aria-label="稍后再看" @click.stop="$emit('watch-later')">
          <el-icon><Clock /></el-icon>
        </button>
      </el-tooltip>
      <div class="cover-mask">
        <span><el-icon><VideoPlay /></el-icon>{{ formatCount(video.viewCount) }}</span>
        <span><el-icon><ChatDotRound /></el-icon>{{ formatCount(video.danmakuCount) }}</span>
        <strong>{{ formatDuration(video.duration) }}</strong>
      </div>
    </div>
    <div class="body">
      <div class="title-row">
        <h3>{{ video.title }}</h3>
        <el-dropdown trigger="click" @command="handleCommand">
          <button class="more-action" type="button" title="更多操作" aria-label="更多操作" @click.stop>
            <el-icon><MoreFilled /></el-icon>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="queue">
                <el-icon><VideoPlay /></el-icon>
                加入播放队列
              </el-dropdown-item>
              <el-dropdown-item command="watch-later">
                <el-icon><Clock /></el-icon>
                稍后再看
              </el-dropdown-item>
              <el-dropdown-item v-if="showFeedback" command="not-interested" divided>
                <el-icon><Close /></el-icon>
                不感兴趣
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      <button class="author" type="button" :disabled="!video.author?.id" @click.stop="openAuthor">
        <el-icon><User /></el-icon>
        <span>{{ video.author?.nickname || '匿名用户' }}</span>
        <em>{{ formatTime(video.createdAt) }}</em>
      </button>
    </div>
  </article>
</template>

<script setup lang="ts">
import { ChatDotRound, Clock, Close, MoreFilled, User, VideoPlay } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import type { VideoInfo } from '@/types'
import { formatCount, formatDuration, formatTime, mediaUrl } from '@/utils/format'

const props = withDefaults(defineProps<{ video: VideoInfo; showFeedback?: boolean }>(), {
  showFeedback: false,
})
const emit = defineEmits<{ open: []; 'watch-later': []; queue: []; 'not-interested': [] }>()
const router = useRouter()

function openAuthor() {
  if (!props.video.author?.id) return
  router.push(`/user/${props.video.author.id}`)
}

function handleCommand(command: string | number | object) {
  if (command === 'watch-later') emit('watch-later')
  if (command === 'queue') emit('queue')
  if (command === 'not-interested') emit('not-interested')
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

.watch-later {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 2;
  display: inline-grid;
  width: 32px;
  height: 32px;
  padding: 0;
  place-items: center;
  border: 0;
  border-radius: 6px;
  background: rgb(0 0 0 / 68%);
  color: #fff;
  cursor: pointer;
  font-size: 17px;
  opacity: 0;
  transform: translateY(-3px);
  transition: opacity 0.16s ease, transform 0.16s ease, background 0.16s ease;
}

.video-card:hover .watch-later,
.watch-later:focus-visible {
  opacity: 1;
  transform: translateY(0);
}

.watch-later:hover {
  background: rgb(0 0 0 / 84%);
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

.title-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 30px;
  align-items: start;
  gap: 4px;
}

h3 {
  display: -webkit-box;
  min-height: 41px;
  margin: 0;
  overflow: hidden;
  color: #18191c;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
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

.more-action {
  display: grid;
  width: 30px;
  height: 30px;
  margin-top: -4px;
  padding: 0;
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: transparent;
  color: #61666d;
  cursor: pointer;
  font-size: 18px;
  opacity: 0;
}

.video-card:hover .more-action,
.more-action:focus-visible,
.more-action[aria-expanded='true'] {
  opacity: 1;
}

.more-action:hover {
  background: #f1f2f3;
  color: #18191c;
}

@media (hover: none) {
  .watch-later,
  .more-action {
    opacity: 1;
    transform: none;
  }
}
</style>
