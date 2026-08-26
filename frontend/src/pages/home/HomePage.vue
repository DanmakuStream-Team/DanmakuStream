<template>
  <main class="home-page">
    <div class="page-shell home-shell">
      <section v-if="continueWatching.length && !isSearching" class="continue-section">
        <div class="section-title-row">
          <div>
            <span class="section-kicker">继续观看</span>
            <h1>接着上次的进度</h1>
          </div>
          <button class="text-action" type="button" @click="router.push('/me/history')">
            查看历史
            <el-icon><ArrowRight /></el-icon>
          </button>
        </div>

        <div class="continue-grid">
          <article
            v-for="record in continueWatching"
            :key="record.video.id"
            class="continue-card"
            @click="openHistoryVideo(record)"
          >
            <div class="continue-cover">
              <img v-if="record.video.coverUrl" :src="mediaUrl(record.video.coverUrl)" :alt="record.video.title">
              <div v-else class="cover-fallback">Danmaku</div>
              <span class="duration">{{ formatDuration(record.video.duration) }}</span>
              <i :style="{ width: `${record.progress}%` }" />
            </div>
            <div class="continue-body">
              <strong>{{ record.video.title }}</strong>
              <span>{{ record.video.author?.nickname || '匿名用户' }} · 已观看 {{ record.progress }}%</span>
            </div>
          </article>
        </div>
      </section>

      <section class="discover-section">
        <div class="section-title-row discover-head">
          <div>
            <span class="section-kicker">{{ isSearching ? '搜索' : '发现' }}</span>
            <h1>{{ isSearching ? `“${keyword.trim()}”的结果` : '为你推荐' }}</h1>
            <p>{{ loadError || (isSearching ? `共找到 ${videoStore.total} 个视频` : activeCategoryText) }}</p>
          </div>
          <div class="feed-actions">
            <el-segmented v-model="sortMode" :options="sortOptions" @change="resetAndLoadVideos" />
            <el-tooltip content="换一批推荐" placement="bottom">
              <button v-if="!isSearching" class="icon-action" type="button" aria-label="换一批推荐" @click="refreshRecommendations">
                <el-icon><Refresh /></el-icon>
              </button>
            </el-tooltip>
          </div>
        </div>

        <div class="category-strip" aria-label="视频分类">
          <button
            v-for="category in categoryList"
            :key="category.value"
            type="button"
            :class="{ active: activeCategory === category.value }"
            @click="selectCategory(category.value)"
          >
            {{ category.label }}
          </button>
        </div>

        <Transition name="feedback">
          <div v-if="lastDismissed" class="feedback-bar" role="status" aria-live="polite">
            <span>已减少类似《{{ lastDismissed.video.title }}》的推荐</span>
            <button type="button" @click="undoNotInterested">撤销</button>
            <button class="feedback-close" type="button" aria-label="关闭提示" @click="clearDismissedFeedback">
              <el-icon><Close /></el-icon>
            </button>
          </div>
        </Transition>

        <div v-if="isInitialLoading" class="feed-layout skeleton-layout" aria-label="正在加载视频">
          <article v-for="index in 8" :key="index" class="skeleton-card" aria-hidden="true">
            <div class="skeleton-cover" />
            <div class="skeleton-line title-line" />
            <div class="skeleton-line meta-line" />
          </article>
        </div>

        <div v-else-if="isSearching" class="search-layout">
          <article v-for="video in videoStore.videoList" :key="video.id" class="search-item" @click="openVideo(video)">
            <div class="search-cover">
              <img v-if="video.coverUrl" :src="mediaUrl(video.coverUrl)" :alt="video.title">
              <div v-else class="cover-fallback">DanmakuStream</div>
              <span class="duration">{{ formatDuration(video.duration) }}</span>
            </div>
            <div class="search-body">
              <h2>{{ video.title }}</h2>
              <button type="button" :disabled="!video.author?.id" @click.stop="openUser(video.author?.id)">
                {{ video.author?.nickname || '匿名用户' }}
              </button>
              <span>{{ formatCount(video.viewCount) }} 次观看 · {{ formatTime(video.createdAt) }}</span>
              <p>{{ video.description || '这个视频暂时没有简介。' }}</p>
            </div>
          </article>
        </div>

        <div v-else-if="videoStore.videoList.length" ref="feedLayoutRef" class="feed-layout">
          <VideoCard
            v-for="video in videoStore.videoList"
            :key="video.id"
            :video="video"
            show-feedback
            @open="openVideo(video)"
            @watch-later="toggleWatchLater(video)"
            @queue="addToQueue(video)"
            @not-interested="markNotInterested(video)"
          />
        </div>

        <div v-if="!isInitialLoading && !videoStore.videoList.length" class="empty-card">
          <el-empty :description="loadError ? '暂时无法加载视频，请稍后重试' : '没有找到相关视频'">
            <el-button v-if="loadError" type="primary" @click="resetAndLoadVideos">重新加载</el-button>
          </el-empty>
        </div>

        <div v-if="videoStore.videoList.length && !loadError" ref="loadMoreRef" class="load-more" aria-live="polite">
          <span v-if="videoStore.loading">正在加载更多...</span>
          <span v-else-if="hasMoreVideos">继续向下浏览</span>
          <span v-else>已经看到全部内容</span>
        </div>
      </section>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowRight, Close, Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import VideoCard from '@/components/common/VideoCard.vue'
