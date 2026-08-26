<template>
  <div class="sc-overlay" aria-live="polite" aria-atomic="true">
    <Transition name="sc-card" mode="out-in">
      <article v-if="active" :key="active.item.id" :class="['sc-card', levelClass]">
        <header>
          <el-avatar :size="34" :src="mediaUrl(active.item.user?.avatar || '')">
            {{ displayName.slice(0, 1) }}
          </el-avatar>
          <div>
            <strong>{{ displayName }}</strong>
            <span>{{ active.item.gift.name }} × {{ active.item.count }} · {{ active.item.value }} 助力</span>
          </div>
          <b>SC</b>
        </header>
        <p>{{ active.item.message }}</p>
        <i :style="{ animationDuration: `${active.remainingMs}ms` }" />
      </article>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import type { LiveMonitorSuperChat } from '@/types'
import { mediaUrl } from '@/utils/format'

const props = defineProps<{ items: LiveMonitorSuperChat[] }>()

type QueueItem = { item: LiveMonitorSuperChat; remainingMs: number }

const active = ref<QueueItem>()
const queue: QueueItem[] = []
const seen = new Set<number>()
let hideTimer: ReturnType<typeof setTimeout> | undefined

const displayName = computed(() => active.value?.item.user?.nickname || active.value?.item.user?.username || '观众')
const levelClass = computed(() => {
  const value = active.value?.item.value || 0
  if (value >= 500) return 'level-high'
  if (value >= 100) return 'level-mid'
  return 'level-base'
})

watch(
  () => props.items.map(item => `${item.id}:${item.displaySeconds}`).join(','),
  () => {
    for (const item of props.items) {
      if (seen.has(item.id) || !item.message || item.displaySeconds <= 0) continue
      seen.add(item.id)
      const createdAt = Date.parse(item.createdAt || '') || Date.now()
      const remainingMs = item.displaySeconds * 1000 - Math.max(0, Date.now() - createdAt)
      if (remainingMs > 0) queue.push({ item, remainingMs })
    }
    showNext()
  },
  { immediate: true },
)

function showNext() {
  if (active.value || !queue.length) return
  active.value = queue.shift()
  if (!active.value) return
  hideTimer = setTimeout(() => {
    active.value = undefined
    hideTimer = undefined
    window.setTimeout(showNext, 180)
  }, active.value.remainingMs)
}

onBeforeUnmount(() => {
  if (hideTimer) clearTimeout(hideTimer)
})
</script>

<style scoped>
.sc-overlay {
  position: absolute;
  top: 18px;
  left: 50%;
  z-index: 12;
  width: min(520px, calc(100% - 32px));
  pointer-events: none;
  transform: translateX(-50%);
}

.sc-card {
  position: relative;
  overflow: hidden;
  border: 1px solid rgb(255 255 255 / 22%);
  border-radius: 6px;
  background: rgb(22 125 180 / 94%);
  color: #fff;
  box-shadow: 0 8px 28px rgb(0 0 0 / 28%);
}

.sc-card.level-mid { background: rgb(214 123 30 / 96%); }
.sc-card.level-high { background: rgb(196 54 66 / 96%); }

.sc-card header {
  display: grid;
  grid-template-columns: 34px minmax(0, 1fr) auto;
  align-items: center;
  gap: 10px;
  padding: 9px 12px;
  background: rgb(0 0 0 / 14%);
}

.sc-card header > div {
  display: grid;
  min-width: 0;
  gap: 2px;
}

.sc-card strong,
.sc-card span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sc-card strong { font-size: 13px; }
.sc-card span { color: rgb(255 255 255 / 76%); font-size: 11px; }
.sc-card b { font-size: 13px; letter-spacing: 0; }

.sc-card p {
  margin: 0;
  padding: 12px 14px 14px;
  overflow-wrap: anywhere;
  font-size: 15px;
  line-height: 1.55;
}

.sc-card > i {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 3px;
  background: rgb(255 255 255 / 78%);
  animation: sc-countdown linear forwards;
  transform-origin: left;
}

@keyframes sc-countdown {
  from { transform: scaleX(1); }
  to { transform: scaleX(0); }
}

.sc-card-enter-active,
.sc-card-leave-active { transition: opacity 0.2s ease, transform 0.2s ease; }
.sc-card-enter-from { opacity: 0; transform: translateY(-10px); }
.sc-card-leave-to { opacity: 0; transform: translateY(-6px); }

@media (max-width: 640px) {
  .sc-overlay { top: 10px; width: calc(100% - 20px); }
  .sc-card p { padding: 10px 12px 12px; font-size: 13px; }
}

@media (prefers-reduced-motion: reduce) {
  .sc-card-enter-active,
  .sc-card-leave-active { transition: none; }
}
</style>
