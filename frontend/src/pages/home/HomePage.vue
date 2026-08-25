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
      <div v-if="isSearching" v-loading="isInitialLoading" class="search-layout">
        <div class="search-header">
          <h2 class="search-result-title">搜索结果</h2>
        </div>
        <article
          v-for="video in videoStore.videoList"
          :key="video.id"
          class="search-item"
          @click="openVideo(video)"
        >
          <div class="search-cover">
            <img v-if="video.coverUrl" :src="mediaUrl(video.coverUrl)" :alt="video.title" />
            <div v-else class="search-cover-fallback">DanmakuStream</div>
            <strong class="search-duration">{{ formatDuration(video.duration) }}</strong>
          </div>
          <div class="search-body">
            <h3>{{ video.title }}</h3>
            <p class="search-author">
              <button class="user-link" type="button" :disabled="!video.author?.id" @click.stop="openUser(video.author?.id)">
                {{ video.author?.nickname || '匿名用户' }}
              </button>
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
      <div v-else ref="feedLayoutRef" v-loading="isInitialLoading" class="feed-layout">
        <div v-if="false && featuredVideo" class="featured-section">
          <article class="feature-card" @click="openVideo(featuredVideo)">
          <div class="feature-cover">
            <img v-if="featuredVideo.coverUrl" :src="mediaUrl(featuredVideo.coverUrl)" :alt="featuredVideo.title" />
            <div v-else class="feature-fallback">DanmakuStream</div>
            <div class="feature-title">
              <h3>{{ featuredVideo.title }}</h3>
              <button class="feature-author" type="button" :disabled="!featuredVideo.author?.id" @click.stop="openUser(featuredVideo.author?.id)">
                {{ featuredVideo.author?.nickname || '匿名用户' }}
              </button>
            </div>
          </div>
          </article>

          <div class="featured-side-grid">
            <article
              v-for="video in featuredSideVideos"
              :key="video.id"
              class="side-video-card"
              @click="openVideo(video)"
            >
              <div class="side-cover">
                <img v-if="video.coverUrl" :src="mediaUrl(video.coverUrl)" :alt="video.title" />
                <div v-else class="side-fallback">Danmaku</div>
              </div>
              <div class="side-video-body">
                <h3>{{ video.title }}</h3>
                <button class="side-author" type="button" :disabled="!video.author?.id" @click.stop="openUser(video.author?.id)">
                  <el-icon><User /></el-icon>
                  <span>{{ video.author?.nickname || '匿名用户' }}</span>
                  <em>{{ formatTime(video.createdAt) }}</em>
                </button>
              </div>
            </article>
          </div>
        </div>

        <VideoCard
          v-for="video in videoStore.videoList"
          :key="video.id"
          :video="video"
          @open="openVideo(video)"
        />

        <div v-if="!videoStore.videoList.length" class="empty-card">
          <el-empty :description="loadError ? '后端暂未连接，或接口返回失败' : '暂无已审核视频'" />
        </div>
      </div>

      <div v-if="videoStore.videoList.length && !loadError" ref="loadMoreRef" class="load-more">
        <span v-if="videoStore.loading">加载中...</span>
        <span v-else-if="hasMoreVideos">继续下滑加载更多</span>
        <span v-else>没有更多了</span>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { User } from '@element-plus/icons-vue'
import VideoCard from '@/components/common/VideoCard.vue'
import { useAuthStore } from '@/store/auth'
import { useVideoStore } from '@/store/video'
import type { VideoInfo } from '@/types'
import { formatCount, formatDuration, formatTime, mediaUrl } from '@/utils/format'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const videoStore = useVideoStore()
const page = ref(1)
const feedLayoutRef = ref<HTMLElement>()
const loadMoreRef = ref<HTMLElement>()
const feedColumnCount = ref(4)
let resizeObserver: ResizeObserver | undefined
let loadMoreObserver: IntersectionObserver | undefined
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
const recommendKey = 'danmaku:recommend-tags'

const featureDescriptions: Record<string, string> = {
  video: '列表 / 搜索 / 详情',
  danmaku: '弹幕互动 / 实时评论 / 视频交流',
}

const activeFeature = computed(() => String(route.query.feature || 'video'))

const activeFeatureText = computed(() => {
  const category = categoryList.value.find(cat => cat.value === activeCategory.value)
  if (category?.value) return `${category.label}频道`
  return featureDescriptions[activeFeature.value] || '视频浏览'
})

const isSearching = computed(() => Boolean(keyword.value.trim()))
const isInitialLoading = computed(() => videoStore.loading && page.value === 1)
const hasMoreVideos = computed(() => videoStore.videoList.length < videoStore.total)
const featuredVideo = computed(() => videoStore.videoList[0])
const featuredSideVideos = computed(() => videoStore.videoList.slice(featuredVideo.value ? 1 : 0, featuredVideo.value ? 7 : 0))
const currentPageSize = computed(() => {
  return isSearching.value ? 20 : 24
})

