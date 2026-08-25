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
		<el-tag type="warning">热度 {{ formatCount(interaction.heat) }}</el-tag>
		<el-button v-if="canManageRoom" @click="router.push(`/live/studio/${roomId}`)">直播工作台</el-button>
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
          :live="true"
		  :danmakus="overlayMessages"
          @error="handlePlayerError"
        />
        <div v-else class="stage-placeholder">
          <el-icon><VideoPlay /></el-icon>
          <strong>{{ loading ? '正在加载直播间' : '等待直播流' }}</strong>
          <span>{{ streamUrl ? 'OBS 开始推流后，HLS 地址通常需要等待几秒才可播放' : '直播间还没有可播放地址' }}</span>
        </div>
        <LiveSuperChatOverlay :items="superChats" />
      </div>

      <aside class="side-col">
        <div class="chat soft-panel">
        <div class="chat-head">
		  <div>
			<h2>直播互动</h2>
			<span>{{ formatCount(interaction.likeCount) }} 赞 · {{ formatCount(interaction.giftValue) }} 助力</span>
		  </div>
          <el-tag :type="connected ? 'success' : 'info'">{{ connected ? '已连接' : '未连接' }}</el-tag>
        </div>
		<el-tabs v-model="sideTab" class="live-tabs" stretch>
		  <el-tab-pane label="弹幕" name="chat">
			<div ref="messagesRef" class="messages">
			  <div v-for="message in messages" :key="message.id" class="message">
				<button v-if="message.author" type="button" @click="router.push(`/user/${message.author?.id}`)">
				  {{ message.author.nickname || message.author.username }}
				</button>
				<span :style="{ color: message.color === '#FFFFFF' ? '#18191c' : message.color }">{{ message.content }}</span>
			  </div>
			  <el-empty v-if="!messages.length" description="暂无弹幕" />
			</div>
		  </el-tab-pane>
		  <el-tab-pane label="助力榜" name="support">
			<div class="rank-list">
			  <button v-for="(item, index) in interaction.supportRank" :key="item.userId" type="button" @click="router.push(`/user/${item.userId}`)">
				<span class="rank-index">{{ index + 1 }}</span>
				<el-avatar :size="30" :src="mediaUrl(item.user?.avatar || '')">{{ item.user?.nickname?.slice(0, 1) || 'U' }}</el-avatar>
				<strong>{{ item.user?.nickname || item.user?.username || '观众' }}</strong>
				<em>{{ formatCount(item.value) }}</em>
			  </button>
			  <el-empty v-if="!interaction.supportRank.length" description="送出礼物后登上助力榜" />
			</div>
		  </el-tab-pane>
		  <el-tab-pane label="热度榜" name="heat">
			<div class="rank-list heat-rank">
			  <button v-for="(item, index) in heatRanking" :key="item.room.id" type="button" @click="router.push(`/live/${item.room.id}`)">
				<span class="rank-index">{{ index + 1 }}</span>
				<div><strong>{{ item.room.title }}</strong><small>{{ item.room.owner?.nickname || '主播' }}</small></div>
				<em>{{ formatCount(item.heat) }}</em>
			  </button>
			  <el-empty v-if="!heatRanking.length" description="暂无热度排行" />
			</div>
		  </el-tab-pane>
		</el-tabs>
		<div v-if="latestGift" class="gift-flash">
		  <strong>{{ latestGift.user?.nickname || '观众' }}</strong>
		  送出 {{ latestGift.gift.name }} × {{ latestGift.count }}
		</div>
		<div v-if="room?.pinnedMessage" class="pinned-message"><strong>置顶</strong><span>{{ room.pinnedMessage }}</span></div>
		<div v-show="sideTab === 'chat'" class="chat-compose">
          <div class="send-box">
			<el-input v-model="text" placeholder="发送实时弹幕" @keyup.enter="send" />
			<el-button type="primary" @click="send">发送</el-button>
          </div>
		  <span class="chat-rule">{{ chatRuleText }}</span>
		  <details class="danmaku-settings">
			<summary>弹幕样式</summary>
			<div class="danmaku-colors">
			  <button
				v-for="c in DANMAKU_COLORS"
				:key="c"
				class="color-dot"
				:class="{ active: color === c }"
				:style="{ background: c }"
				type="button"
				:aria-label="`选择颜色 ${c}`"
				@click="color = c"
			  />
			</div>
			<div class="danmaku-type">
			  <button
				v-for="t in DANMAKU_TYPES"
				:key="t.value"
				class="type-btn"
				:class="{ active: danmakuType === t.value }"
				type="button"
				@click="danmakuType = t.value"
			  >
				{{ t.label }}
			  </button>
			</div>
		  </details>
		</div>
		<div class="interaction-actions">
		  <el-button :type="liked ? 'primary' : 'default'" :loading="liking" @click="toggleLike">
			{{ liked ? '已点赞' : '点赞' }} {{ formatCount(interaction.likeCount) }}
		  </el-button>
		  <el-button type="warning" plain @click="openGiftDialog">赠送礼物</el-button>
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

	<el-dialog v-model="giftVisible" title="礼物与 SC" width="420px">
	  <div class="gift-grid">
		<button v-for="gift in interaction.gifts" :key="gift.key" type="button" :class="{ active: selectedGift?.key === gift.key }" @click="selectedGift = gift">
		  <span>{{ gift.key === 'flower' ? '花' : gift.key === 'star' ? '星' : '火箭' }}</span>
		  <strong>{{ gift.name }}</strong>
		  <small>{{ gift.value }} 助力</small>
		</button>
	  </div>
	  <div class="gift-count"><span>数量</span><el-input-number v-model="giftCount" :min="1" :max="99" /></div>
	  <div class="sc-compose">
		<div><strong>SC 留言</strong><span>选填 · 填写后将在直播画面悬浮 {{ selectedSuperChatDuration }} 秒</span></div>
		<el-input v-model="superChatMessage" type="textarea" :rows="3" maxlength="200" show-word-limit placeholder="写下想让主播看到的话" />
	  </div>
	  <template #footer>
		<el-button @click="giftVisible = false">取消</el-button>
		<el-button type="primary" :loading="sendingGift" :disabled="!selectedGift" @click="sendGift">确认赠送</el-button>
	  </template>
	</el-dialog>
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
import LiveSuperChatOverlay from '@/components/live/LiveSuperChatOverlay.vue'
import { useAuthStore } from '@/store/auth'
import type { Danmaku, LiveGiftDefinition, LiveGiftEvent, LiveInteraction, LiveRoom } from '@/types'
import { formatCount, mediaUrl } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const roomId = computed(() => Number(route.params.id))
const connected = ref(false)
const loading = ref(false)
const ending = ref(false)
const liking = ref(false)
const sendingGift = ref(false)
const room = ref<LiveRoom>()
const messagesRef = ref<HTMLElement>()
const text = ref('')
const color = ref('#FFFFFF')
const danmakuType = ref('scroll')
const viewerCount = ref(0)
const streamReady = ref(false)
const recommendedRooms = ref<LiveRoom[]>([])
const sideTab = ref('chat')
const liked = ref(false)
const giftVisible = ref(false)
const giftCount = ref(1)
const superChatMessage = ref('')
const selectedGift = ref<LiveGiftDefinition>()
const latestGift = ref<LiveGiftEvent>()
const heatRanking = ref<Array<{ room: LiveRoom; heat: number }>>([])
const interaction = ref<LiveInteraction>({ likeCount: 0, giftValue: 0, heat: 0, gifts: [], supportRank: [], superChats: [] })
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
const overlayMessages = ref<Danmaku[]>([])
let ws: DanmakuWebSocket | null = null
let streamTimer: ReturnType<typeof setInterval> | null = null
let lastPlayerErrorAt = 0

