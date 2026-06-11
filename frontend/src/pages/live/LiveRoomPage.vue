<template>
  <main class="page-shell live-page">
    <div class="section-head">
      <div>
        <h1>{{ room?.title || `直播间 ${roomId}` }}</h1>
        <p class="muted">
          <button
            v-if="room?.owner"
            class="owner-link"
            type="button"
            @click="router.push(`/user/${room.owner?.id}`)"
          >
            主播：{{ room.owner.nickname || room.owner.username }}
          </button>
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

      <aside class="side-col">
        <div class="chat soft-panel">
        <div class="chat-head">
          <h2>实时弹幕</h2>
          <el-tag :type="connected ? 'success' : 'info'">{{ connected ? '已连接' : '未连接' }}</el-tag>
        </div>
        <div ref="messagesRef" class="messages">
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
        <div class="danmaku-type">
          <span
            v-for="t in DANMAKU_TYPES"
            :key="t.value"
            class="type-btn"
            :class="{ active: danmakuType === t.value }"
            @click="danmakuType = t.value"
          >
            {{ t.label }}
          </span>
        </div>
        </div>

        <div class="soft-panel recommend-panel">
          <h3>推荐直播间</h3>
          <div class="recommend-list">
            <article
              v-for="item in recommendedRooms"
              :key="item.id"
              class="recommend-item"
              @click="router.push(`/live/${item.id}`)"
            >
              <div class="recommend-cover">
                <img v-if="item.coverUrl" :src="mediaUrl(item.coverUrl)" :alt="item.title" />
                <span v-else>Live</span>
              </div>
              <div class="recommend-body">
                <strong>{{ item.title }}</strong>
                <span>{{ item.owner?.nickname || item.owner?.username || '主播' }} · {{ formatCount(item.viewerCount) }} 人观看</span>
              </div>
            </article>
            <el-empty v-if="!recommendedRooms.length" description="暂无推荐直播间" />
          </div>
        </div>
      </aside>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { VideoPlay } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { DanmakuWebSocket } from '@/api/danmaku'
import { liveApi } from '@/api/live'
import VideoPlayer from '@/components/common/VideoPlayer.vue'
import { useAuthStore } from '@/store/auth'
import type { Danmaku, LiveRoom } from '@/types'
import { formatCount, mediaUrl } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const roomId = computed(() => Number(route.params.id))
const connected = ref(false)
const loading = ref(false)
const ending = ref(false)
const room = ref<LiveRoom>()
const messagesRef = ref<HTMLElement>()
const text = ref('')
const color = ref('#FFFFFF')
const danmakuType = ref('scroll')
const viewerCount = ref(0)
const streamReady = ref(false)
const recommendedRooms = ref<LiveRoom[]>([])
const DANMAKU_COLORS = [
  '#FFFFFF', '#000000',
  '#FF5555', '#55FF55', '#5555FF', '#FFFF55', 
  '#FF55FF', '#55FFFF', '#FF8C00', '#FF69B4', 
  '#00CED1', '#FFD700', '#FF6347'
]
const DANMAKU_TYPES = [
  { label: '滚动', value: 'scroll' },
  { label: '顶部', value: 'top' },
  { label: '底部', value: 'bottom' }
]
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
  loadRecommendations()
  startStreamProbe()
  connectDanmaku()
})

onUnmounted(() => {
  ws?.disconnect()
  stopStreamProbe()
})

watch(() => route.params.id, async () => {
  ws?.disconnect()
  stopStreamProbe()
  connected.value = false
  streamReady.value = false
  messages.value = []
  await loadRoom()
  loadRecommendations()
  startStreamProbe()
  connectDanmaku()
})

async function loadRoom() {
  loading.value = true
  try {
    const res = await liveApi.detail(roomId.value)
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
    roomId: roomId.value,
    token: authStore.token,
    onMessage: (item) => {
      const shouldStickToBottom = isMessagesNearBottom()
      messages.value.push(item)
      if (shouldStickToBottom) scrollMessagesToBottom()
    },
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
  ws?.send(text.value.trim(), color.value, 'medium', danmakuType.value)
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
async function loadRecommendations() {
  try {
    const res = await liveApi.list({ page: 1, pageSize: 8 })
    recommendedRooms.value = res.data.list
      .filter(item => item.id !== roomId.value && item.status === 'live')
      .slice(0, 6)
  } catch {
    recommendedRooms.value = []
  }
}

function isMessagesNearBottom() {
  const el = messagesRef.value
  if (!el) return true
  return el.scrollHeight - el.scrollTop - el.clientHeight < 48
}

async function scrollMessagesToBottom() {
  await nextTick()
  const el = messagesRef.value
  if (!el) return
  el.scrollTop = el.scrollHeight
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

.owner-link {
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font: inherit;
}

.owner-link:hover {
  color: #00aeec;
}

.room-status {
  display: flex;
  gap: 8px;
  align-items: center;
}

.live-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  align-items: start;
  gap: 18px;
}

.stage {
  overflow: hidden;
  height: min(540px, calc(100vh - 190px));
  min-height: 420px;
  padding: 0;
  background: #0b1020;
}

.stage-placeholder {
  display: grid;
  height: 100%;
  min-height: 420px;
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

.side-col {
  display: grid;
  gap: 16px;
  min-width: 0;
}

.chat {
  display: grid;
  grid-template-rows: auto 1fr auto auto;
  height: min(540px, calc(100vh - 190px));
  min-height: 420px;
  min-width: 0;
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
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
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

.danmaku-type {
  display: flex;
  gap: 8px;
  padding-top: 12px;
  justify-content: center;
}

.type-btn {
  padding: 4px 12px;
  border-radius: 16px;
  background: #f0f2f5;
  color: #333;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.type-btn.active {
  background: #165dff;
  color: white;
}

.recommend-panel {
  display: grid;
  gap: 14px;
  padding: 16px;
}

.recommend-panel h3 {
  margin: 0;
}

.recommend-list {
  display: grid;
  gap: 12px;
}

.recommend-item {
  display: grid;
  grid-template-columns: 118px minmax(0, 1fr);
  gap: 10px;
  cursor: pointer;
}

.recommend-cover {
  display: grid;
  overflow: hidden;
  aspect-ratio: 16 / 9;
  place-items: center;
  border-radius: 8px;
  background: #f1f2f3;
  color: #00aeec;
  font-size: 12px;
  font-weight: 900;
}

.recommend-cover img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
}

.recommend-body {
  display: grid;
  align-content: start;
  gap: 6px;
  min-width: 0;
}

.recommend-body strong {
  overflow: hidden;
  color: #18191c;
  font-size: 14px;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recommend-item:hover .recommend-body strong {
  color: #00aeec;
}

.recommend-body span {
  color: #9499a0;
  font-size: 12px;
}

@media (max-width: 920px) {
  .live-grid {
    grid-template-columns: 1fr;
  }
}
</style>