function updateFeedColumnCount() {
  const el = feedLayoutRef.value
  if (!el) return

  const columns = getComputedStyle(el)
    .gridTemplateColumns
    .split(' ')
    .filter(Boolean).length

  feedColumnCount.value = Math.max(columns, 1)
}


async function loadVideos(options: { append?: boolean } = {}) {
  loadError.value = ''
  if (videoStore.loading) return

  const nextPage = options.append ? page.value + 1 : 1
  try {
    await videoStore.fetchVideoList({
      page: nextPage,
      pageSize: currentPageSize.value,
      keyword: keyword.value.trim() || undefined,
      tag: String(route.query.tag || '').trim() || undefined,
      category: activeCategory.value || undefined,
      sort: 'hot',
    }, options.append)
    page.value = nextPage
    rankRecommendedVideos()
    await nextTick()
    observeLoadMore()
  } catch (error: any) {
    loadError.value = '后端服务暂未连接，当前只显示空状态。'
  }
}


function selectCategory(catValue: string) {
  activeCategory.value = catValue
  resetAndLoadVideos()
}

function openUser(userId?: number) {
  if (!userId) return
  router.push(`/user/${userId}`)
}

function openVideo(video: VideoInfo) {
  rememberVideoPreference(video)
  router.push(`/video/${video.id}`)
}

function resetAndLoadVideos() {
  page.value = 1
  loadVideos()
}

function observeLoadMore() {
  loadMoreObserver?.disconnect()
  if (!loadMoreRef.value) return

  loadMoreObserver = new IntersectionObserver((entries) => {
    const entry = entries[0]
    if (!entry?.isIntersecting || !hasMoreVideos.value || videoStore.loading || loadError.value) return
    loadVideos({ append: true })
  }, { rootMargin: '360px 0px' })
  loadMoreObserver.observe(loadMoreRef.value)
}

function rankRecommendedVideos() {
  if (isSearching.value) {
    dedupeVideos()
    return
  }

  const weights = readRecommendWeights()
  const seed = authStore.userInfo?.id || getAnonSeed()
  const unique = dedupeVideos(false)

  videoStore.videoList = unique.sort((a, b) => recommendScore(b, weights, seed) - recommendScore(a, weights, seed))
}

function recommendScore(video: VideoInfo, weights: Record<string, number>, seed: number) {
  const tags = parseTags(video.tags)
  const preferenceScore = tags.reduce((sum, tag) => sum + (weights[tag] || 0), 0)
  const engagementScore = video.likeCount * 5 + video.collectCount * 4 + video.danmakuCount * 2 + video.viewCount
  const freshScore = Date.parse(video.createdAt || '') || 0
  const stableNoise = hashNumber(`${seed}:${video.id}`) % 1000

  return preferenceScore * 10000 + engagementScore * 10 + freshScore / 100000000 + stableNoise / 1000
}

function dedupeVideos(writeBack = true) {
  const map = new Map<number, VideoInfo>()
  for (const video of videoStore.videoList) {
    if (!map.has(video.id)) map.set(video.id, video)
  }
  const list = Array.from(map.values())
  if (writeBack) videoStore.videoList = list
  return list
}

function rememberVideoPreference(video: VideoInfo) {
  const tags = parseTags(video.tags)
  if (!tags.length) return

  const weights = readRecommendWeights()
  for (const tag of tags) {
    weights[tag] = Math.min(20, (weights[tag] || 0) + 1)
  }
  localStorage.setItem(recommendKey, JSON.stringify(weights))
}

function readRecommendWeights() {
  try {
    const parsed = JSON.parse(localStorage.getItem(recommendKey) || '{}')
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {}
    return Object.fromEntries(
      Object.entries(parsed)
        .filter(([key, value]) => key && typeof value === 'number')
        .map(([key, value]) => [key, Math.max(0, Math.min(20, value as number))]),
    ) as Record<string, number>
  } catch {
    return {}
  }
}

function parseTags(tags: VideoInfo['tags']) {
  const list = Array.isArray(tags) ? tags : String(tags || '').split(',')
  return list.map(tag => tag.trim()).filter(Boolean)
}

function getAnonSeed() {
  const key = 'danmaku:recommend-seed'
  const cached = Number(localStorage.getItem(key))
  if (Number.isFinite(cached) && cached > 0) return cached

  const seed = Math.floor(Math.random() * 1000000000)
  localStorage.setItem(key, String(seed))
  return seed
}

function hashNumber(value: string) {
  let hash = 2166136261
  for (let i = 0; i < value.length; i++) {
    hash ^= value.charCodeAt(i)
    hash = Math.imul(hash, 16777619)
  }
  return Math.abs(hash)
}