const streamUrl = computed(() => room.value?.streamUrl || room.value?.playUrl || '')
const superChats = computed(() => interaction.value.superChats || [])
const selectedSuperChatDuration = computed(() => superChatDuration((selectedGift.value?.value || 0) * giftCount.value))
const canManageRoom = computed(() => {
  if (!room.value || !authStore.userInfo) return false
  return room.value.ownerId === authStore.userInfo.id || authStore.isAdmin
})
const chatRuleText = computed(() => {
  const access = room.value?.chatMode === 'members'
    ? '仅付费订阅者可发言'
    : room.value?.chatMode === 'followers' ? '仅关注者可发言' : '登录后可发言'
  return room.value?.slowModeSeconds ? `${access} · 每 ${room.value.slowModeSeconds} 秒一条` : access
})

onMounted(async () => {
  await loadRoom()
	await Promise.all([loadRecommendations(), loadInteraction(), loadHeatRanking(), loadCurrentDanmaku(), loadLikeStatus()])
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
	overlayMessages.value = []
  await loadRoom()
	await Promise.all([loadRecommendations(), loadInteraction(), loadHeatRanking(), loadCurrentDanmaku(), loadLikeStatus()])
  startStreamProbe()
  connectDanmaku()
})

async function loadRoom() {
  loading.value = true
  try {
    const res = await liveApi.detail(roomId.value)
    room.value = res.data
    viewerCount.value = res.data.viewerCount || 0
	interaction.value.likeCount = res.data.likeCount || 0
	interaction.value.giftValue = res.data.giftValue || 0
	interaction.value.heat = res.data.heat || 0
  } catch (error: any) {
    console.warn('直播间加载失败', error)
  } finally {
    loading.value = false
  }
}

