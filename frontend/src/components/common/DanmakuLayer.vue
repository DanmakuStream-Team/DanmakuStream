<template>
  <div class="danmaku-layer">
    <span
      v-for="item in visibleItems"
      :key="item.id"
      class="danmaku"
      :style="{ top: `${item.row * 34 + 18}px`, color: item.color || '#fff' }"
    >
      {{ item.content }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Danmaku } from '@/types'

const props = defineProps<{ items: Danmaku[]; currentTime: number }>()

const visibleItems = computed(() => {
  return props.items
    .filter((item) => Math.abs(item.time - props.currentTime) < 1.2)
    .slice(0, 6)
    .map((item, index) => ({ ...item, row: index % 5 }))
})
</script>

<style scoped>
.danmaku-layer {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
}

.danmaku {
  position: absolute;
  left: 100%;
  min-width: max-content;
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.35);
  font-size: 14px;
  font-weight: 600;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.45);
  animation: fly 6s linear forwards;
}

@keyframes fly {
  to {
    transform: translateX(calc(-100vw - 100%));
  }
}
</style>
