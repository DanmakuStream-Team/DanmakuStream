<template>
  <main class="page-shell affinity-page">
    <section class="affinity-head">
      <div>
        <el-tag effect="light">兴趣画像</el-tag>
        <h1>我的标签相关度</h1>
        <p>根据你最近看过的视频标签计算，看的相关标签视频越多，标签框和文字就越醒目。</p>
      </div>
      <el-button :disabled="!tagItems.length" @click="refresh">刷新</el-button>
    </section>

    <section v-if="tagItems.length" class="soft-panel affinity-panel">
      <div class="tag-cloud">
        <button
          v-for="item in tagItems"
          :key="item.name"
          type="button"
          class="affinity-tag"
          :style="tagStyle(item)"
          @click="router.push({ path: '/', query: { tag: item.name } })"
        >
          <span>{{ item.name }}</span>
          <em>{{ item.score }}</em>
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
import { getUserLibraryRecords } from '@/utils/userLibrary'

interface TagAffinityItem {
  name: string
  score: number
}

const router = useRouter()
const version = ref(0)

const tagItems = computed(() => {
  version.value
  const scores = new Map<string, number>()
  const records = getUserLibraryRecords('history')

  records.forEach((record, index) => {
    const tags = normalizeTags(record.video.tags)
    if (!tags.length) return

    const recencyWeight = Math.max(1, 5 - Math.floor(index / 8))
    const progressWeight = Math.max(1, Math.round((record.progress || 20) / 25))
    const score = recencyWeight + progressWeight

    tags.forEach((tag) => {
      scores.set(tag, (scores.get(tag) || 0) + score)
    })
  })

  return Array.from(scores.entries())
    .map(([name, score]) => ({ name, score }))
    .sort((a, b) => b.score - a.score || a.name.localeCompare(b.name))
    .slice(0, 36)
})

const maxScore = computed(() => Math.max(1, ...tagItems.value.map((item) => item.score)))

function refresh() {
  version.value += 1
}

function tagStyle(item: TagAffinityItem) {
  const ratio = item.score / maxScore.value
  const hue = Math.round(205 - ratio * 145)
  const saturation = Math.round(58 + ratio * 22)
  const lightness = Math.round(92 - ratio * 24)
  const borderLightness = Math.round(74 - ratio * 18)
  const fontSize = 14 + ratio * 20
  const paddingY = 8 + ratio * 10
  const paddingX = 12 + ratio * 18

  return {
    '--tag-bg': `hsl(${hue} ${saturation}% ${lightness}%)`,
    '--tag-border': `hsl(${hue} ${saturation}% ${borderLightness}%)`,
    '--tag-color': `hsl(${hue} 58% 20%)`,
    fontSize: `${fontSize}px`,
    padding: `${paddingY}px ${paddingX}px`,
  }
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

.affinity-panel {
  padding: 24px;
}

.tag-cloud {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 14px;
}

.affinity-tag {
  display: inline-flex;
  align-items: baseline;
  gap: 8px;
  border: 1px solid var(--tag-border);
  border-radius: 8px;
  background: var(--tag-bg);
  color: var(--tag-color);
  cursor: pointer;
  font-weight: 800;
  line-height: 1;
  transition: transform 0.16s ease, filter 0.16s ease;
}

.affinity-tag:hover {
  filter: saturate(1.1);
  transform: translateY(-2px);
}

.affinity-tag em {
  font-size: 0.62em;
  font-style: normal;
  opacity: 0.72;
}

@media (max-width: 760px) {
  .affinity-head {
    display: grid;
    align-items: stretch;
  }

  .tag-cloud {
    gap: 10px;
  }
}
</style>
