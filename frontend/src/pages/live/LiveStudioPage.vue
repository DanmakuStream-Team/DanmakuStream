<template>
  <main class="page-shell studio-page">
    <header class="studio-head">
      <div>
        <span class="eyebrow">直播工作台</span>
        <h1>{{ room?.title || '准备直播' }}</h1>
        <p>在浏览器中组合画面并推送到直播间。</p>
      </div>
      <div class="status-actions">
        <el-tag :type="publishing ? 'success' : previewReady ? 'warning' : 'info'">
          {{ publishing ? '推流中' : previewReady ? '预览中' : '未连接设备' }}
        </el-tag>
        <el-button @click="router.push(`/live/${roomId}`)">查看直播间</el-button>
      </div>
    </header>

    <section class="studio-grid">
      <div class="studio-main">
        <div class="preview-panel">
        <canvas ref="canvasRef" width="1280" height="720" aria-label="直播画面预览" />
        <LiveSuperChatOverlay :items="monitorSuperChats" />
        <div v-if="!previewReady" class="preview-empty">
          <el-icon><VideoCamera /></el-icon>
          <strong>选择画面来源后开始预览</strong>
          <span>浏览器会在获得授权后读取摄像头、麦克风或屏幕。</span>
        </div>
        <div class="preview-status">
          <i :class="{ active: publishing }" />
          {{ publishing ? 'LIVE' : 'PREVIEW' }} · 1280×720
        </div>
        <div class="text-layer-editor">
          <button
            v-for="layer in textLayers"
            :key="layer.id"
            type="button"
            :class="{ selected: selectedTextId === layer.id }"
            :style="{ left: `${layer.x}%`, top: `${layer.y}%`, color: layer.color, fontSize: `${Math.max(12, layer.fontSize / 2)}px` }"
            :aria-label="`拖动文字：${layer.text}`"
            @mousedown="startTextLayerDrag($event, layer)"
            @click="selectedTextId = layer.id"
          >
            {{ layer.text }}
          </button>
        </div>
          <video ref="cameraRef" muted playsinline />
          <video ref="screenRef" muted playsinline />
        </div>
        <LiveMonitorOverlay
          v-if="!monitorPipTarget"
          :visible="true"
          :connected="monitorConnected"
          :messages="monitorMessages"
          :super-chats="monitorSuperChats"
          :embedded="true"
        />
      </div>

      <aside class="control-panel soft-panel">
        <section class="control-section">
          <div class="control-title">
            <h2>画面来源</h2>
            <span>{{ previewReady ? '画面已就绪' : '点击后立即预览' }}</span>
          </div>
          <div class="source-grid">
            <button
              v-for="option in sourceOptions"
              :key="option.value"
              type="button"
              :class="{ active: sourceMode === option.value && previewReady }"
              :disabled="publishing || preparing"
              @click="selectSource(option.value)"
            >
              <el-icon><component :is="option.icon" /></el-icon>
              <span>{{ option.label }}</span>
            </button>
          </div>
        </section>

        <section class="control-section">
          <div class="control-title">
            <h2>画面文字</h2>
            <el-button text type="primary" @click="addTextLayer"><el-icon><Plus /></el-icon>添加</el-button>
          </div>
          <template v-if="selectedTextLayer">
            <el-input v-model="selectedTextLayer.text" maxlength="60" placeholder="输入悬浮文字" />
            <div class="text-layer-actions">
              <label>颜色 <input v-model="selectedTextLayer.color" class="color-input" type="color" aria-label="文字颜色" /></label>
              <span>直接拖动画面中的文字调整位置</span>
              <el-button text type="danger" aria-label="删除当前文字" @click="removeSelectedText"><el-icon><Delete /></el-icon></el-button>
            </div>
          </template>
          <div v-else class="empty-setting">添加一条文字，然后在预览画面中拖动它。</div>
        </section>

        <section class="control-section monitor-setting">
          <div>
            <strong>评论与 SC</strong>
            <span>{{ monitorPipTarget ? '已在跨软件悬浮窗显示' : '可在切换到其他软件后继续查看' }}</span>
          </div>
          <el-button type="primary" plain @click="openMonitorPictureInPicture"><el-icon><FullScreen /></el-icon>{{ monitorPipTarget ? '已悬浮' : '置顶悬浮' }}</el-button>
        </section>

        <section class="control-section audio-row">
          <div>
            <strong>直播声音</strong>
            <span>{{ sourceMode === 'screen' ? '优先使用屏幕共享声音' : '使用麦克风声音' }}</span>
          </div>
          <el-switch v-model="audioEnabled" :disabled="publishing" />
        </section>

        <details class="advanced-settings">
          <summary>更多互动设置</summary>
          <section>
            <el-form-item label="发言权限">
              <el-select v-model="chatSettings.chatMode">
                <el-option label="所有登录用户" value="everyone" />
                <el-option label="仅关注者" value="followers" />
                <el-option label="仅付费订阅者" value="members" />
              </el-select>
            </el-form-item>
            <el-form-item label="慢速模式">
              <el-select v-model="chatSettings.slowModeSeconds">
                <el-option label="关闭" :value="0" />
                <el-option label="每 5 秒一条" :value="5" />
                <el-option label="每 15 秒一条" :value="15" />
                <el-option label="每 30 秒一条" :value="30" />
                <el-option label="每 60 秒一条" :value="60" />
              </el-select>
            </el-form-item>
            <el-input v-model="chatSettings.pinnedMessage" maxlength="200" placeholder="直播间置顶公告，可留空" />
            <el-button :loading="savingChat" @click="saveChatSettings">保存</el-button>
          </section>
        </details>

        <div class="studio-actions">
          <span v-if="!previewReady">先选择一种画面来源</span>
          <el-button v-if="!publishing" type="primary" size="large" :loading="connecting" :disabled="!previewReady" @click="startPublishing">
            开始直播
          </el-button>
          <el-button v-else type="danger" size="large" @click="stopPublishing">停止直播</el-button>
        </div>
      </aside>
    </section>

    <Teleport v-if="monitorPipTarget" :to="monitorPipTarget">
      <LiveMonitorOverlay
        :visible="true"
        :connected="monitorConnected"
        :messages="monitorMessages"
        :super-chats="monitorSuperChats"
        :detached="true"
        @close="closeMonitorPictureInPicture"
      />
    </Teleport>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Delete, FullScreen, Monitor, PictureRounded, Plus, VideoCamera } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { DanmakuWebSocket } from '@/api/danmaku'
