<template>
  <main class="page-shell affinity-page">
    <section class="affinity-head">
      <div>
        <el-tag effect="light">兴趣画像</el-tag>
        <h1>我的标签相关度</h1>
        <p>根据最近观看历史生成关键词云，看的相关标签视频越多，词语越大、越靠近中心。</p>
      </div>
      <div class="head-actions">
        <el-button @click="seedDemoData">生成示例数据</el-button>
        <el-button :disabled="!cloudItems.length" @click="refresh">刷新</el-button>
      </div>
    </section>

    <section v-if="cloudItems.length" class="soft-panel cloud-panel">
      <div class="cloud-canvas">
        <button
          v-for="item in cloudItems"
          :key="item.name"
          type="button"
          class="cloud-word"
          :style="wordStyle(item)"
          :title="`${item.name}: ${item.score}`"
          @click="router.push({ path: '/', query: { tag: item.name } })"
        >
          {{ item.name }}
        </button>
      </div>
    </section>

    <section v-else class="soft-panel empty-panel">
      <el-empty description="还没有足够的观看记录，先去看几个带标签的视频吧" />
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { normalizeTags } from '@/utils/format'
import { getUserLibraryRecords, upsertUserLibraryRecord } from '@/utils/userLibrary'
import type { VideoInfo } from '@/types'

interface TagAffinityItem {
  name: string
  score: number
}

interface CloudItem extends TagAffinityItem {
  x: number
  y: number
  rotate: number
  size: number
  color: string
  weight: number
}

const router = useRouter()
const version = ref(0)
const palette = ['#2276b7', '#f05a28', '#39a65a', '#b04fb3', '#e0a51c', '#5d70c9', '#d84f65', '#33a6a6', '#7f8c8d']
const compactSlots = [
  [50, 50, 0], [40, 48, 0], [60, 50, 0], [50, 39, 0], [49, 61, 0],
  [34, 56, 0], [66, 42, 0], [36, 38, 0], [65, 62, 0], [44, 31, 0],
  [56, 30, 0], [30, 45, 90], [71, 53, 0], [29, 63, 0], [72, 34, 0],
  [42, 68, 0], [58, 69, 90], [23, 52, 0], [77, 64, 0], [24, 35, 0],
  [77, 42, 90], [36, 25, 0], [63, 24, 0], [18, 44, 0], [82, 55, 0],
  [27, 73, 0], [73, 73, 0], [17, 62, 90], [83, 29, 0], [47, 22, 90],
  [53, 78, 0], [33, 76, 90], [67, 20, 0], [16, 29, 0], [86, 69, 90],
  [23, 23, 0], [77, 79, 0], [41, 82, 0], [58, 17, 90], [88, 43, 0],
  [12, 52, 0], [89, 22, 0], [12, 73, 0], [29, 15, 0], [70, 14, 0],
  [46, 14, 0], [54, 87, 0], [20, 82, 90], [81, 85, 0], [10, 34, 0],
] as const
const demoTags = [
  '游戏', '直播', '二次元', '音乐', '科技', '编程', 'Go', 'Vue', '弹幕', '剪辑',
  '动漫', '学习', '生活', '美食', '电影', '开源', '前端', '后端', '算法', '设计',
  '旅行', '校园', '整活', '教程', '赛事', '主播', '热点', '工具', 'AI', '数据库',
]

const tagItems = computed(() => {
  version.value
  const scores = new Map<string, number>()
  const records = getUserLibraryRecords('history')

  records.forEach((record, index) => {
    const tags = normalizeTags(record.video.tags)
    if (!tags.length) return

    const recencyWeight = Math.max(1, 6 - Math.floor(index / 6))
    const progressWeight = Math.max(1, Math.round((record.progress || 20) / 22))
    const score = recencyWeight + progressWeight

    tags.forEach((tag) => {
      scores.set(tag, (scores.get(tag) || 0) + score)
    })
  })

  return Array.from(scores.entries())
    .map(([name, score]) => ({ name, score }))
    .sort((a, b) => b.score - a.score || a.name.localeCompare(b.name))
    .slice(0, 64)
})

const maxScore = computed(() => Math.max(1, ...tagItems.value.map((item) => item.score)))

const cloudItems = computed<CloudItem[]>(() => {
  return tagItems.value.map((item, index) => {
    const weight = item.score / maxScore.value
    const slot = compactSlots[index] || fallbackSlot(index)
    const rotate = index < 12 ? 0 : slot[2]
    const size = 16 + Math.pow(weight, 0.76) * 36
    const color = palette[hashNumber(item.name) % palette.length]

    return {
      ...item,
      x: slot[0],
      y: slot[1],
      rotate,
      size,
      color,
      weight,
    }
  })
})

