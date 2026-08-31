<template>
  <main class="page-shell chat-page">
    <section class="chat-shell">
      <aside class="conversation-panel">
        <div class="panel-head">
          <div><h1>私信</h1><span>{{ totalUnread ? `${totalUnread} 条未读` : '保持联系' }}</span></div>
          <i :class="{ online: connected }" :title="connected ? '实时连接正常' : '正在重新连接'" />
        </div>
        <el-input v-model="conversationKeyword" class="conversation-search" placeholder="搜索会话" clearable />
        <div class="conversation-list">
          <button
            v-for="conversation in filteredConversations"
            :key="conversation.user.id"
            type="button"
            :class="{ active: activeUser?.id === conversation.user.id }"
            @click="openConversation(conversation.user.id)"
          >
            <el-badge :value="conversation.unreadCount" :hidden="conversation.unreadCount === 0" :max="99">
              <el-avatar :size="44" :src="mediaUrl(conversation.user.avatar)">{{ conversation.user.nickname.slice(0, 1) }}</el-avatar>
            </el-badge>
            <span class="conversation-copy">
              <strong>{{ conversation.user.nickname }}</strong>
              <small>{{ messagePreview(conversation.lastMessage) }}</small>
            </span>
            <time>{{ shortTime(conversation.lastMessage.createdAt) }}</time>
          </button>
          <el-empty v-if="!filteredConversations.length" :image-size="72" description="暂无私信会话" />
        </div>
      </aside>

      <section v-if="activeUser" class="message-panel">
        <header class="chat-head">
          <button type="button" @click="router.push(`/user/${activeUser.id}`)">
            <el-avatar :size="36" :src="mediaUrl(activeUser.avatar)">{{ activeUser.nickname.slice(0, 1) }}</el-avatar>
            <span><strong>{{ activeUser.nickname }}</strong><small>查看个人主页</small></span>
          </button>
        </header>

        <div ref="messageScrollRef" v-loading="historyLoading" class="message-scroll" @scroll="handleMessageScroll">
          <button v-if="canLoadMore" class="load-older" type="button" :disabled="historyLoading" @click="loadOlderMessages">
            {{ historyLoading ? '正在加载...' : '查看更早消息' }}
          </button>
          <div v-for="message in messages" :key="message.id" class="message-row" :class="{ mine: message.senderId === currentUserID }">
            <el-avatar v-if="message.senderId !== currentUserID" :size="32" :src="mediaUrl(message.sender.avatar)">
              {{ message.sender.nickname?.slice(0, 1) }}
            </el-avatar>
            <div class="bubble-wrap">
              <p v-if="!message.type || message.type === 'text'">{{ message.content }}</p>
              <el-image
                v-else-if="message.type === 'image'"
                class="message-image"
                :src="mediaUrl(message.mediaUrl)"
                :preview-src-list="[mediaUrl(message.mediaUrl)]"
                fit="cover"
                hide-on-click-modal
              />
              <div v-else-if="message.type === 'video'" class="message-video">
                <video controls preload="metadata" :src="mediaUrl(message.mediaUrl)" />
                <span v-if="message.mediaName">{{ message.mediaName }}</span>
              </div>
              <button
                v-else-if="message.type === 'video_share' && message.video"
                class="shared-video"
                type="button"
                @click="router.push(`/video/${message.video.id}`)"
              >
                <img :src="mediaUrl(message.video.coverUrl)" :alt="message.video.title">
                <span>
                  <strong>{{ message.video.title }}</strong>
                  <small>{{ message.video.author.nickname }} · {{ formatDuration(message.video.duration) }}</small>
                </span>
              </button>
              <time>{{ message.createdAt }}</time>
            </div>
          </div>
          <div v-if="!messages.length && !historyLoading" class="conversation-start">
            <strong>开始和 {{ activeUser.nickname }} 聊天</strong>
            <span>请友善交流，不要发送敏感信息。</span>
          </div>
        </div>

        <footer class="composer">
          <div class="composer-tools">
            <input ref="imageInputRef" type="file" accept="image/jpeg,image/png,image/webp,image/gif" @change="event => selectMedia(event, 'image')">
            <input ref="videoInputRef" type="file" accept="video/mp4,video/webm" @change="event => selectMedia(event, 'video')">
            <el-tooltip content="发送图片">
              <el-button text circle :icon="Picture" :disabled="uploading" aria-label="发送图片" @click="imageInputRef?.click()" />
            </el-tooltip>
            <el-tooltip content="发送视频文件">
              <el-button text circle :icon="VideoCamera" :disabled="uploading" aria-label="发送视频文件" @click="videoInputRef?.click()" />
            </el-tooltip>
            <el-tooltip content="分享站内视频">
              <el-button text circle :icon="Share" :disabled="uploading" aria-label="分享站内视频" @click="openVideoShare" />
            </el-tooltip>
            <span v-if="uploading">正在上传 {{ uploadProgress }}%</span>
          </div>
          <el-input
            v-model="draft"
            type="textarea"
            :rows="3"
            maxlength="2000"
            resize="none"
            placeholder="输入消息，Enter 发送，Shift + Enter 换行"
            @keydown="handleComposerKeydown"
          />
          <div class="composer-foot">
            <span>{{ draft.length }}/2000</span>
            <el-button type="primary" :disabled="!draft.trim()" @click="sendMessage">发送</el-button>
          </div>
        </footer>
      </section>

      <section v-else class="chat-placeholder">
        <div>
          <el-icon><ChatDotRound /></el-icon>
          <h2>选择一个会话</h2>
          <p>也可以从用户主页点击“私信”开始聊天。</p>
        </div>
      </section>
    </section>

    <el-dialog v-model="shareDialogVisible" title="分享站内视频" width="620px">
      <div class="share-search">
        <el-input v-model="shareKeyword" clearable placeholder="搜索公开视频" @keyup.enter="loadShareVideos" />
        <el-button :loading="shareLoading" @click="loadShareVideos">搜索</el-button>
      </div>
      <div v-loading="shareLoading" class="share-video-list">
        <button
          v-for="video in shareVideos"
          :key="video.id"
          type="button"
          :class="{ selected: selectedShareVideo?.id === video.id }"
          @click="selectedShareVideo = video"
        >
          <img :src="mediaUrl(video.coverUrl)" :alt="video.title">
          <span><strong>{{ video.title }}</strong><small>{{ video.author.nickname }} · {{ formatDuration(video.duration) }}</small></span>
        </button>
        <el-empty v-if="!shareVideos.length && !shareLoading" :image-size="72" description="没有找到公开视频" />
      </div>
      <template #footer>
        <el-button @click="shareDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!selectedShareVideo" @click="sendSharedVideo">发送</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ChatDotRound, Picture, Share, VideoCamera } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import {
  ChatWebSocket,
  messageApi,
  type ChatMessageInfo,
  type ChatMessageType,
  type ConversationInfo,
  type MessageSendPayload,
} from '@/api/message'
import { userApi } from '@/api/user'
import { videoApi } from '@/api/video'
import { useAuthStore } from '@/store/auth'
import type { UserInfo, VideoInfo } from '@/types'
import { formatDuration, mediaUrl } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const currentUserID = computed(() => authStore.userInfo?.id || 0)
const conversations = ref<ConversationInfo[]>([])
const activeUser = ref<UserInfo | null>(null)
const messages = ref<ChatMessageInfo[]>([])
const messageTotal = ref(0)
const historyPage = ref(1)
const historyLoading = ref(false)
const conversationKeyword = ref('')
const draft = ref('')
const connected = ref(false)
const messageScrollRef = ref<HTMLElement>()
const imageInputRef = ref<HTMLInputElement>()
const videoInputRef = ref<HTMLInputElement>()
const uploading = ref(false)
const uploadProgress = ref(0)
const shareDialogVisible = ref(false)
const shareKeyword = ref('')
const shareLoading = ref(false)
const shareVideos = ref<VideoInfo[]>([])
const selectedShareVideo = ref<VideoInfo | null>(null)
let socket: ChatWebSocket | null = null
let shouldStickToBottom = true

