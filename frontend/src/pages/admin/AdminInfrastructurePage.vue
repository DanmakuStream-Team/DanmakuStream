<template>
  <main class="page-shell admin-page">
    <div class="section-head">
      <div>
        <h1>服务器监控</h1>
        <p class="muted">存储、流量和在线状态预警。</p>
      </div>
      <el-button :loading="loading" @click="load">刷新</el-button>
    </div>

    <section v-loading="loading" class="metrics-grid">
      <div class="soft-panel metric-card" :class="{ warning: metrics?.storage.warning, critical: metrics?.storage.critical }">
        <div class="metric-title">
          <strong>存储空间</strong>
          <el-tag :type="storageTagType">{{ storageStatus }}</el-tag>
        </div>
        <el-progress
          type="dashboard"
          :percentage="Math.round(metrics?.storage.usagePercent || 0)"
          :color="storageColor"
        />
        <div class="metric-meta">
          <span>已用 {{ formatBytes(metrics?.storage.usedBytes) }}</span>
          <span>剩余 {{ formatBytes(metrics?.storage.freeBytes) }}</span>
          <span>{{ metrics?.storage.path || '-' }}</span>
        </div>
      </div>

      <div class="soft-panel metric-card">
        <div class="metric-title">
          <strong>带宽与流量</strong>
          <el-tag type="success">应用层统计</el-tag>
        </div>
        <div class="traffic-grid">
          <div>
            <span>今日下行</span>
            <strong>{{ formatBytes(metrics?.traffic.todayDownBytes) }}</strong>
          </div>
          <div>
            <span>本月下行</span>
            <strong>{{ formatBytes(metrics?.traffic.monthDownBytes) }}</strong>
          </div>
        </div>
        <p>由 Go 后端 middleware 统计 Docker Compose 部署下的视频、媒体和 API 响应流量。</p>
      </div>

      <div class="soft-panel metric-card">
        <div class="metric-title">
          <strong>在线与并发</strong>
          <el-tag type="success">{{ metrics?.online.liveRoomCount || 0 }} 个直播间</el-tag>
        </div>
        <div class="online-count">{{ metrics?.online.current || 0 }}</div>
        <div class="metric-meta">
          <span>当前在线</span>
          <span>最高并发 {{ metrics?.online.highestConcurrent || 0 }}</span>
        </div>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi, type InfrastructureMetrics } from '@/api/admin'

const loading = ref(false)
const metrics = ref<InfrastructureMetrics>()
const storageColor = computed(() => {
  if (metrics.value?.storage.critical) return '#f56c6c'
  if (metrics.value?.storage.warning) return '#e6a23c'
  return '#fb7299'
})
const storageTagType = computed(() => {
  if (metrics.value?.storage.critical) return 'danger'
  if (metrics.value?.storage.warning) return 'warning'
  return 'success'
})
const storageStatus = computed(() => {
  if (metrics.value?.storage.critical) return '容量危险'
  if (metrics.value?.storage.warning) return '容量预警'
  return '正常'
})

onMounted(load)

async function load() {
  loading.value = true
  try {
    const res = await adminApi.infrastructure()
    metrics.value = res.data
  } catch {
    ElMessage.error('监控数据加载失败')
  } finally {
    loading.value = false
  }
}

function formatBytes(value = 0) {
  if (!value) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = value
  let index = 0
  while (size >= 1024 && index < units.length - 1) {
    size /= 1024
    index += 1
  }
  return `${size.toFixed(size >= 10 ? 1 : 2)} ${units[index]}`
}
</script>

<style scoped>
.admin-page {
  display: grid;
  gap: 18px;
}

.section-head p {
  margin: 8px 0 0;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.metric-card {
  display: grid;
  align-content: start;
  gap: 18px;
  min-height: 280px;
  padding: 22px;
}

.metric-card.critical {
  border-color: rgba(245, 108, 108, 0.38);
}

.metric-card.warning {
  border-color: rgba(230, 162, 60, 0.38);
}

.metric-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.metric-title strong {
  font-size: 18px;
  font-weight: 900;
}

.metric-meta {
  display: grid;
  gap: 8px;
  color: #667085;
  font-size: 13px;
}

.traffic-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.traffic-grid div {
  display: grid;
  gap: 8px;
  padding: 16px;
  border-radius: 8px;
  background: #f6f7f8;
}

.traffic-grid span,
.metric-card p {
  margin: 0;
  color: #667085;
  font-size: 13px;
  line-height: 1.7;
}

.traffic-grid strong,
.online-count {
  color: #18191c;
  font-size: 28px;
  font-weight: 900;
}

.online-count {
  color: #fb7299;
  font-size: 56px;
}

@media (max-width: 980px) {
  .metrics-grid {
    grid-template-columns: 1fr;
  }
}
</style>
