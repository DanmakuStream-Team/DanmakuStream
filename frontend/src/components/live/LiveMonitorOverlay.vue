<template>
  <section
    v-show="visible"
    ref="panelRef"
    :class="['monitor-panel', { detached, embedded }]"
    :style="panelStyle"
    aria-label="主播互动监看窗"
  >
    <header class="monitor-head" @pointerdown="startDrag">
      <div>
        <i :class="{ connected }" />
        <strong>{{ embedded ? '实时弹幕与 SC' : '互动监看' }}</strong>
        <span>{{ connected ? '实时' : '连接中' }}</span>
      </div>
      <div class="monitor-actions">
        <el-tooltip v-if="!detached && !embedded" content="跨软件置顶悬浮" placement="top">
          <button type="button" aria-label="跨软件置顶悬浮" @pointerdown.stop @click="$emit('popout')">
            <el-icon><FullScreen /></el-icon>
          </button>
        </el-tooltip>
        <el-tooltip v-if="!embedded" :content="detached ? '关闭悬浮窗' : '隐藏监看窗'" placement="top">
          <button type="button" :aria-label="detached ? '关闭悬浮窗' : '隐藏监看窗'" @pointerdown.stop @click="$emit('close')">
            <el-icon><Close /></el-icon>
          </button>
        </el-tooltip>
      </div>
    </header>

    <div class="monitor-filter">
      <button v-for="option in filters" :key="option.value" type="button" :class="{ active: filter === option.value }" @click="filter = option.value">
        {{ option.label }}
        <span v-if="option.value === 'sc' && superChats.length">{{ superChats.length }}</span>
      </button>
    </div>

    <div ref="feedRef" class="monitor-feed">
      <article v-for="item in visibleItems" :key="`${item.kind}-${item.id}`" :class="['monitor-item', item.kind]">
        <template v-if="item.kind === 'chat'">
          <el-avatar :size="26" :src="mediaUrl(item.data.author?.avatar || '')">
            {{ displayName(item.data.author).slice(0, 1) }}
          </el-avatar>
          <div>
            <span>{{ displayName(item.data.author) }}</span>
            <p>{{ item.data.content }}</p>
          </div>
        </template>
        <template v-else>
          <span class="sc-mark">SC</span>
          <div>
            <span>{{ displayName(item.data.user) }} · {{ item.data.gift.name }} × {{ item.data.count }}</span>
            <p>{{ item.data.message }}</p>
            <small>{{ item.data.value }} 助力 · 悬浮 {{ item.data.displaySeconds }} 秒</small>
          </div>
          <strong>+{{ item.data.value }}</strong>
        </template>
      </article>
      <div v-if="!visibleItems.length" class="monitor-empty">
        {{ filter === 'sc' ? '暂无 SC' : '等待观众互动' }}
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { Close, FullScreen } from '@element-plus/icons-vue'
import type { CSSProperties } from 'vue'
import type { Danmaku, LiveMonitorSuperChat, UserInfo } from '@/types'
import { mediaUrl } from '@/utils/format'

const props = defineProps<{
  visible: boolean
  connected: boolean
  messages: Danmaku[]
  superChats: LiveMonitorSuperChat[]
  detached?: boolean
  embedded?: boolean
}>()

defineEmits<{ close: []; popout: [] }>()

type Filter = 'all' | 'chat' | 'sc'
type MonitorItem =
  | { kind: 'chat'; id: number; timestamp: number; data: Danmaku }
  | { kind: 'sc'; id: number; timestamp: number; data: LiveMonitorSuperChat }

const panelRef = ref<HTMLElement>()
const feedRef = ref<HTMLElement>()
const filter = ref<Filter>('all')
const offset = ref({ x: 0, y: 0 })
const filters: Array<{ label: string; value: Filter }> = [
  { label: '全部', value: 'all' },
  { label: '评论', value: 'chat' },
  { label: 'SC', value: 'sc' },
]

const panelStyle = computed<CSSProperties>(() => props.embedded ? {} : ({ transform: `translate(${offset.value.x}px, ${offset.value.y}px)` }))
const visibleItems = computed<MonitorItem[]>(() => {
  const chats: MonitorItem[] = props.messages.map((data) => ({
    kind: 'chat', id: data.id, data, timestamp: Date.parse(data.createdAt || '') || data.id,
  }))
  const sc: MonitorItem[] = props.superChats.map((data) => ({
    kind: 'sc', id: data.id, data, timestamp: Date.parse(data.createdAt || '') || data.id,
  }))
  return [...chats, ...sc]
    .filter((item) => filter.value === 'all' || item.kind === filter.value)
    .sort((a, b) => a.timestamp - b.timestamp)
    .slice(-100)
})

watch(() => visibleItems.value.length, async () => {
  const feed = feedRef.value
  const shouldFollow = !feed || feed.scrollHeight - feed.scrollTop - feed.clientHeight < 56
  if (!shouldFollow) return
  await nextTick()
  if (feedRef.value) feedRef.value.scrollTop = feedRef.value.scrollHeight
})

function displayName(user?: UserInfo) {
  return user?.nickname || user?.username || '观众'
}

