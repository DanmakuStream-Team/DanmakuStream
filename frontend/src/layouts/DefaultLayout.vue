<template>
  <el-container class="app-layout">
    <el-header class="topbar">
      <div class="topbar-inner">
        <button class="brand" type="button" @click="router.push('/')">
          <span class="brand-mark">D</span>
          <span>DanmakuStream</span>
        </button>

        <nav class="nav">
          <button
            v-for="item in navItems"
            :key="item.path"
            type="button"
            :class="{ active: route.path === item.path }"
            @click="router.push(item.path)"
          >
            {{ item.label }}
          </button>
        </nav>

        <div class="actions">
          <el-input
            v-model="keyword"
            class="search"
            placeholder="搜索视频"
            clearable
            @keyup.enter="search"
          >
            <template #suffix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-button circle title="上传" @click="router.push('/creator/upload')">
            <el-icon><Upload /></el-icon>
          </el-button>
          <el-dropdown v-if="authStore.isLoggedIn" trigger="click">
            <button class="user-button" type="button">
              <el-avatar :size="32" :src="authStore.userInfo?.avatar">
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
          <el-button v-else type="primary" @click="router.push('/login')">登录</el-button>
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
  { label: '发现', path: '/' },
  { label: '创作', path: '/creator' },
  { label: '管理', path: '/admin' },
]

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
  height: 68px;
  padding: 0;
  border-bottom: 1px solid rgba(20, 32, 51, 0.08);
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(16px);
}

.topbar-inner {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 24px;
  width: min(1180px, calc(100% - 40px));
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
  align-items: center;
  gap: 10px;
  color: #142033;
  font-size: 18px;
  font-weight: 800;
}

.brand-mark {
  display: grid;
  width: 34px;
  height: 34px;
  place-items: center;
  border-radius: 8px;
  background: #165dff;
  color: #fff;
}

.nav {
  display: flex;
  gap: 6px;
}

.nav button {
  height: 36px;
  padding: 0 14px;
  border-radius: 8px;
  color: #667085;
  font-size: 14px;
}

.nav button.active,
.nav button:hover {
  background: #eef4ff;
  color: #165dff;
}

.actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search {
  width: 240px;
}

.user-button {
  display: grid;
  padding: 0;
  place-items: center;
}

.main {
  padding: 28px 0 56px;
}

@media (max-width: 900px) {
  .topbar {
    height: auto;
  }

  .topbar-inner {
    grid-template-columns: 1fr auto;
    gap: 12px;
    padding: 12px 0;
  }

  .nav {
    grid-column: 1 / -1;
    order: 3;
  }

  .search {
    display: none;
  }
}
</style>
