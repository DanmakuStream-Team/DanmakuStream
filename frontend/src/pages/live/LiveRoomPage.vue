<template>
  <main class="page-shell live-page">
    <div class="section-head">
      <div>
        <h1>{{ room?.title || `直播间 ${roomId}` }}</h1>
        <p class="muted">
          <span v-if="room?.owner">主播：{{ room.owner.nickname || room.owner.username }}</span>
          <span v-if="room?.startedAt"> · 开播时间：{{ room.startedAt }}</span>
        </p>
      </div>
      <div class="room-status">
        <el-tag :type="room?.status === 'live' ? 'success' : 'info'">
          {{ room?.status === 'live' ? '直播中' : '未开播' }}
        </el-tag>
        <el-tag type="info">{{ viewerCount }} 人观看</el-tag>
        <el-button
          v-if="canManageRoom"
          type="danger"
          :loading="ending"
          @click="endLiveRoom"
        >
          结束直播
        </el-button>
      </div>
    </div>

    <section class="live-grid">
      <div class="stage soft-panel">
        <VideoPlayer
          v-if="streamReady && streamUrl"
          :url="streamUrl"
          :poster="room?.coverUrl"
          :autoplay="true"
          :danmakus="messages"
          @error="handlePlayerError"
        />
        <div v-else class="stage-placeholder">
          <el-icon><VideoPlay /></el-icon>
          <strong>{{ loading ? '正在加载直播间' : '等待直播流' }}</strong>
          <span>{{ streamUrl ? 'OBS 开始推流后，HLS 地址通常需要等待几秒才可播放' : '直播间还没有可播放地址' }}</span>
        </div>
      </div>

      <aside class="chat soft-panel">
        <div class="chat-head">
          <h2>实时弹幕</h2>
          <el-tag :type="connected ? 'success' : 'info'">{{ connected ? '已连接' : '未连接' }}</el-tag>
        </div>
        <div class="messages">
          <div v-for="message in messages" :key="message.id" class="message" :style="{ color: message.color }">
            {{ message.content }}
          </div>
          <el-empty v-if="!messages.length" description="暂无弹幕" />
        </div>
        <div class="send-box">
          <el-input v-model="text" placeholder="发送弹幕" @keyup.enter="send" />
          <el-button type="primary" @click="send">发送</el-button>
        </div>
        <div class="danmaku-colors">
          <span
            v-for="c in DANMAKU_COLORS"
            :key="c"
            class="color-dot"
            :class="{ active: color === c }"
            :style="{ background: c }"
            @click="color = c"
          />
        </div>
      </aside>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { VideoPlay } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { DanmakuWebSocket } from '@/api/danmaku'
import { liveApi } from '@/api/live'
import VideoPlayer from '@/components/common/VideoPlayer.vue'
import { useAuthStore } from '@/store/auth'
import type { Danmaku, LiveRoom } from '@/types'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const roomId = Number(route.params.id)
const connected = ref(false)
const loading = ref(false)
const ending = ref(false)
const room = ref<LiveRoom>()
const text = ref('')
const color = ref('#FFFFFF')
const viewerCount = ref(0)
const streamReady = ref(false)
const DANMAKU_COLORS = ['#FFFFFF', '#FF5555', '#55FF55', '#5555FF', '#FFFF55', '#FF55FF', '#55FFFF', '#FF8C00', '#FF69B4', '#00CED1', '#FFD700', '#FF6347']
const messages = ref<Danmaku[]>([])
let ws: DanmakuWebSocket | null = null
let streamTimer: ReturnType<typeof setInterval> | null = null
let lastPlayerErrorAt = 0

const streamUrl = computed(() => room.value?.streamUrl || room.value?.playUrl || '')
const canManageRoom = computed(() => {
  if (!room.value || !authStore.userInfo) return false
  return room.value.ownerId === authStore.userInfo.id || authStore.isAdmin
})

onMounted(async () => {
  await loadRoom()
  startStreamProbe()
  connectDanmaku()
})

onUnmounted(() => {
  ws?.disconnect()
  stopStreamProbe()
})