function startDrag(event: PointerEvent) {
  if (props.detached || props.embedded || event.button !== 0 || !panelRef.value) return
  const panel = panelRef.value
  const container = panel.offsetParent as HTMLElement | null
  if (!container) return
  const start = { x: event.clientX, y: event.clientY, offsetX: offset.value.x, offsetY: offset.value.y }
  panel.setPointerCapture(event.pointerId)

  const move = (moveEvent: PointerEvent) => {
    const panelRect = panel.getBoundingClientRect()
    const containerRect = container.getBoundingClientRect()
    const nextX = start.offsetX + moveEvent.clientX - start.x
    const nextY = start.offsetY + moveEvent.clientY - start.y
    offset.value = {
      x: Math.min(Math.max(nextX, containerRect.left - panelRect.left + offset.value.x), containerRect.right - panelRect.right + offset.value.x),
      y: Math.min(Math.max(nextY, containerRect.top - panelRect.top + offset.value.y), containerRect.bottom - panelRect.bottom + offset.value.y),
    }
  }
  const stop = () => {
    panel.removeEventListener('pointermove', move)
    panel.removeEventListener('pointerup', stop)
    panel.removeEventListener('pointercancel', stop)
  }
  panel.addEventListener('pointermove', move)
  panel.addEventListener('pointerup', stop)
  panel.addEventListener('pointercancel', stop)
}
</script>

<style scoped>
.monitor-panel { position: absolute; right: 16px; bottom: 16px; z-index: 8; display: grid; overflow: hidden; width: min(320px, calc(100% - 32px)); height: min(310px, calc(100% - 72px)); grid-template-rows: auto auto minmax(0, 1fr); border: 1px solid rgb(255 255 255 / 16%); border-radius: 6px; background: rgb(15 18 24 / 88%); color: #fff; backdrop-filter: blur(12px); box-shadow: 0 10px 30px rgb(0 0 0 / 28%); }
.monitor-head { display: flex; min-height: 42px; align-items: center; justify-content: space-between; padding: 0 8px 0 12px; border-bottom: 1px solid rgb(255 255 255 / 10%); cursor: move; user-select: none; touch-action: none; }
.monitor-head > div { display: flex; align-items: center; gap: 7px; }
.monitor-head i { width: 7px; height: 7px; border-radius: 50%; background: #6b7280; }
.monitor-head i.connected { background: #22c55e; }
.monitor-head strong { font-size: 13px; }
.monitor-head span { color: rgb(255 255 255 / 56%); font-size: 11px; }
.monitor-head button { display: grid; width: 28px; height: 28px; padding: 0; place-items: center; border: 0; background: transparent; color: rgb(255 255 255 / 72%); cursor: pointer; }
.monitor-actions { display: flex !important; gap: 2px !important; }
.monitor-panel.detached { position: relative; inset: auto; width: 100%; height: 100%; border: 0; border-radius: 0; transform: none !important; box-shadow: none; }
.monitor-panel.detached .monitor-head { cursor: default; }
.monitor-panel.embedded { position: relative; inset: auto; width: 100%; height: 280px; border: 1px solid #e3e6eb; border-radius: 6px; background: #fff; color: #18191c; backdrop-filter: none; transform: none !important; box-shadow: none; }
.monitor-panel.embedded .monitor-head { cursor: default; border-bottom-color: #eef0f3; }
.monitor-panel.embedded .monitor-head i { background: #c9ccd0; }
.monitor-panel.embedded .monitor-head i.connected { background: #22c55e; }
.monitor-panel.embedded .monitor-head span,
.monitor-panel.embedded .monitor-item span { color: #9499a0; }
.monitor-panel.embedded .monitor-filter button { color: #61666d; }
.monitor-panel.embedded .monitor-filter button.active { background: #f1f2f3; color: #18191c; }
.monitor-panel.embedded .monitor-item { border-top-color: #eef0f3; }
.monitor-panel.embedded .monitor-item p { color: #3f444d; }
.monitor-panel.embedded .monitor-item small { color: #9499a0; }
.monitor-panel.embedded .monitor-item.sc { background: #fff8e8; }
.monitor-panel.embedded .monitor-empty { color: #9499a0; }
.monitor-filter { display: flex; gap: 4px; padding: 8px 10px; }
.monitor-filter button { height: 26px; padding: 0 9px; border: 0; border-radius: 4px; background: transparent; color: rgb(255 255 255 / 62%); cursor: pointer; font-size: 12px; }
.monitor-filter button.active { background: rgb(255 255 255 / 12%); color: #fff; }
.monitor-filter span { margin-left: 3px; color: #ffce70; }
.monitor-feed { overflow-y: auto; min-height: 0; padding: 0 10px 10px; scrollbar-width: thin; scrollbar-color: rgb(255 255 255 / 20%) transparent; }
.monitor-item { display: grid; grid-template-columns: 26px minmax(0, 1fr); gap: 8px; padding: 8px 2px; border-top: 1px solid rgb(255 255 255 / 8%); }
.monitor-item > div { min-width: 0; }
.monitor-item span { color: rgb(255 255 255 / 58%); font-size: 11px; }
.monitor-item p { margin: 2px 0 0; overflow-wrap: anywhere; color: rgb(255 255 255 / 90%); font-size: 12px; line-height: 1.5; }
.monitor-item small { display: block; margin-top: 3px; color: rgb(255 255 255 / 42%); font-size: 10px; }
.monitor-item.sc { grid-template-columns: 28px minmax(0, 1fr) auto; margin: 4px 0; padding: 9px 8px; border: 0; border-left: 3px solid #f5b942; border-radius: 4px; background: rgb(245 185 66 / 13%); }
.monitor-item.sc > strong { align-self: center; color: #ffce70; font-size: 12px; }
.sc-mark { display: grid; width: 28px; height: 22px; place-items: center; border-radius: 3px; background: #f5b942; color: #161616 !important; font-size: 10px !important; font-weight: 800; }
.monitor-empty { display: grid; height: 100%; min-height: 100px; place-items: center; color: rgb(255 255 255 / 42%); font-size: 12px; }
@media (max-width: 640px) { .monitor-panel { right: 8px; bottom: 8px; width: min(280px, calc(100% - 16px)); height: min(240px, calc(100% - 16px)); } }
</style>
