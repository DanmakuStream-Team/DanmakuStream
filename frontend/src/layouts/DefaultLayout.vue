<template>
  <el-container class="app-layout">
    <el-header class="topbar">
      <div class="topbar-inner">
        <button class="brand" type="button" @click="router.push('/')">
          <span class="brand-script">Danmaku</span>
          <span class="brand-name">Stream</span>
        </button>

        <nav class="nav">
          <button
            v-for="item in navItems"
            :key="item.key"
            type="button"
            :class="{ active: isActive(item.key) }"
            @click="goNav(item)"
          >
            {{ item.label }}
          </button>
        </nav>

        <div class="actions">
          <el-input
            v-model="keyword"
            class="search"
            placeholder="搜索视频、弹幕、创作者"
            clearable
            @keyup.enter="search"
          >
            <template #suffix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-button class="round-action" circle title="投稿" @click="router.push('/creator/upload')">
            <el-icon><Upload /></el-icon>
          </el-button>
          <el-dropdown v-if="authStore.isLoggedIn" trigger="click">
            <button class="user-button" type="button">
              <el-avatar :size="34" :src="authStore.userInfo?.avatar">
                {{ authStore.userInfo?.nickname?.slice(0, 1) || '我' }}
              </el-avatar>
            </button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push(`/user/${authStore.userInfo?.id}`)">个人主页</el-dropdown-item>
                <el-dropdown-item @click="router.push('/creator')">创作者中心</el-dropdown-item>
                <el-dropdown-item v-if="authStore.isAdmin" @click="router.push('/admin')">管理后台</el-dropdown-item>
                <el-dropdown-item divided @click="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button v-else class="login-btn" type="primary" @click="router.push('/login')">登录</el-button>
        </div>
      </div>
    </el-header>

    <el-main class="main">
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Search, Upload } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const keyword = ref('')

const navItems = [
  { key: 'home', label: '首页', path: '/' },
  { key: 'video', label: '视频', path: '/', query: { feature: 'video' } },
  { key: 'danmaku', label: '弹幕', path: '/', query: { feature: 'danmaku' } },
  { key: 'live', label: '直播', path: '/live/1' },
  { key: 'creator', label: '投稿', path: '/creator/upload' },
  { key: 'admin', label: '审核', path: '/admin' },
]

function goNav(item: (typeof navItems)[number]) {
  router.push({ path: item.path, query: item.query || {} })
}

function isActive(key: string) {
  if (key === 'home') {
    return route.path === '/' && !route.query.feature
  }
  if (key === 'video') {
    return route.path === '/' && route.query.feature === 'video'
  }
  if (key === 'danmaku') {
    return route.path === '/' && route.query.feature === 'danmaku'
  }
  if (key === 'live') {
    return route.path.startsWith('/live')
  }
  if (key === 'creator') {
    return route.path.startsWith('/creator')
  }
  if (key === 'admin') {
    return route.path.startsWith('/admin')
  }
  return false
}

function search() {
  router.push({ path: '/', query: keyword.value ? { keyword: keyword.value } : {} })
}

function logout() {
  authStore.logout()
  router.push('/')
}
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  background: transparent;
}

.topbar {
  position: sticky;
  top: 0;
  z-index: 30;
  height: 64px;
  padding: 0;
  border-bottom: 1px solid rgba(15, 23, 42, 0.08);
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(16px);
}

.topbar-inner {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 24px;
  width: min(1320px, calc(100% - 48px));
  height: 100%;
  margin: 0 auto;
}

.brand,
.nav button,
.user-button {
  border: 0;
  background: transparent;
  cursor: pointer;
}

.brand {
  display: inline-flex;
  align-items: baseline;
  gap: 6px;
  color: #18191c;
}

.brand-script {
  color: #fb7299;
  font-size: 27px;
  font-weight: 900;
  letter-spacing: 0;
}

.brand-name {
  font-size: 19px;
  font-weight: 800;
}

.nav {
  display: flex;
  align-items: center;
  gap: 4px;
}

.nav button {
  height: 36px;
  padding: 0 13px;
  border-radius: 8px;
  color: #61666d;
  font-size: 14px;
}

.nav button:hover,
.nav button.active {
  background: #f1f2f3;
  color: #00aeec;
}

.actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search {
  width: min(360px, 28vw);
}

.round-action {
  color: #61666d;
}

.login-btn {
  min-width: 76px;
  background: #00aeec;
  border-color: #00aeec;
  font-weight: 700;
}

.user-button {
  display: grid;
  padding: 0;
  place-items: center;
}

.main {
  padding: 0 0 48px;
}

@media (max-width: 920px) {
  .topbar {
    height: auto;
  }

  .topbar-inner {
    grid-template-columns: 1fr auto;
    gap: 12px;
    padding: 10px 0;
  }

  .nav {
    grid-column: 1 / -1;
    order: 3;
    overflow-x: auto;
  }

  .search {
    display: none;
  }
}
</style>
