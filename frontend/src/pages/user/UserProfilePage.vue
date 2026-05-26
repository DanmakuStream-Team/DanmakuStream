<template>
  <main class="page-shell profile-page" v-loading="loading">
    <section v-if="user" class="profile-hero soft-panel">
      <el-avatar :size="78" :src="mediaUrl(user.avatar)">{{ user.nickname?.slice(0, 1) }}</el-avatar>
      <div>
        <el-tag>{{ user.role }}</el-tag>
        <h1>{{ user.nickname }}</h1>
        <p>{{ user.bio || '这个用户还没有填写简介。' }}</p>
        <div class="stats">
          <span>{{ user.followCount }} 关注</span>
          <span>{{ user.fanCount }} 粉丝</span>
          <span>{{ user.videoCount || 0 }} 视频</span>
        </div>
      </div>
      <el-button type="primary" @click="follow">{{ user.followed ? '已关注' : '关注' }}</el-button>
    </section>

    <section>
      <div class="section-head">
        <h2>公开视频</h2>
      </div>
      <div v-if="videos.length" class="video-grid">
        <VideoCard v-for="video in videos" :key="video.id" :video="video" @open="router.push(`/video/${video.id}`)" />
      </div>
      <div v-else class="soft-panel empty-panel">
        <el-empty description="暂无公开视频" />
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import VideoCard from '@/components/common/VideoCard.vue'
import { userApi } from '@/api/user'
import { videoApi } from '@/api/video'
import { useAuthStore } from '@/store/auth'
import type { UserInfo, VideoInfo } from '@/types'
import { mediaUrl } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const user = ref<UserInfo | null>(null)
const videos = ref<VideoInfo[]>([])

onMounted(load)

async function load() {
  const id = Number(route.params.id)
  loading.value = true
  try {
    const [profileRes, videosRes] = await Promise.all([
      userApi.profile(id),
      videoApi.userVideos(id, { page: 1, pageSize: 20 }),
    ])
    user.value = profileRes.data
    videos.value = videosRes.data.list
  } finally {
    loading.value = false
  }
}

async function follow() {
  if (!authStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  if (!user.value) return
  const res = await userApi.follow(user.value.id)
  user.value.followed = res.data.followed
  user.value.fanCount += res.data.followed ? 1 : -1
}
</script>

<style scoped>
.profile-page {
  display: grid;
  gap: 26px;
}

.profile-hero {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 22px;
  padding: 28px;
}

h1 {
  margin: 10px 0 8px;
  color: #142033;
}

p {
  margin: 0 0 12px;
  color: #667085;
}

.stats {
  display: flex;
  gap: 14px;
  color: #475467;
}

.video-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 18px;
}

@media (max-width: 820px) {
  .profile-hero,
  .video-grid {
    grid-template-columns: 1fr;
  }
}
</style>