import { liveApi } from '@/api/live'
import LiveMonitorOverlay from '@/components/live/LiveMonitorOverlay.vue'
import LiveSuperChatOverlay from '@/components/live/LiveSuperChatOverlay.vue'
import { useAuthStore } from '@/store/auth'
import type { Danmaku, LiveChatSettings, LiveGiftEvent, LiveMonitorSuperChat, LiveRoom } from '@/types'

type SourceMode = 'camera' | 'screen' | 'screen_camera'
type Corner = 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right'
type TextLayer = { id: number; text: string; x: number; y: number; color: string; fontSize: number }

interface DocumentPictureInPictureAPI {
  requestWindow(options?: { width?: number; height?: number }): Promise<Window>
}

declare global {
  interface Window {
    documentPictureInPicture?: DocumentPictureInPictureAPI
  }
}

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const roomId = computed(() => Number(route.params.id))
const room = ref<LiveRoom>()
const canvasRef = ref<HTMLCanvasElement>()
const cameraRef = ref<HTMLVideoElement>()
const screenRef = ref<HTMLVideoElement>()
const sourceMode = ref<SourceMode>((route.query.mode as SourceMode) || 'camera')
const sourceOptions: Array<{ label: string; value: SourceMode; icon: typeof VideoCamera }> = [
  { label: '摄像头', value: 'camera', icon: VideoCamera },
  { label: '屏幕', value: 'screen', icon: Monitor },
  { label: '画中画', value: 'screen_camera', icon: PictureRounded },
]
const pipCorner = ref<Corner>('bottom-right')
const pipSize = ref(28)
const textLayers = ref<TextLayer[]>([])
const selectedTextId = ref<number>()
const selectedTextLayer = computed(() => textLayers.value.find((layer) => layer.id === selectedTextId.value))
const audioEnabled = ref(true)
const preparing = ref(false)
const connecting = ref(false)
const previewReady = ref(false)
const publishing = ref(false)
const savingChat = ref(false)
const chatSettings = ref<LiveChatSettings>({ chatMode: 'everyone', slowModeSeconds: 0, pinnedMessage: '' })
const monitorConnected = ref(false)
const monitorMessages = ref<Danmaku[]>([])
const monitorSuperChats = ref<LiveMonitorSuperChat[]>([])
const monitorPipTarget = ref<HTMLElement>()

