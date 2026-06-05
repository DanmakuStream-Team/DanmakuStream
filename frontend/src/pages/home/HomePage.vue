<template>
  <main class="home-page bg-white">

    <section class="page-shell pt-6">
      <div v-if="!isSearching" class="mb-5 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h2 class="m-0 text-[28px] font-black text-[#18191c]">推荐视频</h2>
          <!-- 视频分类标签栏 -->
          <div class="category-tabs mt-3 flex flex-wrap gap-2">
            <button
              v-for="cat in categoryList"
              :key="cat.value"
              class="category-btn"
              :class="{ active: activeCategory === cat.value }"
              @click="selectCategory(cat.value)"
            >
              {{ cat.label }}
            </button>
          </div>
          <p class="m-0 mt-1 text-[#9499a0]">{{ loadError || activeFeatureText }}</p>
        </div>
      </div>

      <!-- 搜索模式：列表布局 -->
      <div v-if="isSearching" v-loading="videoStore.loading" class="search-layout">
        <div class="search-header">
          <h2 class="search-result-title">搜索结果</h2>
        </div>
        <article
          v-for="video in videoStore.videoList"
          :key="video.id"
          class="search-item"
          @click="router.push(`/video/${video.id}`)"
        >
          <div class="search-cover">
            <img v-if="video.coverUrl" :src="mediaUrl(video.coverUrl)" :alt="video.title" />
            <div v-else class="search-cover-fallback">DanmakuStream</div>
            <strong class="search-duration">{{ formatDuration(video.duration) }}</strong>
          </div>
          <div class="search-body">
            <h3>{{ video.title }}</h3>
            <p class="search-author">
              <span>{{ video.author?.nickname || '匿名用户' }}</span>
              <em>{{ formatCount(video.viewCount) }} 播放</em>
            </p>
            <p class="search-desc">{{ video.description || '暂无简介' }}</p>
          </div>
        </article>
        <div v-if="!videoStore.videoList.length" class="empty-card">
          <el-empty description="未找到相关视频" />
        </div>
      </div>

      <!-- 正常模式：精选+网格 -->
      <div v-else v-loading="videoStore.loading" class="feed-layout">
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

        <VideoCard
          v-for="video in gridVideos"
          :key="video.id"
          :video="video"
          @open="router.push(`/video/${video.id}`)"
        />

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
import { formatCount, formatDuration, mediaUrl } from '@/utils/format'

const router = useRouter()
const route = useRoute()
const videoStore = useVideoStore()
const page = ref(1)
const pageSize = 21
const categoryList = ref([
  { label: '全部', value: '' },
  { label: '游戏', value: 'game' },
  { label: '科技', value: 'tech' },
  { label: '生活', value: 'life' },
  { label: '音乐', value: 'music' },
  { label: '动漫', value: 'anime' },
  { label: '知识', value: 'knowledge' },
])
const activeCategory = ref('')
const keyword = ref(String(route.query.keyword || ''))
const loadError = ref('')

const backendFeatures = computed(() => {
  const items = [
    { key: 'video', label: '视频浏览', desc: '列表 / 搜索 / 详情' },
    { key: 'upload', label: '投稿上传', desc: '视频 / 封面 / 转码' },
    { key: 'comment', label: '评论互动', desc: '评论 / 回复 / 点赞' },
    { key: 'live', label: '直播间', desc: '播放 / 实时弹幕 / 互动' },
    { key: 'user', label: '用户主页', desc: '资料 / 作品 / 关注' },
  ]
  if (authStore.isAdmin) {
    items.push({ key: 'audit', label: '审核后台', desc: '视频审核 / 弹幕治理' })
  }
  return items
})

const activeFeatureText = computed(() => {
  const feature = backendFeatures.value.find(f => f.key === activeFeature.value)
  return feature?.desc || '视频浏览'
})

const isSearching = computed(() => Boolean(keyword.value.trim()))
const featuredVideo = computed(() => videoStore.videoList[0])
const gridVideos = computed(() => videoStore.videoList.slice(featuredVideo.value ? 1 : 0))


async function loadVideos() {
  loadError.value = ''
  try {
    await videoStore.fetchVideoList({
      page: page.value,
      pageSize,
      keyword: keyword.value.trim() || undefined,
      category: activeCategory.value || undefined,
    })
  } catch (error: any) {
    loadError.value = '后端服务暂未连接，当前只显示空状态。'
  }
}


function selectCategory(catValue: string) {
  activeCategory.value = catValue
  page.value = 1
  loadVideos()
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

watch(() => route.query.feature, () => {
  page.value = 1
  loadVideos()
})
</script>

<style scoped>
.features {
  display: grid;
  grid-template-columns: repeat(7, minmax(120px, 1fr));
  gap: 12px;
}

/* 视频分类标签样式 */
.category-btn {
  padding: 6px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  background: #fff;
  font-size: 14px;
  cursor: pointer;
}
.category-btn.active {
  background: #00aeec;
  color: #fff;
  border-color: #00aeec;
}

.features button {
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

.features button:hover,
.features button.active {
  background: #e3f6ff;
  color: #00aeec;
}

.features button:hover strong,
.features button.active strong {
  color: #00aeec;
}

.inline-search {
  width: 280px;
}

.feed-layout {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 18px;
  min-height: 360px;
}

.feature-card {
  grid-column: span 2;
  grid-row: span 2;
  cursor: pointer;
}

.feature-cover {
  position: relative;
  overflow: hidden;
  height: 100%;
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

.search-layout {
  display: grid;
  gap: 16px;
  min-height: 360px;
}

.search-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 8px;
}

.search-result-title {
  margin: 0;
  font-size: 24px;
  font-weight: 900;
  color: #18191c;
}

.search-item {
  display: grid;
  grid-template-columns: 432px minmax(0, 1fr);
  gap: 28px;
  padding: 14px;
  border-radius: 12px;
  background: #fff;
  cursor: pointer;
  transition: background 0.15s ease;
}


.search-cover {
  position: relative;
  overflow: hidden;
  aspect-ratio: 16 / 9;
  border-radius: 8px;
  background: #f1f2f3;
  flex-shrink: 0;
}

.search-cover img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
}

.search-cover-fallback {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  background: linear-gradient(135deg, #e3f6ff, #fff0f5);
  color: #00aeec;
  font-size: 18px;
  font-weight: 900;
}

.search-duration {
  position: absolute;
  right: 6px;
  bottom: 6px;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.72);
  color: #fff;
  font-size: 12px;
  font-weight: 600;
}

.search-body {
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-width: 0;
}

.search-body h3 {
  margin: 0 0 8px;
  font-size: 21px;
  font-weight: 800;
  color: #18191c;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.search-item:hover .search-body h3 {
  color: #00aeec;
}

.search-author {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 10px;
  font-size: 15px;
  color: #9499a0;
}

.search-author em {
  font-style: normal;
}

.search-desc {
  margin: 0;
  font-size: 14px;
  color: #61666d;
  line-height: 1.55;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.empty-card {
  grid-column: 1 / -1;
  display: grid;
  min-height: 360px;
  place-items: center;
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

  .feature-card {
    grid-column: span 1;
    grid-row: span 1;
  }

  .features {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .features {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .inline-search {
    width: 100%;
  }

  .search-item {
    grid-template-columns: 1fr;
  }
}
</style>
