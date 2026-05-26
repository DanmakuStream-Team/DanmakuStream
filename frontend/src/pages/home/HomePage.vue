<template>
  <main class="home-page">
    <section class="home-banner">
      <div class="banner-inner">
        <div class="banner-brand">DanmakuStream</div>
        <p>在线视频 · 实时弹幕 · 创作投稿</p>
      </div>
    </section>

    <section class="page-shell feature-wrap">
      <div class="feature-heading">
        <button type="button" @click="resetFilter">功能入口</button>
        <p>围绕视频、弹幕、投稿、审核和直播组织页面</p>
      </div>
      <div class="features">
        <button
          v-for="item in backendFeatures"
          :key="item.key"
          type="button"
          :class="{ active: activeFeature === item.key }"
          @click="selectFeature(item)"
        >
          <strong>{{ item.label }}</strong>
          <span>{{ item.desc }}</span>
        </button>
      </div>
      <div class="quick-links">
        <button type="button" @click="router.push('/creator/upload')">上传视频</button>
        <button type="button" @click="router.push('/creator')">我的投稿</button>
        <button type="button" @click="router.push('/live/1')">进入直播</button>
        <button type="button" @click="router.push('/admin/videos')">视频审核</button>
        <button type="button" @click="router.push('/admin/danmaku')">弹幕管理</button>
      </div>
    </section>

    <section class="page-shell feed-section">
      <div class="feed-head">
        <div>
          <h2>推荐视频</h2>
          <p>{{ loadError || activeFeatureText }}</p>
        </div>
        <div class="feed-tools">
          <el-input
            v-model="keyword"
            class="inline-search"
            placeholder="关键词搜索"
            clearable
            @keyup.enter="loadVideos"
          />
          <el-button type="primary" @click="loadVideos">刷新内容</el-button>
        </div>
      </div>

      <div v-loading="videoStore.loading" class="feed-layout">
        <article v-if="featuredVideo" class="feature-card" @click="router.push(`/video/${featuredVideo.id}`)">
          <div class="feature-cover">
            <img v-if="featuredVideo.coverUrl" :src="mediaUrl(featuredVideo.coverUrl)" :alt="featuredVideo.title" />
            <div v-else class="feature-fallback">DanmakuStream</div>
            <div class="feature-title">
              <h3>{{ featuredVideo.title }}</h3>
              <p>{{ featuredVideo.author?.nickname || '匿名用户' }}</p>
            </div>
          </div>
        </article>

        <div v-if="videoStore.videoList.length" class="video-grid">
          <VideoCard
            v-for="video in gridVideos"
            :key="video.id"
            :video="video"
            @open="router.push(`/video/${video.id}`)"
          />
        </div>

        <div v-if="!videoStore.videoList.length" class="empty-card">
          <el-empty :description="loadError ? '后端暂未连接，或接口返回失败' : '暂无已审核视频'" />
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
import { mediaUrl } from '@/utils/format'

const router = useRouter()
const route = useRoute()
const videoStore = useVideoStore()
const page = ref(1)
const pageSize = 20
const keyword = ref(String(route.query.keyword || ''))
const activeFeature = ref(String(route.query.feature || 'video'))
const loadError = ref('')

const backendFeatures = [
  { key: 'video', label: '视频浏览', desc: '列表 / 搜索 / 详情' },
  { key: 'upload', label: '投稿上传', desc: '视频 / 封面 / 转码' },
  { key: 'comment', label: '评论互动', desc: '评论 / 回复 / 点赞' },
  { key: 'danmaku', label: '弹幕系统', desc: '拉取 / 发送 / 屏蔽' },
  { key: 'live', label: '直播间', desc: '播放 / 实时弹幕 / 互动' },
  { key: 'user', label: '用户主页', desc: '资料 / 作品 / 关注' },
  { key: 'audit', label: '审核后台', desc: '视频审核 / 弹幕治理' },
]
const featuredVideo = computed(() => videoStore.videoList[0])
const gridVideos = computed(() => videoStore.videoList.slice(featuredVideo.value ? 1 : 0))
const activeFeatureText = computed(() => {
  const item = backendFeatures.find((feature) => feature.key === activeFeature.value)
  return item ? `${item.label}：${item.desc}` : '展示后端返回的已审核视频。'
})

async function loadVideos() {
  loadError.value = ''
  try {
    await videoStore.fetchVideoList({
      page: page.value,
      pageSize,
      keyword: keyword.value.trim() || undefined,
    })
  } catch (error: any) {
    loadError.value = '后端服务暂未连接，当前只显示空状态。'
  }
}

function selectFeature(item: (typeof backendFeatures)[number]) {
  activeFeature.value = item.key
  if (item.key === 'upload') {
    router.push('/creator/upload')
    return
  }
  if (item.key === 'user') {
    router.push('/login')
    return
  }
  if (item.key === 'audit') {
    router.push('/admin')
    return
  }
  if (item.key === 'live') {
    router.push('/live/1')
    return
  }
  router.push({ path: '/', query: { feature: item.key } })
}

function resetFilter() {
  activeFeature.value = 'video'
  keyword.value = ''
  page.value = 1
  router.push({ path: '/', query: { feature: 'video' } })
}