let cameraStream: MediaStream | null = null
let screenStream: MediaStream | null = null
let mediaRecorder: MediaRecorder | null = null
let publishSocket: WebSocket | null = null
let drawTimer: ReturnType<typeof setInterval> | null = null
let monitorSocket: DanmakuWebSocket | null = null
let monitorPipWindow: Window | null = null
let nextTextLayerId = 1

onMounted(async () => {
  try {
    room.value = (await liveApi.manageDetail(roomId.value)).data
    chatSettings.value = {
      chatMode: room.value.chatMode || 'everyone',
      slowModeSeconds: room.value.slowModeSeconds || 0,
      pinnedMessage: room.value.pinnedMessage || '',
    }
    await loadMonitor()
    connectMonitor()
  } catch {
    ElMessage.error('无法管理这个直播间')
    router.replace('/live')
  }
})

async function saveChatSettings() {
  savingChat.value = true
  try {
    room.value = (await liveApi.updateChatSettings(roomId.value, chatSettings.value)).data
    ElMessage.success('直播互动设置已保存')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '保存失败')
  } finally {
    savingChat.value = false
  }
}

async function loadMonitor() {
  try {
    const snapshot = (await liveApi.monitor(roomId.value)).data
    monitorMessages.value = snapshot.messages
    monitorSuperChats.value = snapshot.superChats
  } catch {
    monitorMessages.value = []
    monitorSuperChats.value = []
  }
}

function connectMonitor() {
  monitorSocket?.disconnect()
  monitorSocket = new DanmakuWebSocket({
    roomId: roomId.value,
    token: authStore.token || '',
    monitor: true,
		onConnectionChange: (connected) => { monitorConnected.value = connected },
    onMessage: (message) => {
      if (monitorMessages.value.some((item) => item.id === message.id)) return
      monitorMessages.value = [...monitorMessages.value.slice(-99), { ...message, createdAt: new Date().toISOString() }]
    },
		onViewerCount: () => undefined,
    onGift: (event) => appendMonitorSuperChat(event),
  })
  monitorSocket.connect()
}

function appendMonitorSuperChat(event: LiveGiftEvent) {
  const item: LiveMonitorSuperChat = {
    id: event.id || Date.now(),
    user: event.user,
    gift: event.gift,
    count: event.count,
    value: event.value,
    message: event.message || '',
    displaySeconds: event.displaySeconds || 0,
    createdAt: event.createdAt || new Date().toISOString(),
  }
  if (monitorSuperChats.value.some((existing) => existing.id === item.id)) return
  monitorSuperChats.value = [...monitorSuperChats.value.slice(-49), item]
}

