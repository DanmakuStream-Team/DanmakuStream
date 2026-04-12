<template>
  <a-layout class="layout">
    <a-layout-header class="header">
      <div class="logo" @click="router.push('/')">灵视 VisionLive</div>
      <div class="search-bar">
        <a-input-search
          v-model="searchKeyword"
          placeholder="搜索视频"
          size="large"
          @search="handleSearch"
        />
      </div>
      <div class="nav-actions">
        <template v-if="authStore.isLoggedIn">
          <a-button @click="router.push('/creator/upload')">投稿</a-button>
          <a-dropdown>
            <a-avatar :image-url="authStore.userInfo?.avatar" />
            <template #content>
              <a-doption @click="router.push(`/user/${authStore.userInfo?.id}`)">个人主页</a-doption>
              <a-doption v-if="authStore.isCreator" @click="router.push('/creator')">创作者中心</a-doption>
              <a-doption v-if="authStore.isAdmin" @click="router.push('/admin')">管理后台</a-doption>
              <a-doption @click="handleLogout">退出登录</a-doption>
            </template>
          </a-dropdown>
        </template>
        <template v-else>
          <a-button @click="router.push('/login')">登录</a-button>
          <a-button type="primary" @click="router.push('/register')">注册</a-button>
        </template>
      </div>
    </a-layout-header>
    <a-layout-content class="content">
      <router-view />
    </a-layout-content>
    <a-layout-footer class="footer">
      灵视 VisionLive © 2026
    </a-layout-footer>
  </a-layout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()
const searchKeyword = ref('')

function handleSearch(keyword: string) {
  router.push({ path: '/', query: { keyword } })
}

function handleLogout() {
  authStore.logout()
  router.push('/')
}
</script>

<style scoped>
.layout { min-height: 100vh; }
.header {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 0 24px;
  background: #fff;
  box-shadow: 0 1px 4px rgba(0,0,0,.1);
}
.logo {
  font-size: 20px;
  font-weight: bold;
  color: #165dff;
  cursor: pointer;
  white-space: nowrap;
}
.search-bar { flex: 1; max-width: 480px; }
.nav-actions { display: flex; align-items: center; gap: 12px; margin-left: auto; }
.content { padding: 24px; background: #f5f6fa; }
.footer { text-align: center; color: #86909c; }
</style>
