<template>
  <main class="page-shell creator-page">
    <div class="section-head">
      <div>
        <h1>创作者后台</h1>
        <p class="muted">查看作品状态、内容表现和近期增长。</p>
      </div>
      <el-button type="primary" @click="router.push('/creator/upload')">上传视频</el-button>
    </div>

    <section v-if="!authStore.isLoggedIn" class="soft-panel empty-panel">
      <el-empty description="登录后查看创作者后台">
        <el-button type="primary" @click="router.push('/login')">去登录</el-button>
      </el-empty>
    </section>

    <template v-else>
      <section class="stats-grid">
        <div v-for="item in stats" :key="item.label" class="soft-panel stat-card">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </div>
      </section>

      <section v-if="authStore.isCreator" class="soft-panel membership-panel">
        <div class="panel-head">
          <div>
            <h2>付费特别关注</h2>
            <p>设置月费和订阅权益。订阅成功的用户会自动成为特别关注。</p>
          </div>
          <el-switch v-model="membershipEnabled" active-text="开放订阅" inactive-text="暂不开放" />
        </div>
        <div class="membership-form">
          <label>
            <span>每月价格</span>
            <el-input-number v-model="membershipPriceYuan" :min="1" :max="1000" :precision="2" :step="1" controls-position="right" />
          </label>
          <label class="benefit-field">
            <span>权益说明</span>
            <el-input v-model="membershipBenefits" maxlength="200" show-word-limit placeholder="例如：特别关注标识、优先接收动态和直播提醒" />
          </label>
          <el-button type="primary" :loading="membershipSaving" @click="saveMembershipPlan">保存设置</el-button>
        </div>
        <p class="membership-note">当前使用项目内演示支付，不会产生真实扣款。</p>
      </section>

      <section class="soft-panel analytics-panel" v-loading="analyticsLoading">
        <div class="panel-head">
          <div>
            <h2>数据趋势</h2>
            <p>{{ analyticsScopeLabel }}的观看、收藏增长和账号开播次数。</p>
          </div>
          <div class="analytics-filters">
            <el-select
              v-model="selectedVideoId"
              class="video-filter"
              clearable
              placeholder="全部作品"
              @change="loadAnalytics"
            >
              <el-option v-for="video in videos" :key="video.id" :label="video.title" :value="video.id" />
            </el-select>
            <el-segmented v-model="analyticsDays" :options="dayOptions" @change="loadAnalytics" />
          </div>
        </div>

        <div class="range-summary">
          <div>
            <span>新增观看</span>
            <strong>{{ formatCount(analytics?.summary.rangeViews || 0) }}</strong>
          </div>
          <div>
            <span>净增收藏</span>
            <strong>{{ signedCount(analytics?.summary.rangeCollects || 0) }}</strong>
          </div>
          <div>
            <span>日均观看</span>
            <strong>{{ formatDecimal(analytics?.summary.averageDailyViews || 0) }}</strong>
          </div>
          <div>
            <span>开播次数</span>
            <strong>{{ formatCount(analytics?.summary.rangeStreams || 0) }}</strong>
          </div>
        </div>

        <div class="charts-grid">
          <article class="chart-card chart-wide">
            <div class="chart-head">
              <strong>观看与收藏曲线</strong>
              <span>每日新增</span>
            </div>
            <MetricLineChart
              :points="analyticsPoints"
              :series="trendSeries"
              ariaLabel="作品每日观看与收藏趋势"
            />
          </article>

          <article class="chart-card">
            <div class="chart-head">
              <strong>增长速度</strong>
              <span>观看与正向收藏之和</span>
            </div>
            <MetricLineChart
              :points="analyticsPoints"
              :series="growthSeries"
              ariaLabel="作品每日增长速度"
            />
          </article>

          <article class="chart-card">
            <div class="chart-head">
              <strong>推流次数</strong>
              <span>账号数据 · 手动与预约开播</span>
            </div>
            <MetricLineChart
              :points="analyticsPoints"
              :series="streamSeries"
              ariaLabel="每日推流次数"
            />
          </article>
        </div>

        <div v-if="analytics?.topVideos.length" class="top-works">
          <div class="subsection-head">
            <div>
              <strong>表现最佳作品</strong>
              <span>按累计观看排序</span>
            </div>
            <button v-if="selectedVideoId" type="button" @click="selectAnalyticsVideo(undefined)">查看全部作品</button>
          </div>
          <div class="top-work-list">
            <button
              v-for="(video, index) in analytics.topVideos"
              :key="video.id"
              type="button"
              class="top-work-row"
              :class="{ active: selectedVideoId === video.id }"
              @click="selectAnalyticsVideo(video.id)"
            >
              <span class="rank">{{ index + 1 }}</span>
              <img v-if="video.coverUrl" :src="mediaUrl(video.coverUrl)" :alt="video.title">
              <span v-else class="rank-cover">D</span>
              <span class="top-work-title">
                <strong>{{ video.title }}</strong>
                <small>{{ statusText(video.status) }}</small>
              </span>
              <span class="top-work-metric"><strong>{{ formatCount(video.viewCount) }}</strong><small>观看</small></span>
              <span class="top-work-metric"><strong>{{ formatCount(video.collectCount) }}</strong><small>收藏</small></span>
              <span class="top-work-metric"><strong>{{ formatCount(video.likeCount) }}</strong><small>点赞</small></span>
            </button>
          </div>
        </div>
        <p class="tracking-note">历史曲线从本功能启用后开始记录，作品累计数据仍完整保留。</p>
      </section>

      <section class="soft-panel works-panel">
        <div class="panel-head works-head">
          <div>
            <h2>作品管理</h2>
            <p>查看审核状态和单个作品表现。</p>
          </div>
          <div class="toolbar">
            <el-segmented v-model="status" :options="statusOptions" />
            <el-input v-model="keyword" class="search" clearable placeholder="搜索标题、简介或标签" />
          </div>
        </div>

        <el-table v-loading="loading" :data="filteredVideos" class="works-table" empty-text="暂无作品">
          <el-table-column label="作品" min-width="320">
            <template #default="{ row }">
              <div class="work-cell">
                <img v-if="row.coverUrl" :src="mediaUrl(row.coverUrl)" :alt="row.title">
                <div v-else class="thumb">D</div>
                <div class="work-info">
                  <strong>{{ row.title }}</strong>
                  <span>{{ row.description || '暂无简介' }}</span>
                </div>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="statusType(row.status)" effect="plain">{{ statusText(row.status) }}</el-tag>
            </template>
          </el-table-column>

          <el-table-column label="数据" min-width="220">
            <template #default="{ row }">
              <div class="metrics">
                <span>{{ formatCount(row.viewCount) }} 播放</span>
                <span>{{ formatCount(row.likeCount) }} 点赞</span>
                <span>{{ formatCount(row.collectCount) }} 收藏</span>
                <span>{{ formatCount(row.danmakuCount) }} 弹幕</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="信息" min-width="180">
            <template #default="{ row }">
              <div class="meta">
                <span>{{ formatDuration(row.duration) }}</span>
                <span>{{ normalizeTags(row.tags).join(' / ') || '无标签' }}</span>
                <span>{{ formatTime(row.createdAt) }}</span>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="操作" width="190" align="right">
            <template #default="{ row }">
              <div class="row-actions">
                <el-button size="small" @click="router.push(`/video/${row.id}`)">查看</el-button>
                <el-button size="small" type="danger" plain :loading="deletingId === row.id" @click="deleteVideo(row)">
                  删除
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </section>
    </template>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import MetricLineChart from '@/components/creator/MetricLineChart.vue'