async function openMonitorPictureInPicture() {
  if (monitorPipWindow && !monitorPipWindow.closed) {
    monitorPipWindow.focus()
    return
  }
  if (!window.documentPictureInPicture) {
    ElMessage.warning('当前浏览器不支持跨软件悬浮窗，请使用最新版 Chrome 或 Edge')
    return
  }
  try {
    const pipWindow = await window.documentPictureInPicture.requestWindow({ width: 360, height: 520 })
    monitorPipWindow = pipWindow
    pipWindow.document.title = '直播评论与 SC'
    document.querySelectorAll('style, link[rel="stylesheet"]').forEach((node) => {
      pipWindow.document.head.appendChild(node.cloneNode(true))
    })
    const baseStyle = pipWindow.document.createElement('style')
    baseStyle.textContent = 'html,body,#live-monitor-pip{width:100%;height:100%;margin:0;overflow:hidden;background:#0f1218;}*{box-sizing:border-box;}'
    pipWindow.document.head.appendChild(baseStyle)
    const target = pipWindow.document.createElement('div')
    target.id = 'live-monitor-pip'
    pipWindow.document.body.appendChild(target)
    monitorPipTarget.value = target
    pipWindow.addEventListener('pagehide', () => {
      if (monitorPipWindow !== pipWindow) return
      monitorPipTarget.value = undefined
      monitorPipWindow = null
    })
  } catch (error) {
    ElMessage.warning(error instanceof Error ? error.message : '悬浮窗打开失败')
  }
}

function closeMonitorPictureInPicture() {
  const pipWindow = monitorPipWindow
  monitorPipTarget.value = undefined
  monitorPipWindow = null
  if (pipWindow && !pipWindow.closed) pipWindow.close()
}

onUnmounted(() => {
  monitorSocket?.disconnect()
  closeMonitorPictureInPicture()
  teardown()
})

async function selectSource(mode: SourceMode) {
  sourceMode.value = mode
  await preparePreview()
}

function addTextLayer() {
  const layer: TextLayer = {
    id: nextTextLayerId++,
    text: '新的悬浮文字',
    x: 50,
    y: 18 + (textLayers.value.length % 5) * 12,
    color: '#ffffff',
    fontSize: 34,
  }
  textLayers.value.push(layer)
  selectedTextId.value = layer.id
  drawFrame()
}

function removeSelectedText() {
  textLayers.value = textLayers.value.filter((layer) => layer.id !== selectedTextId.value)
  selectedTextId.value = textLayers.value.at(-1)?.id
  drawFrame()
}

function startTextLayerDrag(event: MouseEvent, layer: TextLayer) {
  if (event.button !== 0 || !canvasRef.value) return
  event.preventDefault()
  selectedTextId.value = layer.id
  const panel = canvasRef.value.getBoundingClientRect()
  const move = (moveEvent: MouseEvent) => {
    layer.x = Math.min(94, Math.max(6, ((moveEvent.clientX - panel.left) / panel.width) * 100))
    layer.y = Math.min(92, Math.max(8, ((moveEvent.clientY - panel.top) / panel.height) * 100))
    drawFrame()
  }
  const stop = () => {
    window.removeEventListener('mousemove', move)
    window.removeEventListener('mouseup', stop)
  }
  window.addEventListener('mousemove', move)
  window.addEventListener('mouseup', stop)
}

async function preparePreview() {
	if (!navigator.mediaDevices?.getUserMedia) {
	  ElMessage.warning('浏览器开播需要 HTTPS，或在 localhost 上使用')
	  return
	}
  preparing.value = true
  stopSources()
  try {
    if (sourceMode.value === 'camera' || sourceMode.value === 'screen_camera') {
      cameraStream = await navigator.mediaDevices.getUserMedia({
        video: { width: { ideal: 1280 }, height: { ideal: 720 } },
        audio: audioEnabled.value,
      })
      if (cameraRef.value) {
        cameraRef.value.srcObject = cameraStream
        await cameraRef.value.play()
        await waitForVideoData(cameraRef.value)
      }
    }
    if (sourceMode.value === 'screen' || sourceMode.value === 'screen_camera') {
      screenStream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: audioEnabled.value })
      if (screenRef.value) {
        screenRef.value.srcObject = screenStream
        await screenRef.value.play()
        await waitForVideoData(screenRef.value)
      }
      screenStream.getVideoTracks()[0]?.addEventListener('ended', () => {
        if (publishing.value) stopPublishing()
        stopSources()
      })
    }
    previewReady.value = true
    startDrawing()
  } catch (error) {
    stopSources()
    ElMessage.warning(error instanceof Error ? error.message : '没有获得设备权限')
  } finally {
    preparing.value = false
  }
}

