<template>
  <main class="page-shell live-page">
    <div class="section-head">
      <div>
        <h1>直播间 {{ roomId }}</h1>
        <p class="muted">后端直播 API 暂未实现，这里只保留 WebSocket 弹幕入口。</p>
      </div>
      <el-tag :type="connected ? 'success' : 'info'">{{ connected ? '已连接' : '未连接' }}</el-tag>
    </div>

    <section class="live-grid">
      <div class="stage soft-panel">
        <div class="stage-content">
          <el-icon><VideoPlay /></el-icon>
          <strong>直播画面占位</strong>
          <span>接入真实直播流后可替换为播放器</span>
        </div>
      </div>

      <aside class="chat soft-panel">
        <div class="chat-head">
          <h2>实时弹幕</h2>
          <el-tag>{{ messages.length }}</el-tag>
        </div>
        <div class="messages">
          <div v-for="message in messages" :key="message.id" class="message" :style="{ color: message.color }">
            {{ message.content }}
          </div>
          <el-empty v-if="!messages.length" description="暂无弹幕" />
        </div>
        <div class="send-box">
          <el-input v-model="text" placeholder="发送弹幕" @keyup.enter="send" />
          <el-color-picker v-model="color" />
          <el-button type="primary" @click="send">发送</el-button>
        </div>
      </aside>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { VideoPlay } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { DanmakuWebSocket } from '@/api/danmaku'
import { useAuthStore } from '@/store/auth'
import type { Danmaku } from '@/types'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const roomId = Number(route.params.id)
const connected = ref(false)
const text = ref('')
const color = ref('#FFFFFF')
const messages = ref<Danmaku[]>([])
let ws: DanmakuWebSocket | null = null

onMounted(() => {
  if (!authStore.isLoggedIn) return
  ws = new DanmakuWebSocket({
    roomId,
    token: authStore.token,
    onMessage: (item) => messages.value.push(item),
    onViewerCount: () => {},
  })
  ws.connect()
  connected.value = true
})

onUnmounted(() => ws?.disconnect())

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
</script>

<style scoped>
.live-page {
  display: grid;
  gap: 18px;
}

.section-head p {
  margin: 8px 0 0;
}

.live-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  gap: 18px;
}

.stage {
  display: grid;
  min-height: 540px;
  place-items: center;
  background: linear-gradient(135deg, #165dff, #111827);
  color: #fff;
}

.stage-content {
  display: grid;
  justify-items: center;
  gap: 12px;
}

.stage-content .el-icon {
  font-size: 62px;
}

.stage-content strong {
  font-size: 26px;
}

.stage-content span {
  color: rgba(255, 255, 255, 0.74);
}

.chat {
  display: grid;
  grid-template-rows: auto 1fr auto;
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
  grid-template-columns: 1fr auto auto;
  gap: 8px;
}

@media (max-width: 920px) {
  .live-grid {
    grid-template-columns: 1fr;
  }
}
</style>