import { libraryApi, type ServerLibraryRecord } from '@/api/library'
import { useAuthStore } from '@/store/auth'
import { useVideoStore } from '@/store/video'
import type { VideoInfo } from '@/types'
import { formatCount, formatDuration, formatTime, mediaUrl } from '@/utils/format'
import { loadPlayQueue, savePlayQueue } from '@/utils/playQueue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const videoStore = useVideoStore()
const page = ref(1)
const feedLayoutRef = ref<HTMLElement>()
const loadMoreRef = ref<HTMLElement>()
const continueWatching = ref<ServerLibraryRecord[]>([])
const categoryList = [
  { label: '全部', value: '' },
  { label: '游戏', value: 'game' },
  { label: '科技', value: 'tech' },
  { label: '生活', value: 'life' },
  { label: '音乐', value: 'music' },
  { label: '动漫', value: 'anime' },
  { label: '知识', value: 'knowledge' },
]
const sortOptions = [
  { label: '综合', value: 'hot' },
  { label: '最新', value: 'date' },
  { label: '点赞', value: 'like' },
  { label: '收藏', value: 'collect' },
]
const activeCategory = ref('')
const sortMode = ref<'hot' | 'date' | 'like' | 'collect'>('hot')
const keyword = ref(String(route.query.keyword || ''))
const loadError = ref('')
const recommendationNonce = ref(0)
const recommendKey = 'danmaku:recommend-tags'
const dismissedKey = 'danmaku:recommend-dismissed'
const lastDismissed = ref<{
  video: VideoInfo
  index: number
  previousWeights: Record<string, number>
} | null>(null)
let loadMoreObserver: IntersectionObserver | undefined
let feedbackTimer: ReturnType<typeof setTimeout> | undefined

const isSearching = computed(() => Boolean(keyword.value.trim()))
const isInitialLoading = computed(() => videoStore.loading && page.value === 1)
const hasMoreVideos = computed(() => videoStore.videoList.length < videoStore.total)
const currentPageSize = computed(() => isSearching.value ? 20 : 24)
const activeCategoryText = computed(() => {
  const category = categoryList.find(item => item.value === activeCategory.value)
  return category?.value ? `${category.label}频道的热门内容` : '结合热度、时间和你的兴趣排序'
})

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
      sort: sortMode.value,
    }, options.append)
    page.value = nextPage
    rankRecommendedVideos()
    await nextTick()
    observeLoadMore()
  } catch {
    loadError.value = '视频服务暂时不可用'
  }
}

async function loadContinueWatching() {
  if (!authStore.isLoggedIn) {
    continueWatching.value = []
    return
  }
  try {
    const res = await libraryApi.history({ page: 1, pageSize: 8 })
    continueWatching.value = res.data.list.filter(item => item.progress > 0 && item.progress < 95).slice(0, 4)
  } catch {
    continueWatching.value = []
  }
}

function selectCategory(value: string) {
  activeCategory.value = value
  resetAndLoadVideos()
}

function openUser(userId?: number) {
  if (userId) router.push(`/user/${userId}`)
}

function openVideo(video: VideoInfo) {
  rememberVideoPreference(video)
  router.push(`/video/${video.id}`)
}

function openHistoryVideo(record: ServerLibraryRecord) {
  router.push({ path: `/video/${record.video.id}`, query: { t: String(record.position) } })
}

