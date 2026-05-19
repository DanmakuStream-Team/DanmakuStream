<template>
  <div class="max-w-4xl mx-auto">
    <!-- Banner -->
    <div class="h-36 sm:h-48 rounded-2xl bg-gradient-to-r from-rose-400 via-pink-400 to-purple-400" />

    <!-- Profile header -->
    <div class="px-6 -mt-16 relative z-10 flex flex-col sm:flex-row sm:items-end gap-5">
      <el-avatar :size="120" :src="profile?.avatar" class="ring-4 ring-white shadow-lg rounded-full shrink-0">
        <span class="text-4xl font-bold text-pink-500">{{ initial }}</span>
      </el-avatar>

      <div class="flex-1 min-w-0 pb-1">
        <h1 class="text-2xl font-bold text-gray-900 truncate">{{ profile?.nickname || '未知用户' }}</h1>
        <p class="text-sm text-gray-500">@{{ profile?.username }}</p>
        <div class="flex gap-5 mt-2 text-sm">
          <span><strong class="text-gray-900">{{ formatCount(profile?.followCount) }}</strong> 关注</span>
          <span><strong class="text-gray-900">{{ formatCount(profile?.fanCount) }}</strong> 粉丝</span>
        </div>
      </div>

      <div class="shrink-0 self-start sm:self-end">
        <template v-if="isOwner">
          <el-button round>编辑资料</el-button>
        </template>
        <template v-else-if="authStore.isLoggedIn">
          <el-button v-if="!following" type="primary" round @click="toggleFollow">关注</el-button>
          <el-button v-else round @click="toggleFollow">已关注</el-button>
        </template>
      </div>
    </div>

    <!-- Bio -->
    <p class="px-6 mt-4 text-gray-700 leading-relaxed">
      {{ profile?.bio || '这个人很懒，什么都没写' }}
    </p>

    <!-- Divider -->
    <div class="mt-6 border-b border-gray-200" />

    <!-- Content tabs -->
    <div class="px-6 py-6">
      <div v-loading="loading">
        <el-empty v-if="!loading && currentList.length === 0" :description="emptyText" class="py-16" />
        <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-5">
          <VideoCard
            v-for="video in currentList"
            :key="video.id"
            :video="video"
            @click="$router.push(`/video/${video.id}`)"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { userApi } from '@/api/user'
import VideoCard from '@/components/common/VideoCard.vue'
import type { UserInfo, VideoInfo } from '@/types'

const route = useRoute()
const authStore = useAuthStore()

const profile = ref<UserInfo | null>(null)
const loading = ref(false)
const following = ref(false)
const collectedVideos = ref<VideoInfo[]>([])
// TODO: 等后端 likes/collections API 就绪后添加更多 tab 内容

const userId = computed(() => Number(route.params.id))
const isOwner = computed(() => authStore.isLoggedIn && authStore.userInfo?.id === userId.value)
const initial = computed(() => (profile.value?.nickname || '').slice(0, 1).toUpperCase())

// 公开页统一展示收藏视频列表
const currentList = computed(() => collectedVideos.value)

const emptyText = computed(() =>
  isOwner.value ? '你还没有收藏任何视频' : 'TA 还没有收藏任何视频'
)

onMounted(() => fetchProfile())
watch(() => route.params.id, fetchProfile)

async function fetchProfile() {
  const id = userId.value
  if (!id) return

  if (isOwner.value && authStore.userInfo) {
    profile.value = authStore.userInfo
  } else {
    loading.value = true
    try {
      const res = await userApi.getProfile(id)
      profile.value = res.data
    } catch {
      profile.value = null
    } finally {
      loading.value = false
    }
  }

  fetchCollections()
}

async function fetchCollections() {
  loading.value = true
  try {
    // TODO: 替换为 GET /users/:id/collections 或其他后端接口
    collectedVideos.value = []
  } catch {
    collectedVideos.value = []
  } finally {
    loading.value = false
  }
}

function toggleFollow() {
  following.value = !following.value
}

function formatCount(count?: number) {
  if (!count) return '0'
  if (count >= 10000) return `${(count / 10000).toFixed(1)}万`
  return String(count)
}
</script>