const filteredConversations = computed(() => {
  const keyword = conversationKeyword.value.trim().toLowerCase()
  if (!keyword) return conversations.value
  return conversations.value.filter(item => item.user.nickname.toLowerCase().includes(keyword))
})
const totalUnread = computed(() => conversations.value.reduce((sum, item) => sum + item.unreadCount, 0))
const canLoadMore = computed(() => messages.value.length < messageTotal.value)

onMounted(async () => {
  await loadConversations()
  connectSocket()
  await syncRouteConversation()
})

onBeforeUnmount(() => socket?.disconnect())
watch(() => route.params.userId, () => void syncRouteConversation())

async function loadConversations() {
  const res = await messageApi.conversations()
  conversations.value = res.data.list
}

function connectSocket() {
  socket?.disconnect()
  socket = new ChatWebSocket({
    token: authStore.token,
    onMessage: receiveMessage,
    onError: message => ElMessage.error(message),
    onStateChange: value => { connected.value = value },
  })
  socket.connect()
}

async function syncRouteConversation() {
  const userId = Number(route.params.userId)
  if (!userId) {
    activeUser.value = null
    messages.value = []
    return
  }
  if (userId === currentUserID.value) {
    ElMessage.warning('不能给自己发送私信')
    router.replace('/messages')
    return
  }
  const existing = conversations.value.find(item => item.user.id === userId)
  if (existing) {
    activeUser.value = existing.user
  } else {
    try {
      activeUser.value = (await userApi.profile(userId)).data
    } catch {
      router.replace('/messages')
      return
    }
  }
  await loadHistory()
}

