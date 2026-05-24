<template>
  <main class="page-shell home-page">
    <section class="hero">
      <div class="hero-copy">
        <el-tag type="primary" effect="light">在线视频与弹幕社区</el-tag>
        <h1>发现视频，发送弹幕，参与创作</h1>
        <p>简洁的视频浏览体验，适配当前后端的视频、评论、弹幕和审核接口。</p>
        <div class="hero-actions">
          <el-button type="primary" size="large" @click="loadVideos">刷新内容</el-button>
          <el-button size="large" @click="router.push('/creator/upload')">上传视频</el-button>
        </div>
      </div>
      <div class="hero-panel soft-panel">
        <strong>{{ totalText }}</strong>
        <span>已通过审核的视频</span>
      </div>
    </section>

    <section>
      <div class="section-head">
        <div>
          <h2>推荐视频</h2>
          <p class="muted">仅展示后端返回的 approved 视频</p>
        </div>
        <el-input v-model="keyword" class="inline-search" placeholder="关键词搜索" clearable @keyup.enter="loadVideos" />
      </div>

      <div v-loading="videoStore.loading">
        <div v-if="videoStore.videoList.length" class="video-grid">
          <VideoCard
            v-for="video in videoStore.videoList"
            :key="video.id"
            :video="video"
            @open="router.push(`/video/${video.id}`)"
          />
        </div>
        <div v-else class="soft-panel empty-panel">
          <el-empty description="暂无已审核视频" />
        </div>
      </div>

      <div class="pager">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="videoStore.total"
          background
          layout="prev, pager, next"
          @current-change="loadVideos"
        />
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import VideoCard from '@/components/common/VideoCard.vue'
import { useVideoStore } from '@/store/video'
import { formatCount } from '@/utils/format'

const router = useRouter()
const route = useRoute()
const videoStore = useVideoStore()
const page = ref(1)
const pageSize = 20
const keyword = ref(String(route.query.keyword || ''))

const totalText = computed(() => formatCount(videoStore.total))

function loadVideos() {
  videoStore.fetchVideoList({
    page: page.value,
    pageSize,
    keyword: keyword.value.trim() || undefined,
  })
}

onMounted(loadVideos)
watch(() => route.query.keyword, (value) => {
  keyword.value = String(value || '')
  page.value = 1
  loadVideos()
})
</script>

<style scoped>
.home-page {
  display: grid;
  gap: 34px;
}

.hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  align-items: center;
  gap: 28px;
  min-height: 320px;
}

.hero-copy {
  display: grid;
  justify-items: start;
  gap: 16px;
}

.hero h1 {
  max-width: 720px;
  margin: 0;
  color: #111827;
  font-size: clamp(36px, 6vw, 58px);
  line-height: 1.08;
}

.hero p {
  max-width: 560px;
  margin: 0;
  color: #667085;
  font-size: 17px;
  line-height: 1.8;
}

.hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.hero-panel {
  display: grid;
  gap: 8px;
  padding: 28px;
}

.hero-panel strong {
  color: #165dff;
  font-size: 52px;
}

.hero-panel span {
  color: #667085;
}

.inline-search {
  width: 260px;
}

.video-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 18px;
}

.pager {
  display: flex;
  justify-content: center;
  margin-top: 28px;
}

@media (max-width: 980px) {
  .hero,
  .video-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 620px) {
  .hero,
  .video-grid {
    grid-template-columns: 1fr;
  }

  .inline-search {
    width: 100%;
  }
}
</style>