function connectDanmaku() {
  ws = new DanmakuWebSocket({
    roomId: roomId.value,
	token: authStore.token || '',
	onConnectionChange: (value) => { connected.value = value },
    onMessage: (item) => {
	  if (messages.value.some((message) => message.id === item.id)) return
      const shouldStickToBottom = isMessagesNearBottom()
      messages.value.push(item)
	  overlayMessages.value.push(item)
      if (shouldStickToBottom) scrollMessagesToBottom()
    },
    onViewerCount: (count) => {
      viewerCount.value = count
	  interaction.value.heat = count * 10 + interaction.value.likeCount * 2 + interaction.value.giftValue
      connected.value = true
    },
	onLike: (payload) => {
	  interaction.value.likeCount = payload.likeCount
	  interaction.value.heat = payload.heat
	  if (payload.userId === authStore.userInfo?.id) liked.value = payload.liked
	},
	onGift: (payload) => {
	  latestGift.value = payload
	  interaction.value.giftValue = payload.giftValue
	  interaction.value.heat = payload.heat
	  interaction.value.supportRank = payload.supportRank
	  if (payload.message && payload.displaySeconds) {
		interaction.value.superChats = [...interaction.value.superChats.slice(-19), {
		  id: payload.id || Date.now(), user: payload.user, gift: payload.gift,
		  count: payload.count, value: payload.value, message: payload.message,
		  displaySeconds: payload.displaySeconds, createdAt: payload.createdAt || new Date().toISOString(),
		}]
	  }
	  window.setTimeout(() => {
		if (latestGift.value === payload) latestGift.value = undefined
	  }, 4500)
	},
	onChatError: (payload) => ElMessage.warning(payload.message),
  })
  ws.connect()
}

async function loadCurrentDanmaku() {
	try {
	  messages.value = (await liveApi.liveDanmaku(roomId.value)).data
	  scrollMessagesToBottom()
	} catch {
	  messages.value = []
	}
}

async function loadLikeStatus() {
	if (!authStore.isLoggedIn) {
	  liked.value = false
	  return
	}
	try {
	  liked.value = (await liveApi.likeStatus(roomId.value)).data.liked
	} catch {
	  liked.value = false
	}
}

async function loadInteraction() {
	try {
	  interaction.value = (await liveApi.interaction(roomId.value)).data
	  selectedGift.value = interaction.value.gifts[0]
	} catch {
	  interaction.value = { likeCount: 0, giftValue: 0, heat: 0, gifts: [], supportRank: [], superChats: [] }
	}
}

async function loadHeatRanking() {
	try {
	  heatRanking.value = (await liveApi.heatRanking()).data
	} catch {
	  heatRanking.value = []
	}
}

async function toggleLike() {
	if (!authStore.isLoggedIn) {
	  ElMessage.warning('请先登录后点赞')
	  router.push({ path: '/login', query: { redirect: route.fullPath } })
	  return
	}
	liking.value = true
	try {
	  const res = await liveApi.like(roomId.value)
	  liked.value = res.data.liked
	  interaction.value.likeCount = res.data.likeCount
	  interaction.value.heat = res.data.heat
	} catch {
	  ElMessage.error('点赞失败')
	} finally {
	  liking.value = false
	}
}

function openGiftDialog() {
	if (!authStore.isLoggedIn) {
	  ElMessage.warning('请先登录后赠送礼物')
	  router.push({ path: '/login', query: { redirect: route.fullPath } })
	  return
	}
	selectedGift.value ||= interaction.value.gifts[0]
	giftVisible.value = true
}

async function sendGift() {
	if (!selectedGift.value) return
	sendingGift.value = true
	try {
	  const res = await liveApi.sendGift(roomId.value, selectedGift.value.key, giftCount.value, superChatMessage.value.trim())
	  interaction.value.giftValue = res.data.giftValue
	  interaction.value.heat = res.data.heat
	  interaction.value.supportRank = res.data.supportRank
	  giftVisible.value = false
	  giftCount.value = 1
	  superChatMessage.value = ''
	  ElMessage.success(`已送出${res.data.gift.name}`)
	} catch {
	  ElMessage.error('礼物发送失败')
	} finally {
	  sendingGift.value = false
	}
}