function openConversation(userId: number) {
  if (activeUser.value?.id === userId) return
  router.push(`/messages/${userId}`)
}

async function loadHistory() {
  if (!activeUser.value) return
  historyLoading.value = true
  historyPage.value = 1
  try {
    const res = await messageApi.history(activeUser.value.id, { page: 1, pageSize: 50 })
    messages.value = res.data.list
    messageTotal.value = res.data.total
    await messageApi.read(activeUser.value.id)
    markLocalConversationRead(activeUser.value.id)
    await scrollToBottom(false)
  } finally {
    historyLoading.value = false
  }
}

async function loadOlderMessages() {
  if (!activeUser.value || historyLoading.value || !canLoadMore.value) return
  const container = messageScrollRef.value
  const oldHeight = container?.scrollHeight || 0
  historyLoading.value = true
  try {
    const nextPage = historyPage.value + 1
    const res = await messageApi.history(activeUser.value.id, { page: nextPage, pageSize: 50 })
    messages.value = dedupeMessages([...res.data.list, ...messages.value])
    historyPage.value = nextPage
    await nextTick()
    if (container) container.scrollTop = container.scrollHeight - oldHeight
  } finally {
    historyLoading.value = false
  }
}

async function sendMessage() {
  if (!activeUser.value) return
  const content = draft.value.trim()
  if (!content) return
  draft.value = ''
  const sent = await sendPayload({ receiverId: activeUser.value.id, type: 'text', content })
  if (!sent) draft.value = content
}

async function sendPayload(payload: MessageSendPayload) {
  payload.clientMessageId ||= crypto.randomUUID()
  shouldStickToBottom = true
  if (socket?.send(payload)) return true
  try {
    receiveMessage((await messageApi.send(payload)).data)
    return true
  } catch {
    return false
  }
}