function waitForVideoData(video: HTMLVideoElement) {
  if (video.videoWidth && video.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA) return Promise.resolve()
  return new Promise<void>((resolve) => {
    const done = () => {
      video.removeEventListener('loadeddata', done)
      video.removeEventListener('canplay', done)
      resolve()
    }
    video.addEventListener('loadeddata', done, { once: true })
    video.addEventListener('canplay', done, { once: true })
    window.setTimeout(done, 3000)
  })
}

function startDrawing() {
  if (drawTimer) clearInterval(drawTimer)
  drawFrame()
  drawTimer = setInterval(drawFrame, 1000 / 30)
}

function drawFrame() {
  const canvas = canvasRef.value
  const ctx = canvas?.getContext('2d')
  if (!canvas || !ctx) return
  ctx.fillStyle = '#090b10'
  ctx.fillRect(0, 0, canvas.width, canvas.height)

  const mainVideo = sourceMode.value === 'camera' ? cameraRef.value : screenRef.value
  if (mainVideo?.readyState && mainVideo.videoWidth) drawCover(ctx, mainVideo, 0, 0, canvas.width, canvas.height)

  if (sourceMode.value === 'screen_camera' && cameraRef.value?.videoWidth) {
    const width = Math.round(canvas.width * pipSize.value / 100)
    const height = Math.round(width * 9 / 16)
    const margin = 28
    const x = pipCorner.value.includes('right') ? canvas.width - width - margin : margin
    const y = pipCorner.value.includes('bottom') ? canvas.height - height - margin : margin
    ctx.save()
    ctx.shadowColor = 'rgba(0,0,0,.45)'
    ctx.shadowBlur = 18
    ctx.fillStyle = '#111'
    ctx.fillRect(x - 4, y - 4, width + 8, height + 8)
    ctx.restore()
    drawCover(ctx, cameraRef.value, x, y, width, height)
  }

  textLayers.value.forEach((layer) => drawTextLayer(ctx, layer))
}

function drawCover(ctx: CanvasRenderingContext2D, video: HTMLVideoElement, x: number, y: number, width: number, height: number) {
  const scale = Math.max(width / video.videoWidth, height / video.videoHeight)
  const sourceWidth = width / scale
  const sourceHeight = height / scale
  const sourceX = (video.videoWidth - sourceWidth) / 2
  const sourceY = (video.videoHeight - sourceHeight) / 2
  ctx.drawImage(video, sourceX, sourceY, sourceWidth, sourceHeight, x, y, width, height)
}

function drawTextLayer(ctx: CanvasRenderingContext2D, layer: TextLayer) {
  const value = layer.text.trim()
  if (!value) return
  const canvas = canvasRef.value!
  const padding = 16
  ctx.font = `600 ${layer.fontSize}px sans-serif`
  const width = Math.min(ctx.measureText(value).width + padding * 2, canvas.width - 56)
  const height = layer.fontSize + 26
  const centerX = canvas.width * layer.x / 100
  const centerY = canvas.height * layer.y / 100
  const x = Math.min(canvas.width - width - 20, Math.max(20, centerX - width / 2))
  const y = Math.min(canvas.height - height - 20, Math.max(20, centerY - height / 2))
  ctx.fillStyle = 'rgba(0, 0, 0, .58)'
  ctx.fillRect(x, y, width, height)
  ctx.fillStyle = layer.color
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(value, x + width / 2, y + height / 2, width - padding * 2)
  ctx.textAlign = 'start'
}

