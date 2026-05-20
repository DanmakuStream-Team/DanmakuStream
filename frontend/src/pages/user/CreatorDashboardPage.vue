<template>
  <div class="max-w-6xl mx-auto">
    <!-- Header -->
    <div class="flex items-center justify-between mb-8">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">创作者中心</h1>
        <p class="text-gray-500 mt-1">管理你的视频内容</p>
      </div>
      <el-button type="primary" size="large" @click="$router.push('/creator/upload')">
        <el-icon class="mr-1"><Upload /></el-icon>
        上传视频
      </el-button>
    </div>

    <!-- Stats cards -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-8">
      <div class="bg-white rounded-xl p-5 shadow-sm">
        <div class="text-3xl font-bold text-gray-900">{{ myVideos.length }}</div>
        <div class="text-sm text-gray-500 mt-1">已投稿</div>
      </div>
      <div class="bg-white rounded-xl p-5 shadow-sm">
        <div class="text-3xl font-bold text-gray-900">{{ formatCount(totalViews) }}</div>
        <div class="text-sm text-gray-500 mt-1">总播放量</div>
      </div>
      <div class="bg-white rounded-xl p-5 shadow-sm">
        <div class="text-3xl font-bold text-gray-900">{{ formatCount(totalLikes) }}</div>
        <div class="text-sm text-gray-500 mt-1">总点赞</div>
      </div>
      <div class="bg-white rounded-xl p-5 shadow-sm">
        <div class="text-3xl font-bold text-gray-900">{{ formatCount(totalDanmaku) }}</div>
        <div class="text-sm text-gray-500 mt-1">总弹幕</div>
      </div>
    </div>

    <!-- My videos -->
    <div class="bg-white rounded-xl shadow-sm p-6">
      <h2 class="text-lg font-semibold text-gray-900 mb-5">我的投稿</h2>
      <div v-loading="loading">
        <div v-if="myVideos.length > 0" class="grid grid-cols-1 gap-4">
          <div
            v-for="video in myVideos"
            :key="video.id"
            class="flex gap-4 p-3 rounded-xl hover:bg-gray-50 transition-colors cursor-pointer"
            @click="$router.push(`/video/${video.id}`)"
          >
            <img
              :src="video.coverUrl"
              class="w-40 h-24 rounded-lg object-cover shrink-0 bg-gray-200"
            />
            <div class="flex-1 min-w-0 flex flex-col justify-between">
              <div>
                <div class="font-semibold text-gray-900 truncate">{{ video.title }}</div>
                <div class="text-sm text-gray-500 mt-1 line-clamp-2">{{ video.description || '暂无简介' }}</div>
              </div>
              <div class="flex gap-4 text-xs text-gray-400">
                <span>
                  <el-icon><View /></el-icon> {{ formatCount(video.viewCount) }}
                </span>
                <span>
                  <el-icon><Star /></el-icon> {{ formatCount(video.likeCount) }}
                </span>
                <span>{{ formatCount(video.danmakuCount) }} 弹幕</span>
                <span>{{ formatCount(video.collectCount) }} 收藏</span>
              </div>
            </div>
          </div>
        </div>
        <el-empty v-else description="还没有投稿，快去上传第一个视频吧！" class="py-12">
          <el-button type="primary" @click="$router.push('/creator/upload')">上传视频</el-button>
        </el-empty>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAuthStore } from '@/store/auth'
import { videoApi } from '@/api/video'
import type { VideoInfo } from '@/types'

const authStore = useAuthStore()
const loading = ref(false)
const myVideos = ref<VideoInfo[]>([])

const totalViews = computed(() => myVideos.value.reduce((s, v) => s + v.viewCount, 0))
const totalLikes = computed(() => myVideos.value.reduce((s, v) => s + v.likeCount, 0))
const totalDanmaku = computed(() => myVideos.value.reduce((s, v) => s + v.danmakuCount, 0))

onMounted(fetchMyVideos)

async function fetchMyVideos() {
  loading.value = true
  try {
    // TODO: 后端加 /videos?authorId= 后直接传参，不用前端过滤
    const res = await videoApi.getVideoList({ page: 1, pageSize: 100 })
    const uid = authStore.userInfo?.id
    myVideos.value = res.data.list.filter(v => v.author.id === uid)
  } catch {
    myVideos.value = []
  } finally {
    loading.value = false
  }
}

function formatCount(n: number) {
  if (n >= 10000) return `${(n / 10000).toFixed(1)}万`
  return String(n)
}
</script>

<style scoped>
/* 顶部渐变标题 */
.creator-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.25);
}

/* 统计卡片 */
.stat-card {
  background: #fff;
  border: 1px solid #f3f4f6;
}

/* 功能卡片 */
.func-card {
  background: #fff;
  border: 1px solid #f3f4f6;
}
.func-card:hover {
  border-color: #667eea;
}
</style>