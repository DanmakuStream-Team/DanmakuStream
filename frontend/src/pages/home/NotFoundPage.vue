<template>
  <main class="not-found">
    <section class="soft-panel card">
      <el-result title="页面不存在" sub-title="这个地址暂时没有对应内容">
        <template #icon>
          <el-icon><Warning /></el-icon>
        </template>
        <template #extra>
          <el-button type="primary" @click="router.push('/')">返回首页</el-button>
        </template>
      </el-result>
    </section>
  </main>
</template>

<script setup lang="ts">
import { Warning } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style scoped>
.not-found {
  display: grid;
  min-height: 72vh;
  place-items: center;
  padding: 24px;
}

.card {
  width: min(520px, 100%);
  padding: 20px;
}
</style>