import { videoApi } from '@/api/video'
import { membershipApi } from '@/api/membership'
import { useAuthStore } from '@/store/auth'
import type { CreatorAnalytics, VideoInfo, VideoStatus } from '@/types'
import { formatCount, formatDuration, formatTime, mediaUrl, normalizeTags } from '@/utils/format'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const analyticsLoading = ref(false)
const deletingId = ref<number>()
const videos = ref<VideoInfo[]>([])
const analytics = ref<CreatorAnalytics>()
const keyword = ref('')
const status = ref<VideoStatus | ''>('')
const analyticsDays = ref<7 | 30 | 90>(30)
const selectedVideoId = ref<number>()
const membershipEnabled = ref(false)
const membershipPriceYuan = ref(5)
const membershipBenefits = ref('')
const membershipSaving = ref(false)

const statusOptions = [
  { label: '全部', value: '' },
  { label: '待审核', value: 'pending' },
  { label: '已通过', value: 'approved' },
  { label: '已拒绝', value: 'rejected' },
]
const dayOptions = [
  { label: '近 7 天', value: 7 },
  { label: '近 30 天', value: 30 },
  { label: '近 90 天', value: 90 },
]
const trendSeries = [
  { key: 'views' as const, label: '观看', color: '#1677ff' },
  { key: 'collects' as const, label: '收藏', color: '#12a67a' },
]
const growthSeries = [{ key: 'growthSpeed' as const, label: '增长', color: '#e58a18' }]
const streamSeries = [{ key: 'streams' as const, label: '推流', color: '#e54867' }]