async function selectMedia(event: Event, expectedType: 'image' | 'video') {
  if (!activeUser.value || uploading.value) return
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  const maxSize = expectedType === 'image' ? 10 * 1024 * 1024 : 200 * 1024 * 1024
  if (file.size > maxSize) {
    ElMessage.warning(expectedType === 'image' ? '图片不能超过 10MB' : '视频不能超过 200MB')
    return
  }
  uploading.value = true
  uploadProgress.value = 0
  try {
    const uploaded = await messageApi.uploadMedia(file, value => { uploadProgress.value = value })
    const sent = await sendPayload({
      receiverId: activeUser.value.id,
      type: uploaded.data.mediaType,
      content: uploaded.data.mediaType === 'image' ? '[图片]' : '[视频]',
      mediaUrl: uploaded.data.url,
      mediaName: uploaded.data.name,
    })
    if (!sent) ElMessage.error('附件已上传，但消息发送失败，请重试')
  } finally {
    uploading.value = false
    uploadProgress.value = 0
  }
}

async function openVideoShare() {
  selectedShareVideo.value = null
  shareDialogVisible.value = true
  await loadShareVideos()
}

async function loadShareVideos() {
  shareLoading.value = true
  try {
    const res = await videoApi.list({ page: 1, pageSize: 20, keyword: shareKeyword.value.trim(), sort: 'date' })
    shareVideos.value = res.data.list
  } finally {
    shareLoading.value = false
  }
}

async function sendSharedVideo() {
  if (!activeUser.value || !selectedShareVideo.value) return
  const video = selectedShareVideo.value
  const sent = await sendPayload({
    receiverId: activeUser.value.id,
    type: 'video_share',
    content: `[视频分享] ${video.title}`,
    videoId: video.id,
  })
  if (sent) {
    shareDialogVisible.value = false
    selectedShareVideo.value = null
  }
}

function receiveMessage(message: ChatMessageInfo) {
  const otherID = message.senderId === currentUserID.value ? message.receiverId : message.senderId
  const isActive = activeUser.value?.id === otherID
  if (isActive) {
    messages.value = dedupeMessages([...messages.value, message])
    messageTotal.value = Math.max(messageTotal.value + 1, messages.value.length)
    if (message.receiverId === currentUserID.value) void messageApi.read(otherID)
    void scrollToBottom(true)
  }
  upsertConversation(message, otherID, isActive)
}

function upsertConversation(message: ChatMessageInfo, otherID: number, isActive: boolean) {
  let conversation = conversations.value.find(item => item.user.id === otherID)
  if (!conversation) {
    const user = message.senderId === otherID ? message.sender : (message.receiver || activeUser.value)
    if (!user) return
    conversation = { user, lastMessage: message, unreadCount: 0 }
    conversations.value.push(conversation)
  }
  conversation.lastMessage = message
  if (message.receiverId === currentUserID.value && !isActive) conversation.unreadCount += 1
  if (isActive) conversation.unreadCount = 0
  conversations.value = [...conversations.value].sort((a, b) => b.lastMessage.id - a.lastMessage.id)
}

function markLocalConversationRead(userId: number) {
  const conversation = conversations.value.find(item => item.user.id === userId)
  if (conversation) conversation.unreadCount = 0
}

function dedupeMessages(list: ChatMessageInfo[]) {
  const map = new Map<number, ChatMessageInfo>()
  list.forEach(item => map.set(item.id, item))
  return [...map.values()].sort((a, b) => a.id - b.id)
}

function handleComposerKeydown(event: Event | KeyboardEvent) {
  const keyboardEvent = event as KeyboardEvent
  if (keyboardEvent.key !== 'Enter' || keyboardEvent.shiftKey || keyboardEvent.isComposing) return
  keyboardEvent.preventDefault()
  void sendMessage()
}

function handleMessageScroll() {
  const container = messageScrollRef.value
  if (!container) return
  shouldStickToBottom = container.scrollHeight - container.scrollTop - container.clientHeight < 80
}

