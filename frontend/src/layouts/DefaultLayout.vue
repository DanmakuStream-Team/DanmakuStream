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
            v-show="item.key !== 'admin' || authStore.isAdmin"
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
          <el-button class="round-action" circle title="投稿" @click="goUpload">
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

    <div class="layout-body">
      <aside class="sidebar">
        <section class="side-section primary-section">
          <button class="side-item" :class="{ active: isActive('home') }" type="button" @click="router.push('/')">
            <el-icon><HomeFilled /></el-icon>
            <span>首页</span>
          </button>
          <button class="side-item" :class="{ active: isActive('live') }" type="button" @click="router.push('/live/1')">
            <el-icon><VideoCameraFilled /></el-icon>
            <span>直播</span>
          </button>
          <button class="side-item" :class="{ active: isActive('video') }" type="button" @click="router.push({ path: '/', query: { feature: 'video' } })">
            <el-icon><VideoCamera /></el-icon>
            <span>视频</span>
          </button>
          <button class="side-item" :class="{ active: isActive('danmaku') }" type="button" @click="router.push({ path: '/', query: { feature: 'danmaku' } })">
            <el-icon><Bell /></el-icon>
            <span>弹幕</span>
          </button>
          <button class="side-item" :class="{ active: isActive('creator') }" type="button" @click="goUpload">
            <el-icon><Upload /></el-icon>
            <span>投稿</span>
          </button>
        </section>

        <section v-if="!authStore.isLoggedIn" class="side-section login-section">
          <p>登录即可给视频点赞、发表评论并订阅内容。</p>
          <el-button class="side-login" type="primary" plain @click="router.push('/login')">
            <el-icon><UserFilled /></el-icon>
            登录
          </el-button>
        </section>

        <section v-if="authStore.isLoggedIn" class="side-section">
          <button class="side-title action-title" type="button" @click="router.push(`/user/${authStore.userInfo?.id}`)">
            <span>我</span>
            <el-icon><ArrowRight /></el-icon>
          </button>
          <button class="side-item" type="button" @click="router.push(`/user/${authStore.userInfo?.id}`)">
            <el-icon><UserFilled /></el-icon>
            <span>个人主页</span>
          </button>
          <button class="side-item" type="button" @click="router.push('/creator')">
            <el-icon><Notebook /></el-icon>
            <span>创作者中心</span>
          </button>
          <button class="side-item" type="button" @click="goUpload">
            <el-icon><Upload /></el-icon>
            <span>发布视频</span>
          </button>
          <button class="side-item" :class="{ active: isActive('history') }" type="button" @click="router.push('/me/history')">
            <el-icon><Clock /></el-icon>
            <span>历史记录</span>
          </button>
          <button class="side-item" :class="{ active: isActive('liked') }" type="button" @click="router.push('/me/liked')">
            <el-icon><Star /></el-icon>
            <span>赞过的视频</span>
          </button>
          <button class="side-item" :class="{ active: isActive('downloads') }" type="button" @click="router.push('/me/downloads')">
            <el-icon><Download /></el-icon>
            <span>下载内容</span>
          </button>
        </section>

        <section v-if="authStore.isLoggedIn" class="side-section">
          <button class="side-title action-title" type="button" @click="router.push({ path: '/', query: { feature: 'video' } })">
            <span>订阅</span>
            <el-icon><ArrowRight /></el-icon>
          </button>
          <button
            v-for="creator in visibleSubscriptions"
            :key="creator.id"
            class="subscribe-item"
            type="button"
            @click="router.push(`/user/${creator.id}`)"
          >
            <el-avatar :size="34" :src="creator.avatar">{{ creator.nickname.slice(0, 1) }}</el-avatar>
            <span>{{ creator.nickname }}</span>
          </button>
          <button
            v-if="subscriptions.length > displayCount"
            class="side-item more-item"
            type="button"
            @click="displayCount = Math.min(displayCount + 20, subscriptions.length)"
          >
            <el-icon><ArrowDown /></el-icon>
            <span>展开</span>
          </button>
        </section>

        <section class="side-section">
          <h2 class="side-title">探索</h2>
          <button class="side-item" type="button">
            <el-icon><Headset /></el-icon>
            <span>音乐</span>
          </button>
          <button class="side-item" type="button">
            <el-icon><Collection /></el-icon>
            <span>影视</span>
          </button>
          <button class="side-item more-item" type="button">
            <el-icon><ArrowDown /></el-icon>
            <span>展开</span>
          </button>
        </section>
      </aside>

      <el-main class="main">
        <router-view />
      </el-main>
    </div>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  ArrowDown,
  ArrowRight,
  Bell,
  Clock,
  Collection,
  Download,
  Headset,
  HomeFilled,
  Notebook,
  Search,
  Star,
  Upload,
  UserFilled,
  VideoCamera,
  VideoCameraFilled,
} from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { userApi } from '@/api/user'
import type { FolloweeInfo } from '@/api/user'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const keyword = ref('')
const subscriptions = ref<FolloweeInfo[]>([])
const displayCount = ref(5)
const visibleSubscriptions = computed(() => subscriptions.value.slice(0, displayCount.value))