const analyticsPoints = computed(() => analytics.value?.points || [])
const analyticsScopeLabel = computed(() => {
  if (!selectedVideoId.value) return '全部作品'
  return videos.value.find(video => video.id === selectedVideoId.value)?.title || '当前作品'
})
const filteredVideos = computed(() => {
  const word = keyword.value.trim().toLowerCase()
  return videos.value.filter((video) => {
    if (status.value && video.status !== status.value) return false
    if (!word) return true
    const tags = normalizeTags(video.tags).join(' ')
    return [video.title, video.description, tags].some(value => value.toLowerCase().includes(word))
  })
})

const stats = computed(() => [
  { label: '作品总数', value: videos.value.length },
  { label: '待审核', value: videos.value.filter(item => item.status === 'pending').length },
  { label: '总播放', value: formatCount(sumMetric('viewCount')) },
  { label: '总收藏', value: formatCount(sumMetric('collectCount')) },
])

onMounted(() => {
  if (!authStore.isLoggedIn) return
  void Promise.all([load(), loadAnalytics(), authStore.isCreator ? loadMembershipPlan() : Promise.resolve()])
})

async function loadMembershipPlan() {
  const res = await membershipApi.myPlan()
  membershipEnabled.value = res.data.enabled
  membershipPriceYuan.value = res.data.priceCents / 100
  membershipBenefits.value = res.data.benefits
}

async function saveMembershipPlan() {
  membershipSaving.value = true
  try {
    await membershipApi.updateMyPlan({
      enabled: membershipEnabled.value,
      priceCents: Math.round(membershipPriceYuan.value * 100),
      benefits: membershipBenefits.value.trim(),
    })
    ElMessage.success('付费订阅设置已保存')
  } catch (error: any) {
    ElMessage.error(error.message || '付费订阅设置保存失败')
  } finally {
    membershipSaving.value = false
  }
}

async function load() {
  loading.value = true
  try {
    const res = await videoApi.myVideos({ page: 1, pageSize: 100 })
    videos.value = res.data.list
  } finally {
    loading.value = false
  }
}

async function loadAnalytics() {
  analyticsLoading.value = true
  try {
    const res = await videoApi.creatorAnalytics(analyticsDays.value, selectedVideoId.value)
    analytics.value = res.data
  } catch (error: any) {
    ElMessage.error(error.message || '创作数据加载失败')
  } finally {
    analyticsLoading.value = false
  }
}

function selectAnalyticsVideo(videoId?: number) {
  selectedVideoId.value = videoId
  void loadAnalytics()
}

function sumMetric(key: 'viewCount' | 'collectCount') {
  return videos.value.reduce((sum, item) => sum + item[key], 0)
}

function signedCount(value: number) {
  return `${value > 0 ? '+' : ''}${formatCount(value)}`
}

function formatDecimal(value: number) {
  return Intl.NumberFormat('zh-CN', { maximumFractionDigits: 1 }).format(value)
}

function statusText(value: VideoStatus) {
  return ({ pending: '待审核', approved: '已通过', rejected: '已拒绝' } as const)[value]
}

function statusType(value: VideoStatus) {
  return ({ pending: 'warning', approved: 'success', rejected: 'danger' } as const)[value]
}

