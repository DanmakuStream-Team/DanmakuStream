<template>
  <el-container class="app-layout">
    <el-header class="topbar">
      <div class="topbar-inner">
        <div class="topbar-left">
          <button class="brand" type="button" @click="router.push(authStore.isStaff ? '/admin' : '/')">
            <span class="brand-script">Danmaku</span>
            <span class="brand-name">Stream</span>
          </button>

          <nav v-if="visibleNavItems.length" class="nav">
            <button
              v-for="item in visibleNavItems"
              :key="item.key"
              type="button"
              :class="{ active: isActive(item.key) }"
              @click="goNav(item)"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>

        <div class="search-wrap">
          <el-input
            v-model="keyword"
            class="search"
            placeholder="搜索视频、创作者"
            clearable
            @focus="isSearchFocused = true"
            @blur="isSearchFocused = false"
            @keyup.enter="search"
          >
            <template #suffix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <div v-if="showSearchPanel" class="search-panel">
            <div
              v-for="item in visibleSearchHistory"
              :key="item"
              class="search-history-row"
            >
              <button
                class="search-history-item"
                type="button"
                @mousedown.prevent="useSearchHistory(item)"
              >
                <el-icon><Clock /></el-icon>
                <span>{{ item }}</span>
              </button>
              <button
                aria-label="删除搜索历史"
                class="search-history-remove"
                type="button"
                title="删除"
                @mousedown.stop.prevent="removeSearchHistory(item)"
              >
                ×
              </button>
            </div>
          </div>
        </div>

        <div class="actions">
          <el-dropdown v-if="authStore.isLoggedIn" trigger="click" @visible-change="handleNotificationVisible">
            <el-badge :value="unreadCount" :hidden="unreadCount === 0" :max="99">
              <el-button class="round-action" circle title="通知">
                <el-icon><Bell /></el-icon>
              </el-button>
            </el-badge>
            <template #dropdown>
              <div class="notification-panel">
                <div class="notification-head">
                  <strong>通知</strong>
                  <el-button size="small" text :disabled="unreadCount === 0" @click="markAllNotificationsRead">
                    全部已读
                  </el-button>
                </div>
                <button
                  v-for="item in notifications"
                  :key="item.id"
                  class="notification-item"
                  :class="{ unread: !item.read }"
                  type="button"
                  @click="openNotification(item)"
                >
                  <span>{{ item.title }}</span>
                  <em>{{ item.content }}</em>
                  <small>{{ item.createdAt }}</small>
                </button>
                <div v-if="!notifications.length" class="notification-empty">暂无通知</div>
              </div>
            </template>
          </el-dropdown>
          <el-dropdown trigger="click">
            <el-button class="round-action" circle title="创作">
              <el-icon><Upload /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="goUpload">上传视频</el-dropdown-item>
                <el-dropdown-item @click="goLiveStart">开始直播</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-dropdown v-if="authStore.isLoggedIn" trigger="click">
            <button class="user-button" type="button">
              <el-avatar :size="34" :src="authStore.userInfo?.avatar">
                {{ authStore.userInfo?.nickname?.slice(0, 1) || '我' }}
              </el-avatar>
            </button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push(`/user/${authStore.userInfo?.id}`)">个人主页</el-dropdown-item>
                <el-dropdown-item v-if="!authStore.isStaff" @click="router.push('/creator')">创作者中心</el-dropdown-item>
                <el-dropdown-item v-if="authStore.isStaff" @click="router.push('/admin')">审核后台</el-dropdown-item>
                <el-dropdown-item divided @click="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button v-else class="login-btn" type="primary" @click="router.push('/login')">登录</el-button>
        </div>
      </div>
    </el-header>

    <div class="layout-body" :class="{ collapsed: isSidebarCollapsed }">
      <aside class="sidebar" :class="{ collapsed: isSidebarCollapsed }">
        <button
          class="collapse-toggle"
          type="button"
          :title="isSidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
          @click="isSidebarCollapsed = !isSidebarCollapsed"
        >
          <span />
          <span />
          <span />
        </button>

        <section v-if="authStore.isStaff" class="side-section primary-section">
          <button
            v-for="item in visibleAdminNavItems"
            :key="item.key"
            class="side-item"
            :class="{ active: isActive(item.key) }"
            type="button"
            @click="router.push(item.path)"
          >
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.label }}</span>
          </button>
        </section>

        <section v-if="!authStore.isStaff" class="side-section primary-section">
          <button class="side-item" :class="{ active: isActive('home') }" type="button" @click="router.push('/')">
            <el-icon><HomeFilled /></el-icon>
            <span>首页</span>
          </button>
          <button class="side-item" :class="{ active: isActive('live') }" type="button" @click="router.push('/live')">
            <el-icon><VideoCameraFilled /></el-icon>
            <span>直播</span>
          </button>
          <button class="side-item" :class="{ active: isActive('video') }" type="button" @click="router.push({ path: '/', query: { feature: 'video' } })">
            <el-icon><VideoCamera /></el-icon>
            <span>视频</span>
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

        <section v-if="authStore.isLoggedIn && !authStore.isStaff" class="side-section">
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
          <button class="side-item" :class="{ active: isActive('tags') }" type="button" @click="router.push('/me/tags')">
            <el-icon><CollectionTag /></el-icon>
            <span>标签相关度</span>
          </button>
          <button class="side-item" :class="{ active: isActive('liked') }" type="button" @click="router.push('/me/liked')">
            <el-icon><ThumbUpIcon /></el-icon>
            <span>赞过的视频</span>
          </button>
          <button class="side-item" :class="{ active: isActive('collections') }" type="button" @click="router.push('/me/collections')">
            <el-icon><Star /></el-icon>
            <span>收藏内容</span>
          </button>
          <button class="side-item" :class="{ active: isActive('downloads') }" type="button" @click="router.push('/me/downloads')">
            <el-icon><Download /></el-icon>
            <span>下载内容</span>
          </button>
        </section>

        <section v-if="authStore.isLoggedIn && !authStore.isStaff" class="side-section">
          <button class="side-title action-title" :class="{ active: isActive('subscriptions') }" type="button" @click="router.push('/subscriptions')">
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
  CollectionTag,
  Download,
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
import ThumbUpIcon from '@/components/icons/ThumbUpIcon.vue'
import { notificationApi } from '@/api/notification'
import type { NotificationInfo } from '@/types'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const keyword = ref('')
const isSearchFocused = ref(false)
const isSidebarCollapsed = ref(false)
const searchHistory = ref<string[]>([])
const subscriptions = ref<FolloweeInfo[]>([])
const notifications = ref<NotificationInfo[]>([])
const unreadCount = ref(0)
const displayCount = ref(5)
const searchHistoryKey = 'danmaku:search-history'
const visibleSubscriptions = computed(() => subscriptions.value.slice(0, displayCount.value))
const visibleSearchHistory = computed(() => searchHistory.value.slice(0, 8))
const showSearchPanel = computed(() => isSearchFocused.value && visibleSearchHistory.value.length > 0)