async function scrollToBottom(onlyWhenSticky: boolean) {
  if (onlyWhenSticky && !shouldStickToBottom) return
  await nextTick()
  const container = messageScrollRef.value
  if (container) container.scrollTop = container.scrollHeight
}

function shortTime(value: string) {
  const today = new Date().toISOString().slice(0, 10)
  return value.startsWith(today) ? value.slice(11, 16) : value.slice(5, 10)
}

function messagePreview(message: ChatMessageInfo) {
  const type = (message.type || 'text') as ChatMessageType
  if (type === 'image') return '[图片]'
  if (type === 'video') return '[视频]'
  if (type === 'video_share') return message.video ? `[视频分享] ${message.video.title}` : '[视频分享]'
  return message.content
}
</script>

<style scoped>
.chat-page { padding-top: 24px; }
.chat-shell { display: grid; grid-template-columns: 310px minmax(0, 1fr); height: calc(100vh - 116px); min-height: 560px; overflow: hidden; border: 1px solid #e4e7ec; border-radius: 8px; background: #fff; }
.conversation-panel { display: grid; grid-template-rows: auto auto minmax(0, 1fr); min-width: 0; border-right: 1px solid #e4e7ec; }
.panel-head { display: flex; align-items: center; justify-content: space-between; padding: 20px 18px 12px; }
.panel-head h1 { margin: 0; color: #18191c; font-size: 22px; }
.panel-head span { color: #98a2b3; font-size: 12px; }
.panel-head i { width: 9px; height: 9px; border-radius: 50%; background: #d0d5dd; }
.panel-head i.online { background: #12b76a; }
.conversation-search { width: auto; margin: 0 14px 12px; }
.conversation-list { overflow-y: auto; }
.conversation-list > button { display: grid; grid-template-columns: 48px minmax(0, 1fr) auto; width: 100%; min-height: 72px; align-items: center; gap: 10px; padding: 10px 14px; border: 0; background: transparent; cursor: pointer; text-align: left; }
.conversation-list > button:hover, .conversation-list > button.active { background: #f5f7f9; }
.conversation-list > button.active { box-shadow: inset 3px 0 #00aeec; }
.conversation-copy { display: grid; min-width: 0; gap: 5px; }
.conversation-copy strong, .conversation-copy small { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.conversation-copy strong { color: #18191c; font-size: 14px; }
.conversation-copy small, .conversation-list time { color: #98a2b3; font-size: 12px; }
.message-panel { display: grid; min-width: 0; grid-template-rows: 64px minmax(0, 1fr) auto; }
.chat-head { display: flex; align-items: center; padding: 0 20px; border-bottom: 1px solid #e9ecf0; }
.chat-head button { display: flex; align-items: center; gap: 10px; padding: 0; border: 0; background: transparent; cursor: pointer; text-align: left; }
.chat-head span { display: grid; gap: 2px; }
.chat-head strong { color: #18191c; font-size: 14px; }
.chat-head small { color: #98a2b3; font-size: 11px; }
.message-scroll { position: relative; overflow-x: hidden; overflow-y: auto; padding: 22px 24px; background: #f7f8fa; }
.message-row { display: flex; align-items: flex-end; gap: 8px; margin: 12px 0; }
.message-row.mine { justify-content: flex-end; }
.bubble-wrap { display: grid; max-width: min(68%, 620px); gap: 4px; }
.bubble-wrap p { margin: 0; padding: 9px 12px; border: 1px solid #e4e7ec; border-radius: 8px; background: #fff; color: #242b35; line-height: 1.55; overflow-wrap: anywhere; white-space: pre-wrap; }
.message-row.mine .bubble-wrap p { border-color: #00aeec; background: #00aeec; color: #fff; }
.message-image { width: min(360px, 52vw); max-height: 320px; border-radius: 8px; background: #e9edf2; cursor: zoom-in; }
.message-image :deep(img) { max-height: 320px; }
.message-video { display: grid; width: min(440px, 58vw); overflow: hidden; border: 1px solid #e4e7ec; border-radius: 8px; background: #fff; }
.message-video video { display: block; width: 100%; max-height: 340px; background: #111; }
.message-video span { overflow: hidden; padding: 8px 10px; color: #667085; font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.shared-video { display: grid; width: min(400px, 56vw); grid-template-columns: 144px minmax(0, 1fr); overflow: hidden; padding: 0; border: 1px solid #e4e7ec; border-radius: 8px; background: #fff; cursor: pointer; text-align: left; }
.shared-video:hover { border-color: #00aeec; }
.shared-video img { width: 144px; height: 82px; object-fit: cover; }
.shared-video > span { display: grid; min-width: 0; align-content: center; gap: 7px; padding: 10px 12px; }
.shared-video strong, .shared-video small { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.shared-video strong { color: #18191c; font-size: 13px; }
.shared-video small { color: #98a2b3; font-size: 11px; }
.bubble-wrap time { color: #98a2b3; font-size: 10px; }
.message-row.mine time { text-align: right; }
.load-older { display: block; margin: 0 auto 14px; border: 0; background: transparent; color: #667085; cursor: pointer; font-size: 12px; }
.conversation-start, .chat-placeholder { display: grid; place-items: center; color: #98a2b3; text-align: center; }
.conversation-start { min-height: 180px; align-content: center; gap: 6px; }
.conversation-start strong { color: #475467; }
.conversation-start > * { width: 100%; min-width: 0; max-width: 100%; overflow-wrap: anywhere; white-space: normal; }
.composer { padding: 12px 16px 14px; border-top: 1px solid #e4e7ec; }
.composer-tools { display: flex; min-height: 32px; align-items: center; gap: 2px; margin-bottom: 6px; }
.composer-tools input { display: none; }
.composer-tools > span { margin-left: 8px; color: #667085; font-size: 11px; }
.composer-foot { display: flex; align-items: center; justify-content: space-between; margin-top: 8px; }
.composer-foot span { color: #98a2b3; font-size: 11px; }
.chat-placeholder > div { display: grid; place-items: center; }
.chat-placeholder .el-icon { color: #c5cad1; font-size: 52px; }
.chat-placeholder h2 { margin: 14px 0 6px; color: #475467; font-size: 18px; }
.chat-placeholder p { margin: 0; font-size: 13px; }
.share-search { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 8px; margin-bottom: 14px; }
.share-video-list { display: grid; max-height: 420px; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; overflow-y: auto; }
.share-video-list > button { display: grid; min-width: 0; grid-template-columns: 112px minmax(0, 1fr); align-items: center; gap: 10px; overflow: hidden; padding: 6px; border: 1px solid #e4e7ec; border-radius: 6px; background: #fff; cursor: pointer; text-align: left; }
.share-video-list > button:hover, .share-video-list > button.selected { border-color: #00aeec; background: #f2fbff; }
.share-video-list img { width: 112px; aspect-ratio: 16 / 9; border-radius: 4px; object-fit: cover; }
.share-video-list span { display: grid; min-width: 0; gap: 6px; }
.share-video-list strong, .share-video-list small { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.share-video-list strong { color: #18191c; font-size: 12px; }
.share-video-list small { color: #98a2b3; font-size: 10px; }
@media (max-width: 760px) { .chat-shell { grid-template-columns: 108px minmax(0, 1fr); } .panel-head { padding: 16px 10px 10px; } .panel-head span, .conversation-search, .conversation-copy, .conversation-list time { display: none; } .conversation-list > button { grid-template-columns: 1fr; justify-items: center; padding: 10px; } .share-video-list { grid-template-columns: 1fr; } .shared-video { grid-template-columns: 110px minmax(0, 1fr); } .shared-video img { width: 110px; height: 68px; } }
</style>
