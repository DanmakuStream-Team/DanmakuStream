<template>
  <main class="page-shell danmaku-admin-page">
    <div class="section-head">
      <h1>弹幕管理</h1>
      <el-input v-model="keyword" class="search" placeholder="搜索弹幕内容" clearable @keyup.enter="load" />
    </div>

    <section class="soft-panel list-panel" v-loading="loading">
      <div v-if="items.length" class="rows">
        <div v-for="item in items" :key="item.id" class="row">
          <div class="badge">#{{ item.id }}</div>
          <div>
            <strong>{{ item.content }}</strong>
            <span>视频 {{ item.videoId }} · {{ item.time }} 秒 · 用户 {{ item.userId }}</span>
          </div>
          <el-tag :type="item.blocked ? 'danger' : 'success'">{{ item.blocked ? '已屏蔽' : '正常' }}</el-tag>
          <el-button type="danger" :disabled="item.blocked" @click="block(item)">屏蔽</el-button>
        </div>
      </div>
      <el-empty v-else description="暂无弹幕" />
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { danmakuApi } from '@/api/danmaku'
import type { Danmaku } from '@/types'

const loading = ref(false)
const keyword = ref('')
const items = ref<Danmaku[]>([])

onMounted(load)

async function load() {
  loading.value = true
  try {
    const res = await danmakuApi.adminList({
      page: 1,
      pageSize: 50,
      keyword: keyword.value.trim() || undefined,
    })
    items.value = res.data.list
  } finally {
    loading.value = false
  }
}

async function block(item: Danmaku) {
  await danmakuApi.block(item.id)
  item.blocked = true
  ElMessage.success('弹幕已屏蔽')
}
</script>

<style scoped>
.danmaku-admin-page {
  display: grid;
  gap: 18px;
}

.search {
  width: 280px;
}

.list-panel {
  padding: 18px;
}

.rows {
  display: grid;
  gap: 12px;
}

.row {
  display: grid;
  grid-template-columns: 64px minmax(0, 1fr) auto auto;
  align-items: center;
  gap: 14px;
  padding: 12px;
  border-radius: 8px;
  background: #f7f9fc;
}

.badge {
  display: grid;
  height: 42px;
  place-items: center;
  border-radius: 8px;
  background: #eef4ff;
  color: #165dff;
  font-weight: 800;
}

.row strong,
.row span {
  display: block;
}

.row span {
  margin-top: 6px;
  color: #667085;
  font-size: 13px;
}

@media (max-width: 760px) {
  .search {
    width: 100%;
  }

  .row {
    grid-template-columns: 64px 1fr;
  }
}
</style>