async function startPublishing() {
  if (!previewReady.value) await preparePreview()
  if (!previewReady.value || !canvasRef.value || !authStore.token) return
  connecting.value = true
  try {
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
    const socket = new WebSocket(`${protocol}://${location.host}/ws/live-publish/${roomId.value}?token=${encodeURIComponent(authStore.token)}`)
    socket.binaryType = 'arraybuffer'
    await new Promise<void>((resolve, reject) => {
      socket.onopen = () => resolve()
      socket.onerror = () => reject(new Error('推流服务连接失败'))
    })
    publishSocket = socket

    const output = canvasRef.value.captureStream(30)
    if (audioEnabled.value) {
      const audioTrack = cameraStream?.getAudioTracks()[0] || screenStream?.getAudioTracks()[0]
      if (audioTrack) output.addTrack(audioTrack)
    }
    const mimeType = ['video/webm;codecs=vp8,opus', 'video/webm;codecs=vp8', 'video/webm']
      .find((value) => MediaRecorder.isTypeSupported(value))
    if (!mimeType) throw new Error('当前浏览器不支持直播编码')

    mediaRecorder = new MediaRecorder(output, { mimeType, videoBitsPerSecond: 3_000_000 })
    mediaRecorder.ondataavailable = async (event) => {
      if (!event.data.size || publishSocket?.readyState !== WebSocket.OPEN) return
      publishSocket.send(await event.data.arrayBuffer())
    }
    mediaRecorder.onerror = () => {
      ElMessage.error('浏览器编码中断')
      stopPublishing()
    }
    socket.onclose = () => {
	  const wasPublishing = publishing.value
      publishing.value = false
	  if (mediaRecorder?.state === 'recording') mediaRecorder.stop()
	  mediaRecorder = null
	  if (wasPublishing) ElMessage.warning('推流连接已断开')
    }
    mediaRecorder.start(1000)
    publishing.value = true
    ElMessage.success('直播画面正在推送')
  } catch (error) {
    publishSocket?.close()
    publishSocket = null
    ElMessage.error(error instanceof Error ? error.message : '开始推流失败')
  } finally {
    connecting.value = false
  }
}

function stopPublishing() {
  if (mediaRecorder?.state !== 'inactive') mediaRecorder?.stop()
  mediaRecorder = null
  publishing.value = false
  setTimeout(() => {
    publishSocket?.close()
    publishSocket = null
  }, 250)
  ElMessage.success('已停止推流，直播间仍然保留')
}

function stopSources() {
  if (publishing.value) stopPublishing()
  cameraStream?.getTracks().forEach((track) => track.stop())
  screenStream?.getTracks().forEach((track) => track.stop())
  cameraStream = null
  screenStream = null
  if (cameraRef.value) cameraRef.value.srcObject = null
  if (screenRef.value) screenRef.value.srcObject = null
  if (drawTimer) clearInterval(drawTimer)
  drawTimer = null
  previewReady.value = false
}

function teardown() {
  stopSources()
  publishSocket?.close()
}
</script>