onMounted(loadVideos)
watch(() => route.query.keyword, (value) => {
  keyword.value = String(value || '')
  page.value = 1
  loadVideos()
})

watch(() => route.query.feature, (value) => {
  activeFeature.value = String(value || 'video')
  page.value = 1
  loadVideos()
})
</script>

<style scoped>
.home-page {
  background: #fff;
}

.home-banner {
  height: 174px;
  margin-top: -64px;
  padding-top: 64px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0) 45%, #fff 100%),
    linear-gradient(120deg, rgba(0, 174, 236, 0.42), rgba(251, 114, 153, 0.28)),
    url("https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1600&q=80");
  background-position: center;
  background-size: cover;
}

.banner-inner {
  display: grid;
  align-content: end;
  width: min(1320px, calc(100% - 48px));
  height: 100%;
  margin: 0 auto;
  padding-bottom: 22px;
  color: #fff;
  text-shadow: 0 2px 12px rgba(0, 0, 0, 0.28);
}

.banner-brand {
  font-size: 38px;
  font-weight: 900;
}

.banner-inner p {
  margin: 4px 0 0;
  font-size: 15px;
}

.feature-wrap {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
  padding: 18px 0 20px;
  border-bottom: 1px solid #f1f2f3;
}

.feature-heading {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
}

.feature-heading button {
  border: 0;
  background: transparent;
  color: #18191c;
  cursor: pointer;
  font-size: 18px;
  font-weight: 800;
}

.feature-heading button:hover {
  color: #00aeec;
}

.feature-heading p {
  margin: 0;
  color: #9499a0;
  font-size: 13px;
}

.features {
  display: grid;
  grid-template-columns: repeat(7, minmax(120px, 1fr));
  gap: 12px;
}

.features button,
.quick-links button {
  border: 0;
  border-radius: 8px;
  background: #f6f7f8;
  color: #61666d;
  cursor: pointer;
}

.features button {
  display: grid;
  gap: 4px;
  min-height: 62px;
  padding: 9px 12px;
  text-align: left;
}

.features strong {
  color: #18191c;
  font-size: 15px;
}

.features span {
  color: #9499a0;
  font-size: 12px;
}

.quick-links button {
  height: 34px;
  padding: 0 12px;
  font-size: 14px;
}

.features button:hover,
.features button.active,
.quick-links button:hover {
  background: #e3f6ff;
  color: #00aeec;
}

.features button:hover strong,
.features button.active strong {
  color: #00aeec;
}

.quick-links {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

/*
.channels button,
.channel-right button {
  height: 34px;
  border: 0;
  border-radius: 7px;
  background: #f6f7f8;
  color: #61666d;
  cursor: pointer;
  font-size: 14px;
}

.channels button:hover,
.channels button.active,
.channel-right button:hover {
  background: #e3f6ff;
  color: #00aeec;
}

.channel-right {
  display: grid;
  grid-template-columns: repeat(3, auto);
  gap: 12px;
}
*/

.feed-section {
  padding-top: 26px;
}

.feed-head {
  display: flex;
  align-items: end;
  justify-content: space-between;
  gap: 18px;
  margin-bottom: 18px;
}

.feed-head h2 {
  margin: 0;
  color: #18191c;
  font-size: 24px;
}

.feed-head p {
  margin: 8px 0 0;
  color: #9499a0;
}

.feed-tools {
  display: flex;
  align-items: center;
  gap: 10px;
}

.inline-search {
  width: 280px;
}

.feed-layout {
  display: grid;
  grid-template-columns: 430px minmax(0, 1fr);
  gap: 22px;
  min-height: 360px;
}

.feature-card {
  cursor: pointer;
}

.feature-cover {
  position: relative;
  overflow: hidden;
  height: 100%;
  min-height: 320px;
  border-radius: 10px;
  background: #f1f2f3;
}

.feature-cover img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
}

.feature-fallback {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  background: linear-gradient(135deg, #e3f6ff, #fff0f5);
  color: #00aeec;
  font-size: 28px;
  font-weight: 900;
}

.feature-title {
  position: absolute;
  inset: auto 0 0;
  padding: 70px 18px 18px;
  background: linear-gradient(180deg, transparent, rgba(0, 0, 0, 0.74));
  color: #fff;
}

.feature-title h3 {
  margin: 0;
  font-size: 22px;
  line-height: 1.35;
}

.feature-title p {
  margin: 8px 0 0;
  color: rgba(255, 255, 255, 0.76);
}

.video-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 20px 18px;
}

.empty-card {
  grid-column: 1 / -1;
  display: grid;
  min-height: 360px;
  place-items: center;
  border: 1px solid #f1f2f3;
  border-radius: 10px;
  background: #fff;
}

.pager {
  display: flex;
  justify-content: center;
  margin-top: 28px;
}

@media (max-width: 1100px) {
  .feed-layout {
    grid-template-columns: 1fr;
  }

  .features {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .feature-heading {
    align-items: flex-start;
    flex-direction: column;
    gap: 6px;
  }

  .features,
  .video-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .feed-head,
  .feed-tools {
    align-items: stretch;
    flex-direction: column;
  }

  .inline-search {
    width: 100%;
  }
}
</style>
