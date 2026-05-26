<template>
  <main class="page-shell admin-page">
    <div class="section-head">
      <div>
        <h1>管理后台</h1>
        <p class="muted">处理视频审核和弹幕治理。</p>
      </div>
    </div>

    <section class="quick-grid">
      <button class="soft-panel quick-card" type="button" @click="router.push('/admin/videos')">
        <strong>视频审核</strong>
        <span>查看 pending / approved / rejected 视频</span>
      </button>
      <button class="soft-panel quick-card" type="button" @click="router.push('/admin/danmaku')">
        <strong>弹幕管理</strong>
        <span>检索弹幕并屏蔽违规内容</span>
      </button>
    </section>
  </main>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style scoped>
.admin-page {
  display: grid;
  gap: 18px;
}

.section-head p {
  margin: 8px 0 0;
}

.quick-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px;
}

.quick-card {
  display: grid;
  gap: 10px;
  min-height: 180px;
  padding: 24px;
  border: 1px solid rgba(20, 32, 51, 0.08);
  cursor: pointer;
  text-align: left;
}

.quick-card strong {
  color: #142033;
  font-size: 24px;
}

.quick-card span {
  color: #667085;
}

@media (max-width: 720px) {
  .quick-grid {
    grid-template-columns: 1fr;
  }
}
</style>
