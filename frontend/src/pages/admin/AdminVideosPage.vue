<template>
  <main class="page-shell admin-list-page">
    <div class="section-head">
      <h1>视频审核</h1>
      <el-input v-model="keyword" class="search" placeholder="搜索标题或简介" clearable @keyup.enter="load" />
    </div>

    <section class="soft-panel list-panel" v-loading="loading">
      <div class="toolbar">
        <el-segmented v-model="status" :options="statusOptions" @change="load" />
      </div>
      <div v-if="videos.length" class="rows">
        <div v-for="video in videos" :key="video.id" class="row">
          <img v-if="video.coverUrl" :src="mediaUrl(video.coverUrl)" :alt="video.title" />
          <div v-else class="thumb">D</div>
          <div>
            <strong>{{ video.title }}</strong>
            <span>{{ video.author?.nickname || '匿名用户' }} · {{ formatCount(video.viewCount) }} 播放</span>
          </div>
          <el-tag :type="statusType(video.status)">{{ statusText(video.status) }}</el-tag>
          <el-select v-model="video.status" size="small" @change="updateStatus(video.id, video.status)">
            <el-option label="待审核" value="pending" />
            <el-option label="通过" value="approved" />
            <el-option label="拒绝" value="rejected" />
          </el-select>
        </div>
      </div>
      <el-empty v-else description="暂无视频" />
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { videoApi } from '@/api/video'
import type { VideoInfo, VideoStatus } from '@/types'
import { formatCount, mediaUrl } from '@/utils/format'

const loading = ref(false)
const videos = ref<VideoInfo[]>([])
const keyword = ref('')
const status = ref('')
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

async function updateStatus(id: number, nextStatus: VideoStatus) {
  await videoApi.adminUpdateStatus(id, nextStatus)
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
  grid-template-columns: 110px minmax(0, 1fr) auto 130px;
  align-items: center;
  gap: 14px;
  padding: 12px;
  border-radius: 8px;
  background: #f7f9fc;
}

.row img,
.thumb {
  width: 110px;
  height: 62px;
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

.row strong,
.row span {
  display: block;
}

.row span {
  margin-top: 6px;
  color: #667085;
  font-size: 13px;
}

@media (max-width: 820px) {
  .search {
    width: 100%;
  }

  .row {
    grid-template-columns: 90px 1fr;
  }
}
</style>