async function loadFollowing() {
  if (!authStore.isLoggedIn) return
  try {
    const res = await userApi.following()
    subscriptions.value = res.data.list
  } catch {
    subscriptions.value = []
  }
}

async function loadNotifications() {
  if (!authStore.isLoggedIn) return
  try {
    const res = await notificationApi.list({ page: 1, pageSize: 8 })
    notifications.value = res.data.list
    unreadCount.value = res.data.unreadCount
  } catch {
    notifications.value = []
    unreadCount.value = 0
  }
}

function loadSearchHistory() {
  try {
    const parsed = JSON.parse(localStorage.getItem(searchHistoryKey) || '[]')
    searchHistory.value = Array.isArray(parsed)
      ? parsed.filter((item): item is string => typeof item === 'string' && Boolean(item.trim())).slice(0, 20)
      : []
  } catch {
    searchHistory.value = []
  }
}

function saveSearchHistory(value: string) {
  const normalized = value.trim()
  if (!normalized) return
  searchHistory.value = [
    normalized,
    ...searchHistory.value.filter(item => item !== normalized),
  ].slice(0, 20)
  localStorage.setItem(searchHistoryKey, JSON.stringify(searchHistory.value))
}

function removeSearchHistory(value: string) {
  searchHistory.value = searchHistory.value.filter(item => item !== value)
  localStorage.setItem(searchHistoryKey, JSON.stringify(searchHistory.value))
}