async function loadRoom() {
  loading.value = true
  try {
    const res = await liveApi.detail(roomId)
    room.value = res.data
    viewerCount.value = res.data.viewerCount || 0
  } catch (error: any) {
    ElMessage.error(error.message || '直播间加载失败')
  } finally {
    loading.value = false
  }
}

function connectDanmaku() {
  if (!authStore.isLoggedIn || !authStore.token) return
  ws = new DanmakuWebSocket({
    roomId,
    token: authStore.token,
    onMessage: (item) => messages.value.push(item),
    onViewerCount: (count) => {
      viewerCount.value = count
      connected.value = true
    },
  })
  ws.connect()
}

function startStreamProbe() {
  stopStreamProbe()
  checkStreamReady()
  streamTimer = setInterval(checkStreamReady, 3000)
}

function stopStreamProbe() {
  if (!streamTimer) return
  clearInterval(streamTimer)
  streamTimer = null
}

async function endLiveRoom() {
  if (!room.value) return
  ending.value = true
  try {
    await liveApi.end(room.value.id)
    ws?.disconnect()
    stopStreamProbe()
    ElMessage.success('直播已结束')
    router.push('/live')
  } catch {
    ElMessage.error('结束直播失败')
  } finally {
    ending.value = false
  }
}

async function checkStreamReady() {
  if (!streamUrl.value) return
  try {
    const res = await fetch(`${streamUrl.value}?_t=${Date.now()}`, { cache: 'no-store' })
    const text = await res.text()
    if (res.ok && text.includes('#EXTM3U')) {
      streamReady.value = true
      stopStreamProbe()
      return
    }
  } catch {
    // HLS is not ready yet. Keep polling quietly.
  }
  streamReady.value = false
}

function send() {
  if (!authStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  if (!text.value.trim()) return
  ws?.send(text.value.trim(), color.value)
  messages.value.push({
    id: Date.now(),
    videoId: roomId,
    userId: authStore.userInfo?.id || 0,
    content: text.value.trim(),
    time: 0,
    color: color.value,
    fontSize: 'medium',
    type: 'scroll',
  })
  text.value = ''
}

function handlePlayerError() {
  streamReady.value = false
  startStreamProbe()
  const now = Date.now()
  if (now - lastPlayerErrorAt > 5000) {
    ElMessage.warning('直播流暂未就绪，请确认 OBS 已开始推流')
    lastPlayerErrorAt = now
  }
}
</script>

<style scoped>
.live-page {
  display: grid;
  gap: 18px;
}

.section-head p {
  margin: 8px 0 0;
}

.room-status {
  display: flex;
  gap: 8px;
  align-items: center;
}

.live-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 18px;
}

.stage {
  overflow: hidden;
  min-height: 540px;
  padding: 0;
  background: #0b1020;
}

.stage-placeholder {
  display: grid;
  min-height: 540px;
  place-items: center;
  align-content: center;
  gap: 12px;
  color: #fff;
  text-align: center;
}

.stage-placeholder .el-icon {
  font-size: 62px;
}

.stage-placeholder strong {
  font-size: 26px;
}

.stage-placeholder span {
  color: rgba(255, 255, 255, 0.74);
}

.chat {
  display: grid;
  grid-template-rows: auto 1fr auto auto;
  min-height: 540px;
  padding: 16px;
}

.chat-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.chat-head h2 {
  margin: 0;
}

.messages {
  display: grid;
  align-content: start;
  gap: 10px;
  overflow: auto;
  padding: 16px 0;
}

.message {
  padding: 9px 10px;
  border-radius: 8px;
  background: #f7f9fc;
}

.send-box {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px;
}

.danmaku-colors {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
  padding-top: 12px;
}

.color-dot {
  width: 22px;
  height: 22px;
  border: 2px solid transparent;
  border-radius: 50%;
  cursor: pointer;
  transition: border-color 0.15s, transform 0.15s;
}

.color-dot:hover {
  transform: scale(1.15);
}

.color-dot.active {
  border-color: #165dff;
  transform: scale(1.1);
}

@media (max-width: 920px) {
  .live-grid {
    grid-template-columns: 1fr;
  }
}
</style>