async function toggleWatchLater(video: VideoInfo) {
  if (!authStore.isLoggedIn) {
    ElMessage.warning('登录后可以使用稍后再看')
    router.push({ path: '/login', query: { redirect: route.fullPath } })
    return
  }
  try {
    const res = await libraryApi.toggleWatchLater(video.id)
    ElMessage.success(res.data.saved ? '已加入稍后再看' : '已从稍后再看移除')
  } catch (error: any) {
    ElMessage.error(error.message || '稍后再看操作失败')
  }
}

function addToQueue(video: VideoInfo) {
  const queue = loadPlayQueue()
  if (queue.some(item => item.id === video.id)) {
    ElMessage.info('这个视频已经在播放队列中')
    return
  }
  savePlayQueue([...queue, video])
  ElMessage.success('已加入播放队列')
}

function markNotInterested(video: VideoInfo) {
  const index = videoStore.videoList.findIndex(item => item.id === video.id)
  if (index < 0) return

  const weights = readRecommendWeights()
  const previousWeights = { ...weights }
  parseTags(video.tags).forEach((tag) => {
    weights[tag] = Math.max(-5, (weights[tag] || 0) - 2)
  })
  localStorage.setItem(recommendKey, JSON.stringify(weights))

  const dismissedIds = readDismissedVideoIds()
  dismissedIds.add(video.id)
  saveDismissedVideoIds(dismissedIds)
  videoStore.videoList.splice(index, 1)
  lastDismissed.value = { video, index, previousWeights }

  if (feedbackTimer) clearTimeout(feedbackTimer)
  feedbackTimer = setTimeout(clearDismissedFeedback, 7000)
}

function undoNotInterested() {
  const dismissed = lastDismissed.value
  if (!dismissed) return
  const dismissedIds = readDismissedVideoIds()
  dismissedIds.delete(dismissed.video.id)
  saveDismissedVideoIds(dismissedIds)
  localStorage.setItem(recommendKey, JSON.stringify(dismissed.previousWeights))
  if (!videoStore.videoList.some(item => item.id === dismissed.video.id)) {
    videoStore.videoList.splice(Math.min(dismissed.index, videoStore.videoList.length), 0, dismissed.video)
  }
  clearDismissedFeedback()
}

function clearDismissedFeedback() {
  lastDismissed.value = null
  if (feedbackTimer) clearTimeout(feedbackTimer)
  feedbackTimer = undefined
}

function resetAndLoadVideos() {
  page.value = 1
  void loadVideos()
}

function refreshRecommendations() {
  localStorage.setItem('danmaku:recommend-seed', String(Math.floor(Math.random() * 1_000_000_000)))
  recommendationNonce.value = Math.floor(Math.random() * 1_000_000_000)
  resetAndLoadVideos()
}

function observeLoadMore() {
  loadMoreObserver?.disconnect()
  if (!loadMoreRef.value) return
  loadMoreObserver = new IntersectionObserver((entries) => {
    const entry = entries[0]
    if (!entry?.isIntersecting || !hasMoreVideos.value || videoStore.loading || loadError.value) return
    void loadVideos({ append: true })
  }, { rootMargin: '360px 0px' })
  loadMoreObserver.observe(loadMoreRef.value)
}

function rankRecommendedVideos() {
  const unique = dedupeVideos()
  if (isSearching.value || sortMode.value !== 'hot') {
    videoStore.videoList = unique
    return
  }
  const weights = readRecommendWeights()
  const seed = (authStore.userInfo?.id || getAnonSeed()) + recommendationNonce.value
  videoStore.videoList = unique.sort((a, b) => recommendScore(b, weights, seed) - recommendScore(a, weights, seed))
}

function recommendScore(video: VideoInfo, weights: Record<string, number>, seed: number) {
  const preferenceScore = parseTags(video.tags).reduce((sum, tag) => sum + (weights[tag] || 0), 0)
  const engagementScore = video.likeCount * 5 + video.collectCount * 4 + video.danmakuCount * 2 + video.viewCount
  const freshScore = Date.parse(video.createdAt || '') || 0
  return preferenceScore * 10_000 + engagementScore * 10 + freshScore / 100_000_000 + hashNumber(`${seed}:${video.id}`) / 1_000_000
}

function dedupeVideos() {
  const map = new Map<number, VideoInfo>()
  const dismissedIds = readDismissedVideoIds()
  videoStore.videoList.forEach((video) => {
    if (!dismissedIds.has(video.id) && !map.has(video.id)) map.set(video.id, video)
  })
  return [...map.values()]
}

