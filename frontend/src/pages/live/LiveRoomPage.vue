<template>
  <div class="max-w-6xl mx-auto flex flex-col h-[calc(100vh-120px)]">
    <div v-loading="loading" class="flex-1 flex flex-col min-h-0">
      <div v-if="room" class="flex gap-4 flex-1 min-h-0">
        <!-- Player -->
        <div class="flex-1 min-w-0">
          <div class="bg-black rounded-xl overflow-hidden aspect-video relative">
            <div ref="playerContainer" class="w-full h-full" />
            <div class="absolute top-3 left-3 flex items-center gap-2 z-10">
              <el-tag type="danger" size="small">直播中</el-tag>
              <span class="text-white text-xs bg-black/50 px-2 py-0.5 rounded-full flex items-center gap-1">
                <el-icon><View /></el-icon>
                {{ viewerCount }} 人在看
              </span>
            </div>
          </div>
        </div>

        <!-- Danmaku panel -->
        <div class="w-72 flex flex-col border border-gray-200 rounded-xl overflow-hidden flex-shrink-0">
          <div class="px-3 py-2.5 font-semibold bg-gray-50 border-b border-gray-200 flex items-center gap-2">
            <span>弹幕列表</span>
            <el-tag size="small" type="primary">{{ danmakuList.length }}</el-tag>
          </div>
          <div ref="danmakuListEl" class="flex-1 overflow-y-auto px-3 py-2 space-y-1.5 min-h-0">
            <div
              v-for="(d, i) in danmakuList"
              :key="i"
              class="text-sm leading-relaxed break-all"
              :style="{ color: d.color }"
            >
              <span class="font-semibold opacity-80 mr-1">{{ d.username ?? '观众' }}:</span>
              <span>{{ d.content }}</span>
            </div>
          </div>
        </div>
      </div>

      <el-result
        v-else-if="!loading"
        icon="warning"
        title="直播间不存在或尚未开播"
      >
        <template #extra>
          <el-button type="primary" @click="$router.back()">返回</el-button>
        </template>
      </el-result>
    </div>

    <!-- Danmaku input -->
    <div v-if="room" class="mt-3 bg-white border border-gray-200 rounded-xl px-4 py-3 flex gap-2 items-center">
      <el-input
        v-model="danmakuInput"
        placeholder="发送弹幕..."
        :maxlength="100"
        class="flex-1"
        :disabled="!authStore.isLoggedIn"
        @keydown.enter="sendDanmaku"
      />
      <el-color-picker v-model="danmakuColor" size="small" />
      <el-tooltip :content="authStore.isLoggedIn ? '' : '请先登录'" :disabled="authStore.isLoggedIn">
        <el-button type="primary" :disabled="!authStore.isLoggedIn" @click="sendDanmaku">发送</el-button>
      </el-tooltip>
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
  await new Promise(r => setTimeout(r, 300))
  room.value = { id, title: `直播间 ${id}` }
  loading.value = false

  wsClient = new DanmakuWebSocket({
    roomId: id,
    token: authStore.token,
    onMessage: (d: Danmaku) => {
      danmakuList.value.push({ content: d.content, color: d.color ?? '#FFFFFF' })
      scrollDanmakuList()
    },
    onViewerCount: (count: number) => { viewerCount.value = count },
  })
  wsClient.connect()
})

onUnmounted(() => { wsClient?.disconnect() })

function sendDanmaku() {
  const text = danmakuInput.value.trim()
  if (!text || !authStore.isLoggedIn) return
  wsClient?.send(text, danmakuColor.value)
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