function superChatDuration(value: number) {
	if (value >= 1000) return 120
	if (value >= 500) return 90
	if (value >= 200) return 60
	if (value >= 50) return 30
	return 15
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
  const sent = ws?.send(text.value.trim(), color.value, 'medium', danmakuType.value)
  if (!sent) {
    ElMessage.warning('实时弹幕正在重连，请稍后再试')
    return
  }
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
  position: relative;
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
	grid-template-rows: auto minmax(0, 1fr) auto auto auto;
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

.chat-head > div { display: grid; gap: 4px; }
.chat-head span { color: #9499a0; font-size: 12px; }
.live-tabs { min-height: 0; overflow: hidden; }
.live-tabs :deep(.el-tabs__header) { margin: 8px 0 0; }
.live-tabs :deep(.el-tabs__content), .live-tabs :deep(.el-tab-pane) { height: 100%; min-height: 0; }
.live-tabs :deep(.el-tabs__content) { overflow: hidden; }

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
	display: flex;
	gap: 6px;
	padding: 7px 8px;
  border-radius: 8px;
  background: #f7f9fc;
}
.message button { flex: none; padding: 0; border: 0; background: transparent; color: #00aeec; font: inherit; font-weight: 700; cursor: pointer; }
.message span { min-width: 0; overflow-wrap: anywhere; }
.chat-compose { display: grid; gap: 0; }
.interaction-actions { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; padding-top: 10px; border-top: 1px solid #eef0f3; }
.interaction-actions .el-button { margin: 0; }
.gift-flash { margin: 0 0 8px; padding: 8px 10px; border-left: 3px solid #f7ba2a; background: #fff8e8; color: #6c4d00; font-size: 13px; }
.pinned-message { display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 8px; margin-bottom: 8px; padding: 8px 10px; border-left: 3px solid #00aeec; background: #f2fbff; color: #344054; font-size: 12px; }
.pinned-message strong { color: #00aeec; }
.pinned-message span { overflow-wrap: anywhere; }
.chat-rule { padding-top: 6px; color: #9499a0; font-size: 11px; }
.rank-list { display: grid; align-content: start; gap: 4px; height: 100%; overflow-y: auto; padding: 10px 0; }
.rank-list > button { display: grid; grid-template-columns: 24px 30px minmax(0, 1fr) auto; align-items: center; gap: 8px; min-height: 42px; padding: 6px; border: 0; border-radius: 4px; background: transparent; text-align: left; cursor: pointer; }
.rank-list > button:hover { background: #f6f7f8; }
.rank-index { color: #9499a0; font-size: 12px; text-align: center; }
.rank-list button:nth-child(-n+3) .rank-index { color: #f59a23; font-weight: 800; }
.rank-list strong { overflow: hidden; color: #18191c; font-size: 13px; text-overflow: ellipsis; white-space: nowrap; }
.rank-list em { color: #f59a23; font-size: 12px; font-style: normal; font-weight: 700; }
.heat-rank > button { grid-template-columns: 24px minmax(0, 1fr) auto; }
.heat-rank div { display: grid; min-width: 0; gap: 2px; }
.heat-rank small { color: #9499a0; font-size: 11px; }
.gift-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
.gift-grid button { display: grid; justify-items: center; gap: 5px; padding: 14px 8px; border: 1px solid #e3e5e7; border-radius: 6px; background: #fff; cursor: pointer; }
.gift-grid button.active { border-color: #00aeec; background: #f2fbff; }
.gift-grid button span { display: grid; width: 42px; height: 42px; place-items: center; border-radius: 50%; background: #fff1c8; color: #bd7900; font-weight: 800; }
.gift-grid button strong { font-size: 14px; }
.gift-grid button small { color: #9499a0; }
.gift-count { display: flex; align-items: center; justify-content: space-between; margin-top: 18px; }
.sc-compose { display: grid; gap: 9px; margin-top: 18px; }
.sc-compose > div { display: grid; gap: 3px; }
.sc-compose strong { font-size: 13px; }
.sc-compose span { color: #9499a0; font-size: 11px; }

.send-box {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px;
}

.danmaku-settings {
  margin-top: 8px;
  border-top: 1px solid #eef0f3;
}

.danmaku-settings summary {
  padding: 8px 0 2px;
  color: #9499a0;
  cursor: pointer;
  font-size: 11px;
  list-style-position: inside;
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
  padding: 0;
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
  border: 0;
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

  .stage {
    width: 100%;
    height: auto;
    min-height: 0;
    aspect-ratio: 16 / 9;
  }

  .stage-placeholder {
    min-height: 0;
  }

  .room-status {
    flex-wrap: wrap;
  }
}
</style>