function readDismissedVideoIds() {
  try {
    const value = JSON.parse(localStorage.getItem(dismissedKey) || '[]')
    return new Set<number>(Array.isArray(value) ? value.filter(id => Number.isInteger(id)) : [])
  } catch {
    return new Set<number>()
  }
}

function saveDismissedVideoIds(ids: Set<number>) {
  localStorage.setItem(dismissedKey, JSON.stringify([...ids].slice(-200)))
}

function rememberVideoPreference(video: VideoInfo) {
  const tags = parseTags(video.tags)
  if (!tags.length) return
  const weights = readRecommendWeights()
  tags.forEach(tag => { weights[tag] = Math.min(20, (weights[tag] || 0) + 1) })
  localStorage.setItem(recommendKey, JSON.stringify(weights))
}

function readRecommendWeights() {
  try {
    const parsed = JSON.parse(localStorage.getItem(recommendKey) || '{}')
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {}
    return Object.fromEntries(Object.entries(parsed).filter(([key, value]) => key && typeof value === 'number')) as Record<string, number>
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
  const seed = Math.floor(Math.random() * 1_000_000_000)
  localStorage.setItem(key, String(seed))
  return seed
}

function hashNumber(value: string) {
  let hash = 2166136261
  for (let index = 0; index < value.length; index++) {
    hash ^= value.charCodeAt(index)
    hash = Math.imul(hash, 16777619)
  }
  return Math.abs(hash % 1000)
}

onMounted(() => {
  void Promise.all([loadVideos(), loadContinueWatching()])
})

onBeforeUnmount(() => {
  loadMoreObserver?.disconnect()
  if (feedbackTimer) clearTimeout(feedbackTimer)
})

watch(() => route.query.keyword, (value) => {
  keyword.value = String(value || '')
  resetAndLoadVideos()
})
watch(() => route.query.tag, resetAndLoadVideos)
watch(() => authStore.isLoggedIn, () => void loadContinueWatching())
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  background: #fff;
}

.home-shell {
  display: grid;
  gap: 34px;
  padding-top: 28px;
}

.continue-section,
.discover-section {
  min-width: 0;
}

.section-title-row {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 16px;
}

.section-kicker {
  display: block;
  margin-bottom: 4px;
  color: #00aeec;
  font-size: 12px;
  font-weight: 800;
}

.section-title-row h1 {
  margin: 0;
  color: #18191c;
  font-size: 25px;
  line-height: 1.25;
}

.discover-head p {
  margin: 6px 0 0;
  color: #9499a0;
  font-size: 13px;
}

.text-action,
.icon-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  color: #61666d;
  cursor: pointer;
}

.text-action {
  gap: 5px;
  padding: 8px 0;
  font-size: 13px;
}

.text-action:hover,
.icon-action:hover {
  color: #00aeec;
}

.continue-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.continue-card {
  display: grid;
  grid-template-columns: 42% minmax(0, 1fr);
  align-items: center;
  min-width: 0;
  overflow: hidden;
  border: 1px solid #e7e9ed;
  border-radius: 8px;
  background: #fff;
  cursor: pointer;
}

.continue-card:hover {
  border-color: #cfd4dc;
}

.continue-cover,
.search-cover {
  position: relative;
  overflow: hidden;
  aspect-ratio: 16 / 9;
  background: #f1f2f3;
}

.continue-cover img,
.search-cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.continue-cover i {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 3px;
  background: #fb7299;
}

.duration {
  position: absolute;
  right: 6px;
  bottom: 6px;
  padding: 2px 5px;
  border-radius: 4px;
  background: rgb(0 0 0 / 72%);
  color: #fff;
  font-size: 11px;
}

.continue-body {
  display: grid;
  min-width: 0;
  gap: 7px;
  padding: 10px 12px;
}