onMounted(async () => {
  await nextTick()
  updateFeedColumnCount()
  if (feedLayoutRef.value) {
    resizeObserver = new ResizeObserver(() => {
      const previousPageSize = currentPageSize.value
      updateFeedColumnCount()
      if (currentPageSize.value !== previousPageSize) {
        page.value = 1
        loadVideos()
      }
    })
    resizeObserver.observe(feedLayoutRef.value)
  }
  loadVideos()
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  loadMoreObserver?.disconnect()
})
watch(() => route.query.keyword, (value) => {
  keyword.value = String(value || '')
  resetAndLoadVideos()
})

watch(() => route.query.tag, () => {
  resetAndLoadVideos()
})

watch(() => route.query.feature, () => {
  resetAndLoadVideos()
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
  font-size: 13px;
  white-space: nowrap;
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
  font-size: 14px;
}

.features span {
  color: #9499a0;
  font-size: 11px;
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
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 18px;
  min-height: 360px;
}

.featured-section {
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: minmax(0, 2fr) minmax(0, 3fr);
  gap: 18px;
  align-items: stretch;
}

.feature-card {
  min-width: 0;
  min-height: 100%;
  cursor: pointer;
}

.feature-cover {
  position: relative;
  overflow: hidden;
  aspect-ratio: 16 / 9;
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
  font-size: 26px;
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
  overflow: hidden;
  font-size: 20px;
  line-height: 1.35;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feature-author {
  display: block;
  overflow: hidden;
  margin: 8px 0 0;
  padding: 0;
  border: 0;
  background: transparent;
  color: rgba(255, 255, 255, 0.76);
  cursor: pointer;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feature-author:not(:disabled):hover {
  color: #fff;
}

.feature-author:disabled {
  cursor: default;
}

.featured-side-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  grid-template-rows: repeat(2, minmax(0, 1fr));
  gap: 18px;
  min-width: 0;
}

.side-video-card {
  display: block;
  min-width: 0;
  min-height: 0;
  padding: 0;
  border-radius: 0;
  background: transparent;
  cursor: pointer;
  transition: none;
}

.side-video-card:hover {
  background: transparent;
  transform: none;
}

.side-cover {
  position: relative;
  overflow: hidden;
  aspect-ratio: 16 / 9;
  border-radius: 8px;
  background: #f1f2f3;
}

.side-cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.25s ease;
}

.side-video-card:hover .side-cover img {
  transform: scale(1.04);
}

.side-fallback {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  color: #00aeec;
  font-size: 12px;
  font-weight: 900;
}

.side-video-body {
  display: grid;
  gap: 6px;
  min-width: 0;
  padding: 8px 2px 0;
}

.side-video-body h3 {
  display: block;
  min-height: 20px;
  margin: 0;
  overflow: hidden;
  color: #18191c;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.45;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.side-video-card:hover .side-video-body h3 {
  color: #00aeec;
}

.side-author {
  display: flex;
  align-items: center;
  gap: 5px;
  min-width: 0;
  padding: 0;
  border: 0;
  background: transparent;
  color: #9499a0;
  cursor: pointer;
  font-size: 12px;
  text-align: left;
}

.side-author .el-icon {
  flex-shrink: 0;
  font-size: 13px;
}

.side-author:hover {
  color: #00aeec;
}

.side-author:disabled {
  cursor: default;
}

.side-author:disabled:hover {
  color: #9499a0;
}

.side-author span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.side-author em {
  flex-shrink: 0;
  font-style: normal;
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
  font-size: 22px;
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
  font-size: 16px;
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
  font-size: 11px;
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
  overflow: hidden;
  font-size: 19px;
  font-weight: 800;
  color: #18191c;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-item:hover .search-body h3 {
  color: #00aeec;
}

.search-author {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 0 0 10px;
  font-size: 14px;
  color: #9499a0;
}

.search-author em {
  font-style: normal;
}

.user-link {
  max-width: 180px;
  overflow: hidden;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font: inherit;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-link:not(:disabled):hover {
  color: #00aeec;
}

.user-link:disabled {
  cursor: default;
}

.search-desc {
  margin: 0;
  font-size: 13px;
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

.load-more {
  display: grid;
  min-height: 68px;
  place-items: center;
  margin-top: 22px;
  color: #9499a0;
  font-size: 14px;
  font-weight: 800;
}

@media (max-width: 1100px) {
  .feed-layout {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .featured-section {
    grid-template-columns: 1fr;
  }

  .featured-side-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .features {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .feed-layout {
    grid-template-columns: 1fr;
  }

  .features {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .inline-search {
    width: 100%;
  }

  .search-item {
    grid-template-columns: 1fr;
  }

  .featured-section {
    grid-template-columns: 1fr;
  }

  .featured-side-grid {
    grid-template-columns: 1fr;
  }
}
</style>