function useSearchHistory(value: string) {
  keyword.value = value
  search()
}

onMounted(() => {
  loadFollowing()
  loadNotifications()
  loadSearchHistory()
  keyword.value = String(route.query.keyword || '')
})

watch(() => authStore.isLoggedIn, (loggedIn) => {
  if (loggedIn) {
    loadFollowing()
    loadNotifications()
  } else {
    subscriptions.value = []
    notifications.value = []
    unreadCount.value = 0
  }
})

watch(() => route.query.keyword, (value) => {
  keyword.value = String(value || '')
})

const navItems = [
  { key: 'home', label: '首页', path: '/' },
  { key: 'video', label: '视频', path: '/', query: { feature: 'video' } },
  { key: 'live', label: '直播', path: '/live' },
  { key: 'creator', label: '投稿', path: '/creator/upload' },
  { key: 'admin', label: '审核', path: '/admin' },
]

const visibleNavItems = computed(() => {
  if (authStore.isStaff) return []
  return navItems.filter((item) => item.key !== 'admin')
})

const adminNavItems = [
  { key: 'admin-dashboard', label: '后台首页', path: '/admin', icon: HomeFilled, adminOnly: false },
  { key: 'admin-videos', label: '视频审核', path: '/admin/videos', icon: VideoCamera, adminOnly: false },
  { key: 'admin-danmaku', label: '弹幕管理', path: '/admin/danmaku', icon: CollectionTag, adminOnly: false },
  { key: 'admin-users', label: '用户管理', path: '/admin/users', icon: UserFilled, adminOnly: true },
  { key: 'admin-operations', label: '运营管理', path: '/admin/operations', icon: Notebook, adminOnly: true },
  { key: 'admin-infrastructure', label: '系统状态', path: '/admin/infrastructure', icon: Notebook, adminOnly: true },
]

const visibleAdminNavItems = computed(() => {
  return adminNavItems.filter((item) => !item.adminOnly || authStore.isAdmin)
})

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
  if (key === 'live') {
    return route.path.startsWith('/live')
  }
  if (key === 'creator') {
    return route.path.startsWith('/creator')
  }
  if (key === 'admin') {
    return route.path.startsWith('/admin')
  }
  if (key === 'admin-dashboard') {
    return route.path === '/admin'
  }
  if (key === 'admin-videos') {
    return route.path.startsWith('/admin/videos')
  }
  if (key === 'admin-danmaku') {
    return route.path.startsWith('/admin/danmaku')
  }
  if (key === 'admin-users') {
    return route.path.startsWith('/admin/users')
  }
  if (key === 'admin-operations') {
    return route.path.startsWith('/admin/operations')
  }
  if (key === 'admin-infrastructure') {
    return route.path.startsWith('/admin/infrastructure')
  }
  if (key === 'history') {
    return route.path === '/me/history'
  }
  if (key === 'tags') {
    return route.path === '/me/tags'
  }
  if (key === 'liked') {
    return route.path === '/me/liked'
  }
  if (key === 'collections') {
    return route.path === '/me/collections'
  }
  if (key === 'downloads') {
    return route.path === '/me/downloads'
  }
  if (key === 'subscriptions') {
    return route.path === '/subscriptions'
  }
  return false
}

function search() {
  const value = keyword.value.trim()
  saveSearchHistory(value)
  isSearchFocused.value = false
  router.push({ path: '/', query: value ? { keyword: value } : {} })
}

function goUpload() {
  if (authStore.isLoggedIn) {
    router.push('/creator/upload')
    return
  }
  ElMessage.warning('请先登录后再投稿')
  router.push({ path: '/login', query: { redirect: '/creator/upload' } })
}

function goLiveStart() {
  if (authStore.isLoggedIn) {
    router.push({ path: '/live', query: { create: '1' } })
    return
  }
  ElMessage.warning('请先登录后再开播')
  router.push({ path: '/login', query: { redirect: '/live?create=1' } })
}