async function loadFollowing() {
  if (!authStore.isLoggedIn) return
  try {
    const res = await userApi.following()
    subscriptions.value = res.data.list
  } catch {
    subscriptions.value = []
  }
}

onMounted(loadFollowing)
watch(() => authStore.isLoggedIn, (loggedIn) => {
  if (loggedIn) loadFollowing()
  else subscriptions.value = []
})

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
  if (key === 'history') {
    return route.path === '/me/history'
  }
  if (key === 'liked') {
    return route.path === '/me/liked'
  }
  if (key === 'downloads') {
    return route.path === '/me/downloads'
  }
  return false
}

function search() {
  router.push({ path: '/', query: keyword.value ? { keyword: keyword.value } : {} })
}

function goUpload() {
  if (authStore.isLoggedIn) {
    router.push('/creator/upload')
    return
  }
  ElMessage.warning('请先登录后再投稿')
  router.push({ path: '/login', query: { redirect: '/creator/upload' } })
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
  border-bottom: 1px solid rgba(15, 23, 42, 0.08);
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(16px);
}

.topbar-inner {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 24px;
  width: 100%;
  height: 100%;
  padding: 0 var(--topbar-x);
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
  font-size: 32px;
  font-weight: 900;
  letter-spacing: 0;
}

.brand-name {
  font-size: 23px;
  font-weight: 900;
}

.nav {
  display: flex;
  align-items: center;
  gap: 4px;
}

.nav button {
  height: 40px;
  padding: 0 13px;
  border-radius: 8px;
  color: #61666d;
  font-size: 16px;
  font-weight: 800;
}

.nav button:hover,
.nav button.active {
  background: #f1f2f3;
  color: #00aeec;
}

.actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  width: 100%;
}

.search {
  width: 62%;
}

.round-action {
  color: #61666d;
}

.login-btn {
  min-width: 76px;
  background: #00aeec;
  border-color: #00aeec;
  font-size: 16px;
  font-weight: 900;
}

.user-button {
  display: grid;
  padding: 0;
  place-items: center;
}

.layout-body {
  display: grid;
  grid-template-columns: 13% minmax(0, 1fr);
  gap: 1.5%;
  width: 100%;
}

.sidebar {
  position: sticky;
  top: 68px;
  align-self: start;
  height: calc(100vh - 68px);
  padding: 18px 4% 28px 5%;
  overflow-y: auto;
  border-right: 1px solid #f1f2f3;
  background: #fff;
}

.side-section {
  display: grid;
  gap: 8px;
  padding: 16px 0;
  border-bottom: 1px solid #e7e7e7;
}

.primary-section {
  padding-top: 0;
}

.side-section:last-child {
  border-bottom: 0;
}

.side-item,
.subscribe-item,
.action-title {
  display: grid;
  grid-template-columns: 38px minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
  width: 100%;
  min-height: 48px;
  padding: 0 16px;
  border: 0;
  border-radius: 12px;
  background: transparent;
  color: #18191c;
  text-align: left;
  cursor: pointer;
}

.side-item:hover,
.subscribe-item:hover,
.action-title:hover,
.side-item.active {
  background: #f1f2f3;
}

.side-item.active {
  font-weight: 900;
}

.side-item .el-icon {
  justify-self: center;
  color: #111;
  font-size: 27px;
}

.side-item span,
.subscribe-item span,
.side-title {
  overflow: hidden;
  font-size: 20px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.side-title {
  margin: 0;
  padding: 0 16px 6px;
}

.action-title {
  grid-template-columns: minmax(0, 1fr) auto;
  min-height: 42px;
  padding: 0 16px;
}

.action-title span {
  font-size: 21px;
  font-weight: 900;
}

.action-title .el-icon {
  color: #111;
  font-size: 17px;
}

.login-section {
  gap: 14px;
  padding: 22px 16px 24px;
}

.login-section p {
  margin: 0;
  color: #333;
  font-size: 18px;
  font-weight: 700;
  line-height: 1.7;
}

.side-login {
  justify-self: start;
  height: 44px;
  padding: 0 18px;
  border-radius: 999px;
  font-size: 18px;
  font-weight: 900;
}

.subscribe-item {
  grid-template-columns: 38px minmax(0, 1fr);
  min-height: 50px;
}

.subscribe-item span {
  font-size: 18px;
  font-weight: 700;
}

.unread-dot,
.live-dot {
  justify-self: center;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #2f66d4;
}

.live-dot {
  width: 11px;
  height: 11px;
  background: #ff2f55;
  box-shadow: 0 0 0 4px rgba(255, 47, 85, 0.12);
}

.more-item {
  color: #333;
}

.main {
  min-width: 0;
  padding: 0 0 48px;
}

@media (max-width: 920px) {
  .topbar {
    height: auto;
  }

  .topbar-inner {
    grid-template-columns: 1fr auto;
    gap: 12px;
    padding: 10px var(--topbar-x);
  }

  .nav {
    grid-column: 1 / -1;
    order: 3;
    overflow-x: auto;
  }

  .search {
    display: none;
  }

  .layout-body {
    display: block;
  }

  .sidebar {
    display: none;
  }
}
</style>
