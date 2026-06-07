<template>
  <main class="page-shell library-page">
    <section class="library-head">
      <div>
        <el-tag effect="light">{{ activeMeta.tag }}</el-tag>
        <h1>{{ activeMeta.title }}</h1>
        <p>{{ activeMeta.description }}</p>
      </div>
      <el-button :icon="Delete" :disabled="!records.length" @click="clearRecords">清空</el-button>
    </section>

    <section v-if="records.length" class="library-list">
      <article v-for="record in records" :key="record.video.id" class="library-item soft-panel">
        <button class="cover-button" type="button" @click="router.push(`/video/${record.video.id}`)">
          <img v-if="record.video.coverUrl" :src="mediaUrl(record.video.coverUrl)" :alt="record.video.title" />
          <span v-else>Danmaku</span>
        </button>
        <div class="item-main">
          <div>
            <h2>{{ record.video.title }}</h2>
            <p>{{ record.video.description || '这个视频暂无简介。' }}</p>
          </div>
          <div class="meta-row">
            <span><el-icon><User /></el-icon>{{ record.video.author?.nickname || '匿名用户' }}</span>
            <span><el-icon><VideoPlay /></el-icon>{{ formatCount(record.video.viewCount) }}</span>
            <span><el-icon><Star /></el-icon>{{ formatCount(record.video.likeCount) }}</span>
            <span>{{ formatTime(record.savedAt) }}</span>
          </div>
          <div v-if="kind === 'history'" class="progress-line">
            <el-progress :percentage="record.progress || 0" :show-text="false" />
          </div>
        </div>
        <div class="item-actions">
          <el-button type="primary" @click="router.push(`/video/${record.video.id}`)">播放</el-button>
          <el-button v-if="kind !== 'downloads'" :icon="Download" @click="downloadVideo(record.video)">下载</el-button>
          <el-button :icon="Close" text @click="removeRecord(record.video.id)" />
        </div>
      </article>
    </section>

    <section v-else class="soft-panel empty-panel">
      <el-empty :description="activeMeta.empty" />
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Close, Delete, Download, Star, User, VideoPlay } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { VideoInfo } from '@/types'
import { formatCount, formatTime, mediaUrl } from '@/utils/format'
import {
  clearUserLibraryRecords,
  getUserLibraryRecords,
  removeUserLibraryRecord,
  upsertUserLibraryRecord,
  type UserLibraryKind,
  type UserLibraryRecord,
} from '@/utils/userLibrary'

const route = useRoute()
const router = useRouter()
const records = ref<UserLibraryRecord[]>([])

const kind = computed<UserLibraryKind>(() => {
  const value = route.params.kind
  return value === 'liked' || value === 'collections' || value === 'downloads' ? value : 'history'
})

const activeMeta = computed(() => {
  const map = {
    history: {
      tag: '观看记录',
      title: '历史记录',
      description: '继续观看最近打开过的视频，进度会保存在本机浏览器。',
      empty: '暂无历史记录',
    },
    liked: {
      tag: '互动记录',
      title: '赞过的视频',
      description: '这里会收纳你在本机点赞过的视频，方便回看。',
      empty: '暂无赞过的视频',
    },
    collections: {
      tag: '收藏夹',
      title: '收藏内容',
      description: '这里会收纳你在本机收藏过的视频，方便集中查看。',
      empty: '暂无收藏内容',
    },
    downloads: {
      tag: '离线内容',
      title: '下载内容',
      description: '管理从视频详情页下载过的内容。',
      empty: '暂无下载内容',
    },
  }
  return map[kind.value]
})

watch(kind, loadRecords, { immediate: true })

function loadRecords() {
  records.value = getUserLibraryRecords(kind.value)
}

function removeRecord(videoId: number) {
  removeUserLibraryRecord(kind.value, videoId)
  loadRecords()
}

async function clearRecords() {
  await ElMessageBox.confirm(`确定清空${activeMeta.value.title}吗？`, '清空记录', {
    type: 'warning',
    confirmButtonText: '清空',
    cancelButtonText: '取消',
  })
  clearUserLibraryRecords(kind.value)
  loadRecords()
}

function downloadVideo(video: VideoInfo) {
  if (!video.videoUrl) {
    ElMessage.warning('当前视频没有可下载地址')
    return
  }
  upsertUserLibraryRecord('downloads', video)
  loadRecords()
  const link = document.createElement('a')
  link.href = mediaUrl(video.videoUrl)
  link.download = `${video.title || 'danmaku-video'}.mp4`
  link.target = '_blank'
  link.rel = 'noreferrer'
  link.click()
}
</script>

<style scoped>
.library-page {
  display: grid;
  gap: 20px;
  padding-top: 22px;
}

.library-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18px;
}

.library-head h1 {
  margin: 10px 0 8px;
  color: #18191c;
  font-size: 32px;
  font-weight: 900;
  line-height: 1.2;
}

.library-head p,
.item-main p {
  margin: 0;
  color: #667085;
  line-height: 1.7;
}

.library-list {
  display: grid;
  gap: 14px;
}

.library-item {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr) auto;
  gap: 18px;
  align-items: center;
  padding: 14px;
}

.cover-button {
  display: block;
  overflow: hidden;
  width: 100%;
  aspect-ratio: 16 / 9;
  border: 0;
  border-radius: 8px;
  background: #f1f2f3;
  color: #00aeec;
  cursor: pointer;
  font-weight: 900;
}

.cover-button img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-button span {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
}

.item-main {
  display: grid;
  min-width: 0;
  gap: 12px;
}

.item-main h2 {
  display: block;
  margin: 0 0 6px;
  overflow: hidden;
  color: #18191c;
  font-size: 18px;
  font-weight: 800;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-main p {
  display: -webkit-box;
  overflow: hidden;
  font-size: 13px;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.meta-row {
  display: flex;
  flex-wrap: nowrap;
  gap: 12px;
  overflow: hidden;
  color: #9499a0;
  font-size: 12px;
}

.meta-row span {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  gap: 4px;
}

.progress-line {
  max-width: 360px;
}

.item-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

@media (max-width: 920px) {
  .library-head,
  .library-item {
    align-items: stretch;
    grid-template-columns: 1fr;
  }

  .library-head {
    display: grid;
  }

  .item-actions {
    flex-direction: row;
    flex-wrap: wrap;
  }
}
</style>