function handleNotificationVisible(visible: boolean) {
  if (visible) loadNotifications()
}

async function openNotification(item: NotificationInfo) {
  if (!item.read) {
    await notificationApi.read(item.id)
    item.read = true
    unreadCount.value = Math.max(0, unreadCount.value - 1)
  }
  if (item.link) {
    router.push(normalizeNotificationLink(item.link))
  }
}

async function markAllNotificationsRead() {
  await notificationApi.readAll()
  notifications.value = notifications.value.map((item) => ({ ...item, read: true }))
  unreadCount.value = 0
}

function normalizeNotificationLink(link: string) {
  if (link === '/live-schedules') return '/live'
  return link || '/'
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
  grid-template-columns: minmax(320px, 1fr) minmax(320px, 640px) minmax(92px, 1fr);
  align-items: center;
  gap: 24px;
  width: 100%;
  height: 100%;
  padding: 0 var(--topbar-x);
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 24px;
  min-width: 0;
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
  flex-shrink: 0;
  gap: 6px;
  white-space: nowrap;
  color: #18191c;
}

.brand-script {
  color: #fb7299;
  font-size: 30px;
  font-weight: 900;
  letter-spacing: 0;
}

.brand-name {
  font-size: 21px;
  font-weight: 900;
}

.nav {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
  overflow: hidden;
}

.nav button {
  flex-shrink: 0;
  height: 40px;
  padding: 0 13px;
  border-radius: 8px;
  color: #61666d;
  font-size: 15px;
  font-weight: 800;
  white-space: nowrap;
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
  flex-wrap: nowrap;
  gap: 10px;
  min-width: 92px;
  width: 100%;
}

.search-wrap {
  position: relative;
  width: min(100%, 640px);
  justify-self: center;
}

.search {
  width: 100%;
}

.search :deep(.el-input__wrapper) {
  min-height: 42px;
  border-radius: 999px;
  padding: 0 16px;
}

.search-panel {
  position: absolute;
  top: calc(100% + 8px);
  left: 0;
  z-index: 40;
  width: 100%;
  max-height: 294px;
  overflow: hidden;
  padding: 8px 0;
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 14px 38px rgba(15, 23, 42, 0.14);
}

.search-panel::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 34px;
  pointer-events: none;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0), #fff);
  content: '';
}

.search-history-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 34px;
  align-items: center;
  width: 100%;
  min-height: 40px;
  padding: 0 8px 0 14px;
}

.search-history-row:hover {
  background: #f6f7f8;
}

.search-history-item {
  display: grid;
  grid-template-columns: 24px minmax(0, 1fr);
  align-items: center;
  gap: 10px;
  min-width: 0;
  min-height: 40px;
  padding: 0;
  border: 0;
  background: transparent;
  color: #18191c;
  text-align: left;
  cursor: pointer;
}

.search-history-item .el-icon {
  color: #9499a0;
  font-size: 16px;
}

.search-history-item span {
  overflow: hidden;
  font-size: 14px;
  font-weight: 700;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-history-remove {
  display: grid;
  width: 26px;
  height: 26px;
  justify-self: center;
  place-items: center;
  border: 0;
  border-radius: 50%;
  background: transparent;
  color: #9499a0;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
}

.search-history-remove:hover {
  background: #e7e7e7;
  color: #18191c;
}

.round-action {
  color: #61666d;
}

.notification-panel {
  width: 320px;
  max-height: 430px;
  overflow-y: auto;
  padding: 10px;
}

.notification-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 2px 4px 8px;
}

.notification-head strong {
  color: #18191c;
  font-size: 16px;
  font-weight: 900;
}

.notification-item {
  display: grid;
  gap: 4px;
  width: 100%;
  padding: 10px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: #18191c;
  text-align: left;
  cursor: pointer;
}

.notification-item:hover {
  background: #f6f7f8;
}

.notification-item.unread {
  background: rgba(0, 174, 236, 0.08);
}

.notification-item span,
.notification-item em,
.notification-item small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notification-item span {
  font-size: 14px;
  font-weight: 900;
}