async function deleteVideo(video: VideoInfo) {
  try {
    await ElMessageBox.confirm(`确定删除《${video.title}》吗？删除后不可恢复。`, '删除作品', {
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
    ElMessage.success('作品已删除')
    await Promise.all([load(), loadAnalytics()])
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

.section-head p,
.panel-head p {
  margin: 6px 0 0;
}

.panel-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.panel-head h2 {
  margin: 0;
  color: #101828;
  font-size: 18px;
}

.panel-head p,
.tracking-note {
  color: #667085;
  font-size: 13px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.stat-card {
  display: grid;
  gap: 8px;
  padding: 16px;
}

.stat-card span,
.range-summary span {
  color: #667085;
  font-size: 13px;
}

.stat-card strong {
  color: #101828;
  font-size: 24px;
}

.analytics-panel,
.membership-panel,
.works-panel {
  padding: 18px;
}

.membership-form {
  display: grid;
  grid-template-columns: 180px minmax(280px, 1fr) auto;
  align-items: end;
  gap: 16px;
  margin-top: 18px;
}

.membership-form label {
  display: grid;
  gap: 8px;
}

.membership-form label > span {
  color: #475467;
  font-size: 13px;
  font-weight: 700;
}

.membership-note {
  margin: 12px 0 0;
  color: #98a2b3;
  font-size: 12px;
}

.analytics-filters {
  display: flex;
  align-items: center;
  gap: 10px;
}

.video-filter {
  width: 210px;
}

.range-summary {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1px;
  margin: 18px 0;
  overflow: hidden;
  border: 1px solid #e4e7ec;
  border-radius: 6px;
  background: #e4e7ec;
}

.range-summary div {
  display: grid;
  gap: 6px;
  padding: 12px 14px;
  background: #fff;
}

.range-summary strong {
  color: #101828;
  font-size: 18px;
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.chart-card {
  min-width: 0;
  padding: 14px;
  border: 1px solid #e4e7ec;
  border-radius: 6px;
}

.chart-wide {
  grid-column: 1 / -1;
}

.chart-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.chart-head strong {
  color: #344054;
  font-size: 14px;
}

.chart-head span {
  color: #98a2b3;
  font-size: 12px;
}

.tracking-note {
  margin: 12px 0 0;
}

.top-works {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e4e7ec;
}

.subsection-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.subsection-head div {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.subsection-head strong {
  color: #344054;
  font-size: 14px;
}

.subsection-head span,
.subsection-head button {
  color: #667085;
  font-size: 12px;
}

.subsection-head button {
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
}

.subsection-head button:hover {
  color: #1677ff;
}

.top-work-list {
  display: grid;
}

.top-work-row {
  display: grid;
  grid-template-columns: 28px 96px minmax(160px, 1fr) repeat(3, minmax(68px, 92px));
  align-items: center;
  gap: 12px;
  width: 100%;
  min-height: 70px;
  padding: 8px 10px;
  border: 0;
  border-top: 1px solid #eef1f5;
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.top-work-row:first-child {
  border-top: 0;
}

.top-work-row:hover,
.top-work-row.active {
  background: #f7f9fc;
}

.top-work-row.active {
  box-shadow: inset 3px 0 #1677ff;
}

.rank {
  color: #98a2b3;
  font-size: 13px;
  text-align: center;
}

.top-work-row img,
.rank-cover {
  width: 96px;
  aspect-ratio: 16 / 9;
  border-radius: 4px;
  object-fit: cover;
}

.rank-cover {
  display: grid;
  place-items: center;
  background: #eef3fb;
  color: #165dff;
  font-weight: 700;
}

.top-work-title,
.top-work-metric {
  display: grid;
  min-width: 0;
  gap: 4px;
}

.top-work-title strong {
  overflow: hidden;
  color: #344054;
  font-size: 13px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.top-work-title small,
.top-work-metric small {
  color: #98a2b3;
  font-size: 11px;
}

.top-work-metric strong {
  color: #475467;
  font-size: 13px;
}

.works-head {
  align-items: center;
  margin-bottom: 16px;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search {
  width: 260px;
}

.works-table {
  width: 100%;
}

.work-cell {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 12px;
}

.work-cell img,
.thumb {
  width: 112px;
  aspect-ratio: 16 / 9;
  flex: 0 0 auto;
  border-radius: 6px;
  object-fit: cover;
}

.thumb {
  display: grid;
  place-items: center;
  background: #eef3fb;
  color: #165dff;
  font-weight: 800;
}

.work-info {
  display: grid;
  min-width: 0;
  gap: 6px;
}

.work-info strong,
.work-info span,
.meta span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.work-info span,
.meta,
.metrics {
  color: #667085;
  font-size: 13px;
}

.metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 12px;
}

.meta {
  display: grid;
  gap: 5px;
}

.row-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 1080px) {
  .works-head,
  .toolbar,
  .analytics-filters {
    align-items: stretch;
    flex-direction: column;
  }

  .search,
  .video-filter {
    width: 100%;
  }

  .membership-form {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .stats-grid,
  .range-summary,
  .charts-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .chart-card,
  .chart-wide {
    grid-column: 1 / -1;
  }

  .panel-head {
    align-items: stretch;
    flex-direction: column;
  }

  .top-work-row {
    grid-template-columns: 24px 80px minmax(130px, 1fr) repeat(2, 60px);
  }

  .top-work-row img,
  .rank-cover {
    width: 80px;
  }

  .top-work-metric:last-child {
    display: none;
  }
}

@media (max-width: 520px) {
  .stats-grid,
  .range-summary {
    grid-template-columns: 1fr;
  }

  .work-cell img,
  .thumb {
    width: 88px;
  }

  .top-work-row {
    grid-template-columns: 24px 72px minmax(100px, 1fr) 54px;
    gap: 8px;
  }

  .top-work-row img,
  .rank-cover {
    width: 72px;
  }

  .top-work-metric:nth-last-child(-n + 2) {
    display: none;
  }
}
</style>
