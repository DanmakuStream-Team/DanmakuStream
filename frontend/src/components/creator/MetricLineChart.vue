<template>
  <div class="metric-chart">
    <div class="legend" aria-label="图例">
      <span v-for="item in series" :key="item.key">
        <i :style="{ backgroundColor: item.color }" />{{ item.label }}
      </span>
    </div>

    <svg
      ref="svgRef"
      class="chart-svg"
      viewBox="0 0 720 220"
      role="img"
      :aria-label="ariaLabel"
      @pointermove="handlePointerMove"
      @pointerleave="activeIndex = -1"
    >
      <g class="grid-lines">
        <template v-for="tick in yTicks" :key="tick.y">
          <line :x1="padding.left" :x2="width - padding.right" :y1="tick.y" :y2="tick.y" />
          <text :x="padding.left - 9" :y="tick.y + 4" text-anchor="end">{{ formatValue(tick.value) }}</text>
        </template>
      </g>

      <g class="x-labels">
        <text
          v-for="label in xLabels"
          :key="label.index"
          :x="xAt(label.index)"
          :y="height - 8"
          text-anchor="middle"
        >{{ label.text }}</text>
      </g>

      <path
        v-for="item in chartSeries"
        :key="item.key"
        class="metric-line"
        :d="item.path"
        :stroke="item.color"
      />

      <template v-if="activeIndex >= 0">
        <line class="cursor-line" :x1="xAt(activeIndex)" :x2="xAt(activeIndex)" :y1="padding.top" :y2="plotBottom" />
        <circle
          v-for="item in chartSeries"
          :key="item.key"
          :cx="xAt(activeIndex)"
          :cy="yAt(item.values[activeIndex])"
          r="4"
          :fill="item.color"
        />
      </template>
    </svg>

    <div v-if="activePoint" class="chart-tooltip" :style="tooltipStyle">
      <strong>{{ activePoint.date.slice(5).replace('-', '/') }}</strong>
      <span v-for="item in chartSeries" :key="item.key">
        <i :style="{ backgroundColor: item.color }" />{{ item.label }} {{ formatValue(item.values[activeIndex]) }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { CreatorAnalyticsPoint } from '@/types'

type NumericMetric = 'views' | 'collects' | 'growthSpeed' | 'streams'

interface SeriesOption {
  key: NumericMetric
  label: string
  color: string
}

const props = defineProps<{
  points: CreatorAnalyticsPoint[]
  series: SeriesOption[]
  ariaLabel: string
}>()

const width = 720
const height = 220
const padding = { top: 18, right: 18, bottom: 34, left: 50 }
const plotBottom = height - padding.bottom
const svgRef = ref<SVGElement>()
const activeIndex = ref(-1)

const valueBounds = computed(() => {
  const values = props.series.flatMap(item => props.points.map(point => Math.max(0, point[item.key])))
  const rawValues = props.series.flatMap(item => props.points.map(point => point[item.key]))
  return {
    min: Math.min(0, ...rawValues),
    max: Math.max(1, ...values),
  }
})

const yTicks = computed(() => Array.from({ length: 4 }, (_, index) => {
  const ratio = index / 3
  return {
    y: padding.top + (plotBottom - padding.top) * ratio,
    value: Math.round(valueBounds.value.max - (valueBounds.value.max - valueBounds.value.min) * ratio),
  }
}))

const chartSeries = computed(() => props.series.map(item => {
  const values = props.points.map(point => point[item.key])
  const path = values.map((value, index) => `${index === 0 ? 'M' : 'L'} ${xAt(index)} ${yAt(value)}`).join(' ')
  return { ...item, values, path }
}))

const xLabels = computed(() => {
  if (!props.points.length) return []
  const count = props.points.length <= 7 ? props.points.length : 6
  const indexes = new Set<number>()
  for (let i = 0; i < count; i++) {
    indexes.add(Math.round((props.points.length - 1) * i / Math.max(1, count - 1)))
  }
  return [...indexes].map(index => ({ index, text: props.points[index].date.slice(5).replace('-', '/') }))
})

const activePoint = computed(() => activeIndex.value >= 0 ? props.points[activeIndex.value] : undefined)
const tooltipStyle = computed(() => {
  const ratio = props.points.length <= 1 ? 0 : activeIndex.value / (props.points.length - 1)
  return ratio > 0.68 ? { right: '14px' } : { left: `${Math.max(14, ratio * 82)}%` }
})

function xAt(index: number) {
  if (props.points.length <= 1) return padding.left
  return padding.left + index * (width - padding.left - padding.right) / (props.points.length - 1)
}

function yAt(value: number) {
  const range = valueBounds.value.max - valueBounds.value.min
  return plotBottom - (value - valueBounds.value.min) / range * (plotBottom - padding.top)
}

function handlePointerMove(event: PointerEvent) {
  if (!svgRef.value || !props.points.length) return
  const rect = svgRef.value.getBoundingClientRect()
  const relativeX = (event.clientX - rect.left) / rect.width * width
  const ratio = (relativeX - padding.left) / (width - padding.left - padding.right)
  activeIndex.value = Math.max(0, Math.min(props.points.length - 1, Math.round(ratio * (props.points.length - 1))))
}

function formatValue(value: number) {
  return Intl.NumberFormat('zh-CN', { maximumFractionDigits: 1 }).format(value)
}
</script>

<style scoped>
.metric-chart {
  position: relative;
  min-width: 0;
}

.legend {
  display: flex;
  min-height: 24px;
  align-items: center;
  gap: 16px;
  color: #667085;
  font-size: 12px;
}

.legend span,
.chart-tooltip span {
  display: flex;
  align-items: center;
  gap: 6px;
}

.legend i,
.chart-tooltip i {
  width: 8px;
  height: 8px;
  border-radius: 2px;
}

.chart-svg {
  display: block;
  width: 100%;
  min-height: 220px;
  overflow: visible;
  touch-action: none;
}

.grid-lines line {
  stroke: #e8ecf2;
  stroke-width: 1;
}

.grid-lines text,
.x-labels text {
  fill: #98a2b3;
  font-size: 11px;
}

.metric-line {
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2.5;
}

.cursor-line {
  stroke: #cbd2dc;
  stroke-dasharray: 4 4;
}

.chart-tooltip {
  position: absolute;
  top: 38px;
  z-index: 2;
  display: grid;
  min-width: 126px;
  gap: 7px;
  padding: 10px;
  border: 1px solid #e4e7ec;
  border-radius: 6px;
  background: rgb(255 255 255 / 96%);
  box-shadow: 0 6px 18px rgb(16 24 40 / 10%);
  color: #475467;
  font-size: 12px;
  pointer-events: none;
}

.chart-tooltip strong {
  color: #101828;
}
</style>