.notification-item em {
  color: #61666d;
  font-size: 13px;
  font-style: normal;
}

.notification-item small {
  color: #9499a0;
  font-size: 12px;
}

.notification-empty {
  display: grid;
  min-height: 96px;
  place-items: center;
  color: #9499a0;
  font-size: 14px;
  font-weight: 800;
}

.login-btn {
  min-width: 76px;
  background: #00aeec;
  border-color: #00aeec;
  font-size: 15px;
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
  transition: grid-template-columns 0.18s ease;
}

.layout-body.collapsed {
  grid-template-columns: 76px minmax(0, 1fr);
  gap: 0;
}

.sidebar {
  position: sticky;
  top: 68px;
  align-self: start;
  height: calc(100vh - 68px);
  padding: 22px 4% 28px 5%;
  overflow-y: auto;
  border-right: 1px solid #f1f2f3;
  background: #fff;
  scrollbar-width: none;
  transition: padding 0.18s ease;
}

.sidebar::-webkit-scrollbar {
  width: 0;
  height: 0;
}

.sidebar.collapsed {
  padding: 22px 8px 24px;
  overflow-x: hidden;
}

.collapse-toggle {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  right: -18px;
  z-index: 2;
  display: grid;
  align-content: center;
  gap: 4px;
  width: 36px;
  height: 36px;
  padding: 0 9px;
  border: 1px solid rgba(251, 114, 153, 0.22);
  border-radius: 50%;
  background: #fff;
  color: #fb7299;
  cursor: pointer;
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.1);
  transition:
    background 0.16s ease,
    border-color 0.16s ease,
    box-shadow 0.16s ease;
}

.collapse-toggle:hover {
  border-color: rgba(251, 114, 153, 0.42);
  background: rgba(251, 114, 153, 0.08);
  box-shadow: 0 6px 18px rgba(251, 114, 153, 0.16);
}

.collapse-toggle span {
  display: block;
  width: 100%;
  height: 2px;
  border-radius: 999px;
  background: currentColor;
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

.sidebar.collapsed .side-item,
.sidebar.collapsed .subscribe-item,
.sidebar.collapsed .action-title {
  grid-template-columns: 1fr;
  justify-items: center;
  gap: 0;
  min-height: 46px;
  padding: 0;
}

.sidebar.collapsed .side-item span,
.sidebar.collapsed .subscribe-item span,
.sidebar.collapsed .side-title,
.sidebar.collapsed .action-title span,
.sidebar.collapsed .action-title .el-icon,
.sidebar.collapsed .login-section p {
  display: none;
}

.sidebar.collapsed .login-section {
  padding: 16px 0;
}

.sidebar.collapsed .side-login {
  justify-self: center;
  width: 46px;
  min-width: 0;
  padding: 0;
}

.sidebar.collapsed .side-login span {
  display: none;
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
  font-size: 25px;
}

.side-item span,
.subscribe-item span,
.side-title {
  overflow: hidden;
  font-size: 18px;
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
  font-size: 19px;
  font-weight: 900;
}

.action-title .el-icon {
  color: #111;
  font-size: 16px;
}

.login-section {
  gap: 14px;
  padding: 22px 16px 24px;
}

.login-section p {
  margin: 0;
  color: #333;
  font-size: 16px;
  font-weight: 700;
  line-height: 1.7;
}

.side-login {
  justify-self: start;
  height: 44px;
  padding: 0 18px;
  border-radius: 999px;
  font-size: 16px;
  font-weight: 900;
}

.subscribe-item {
  grid-template-columns: 38px minmax(0, 1fr);
  min-height: 50px;
}

.sidebar.collapsed .subscribe-item {
  grid-template-columns: 1fr;
}

.subscribe-item span {
  font-size: 16px;
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

  .topbar-left {
    gap: 16px;
    min-width: 0;
  }

  .search-wrap {
    grid-column: 1 / -1;
    order: 3;
    width: 100%;
  }

  .nav {
    grid-column: 1 / -1;
    order: 4;
    overflow-x: auto;
  }

  .layout-body {
    display: block;
  }

  .sidebar {
    display: none;
  }
}
</style>
