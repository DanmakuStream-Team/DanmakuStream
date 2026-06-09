<template>
  <main class="page-shell creator-page">
    <div class="section-head">
      <div>
        <h1>创作者中心</h1>
        <p class="muted">管理自己上传的视频和审核状态。</p>
      </div>
      <el-button type="primary" @click="router.push('/creator/upload')">上传视频</el-button>
    </div>

    <section v-if="!authStore.isLoggedIn" class="soft-panel empty-panel">
      <el-empty description="登录后查看创作者中心">
        <el-button type="primary" @click="router.push('/login')">去登录</el-button>
      </el-empty>
    </section>

    <section v-else class="soft-panel list-panel" v-loading="loading">
      <div class="toolbar">
        <el-segmented v-model="status" :options="statusOptions" @change="load" />
      </div>
      <div v-if="videos.length" class="video-list">
        <div v-for="video in videos" :key="video.id" class="video-row">
          <img v-if="video.coverUrl" :src="mediaUrl(video.coverUrl)" :alt="video.title" />
          <div v-else class="thumb">D</div>
          <div>
            <strong>{{ video.title }}</strong>
            <span>{{ formatCount(video.viewCount) }} 播放 · {{ formatCount(video.danmakuCount) }} 弹幕</span>
          </div>
          <el-tag :type="statusType(video.status)">{{ statusText(video.status) }}</el-tag>
          <div class="row-actions">
            <el-button @click="router.push(`/video/${video.id}`)">查看</el-button>
            <el-button
              type="danger"
              plain
              :loading="deletingId === video.id"
              @click="deleteVideo(video)"
            >
              删除
            </el-button>
          </div>
        </div>
      </div>
      <el-empty v-else description="暂无视频" />
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { videoApi } from '@/api/video'
import { useAuthStore } from '@/store/auth'
import type { VideoInfo, VideoStatus } from '@/types'
import { formatCount, mediaUrl } from '@/utils/format'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const deletingId = ref<number>()
const videos = ref<VideoInfo[]>([])
const status = ref('')
const statusOptions = [
  { label: '全部', value: '' },
  { label: '待审核', value: 'pending' },
  { label: '已通过', value: 'approved' },
  { label: '已拒绝', value: 'rejected' },
]

onMounted(load)

async function load() {
  if (!authStore.isLoggedIn) return
  loading.value = true
  try {
    const res = await videoApi.myVideos({ page: 1, pageSize: 50, status: status.value as VideoStatus | '' })
    videos.value = res.data.list
  } finally {
    loading.value = false
  }
}

function statusText(value: VideoStatus) {
  return ({ pending: '待审核', approved: '已通过', rejected: '已拒绝' } as const)[value]
}

function statusType(value: VideoStatus) {
  return ({ pending: 'warning', approved: 'success', rejected: 'danger' } as const)[value]
}

async function deleteVideo(video: VideoInfo) {
  try {
    await ElMessageBox.confirm(`确定删除《${video.title}》吗？删除后不可恢复。`, '删除视频', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      confirmButtonClass: 'el-button--danger',
    })
  } catch {
    return
  }

  deletingId.value = video.id
  try {
    await videoApi.remove(video.id)
    ElMessage.success('视频已删除')
    await load()
  } catch (error: any) {
    ElMessage.error(error.message || '删除失败')
  } finally {
    deletingId.value = undefined
  }
}
</script>

<style scoped>
.creator-page {
  display: grid;
  gap: 18px;
}

.section-head p {
  margin: 8px 0 0;
}

.list-panel {
  padding: 18px;
}

.toolbar {
  margin-bottom: 16px;
}

.video-list {
  display: grid;
  gap: 12px;
}

.video-row {
  display: grid;
  grid-template-columns: 104px minmax(0, 1fr) auto auto;
  align-items: center;
  gap: 14px;
  padding: 12px;
  border-radius: 8px;
  background: #f7f9fc;
}

.video-row img,
.thumb {
  width: 104px;
  height: 58px;
  border-radius: 8px;
  object-fit: cover;
}

.thumb {
  display: grid;
  place-items: center;
  background: #e8edf5;
  color: #165dff;
  font-weight: 800;
}

.video-row strong,
.video-row span {
  display: block;
}

.video-row span {
  margin-top: 6px;
  color: #667085;
  font-size: 13px;
}

.row-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 720px) {
  .video-row {
    grid-template-columns: 90px 1fr;
  }

  .row-actions {
    grid-column: 1 / -1;
    justify-content: flex-start;
  }
}
</style>
