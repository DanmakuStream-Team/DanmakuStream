<template>
  <main class="page-shell admin-list-page">
    <div class="section-head">
      <h1>视频审核</h1>
      <el-input v-model="keyword" class="search" placeholder="搜索标题或简介" clearable @keyup.enter="load" />
    </div>

    <section v-loading="loading" class="soft-panel list-panel">
      <div class="toolbar">
        <el-segmented v-model="status" :options="statusOptions" @change="load" />
      </div>

      <div v-if="videos.length" class="rows">
        <div v-for="video in videos" :key="video.id" class="row">
          <button class="thumb-button" type="button" @click="previewVideo(video)">
            <img v-if="video.coverUrl" :src="mediaUrl(video.coverUrl)" :alt="video.title">
            <span v-else>D</span>
          </button>

          <div class="video-main">
            <strong>{{ video.title }}</strong>
            <span>{{ video.author?.nickname || '匿名用户' }} · {{ formatCount(video.viewCount) }} 播放</span>
            <p>{{ video.description || '暂无简介' }}</p>
          </div>

          <el-tag :type="statusType(video.status)">{{ statusText(video.status) }}</el-tag>

          <div class="row-actions">
            <el-button @click="previewVideo(video)">查看视频</el-button>
            <el-select v-model="video.status" size="small" @change="updateStatus(video.id, video.status)">
              <el-option label="待审核" value="pending" />
              <el-option label="通过" value="approved" />
              <el-option label="拒绝" value="rejected" />
            </el-select>
          </div>
        </div>
      </div>

      <el-empty v-else description="暂无视频" />
    </section>

    <el-dialog v-model="previewVisible" :title="previewVideoInfo?.title || '视频预览'" width="860px" destroy-on-close>
      <VideoPlayer
        v-if="previewVideoInfo"
        :url="previewVideoInfo.videoUrl"
        :poster="previewVideoInfo.coverUrl"
        :danmakus="[]"
      />
      <template #footer>
        <div class="dialog-footer">
          <el-tag v-if="previewVideoInfo" :type="statusType(previewVideoInfo.status)">
            {{ statusText(previewVideoInfo.status) }}
          </el-tag>
          <el-button @click="previewVisible = false">关闭</el-button>
          <el-button
            v-if="previewVideoInfo"
            type="success"
            @click="updateStatus(previewVideoInfo.id, 'approved')"
          >
            通过
          </el-button>
          <el-button
            v-if="previewVideoInfo"
            type="danger"
            @click="updateStatus(previewVideoInfo.id, 'rejected')"
          >
            拒绝
          </el-button>
        </div>
      </template>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import VideoPlayer from '@/components/common/VideoPlayer.vue'
import { videoApi } from '@/api/video'
import type { VideoInfo, VideoStatus } from '@/types'
import { formatCount, mediaUrl } from '@/utils/format'

const loading = ref(false)
const videos = ref<VideoInfo[]>([])
const keyword = ref('')
const status = ref('')
const previewVisible = ref(false)
const previewVideoInfo = ref<VideoInfo | null>(null)
const statusOptions = [
  { label: '全部', value: '' },
  { label: '待审核', value: 'pending' },
  { label: '已通过', value: 'approved' },
  { label: '已拒绝', value: 'rejected' },
]

onMounted(load)

async function load() {
  loading.value = true
  try {
    const res = await videoApi.adminList({
      page: 1,
      pageSize: 50,
      status: status.value as VideoStatus | '',
      keyword: keyword.value.trim() || undefined,
    })
    videos.value = res.data.list
  } finally {
    loading.value = false
  }
}

function previewVideo(video: VideoInfo) {
  previewVideoInfo.value = video
  previewVisible.value = true
}

async function updateStatus(id: number, nextStatus: VideoStatus) {
  await videoApi.adminUpdateStatus(id, nextStatus)
  const target = videos.value.find((item) => item.id === id)
  if (target) target.status = nextStatus
  if (previewVideoInfo.value?.id === id) previewVideoInfo.value.status = nextStatus
  ElMessage.success('审核状态已更新')
}

function statusText(value: VideoStatus) {
  return ({ pending: '待审核', approved: '已通过', rejected: '已拒绝' } as const)[value]
}

function statusType(value: VideoStatus) {
  return ({ pending: 'warning', approved: 'success', rejected: 'danger' } as const)[value]
}
</script>

<style scoped>
.admin-list-page {
  display: grid;
  gap: 18px;
}

.search {
  width: 280px;
}

.list-panel {
  padding: 18px;
}

.toolbar {
  margin-bottom: 16px;
}

.rows {
  display: grid;
  gap: 12px;
}

.row {
  display: grid;
  grid-template-columns: 132px minmax(0, 1fr) auto 220px;
  align-items: center;
  gap: 14px;
  padding: 12px;
  border-radius: 8px;
  background: #f7f9fc;
}

.thumb-button {
  display: block;
  overflow: hidden;
  width: 132px;
  aspect-ratio: 16 / 9;
  border: 0;
  border-radius: 8px;
  background: #e8edf5;
  color: #165dff;
  cursor: pointer;
  font-weight: 800;
}

.thumb-button img,
.thumb-button span {
  display: block;
  width: 100%;
  height: 100%;
}

.thumb-button img {
  object-fit: cover;
}

.thumb-button span {
  display: grid;
  place-items: center;
}

.video-main {
  display: grid;
  min-width: 0;
  gap: 6px;
}

.video-main strong,
.video-main span,
.video-main p {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-main span,
.video-main p {
  margin: 0;
  color: #667085;
  font-size: 13px;
}

.row-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

.row-actions .el-select {
  width: 110px;
}

.dialog-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 920px) {
  .search {
    width: 100%;
  }

  .row {
    grid-template-columns: 100px 1fr;
  }

  .thumb-button {
    width: 100px;
  }

  .row-actions {
    grid-column: 1 / -1;
    justify-content: flex-start;
  }
}
</style>