.continue-body strong {
  overflow: hidden;
  color: #18191c;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.continue-body span {
  overflow: hidden;
  color: #9499a0;
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feed-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.icon-action {
  width: 34px;
  height: 34px;
  border: 1px solid #e4e7ec;
  border-radius: 50%;
  font-size: 17px;
}

.category-strip {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  overflow-x: auto;
  padding-bottom: 2px;
  scrollbar-width: none;
}

.category-strip::-webkit-scrollbar {
  display: none;
}

.category-strip button {
  min-width: 64px;
  height: 34px;
  padding: 0 14px;
  flex: 0 0 auto;
  border: 1px solid #e4e7ec;
  border-radius: 6px;
  background: #fff;
  color: #475467;
  cursor: pointer;
  font-size: 13px;
}

.category-strip button:hover {
  border-color: #b8c0cc;
}

.category-strip button.active {
  border-color: #18191c;
  background: #18191c;
  color: #fff;
}

.feedback-bar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto 30px;
  min-height: 42px;
  align-items: center;
  gap: 12px;
  margin: -6px 0 18px;
  padding: 6px 8px 6px 14px;
  border: 1px solid #e4e7ec;
  border-radius: 6px;
  background: #f8f9fa;
  color: #475467;
  font-size: 13px;
}

.feedback-bar > span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feedback-bar button {
  padding: 6px 8px;
  border: 0;
  background: transparent;
  color: #00aeec;
  cursor: pointer;
  font-weight: 700;
}

.feedback-bar .feedback-close {
  display: grid;
  width: 30px;
  height: 30px;
  padding: 0;
  place-items: center;
  border-radius: 50%;
  color: #9499a0;
}

.feedback-bar .feedback-close:hover {
  background: #e9ecf0;
  color: #18191c;
}

.feedback-enter-active,
.feedback-leave-active {
  transition: opacity 0.16s ease, transform 0.16s ease;
}

.feedback-enter-from,
.feedback-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.feed-layout {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 26px 18px;
  min-height: 360px;
}

.skeleton-card {
  min-width: 0;
  animation: skeleton-pulse 1.25s ease-in-out infinite alternate;
}

.skeleton-cover,
.skeleton-line {
  background: #eef0f3;
}

.skeleton-cover {
  aspect-ratio: 16 / 9;
  border-radius: 8px;
}

.skeleton-line {
  height: 12px;
  margin-top: 10px;
  border-radius: 4px;
}

.title-line {
  width: 84%;
}

.meta-line {
  width: 56%;
  height: 10px;
  margin-top: 8px;
}

@keyframes skeleton-pulse {
  from { opacity: 0.58; }
  to { opacity: 1; }
}

@media (prefers-reduced-motion: reduce) {
  .skeleton-card {
    animation: none;
  }
}

.search-layout {
  display: grid;
  gap: 8px;
  min-height: 360px;
}

.search-item {
  display: grid;
  grid-template-columns: minmax(260px, 360px) minmax(0, 1fr);
  gap: 22px;
  padding: 12px 0;
  border-bottom: 1px solid #eef0f3;
  cursor: pointer;
}

.search-cover {
  border-radius: 8px;
}

.cover-fallback {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  background: #f2f6fa;
  color: #00aeec;
  font-weight: 800;
}

.search-body {
  display: flex;
  min-width: 0;
  justify-content: center;
  flex-direction: column;
}

.search-body h2 {
  margin: 0 0 10px;
  overflow: hidden;
  color: #18191c;
  font-size: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-item:hover h2 {
  color: #00aeec;
}

.search-body button {
  align-self: flex-start;
  padding: 0;
  border: 0;
  background: transparent;
  color: #61666d;
  cursor: pointer;
  font-size: 13px;
}

.search-body span {
  margin-top: 5px;
  color: #9499a0;
  font-size: 12px;
}

.search-body p {
  display: -webkit-box;
  margin: 14px 0 0;
  overflow: hidden;
  color: #667085;
  font-size: 13px;
  line-height: 1.6;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.empty-card {
  display: grid;
  min-height: 340px;
  place-items: center;
}

.load-more {
  display: grid;
  min-height: 76px;
  place-items: center;
  color: #9499a0;
  font-size: 13px;
}

@media (max-width: 1200px) {
  .feed-layout,
  .continue-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 860px) {
  .feed-layout,
  .continue-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .discover-head {
    align-items: flex-start;
    flex-direction: column;
  }

  .search-item {
    grid-template-columns: minmax(220px, 42%) minmax(0, 1fr);
  }
}

@media (max-width: 560px) {
  .home-shell {
    gap: 26px;
    padding-top: 20px;
  }

  .section-title-row h1 {
    font-size: 22px;
  }

  .feed-layout,
  .continue-grid,
  .search-item {
    grid-template-columns: 1fr;
  }

  .continue-card {
    grid-template-columns: 42% minmax(0, 1fr);
  }

  .feed-actions {
    width: 100%;
    overflow-x: auto;
  }

  .search-body {
    padding: 4px 2px 14px;
  }
}
</style>
