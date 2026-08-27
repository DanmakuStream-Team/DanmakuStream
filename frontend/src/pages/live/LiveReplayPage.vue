<template>
  <main class="page-shell replay-page">
    <section v-loading="loading" class="replay-layout">
      <div class="main-column">
        <div class="stage soft-panel">
          <VideoPlayer
            v-if="replay?.status === 'ready' && replay.replayUrl"
            ref="playerRef"
            :url="replay.replayUrl"
            :poster="replay.coverUrl"
            :danmakus="danmakus"
            @error="ElMessage.error($event)"
          />
          <div v-else class="replay-placeholder">
            <el-icon><VideoPlay /></el-icon>
            <strong>{{ replay?.status === 'unavailable' ? '回放不可用' : '回放正在生成' }}</strong>
            <span>{{ replay?.status === 'unavailable' ? '该场直播没有留下可播放文件' : '直播结束后需要等待几秒整理播放列表' }}</span>
            <el-button v-if="replay?.status === 'processing'" :loading="refreshing" @click="refreshReplay">刷新状态</el-button>
          </div>
        </div>

        <section v-if="replay" class="replay-info">
          <div>
            <el-tag type="info">直播回放</el-tag>
            <h1>{{ replay.title }}</h1>
            <p>{{ replay.startedAt }} 开播 · {{ formatDuration(replay.duration) }} · 峰值 {{ formatCount(replay.viewerPeak) }} 人观看</p>
          </div>
          <button v-if="replay.owner" class="owner-card" type="button" @click="router.push(`/user/${replay.owner?.id}`)">
            <el-avatar :size="44" :src="mediaUrl(replay.owner.avatar)">{{ replay.owner.nickname.slice(0, 1) }}</el-avatar>
            <span><strong>{{ replay.owner.nickname }}</strong><small>查看主播主页</small></span>
          </button>
        </section>
      </div>

      <aside class="side-column soft-panel">
        <div class="side-head">
          <h2>历史弹幕</h2>
          <span>{{ danmakus.length }} 条</span>
        </div>
        <div class="danmaku-list">
          <button v-for="item in danmakus" :key="item.id" type="button" @click="seekTo(item.time)">
            <time>{{ formatDuration(item.time) }}</time>
            <span :style="{ color: readableColor(item.color) }">{{ item.content }}</span>
          </button>
          <el-empty v-if="!danmakus.length" :image-size="72" description="这场直播暂无弹幕" />
        </div>
      </aside>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { VideoPlay } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { liveApi } from '@/api/live'
import VideoPlayer from '@/components/common/VideoPlayer.vue'
import type { Danmaku, LiveReplay } from '@/types'
import { formatCount, formatDuration, mediaUrl } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const replayId = computed(() => Number(route.params.id))
const playerRef = ref<InstanceType<typeof VideoPlayer>>()
const replay = ref<LiveReplay>()
const danmakus = ref<Danmaku[]>([])
const loading = ref(false)
const refreshing = ref(false)
let statusTimer: number | undefined

onMounted(load)
onBeforeUnmount(stopStatusPolling)
watch(() => route.params.id, load)

async function load() {
  stopStatusPolling()
  loading.value = true
  try {
    const [detail, items] = await Promise.all([
      liveApi.replayDetail(replayId.value),
      liveApi.replayDanmaku(replayId.value),
    ])
    replay.value = detail.data
    danmakus.value = items.data
    if (replay.value.status === 'processing') startStatusPolling()
  } finally {
    loading.value = false
  }
}

async function refreshReplay() {
  refreshing.value = true
  try {
    replay.value = (await liveApi.replayDetail(replayId.value)).data
    if (replay.value.status !== 'processing') stopStatusPolling()
  } finally {
    refreshing.value = false
  }
}

function startStatusPolling() {
  stopStatusPolling()
  statusTimer = window.setInterval(() => void refreshReplay(), 4000)
}

function stopStatusPolling() {
  if (statusTimer) window.clearInterval(statusTimer)
  statusTimer = undefined
}

function seekTo(time: number) {
  playerRef.value?.seek(time)
  void playerRef.value?.play()
}

function readableColor(color: string) {
  return color?.toUpperCase() === '#FFFFFF' ? '#475467' : color
}
</script>

<style scoped>
.replay-page { padding-top: 24px; }
.replay-layout { display: grid; grid-template-columns: minmax(0, 1fr) 320px; align-items: start; gap: 18px; }
.main-column { display: grid; min-width: 0; gap: 18px; }
.stage { overflow: hidden; min-height: 420px; padding: 0; background: #05070d; }
.replay-placeholder { display: grid; min-height: 420px; place-items: center; align-content: center; gap: 10px; color: #fff; text-align: center; }
.replay-placeholder .el-icon { font-size: 58px; }
.replay-placeholder strong { font-size: 23px; }
.replay-placeholder span { color: rgb(255 255 255 / 68%); font-size: 13px; }
.replay-info { display: flex; align-items: center; justify-content: space-between; gap: 20px; }
.replay-info h1 { margin: 10px 0 7px; color: #18191c; font-size: 25px; }
.replay-info p { margin: 0; color: #9499a0; font-size: 13px; }
.owner-card { display: flex; min-width: 190px; align-items: center; gap: 10px; padding: 10px; border: 0; background: transparent; cursor: pointer; text-align: left; }
.owner-card span { display: grid; gap: 4px; }
.owner-card strong { color: #18191c; }
.owner-card small { color: #9499a0; }
.side-column { display: grid; height: min(560px, calc(100vh - 130px)); grid-template-rows: auto minmax(0, 1fr); padding: 16px; }
.side-head { display: flex; align-items: center; justify-content: space-between; padding-bottom: 12px; border-bottom: 1px solid #edf0f3; }
.side-head h2 { margin: 0; font-size: 17px; }
.side-head span { color: #9499a0; font-size: 12px; }
.danmaku-list { overflow-y: auto; padding-top: 8px; }
.danmaku-list > button { display: grid; width: 100%; grid-template-columns: 44px minmax(0, 1fr); gap: 8px; padding: 8px 4px; border: 0; background: transparent; cursor: pointer; text-align: left; }
.danmaku-list > button:hover { background: #f6f8fa; }
.danmaku-list time { color: #98a2b3; font-size: 11px; }
.danmaku-list span { overflow: hidden; font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }
@media (max-width: 900px) { .replay-layout { grid-template-columns: 1fr; } .side-column { height: 380px; } .replay-info { align-items: flex-start; flex-direction: column; } }
</style>
