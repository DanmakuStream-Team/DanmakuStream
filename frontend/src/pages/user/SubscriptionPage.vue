<template>
  <main class="page-shell subscription-page">
    <section class="section-head subscription-head">
      <div>
        <h1>订阅</h1>
        <p>按发布时间查看你关注的博主最近发布的视频。</p>
      </div>
      <el-button :loading="loading" @click="loadSubscriptionVideos">刷新</el-button>
    </section>

    <section v-if="followees.length" class="creator-strip">
      <button
        v-for="creator in followees"
        :key="creator.id"
        class="creator-chip"
        type="button"
        @click="router.push(`/user/${creator.id}`)"
      >
        <el-avatar :size="30" :src="mediaUrl(creator.avatar)">
          {{ creator.nickname.slice(0, 1) }}
        </el-avatar>
        <span>{{ creator.nickname }}</span>
      </button>
    </section>

    <section v-loading="loading" class="subscription-body">
      <div v-if="videos.length" class="video-grid">
        <VideoCard
          v-for="video in videos"
          :key="video.id"
          :video="video"
          @open="router.push(`/video/${video.id}`)"
        />
      </div>

      <div v-else class="soft-panel empty-panel">
        <el-empty :description="emptyText" />
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import VideoCard from '@/components/common/VideoCard.vue'
import { userApi, type FolloweeInfo } from '@/api/user'
import { videoApi } from '@/api/video'
import type { VideoInfo } from '@/types'
import { mediaUrl } from '@/utils/format'

const router = useRouter()
const loading = ref(false)
const followees = ref<FolloweeInfo[]>([])
const videos = ref<VideoInfo[]>([])

const emptyText = computed(() => {
  if (!followees.value.length) return '你还没有订阅任何博主'
  return '你订阅的博主暂时没有发布视频'
})

onMounted(loadSubscriptionVideos)

async function loadCreatorVideos(creatorId: number) {
  const pageSize = 50
  let page = 1
  const result: VideoInfo[] = []

  while (true) {
    const res = await videoApi.userVideos(creatorId, { page, pageSize })
    result.push(...res.data.list)
    if (result.length >= res.data.total || res.data.list.length < pageSize) break
    page += 1
  }

  return result
}

async function loadSubscriptionVideos() {
  loading.value = true
  try {
    const followingRes = await userApi.following()
    followees.value = followingRes.data.list

    const videoGroups = await Promise.all(
      followees.value.map(creator => loadCreatorVideos(creator.id)),
    )

    videos.value = videoGroups
      .flat()
      .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
  } catch {
    followees.value = []
    videos.value = []
    ElMessage.error('订阅内容加载失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.subscription-page {
  display: grid;
  gap: 18px;
  padding-top: 22px;
}

.subscription-head {
  margin-bottom: 0;
}

.subscription-head p {
  margin: 8px 0 0;
  color: #667085;
  line-height: 1.7;
}

.creator-strip {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding: 2px 0 8px;
  scrollbar-color: rgba(251, 114, 153, 0.34) transparent;
  scrollbar-width: thin;
}

.creator-chip {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  gap: 8px;
  height: 40px;
  padding: 0 12px 0 6px;
  border: 1px solid #f1f2f3;
  border-radius: 999px;
  background: #fff;
  color: #18191c;
  cursor: pointer;
}

.creator-chip:hover {
  border-color: rgba(251, 114, 153, 0.38);
  background: rgba(251, 114, 153, 0.08);
  color: #fb7299;
}

.creator-chip span {
  max-width: 120px;
  overflow: hidden;
  font-size: 13px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.subscription-body {
  min-height: 360px;
}

.video-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 18px;
}
</style>
