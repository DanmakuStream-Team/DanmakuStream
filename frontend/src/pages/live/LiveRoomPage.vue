<template>
  <div class="live-room">
    <a-spin :loading="loading">
      <div class="main-content" v-if="room">
        <!-- 直播播放器 -->
        <div class="player-section">
          <div class="player-wrapper">
            <div ref="playerContainer" class="player-container">
              <div class="live-badge">
                <a-tag color="red" size="small">直播中</a-tag>
                <span class="viewer-count">
                  <icon-eye /> {{ viewerCount }} 人在看
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧弹幕列表 -->
        <div class="danmaku-panel">
          <div class="panel-header">
            <span>弹幕列表</span>
            <a-tag color="arcoblue" size="small">{{ danmakuList.length }}</a-tag>
          </div>
          <div class="danmaku-list" ref="danmakuListEl">
            <div
              v-for="(d, i) in danmakuList"
              :key="i"
              class="danmaku-item"
              :style="{ color: d.color }"
            >
              <span class="danmaku-user">{{ d.username ?? '观众' }}:</span>
              <span class="danmaku-content">{{ d.content }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 未开播占位 -->
      <a-result
        v-else-if="!loading"
        status="warning"
        title="直播间不存在或尚未开播"
      >
        <template #extra>
          <a-button type="primary" @click="$router.back()">返回</a-button>
        </template>
      </a-result>
    </a-spin>

    <!-- 弹幕输入栏 -->
    <div class="danmaku-control" v-if="room">
      <a-input
        v-model="danmakuInput"
        placeholder="发送弹幕..."
        :max-length="100"
        @keydown.enter="sendDanmaku"
        :disabled="!authStore.isLoggedIn"
        class="danmaku-input"
      />
      <a-color-picker v-model="danmakuColor" size="small" />
      <a-tooltip :content="authStore.isLoggedIn ? '' : '请先登录'">
        <a-button
          type="primary"
          @click="sendDanmaku"
          :disabled="!authStore.isLoggedIn"
        >
          发送
        </a-button>
      </a-tooltip>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { DanmakuWebSocket } from '@/api/danmaku'
import type { Danmaku } from '@/types'

interface LiveRoom {
  id: number
  title: string
  streamUrl?: string
}

interface DanmakuItem extends Partial<Danmaku> {
  username?: string
  content: string
  color: string
}

const route = useRoute()
const authStore = useAuthStore()

const playerContainer = ref<HTMLElement>()
const danmakuListEl = ref<HTMLElement>()
const loading = ref(true)
const room = ref<LiveRoom | null>(null)
const viewerCount = ref(0)
const danmakuList = ref<DanmakuItem[]>([])
const danmakuInput = ref('')
const danmakuColor = ref('#FFFFFF')

let wsClient: DanmakuWebSocket | null = null

onMounted(async () => {
  const id = Number(route.params.id)

  // 模拟加载直播间信息（后续替换为真实 API）
  await new Promise(r => setTimeout(r, 300))
  room.value = { id, title: `直播间 ${id}` }
  loading.value = false

  // 建立 WebSocket 连接
  wsClient = new DanmakuWebSocket({
    roomId: id,
    token: authStore.token,
    onMessage: (d: Danmaku) => {
      danmakuList.value.push({
        content: d.content,
        color: d.color ?? '#FFFFFF',
      })
      scrollDanmakuList()
    },
    onViewerCount: (count: number) => {
      viewerCount.value = count
    },
  })
  wsClient.connect()
})

onUnmounted(() => {
  wsClient?.disconnect()
})

function sendDanmaku() {
  const text = danmakuInput.value.trim()
  if (!text || !authStore.isLoggedIn) return
  wsClient?.send(text, danmakuColor.value)
  // 乐观更新：立即显示自己发的弹幕
  danmakuList.value.push({
    content: text,
    color: danmakuColor.value,
    username: authStore.userInfo?.username ?? '我',
  })
  danmakuInput.value = ''
  scrollDanmakuList()
}

async function scrollDanmakuList() {
  await nextTick()
  if (danmakuListEl.value) {
    danmakuListEl.value.scrollTop = danmakuListEl.value.scrollHeight
  }
}
</script>

<style scoped>
.live-room {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
}

.main-content {
  display: flex;
  gap: 16px;
  flex: 1;
  min-height: 0;
}

/* 播放器 */
.player-section {
  flex: 1;
  min-width: 0;
}

.player-wrapper {
  background: #000;
  border-radius: 8px;
  overflow: hidden;
  aspect-ratio: 16/9;
  position: relative;
}

.player-container {
  width: 100%;
  height: 100%;
}

.live-badge {
  position: absolute;
  top: 12px;
  left: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  z-index: 10;
}

.viewer-count {
  color: #fff;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(0, 0, 0, 0.5);
  padding: 2px 8px;
  border-radius: 12px;
}

/* 弹幕列表面板 */
.danmaku-panel {
  width: 280px;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--color-border, #e5e6e7);
  border-radius: 8px;
  overflow: hidden;
}

.panel-header {
  padding: 10px 12px;
  font-weight: 600;
  border-bottom: 1px solid var(--color-border, #e5e6e7);
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--color-bg-2, #f7f8fa);
}

.danmaku-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.danmaku-item {
  font-size: 13px;
  line-height: 1.5;
  word-break: break-all;
}

.danmaku-user {
  font-weight: 600;
  margin-right: 4px;
  opacity: 0.8;
}

/* 弹幕输入栏 */
.danmaku-control {
  position: sticky;
  bottom: 0;
  background: var(--color-bg-1, #fff);
  padding: 12px;
  display: flex;
  gap: 8px;
  align-items: center;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.1);
  margin-top: 12px;
  border-radius: 8px 8px 0 0;
}

.danmaku-input {
  flex: 1;
}
</style>