<style scoped>
.studio-page { display: grid; gap: 18px; }
.studio-head { display: flex; align-items: flex-end; justify-content: space-between; gap: 20px; }
.studio-head h1 { margin: 4px 0 6px; font-size: 28px; }
.studio-head p, .control-title span, .audio-row span { margin: 0; color: #9499a0; }
.eyebrow { color: #00aeec; font-size: 13px; font-weight: 700; }
.status-actions { display: flex; align-items: center; gap: 10px; }
.studio-grid { display: grid; grid-template-columns: minmax(0, 1fr) 360px; gap: 18px; align-items: start; }
.studio-main { display: grid; min-width: 0; gap: 14px; }
.preview-panel { position: relative; overflow: hidden; aspect-ratio: 16 / 9; border-radius: 8px; background: #090b10; }
.preview-panel canvas { width: 100%; height: 100%; display: block; }
.preview-panel video { position: absolute; width: 1px; height: 1px; opacity: 0; pointer-events: none; }
.preview-empty { position: absolute; inset: 0; display: grid; place-content: center; justify-items: center; gap: 10px; color: #fff; text-align: center; }
.preview-empty .el-icon { font-size: 52px; }
.preview-empty span { color: rgba(255,255,255,.65); font-size: 13px; }
.preview-status { position: absolute; top: 16px; right: 16px; display: flex; align-items: center; gap: 7px; padding: 6px 9px; border-radius: 4px; background: rgba(0,0,0,.6); color: #fff; font-size: 12px; }
.preview-status i { width: 7px; height: 7px; border-radius: 50%; background: #9499a0; }
.preview-status i.active { background: #ff4d4f; }
.text-layer-editor { position: absolute; inset: 0; z-index: 7; pointer-events: none; }
.text-layer-editor button { position: absolute; max-width: 80%; overflow: hidden; padding: 8px 14px; border: 1px solid transparent; border-radius: 4px; background: transparent; cursor: grab; font-weight: 600; text-overflow: ellipsis; white-space: nowrap; transform: translate(-50%, -50%); pointer-events: auto; touch-action: none; }
.text-layer-editor button:active { cursor: grabbing; }
.text-layer-editor button.selected { border-color: #00aeec; outline: 1px solid rgb(0 174 236 / 28%); }
.text-layer-editor button { color: transparent !important; }
.control-panel { display: grid; gap: 0; padding: 0 18px; }
.control-section { display: grid; gap: 12px; padding: 18px 0; border-bottom: 1px solid #eef0f3; }
.control-title { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.control-title h2 { margin: 0; font-size: 16px; }
.control-title span { font-size: 12px; }
.control-section :deep(.el-form-item) { margin: 0; }
.source-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px; }
.source-grid button { display: grid; min-width: 0; min-height: 64px; padding: 8px 4px; place-items: center; gap: 5px; border: 1px solid #e3e6eb; border-radius: 6px; background: #fff; color: #61666d; cursor: pointer; }
.source-grid button:hover, .source-grid button.active { border-color: #00aeec; background: #f1fbff; color: #00aeec; }
.source-grid button:disabled { cursor: not-allowed; opacity: .55; }
.source-grid .el-icon { font-size: 22px; }
.source-grid span { overflow: hidden; max-width: 100%; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.text-layer-actions { display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 10px; }
.text-layer-actions label { display: flex; align-items: center; gap: 6px; color: #61666d; font-size: 12px; }
.text-layer-actions > span, .empty-setting { color: #9499a0; font-size: 12px; }
.color-input { width: 42px; height: 32px; padding: 2px; border: 1px solid #dcdfe6; border-radius: 4px; background: #fff; }
.audio-row { grid-template-columns: 1fr auto; align-items: center; }
.audio-row div { display: grid; gap: 4px; }
.audio-row span { font-size: 12px; }
.monitor-setting { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: center; gap: 12px; padding-top: 2px; }
.monitor-setting > div { display: grid; gap: 3px; }
.monitor-setting strong { font-size: 13px; }
.monitor-setting span { color: #9499a0; font-size: 12px; }
.advanced-settings { padding: 14px 0; border-bottom: 1px solid #eef0f3; }
.advanced-settings summary { color: #61666d; cursor: pointer; font-size: 13px; }
.advanced-settings section { display: grid; gap: 12px; padding-top: 14px; }
.studio-actions { display: grid; gap: 9px; padding: 18px 0; }
.studio-actions > span { color: #9499a0; font-size: 12px; text-align: center; }
.studio-actions .el-button { margin: 0; }
@media (max-width: 980px) { .studio-grid { grid-template-columns: 1fr; } }
@media (max-width: 640px) { .studio-head { align-items: flex-start; flex-direction: column; } .control-panel { padding: 0 14px; } .monitor-setting { align-items: start; grid-template-columns: 1fr; } }
</style>
