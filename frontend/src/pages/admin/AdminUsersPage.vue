<template>
  <main class="page-shell admin-users-page">
    <div class="section-head">
      <div>
        <h1>用户与权限</h1>
        <p class="muted">查看用户画像，分配超级管理员、版主和普通用户角色。</p>
      </div>
      <el-input v-model="keyword" class="search" placeholder="搜索用户" clearable @keyup.enter="load" />
    </div>

    <section class="soft-panel list-panel" v-loading="loading">
      <div v-if="users.length" class="user-rows">
        <article v-for="user in users" :key="user.id" class="user-row">
          <el-avatar :size="42" :src="mediaUrl(user.avatar)">
            {{ user.nickname.slice(0, 1) }}
          </el-avatar>
          <div class="user-main">
            <strong>{{ user.nickname }}</strong>
            <span>{{ user.username }} · 注册 {{ formatTime(user.createdAt) }}</span>
          </div>
          <div class="user-stats">
            <span>{{ user.videoCount }} 视频</span>
            <span>{{ user.danmakuCount }} 弹幕</span>
            <span>{{ user.fanCount }} 粉丝</span>
          </div>
          <el-select v-model="user.role" size="small" @change="updateRole(user)">
            <el-option label="普通用户" value="user" />
            <el-option label="内容审核员/版主" value="moderator" />
            <el-option label="超级管理员" value="admin" />
          </el-select>
        </article>
      </div>
      <el-empty v-else description="暂无用户" />

      <div class="pager">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          background
          layout="prev, pager, next"
          @current-change="load"
        />
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi, type AdminUserItem } from '@/api/admin'
import { formatTime, mediaUrl } from '@/utils/format'

const loading = ref(false)
const users = ref<AdminUserItem[]>([])
const keyword = ref('')
const page = ref(1)
const pageSize = 20
const total = ref(0)

onMounted(load)

async function load() {
  loading.value = true
  try {
    const res = await adminApi.users({
      page: page.value,
      pageSize,
      keyword: keyword.value.trim() || undefined,
    })
    users.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

async function updateRole(user: AdminUserItem) {
  await adminApi.updateUserRole(user.id, user.role)
  ElMessage.success('角色已更新')
}
</script>

<style scoped>
.admin-users-page {
  display: grid;
  gap: 18px;
}

.section-head p {
  margin: 8px 0 0;
}

.search {
  width: 280px;
}

.list-panel {
  padding: 18px;
}

.user-rows {
  display: grid;
  gap: 12px;
}

.user-row {
  display: grid;
  grid-template-columns: 42px minmax(0, 1.3fr) minmax(280px, 1fr) 170px;
  align-items: center;
  gap: 14px;
  padding: 12px;
  border-radius: 8px;
  background: #f7f9fc;
}

.user-main,
.user-stats {
  min-width: 0;
}

.user-main strong,
.user-main span {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-main strong {
  color: #18191c;
  font-weight: 900;
}

.user-main span,
.user-stats span {
  color: #667085;
  font-size: 13px;
}

.user-stats {
  display: flex;
  gap: 12px;
}

.pager {
  display: flex;
  justify-content: center;
  margin-top: 18px;
}

@media (max-width: 920px) {
  .search {
    width: 100%;
  }

  .user-row {
    grid-template-columns: 42px minmax(0, 1fr);
  }
}
</style>