function fallbackSlot(index: number) {
  const offset = index - compactSlots.length
  const col = offset % 8
  const row = Math.floor(offset / 8)
  const x = 14 + col * 10
  const y = 14 + (row % 7) * 12
  const rotate = (index + row) % 4 === 0 ? 90 : 0
  return [x, y, rotate] as const
}

function refresh() {
  version.value += 1
}

function seedDemoData() {
  for (let i = 0; i < 42; i++) {
    const tagCount = 2 + (hashNumber(`count:${i}:${Date.now()}`) % 4)
    const tags = pickDemoTags(i, tagCount)
    const video: VideoInfo = {
      id: 900000 + i,
      title: `示例视频 ${i + 1}`,
      description: '用于生成标签相关度词云的本地示例数据',
      coverUrl: '',
      videoUrl: '',
      duration: 180 + i * 7,
      viewCount: 1000 + i * 31,
      likeCount: 20 + i,
      collectCount: 8 + i,
      danmakuCount: 12 + i,
      status: 'approved',
      category: 'demo',
      author: {
        id: 1,
        username: 'demo',
        nickname: '示例用户',
        avatar: '',
        bio: '',
        role: 'user',
        followCount: 0,
        fanCount: 0,
      },
      tags: tags.join(','),
      createdAt: new Date(Date.now() - i * 3600_000).toISOString(),
    }
    upsertUserLibraryRecord('history', video, 20 + (hashNumber(`progress:${i}:${Date.now()}`) % 80))
  }
  refresh()
}

function pickDemoTags(seed: number, count: number) {
  const list: string[] = []
  let cursor = hashNumber(`${seed}:${Date.now()}`)
  while (list.length < count) {
    const tag = demoTags[cursor % demoTags.length]
    if (!list.includes(tag)) list.push(tag)
    cursor = hashNumber(`${cursor}:${tag}`)
  }
  return list
}

function wordStyle(item: CloudItem) {
  return {
    left: `${item.x}%`,
    top: `${item.y}%`,
    color: item.color,
    fontSize: `${item.size}px`,
    fontWeight: String(Math.round(620 + item.weight * 280)),
    opacity: String(0.62 + item.weight * 0.38),
    transform: `translate(-50%, -50%) rotate(${item.rotate}deg)`,
    zIndex: String(Math.round(10 + item.weight * 20)),
  }
}

function hashNumber(value: string) {
  let hash = 2166136261
  for (let i = 0; i < value.length; i++) {
    hash ^= value.charCodeAt(i)
    hash = Math.imul(hash, 16777619)
  }
  return Math.abs(hash)
}
</script>

<style scoped>
.affinity-page {
  display: grid;
  gap: 20px;
  padding-top: 22px;
}

.affinity-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18px;
}

.head-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: flex-end;
}

.affinity-head h1 {
  margin: 10px 0 8px;
  color: #18191c;
  font-size: 32px;
  font-weight: 900;
  line-height: 1.2;
}

.affinity-head p {
  margin: 0;
  color: #667085;
  line-height: 1.7;
}

.cloud-panel {
  padding: 18px;
}

.cloud-canvas {
  position: relative;
  overflow: hidden;
  max-width: 980px;
  min-height: 460px;
  margin: 0 auto;
  border: 1px dashed #d0d5dd;
  border-radius: 8px;
  background:
    radial-gradient(circle at center, rgb(0 174 236 / 8%), transparent 48%),
    #fff;
}

.cloud-canvas::before {
  position: absolute;
  inset: 36px 9%;
  border: 1px dashed rgb(148 163 184 / 38%);
  border-radius: 50%;
  content: '';
}

.cloud-word {
  position: absolute;
  max-width: 220px;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
  font-family: "Microsoft YaHei", "PingFang SC", Arial, sans-serif;
  line-height: 1;
  text-align: center;
  text-shadow: 0 1px 0 rgb(255 255 255 / 80%);
  white-space: nowrap;
  transition: filter 0.16s ease, transform 0.16s ease;
}

.cloud-word:hover {
  filter: saturate(1.25) brightness(0.95);
}

@media (max-width: 760px) {
  .affinity-head {
    display: grid;
    align-items: stretch;
  }

  .cloud-canvas {
    min-height: 420px;
  }
}
</style>
