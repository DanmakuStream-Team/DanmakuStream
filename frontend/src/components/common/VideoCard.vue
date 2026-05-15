<template>
  <div
    class="group bg-white rounded-xl overflow-hidden shadow-sm hover:shadow-md transition-shadow cursor-pointer"
    @click="$emit('click')"
  >
    <div class="relative">
      <img :src="video.coverUrl" :alt="video.title" class="w-full aspect-video object-cover" />
      <span class="absolute bottom-2 right-2 bg-black/70 text-white text-xs px-1.5 py-0.5 rounded">
        {{ formatDuration(video.duration) }}
      </span>
    </div>
    <div class="p-3">
      <div class="text-sm font-semibold text-gray-900 line-clamp-2 leading-snug mb-2">
        {{ video.title }}
      </div>
      <div class="flex items-center justify-between text-xs text-gray-500">
        <div class="flex items-center gap-1.5 min-w-0">
          <el-avatar :size="18" :src="video.author.avatar">
            {{ video.author.nickname?.slice(0, 1) }}
          </el-avatar>
          <span class="truncate">{{ video.author.nickname }}</span>
        </div>
        <span class="flex-shrink-0">{{ formatCount(video.viewCount) }} 播放</span>
      </div>
    </div>
  </div>
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
