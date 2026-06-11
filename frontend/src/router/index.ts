import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      { path: '', name: 'Home', component: () => import('@/pages/home/HomePage.vue') },
      { path: 'video/:id', name: 'VideoDetail', component: () => import('@/pages/video/VideoDetailPage.vue') },
      { path: 'live', name: 'LiveList', component: () => import('@/pages/live/LiveListPage.vue') },
      { path: 'live/:id', name: 'LiveRoom', component: () => import('@/pages/live/LiveRoomPage.vue') },
      { path: 'user/:id', name: 'UserProfile', component: () => import('@/pages/user/UserProfilePage.vue') },
      { path: 'subscriptions', name: 'Subscriptions', component: () => import('@/pages/user/SubscriptionPage.vue'), meta: { requiresAuth: true } },
      { path: 'me/:kind(history|liked|collections|downloads)', name: 'UserLibrary', component: () => import('@/pages/user/UserLibraryPage.vue'), meta: { requiresAuth: true } },
      { path: 'me/tags', name: 'TagAffinity', component: () => import('@/pages/user/TagAffinityPage.vue'), meta: { requiresAuth: true } },
      { path: 'creator', name: 'CreatorDashboard', component: () => import('@/pages/user/CreatorDashboardPage.vue'), meta: { requiresAuth: true } },
      { path: 'creator/upload', name: 'VideoUpload', component: () => import('@/pages/video/VideoUploadPage.vue'), meta: { requiresAuth: true } },
      { path: 'admin', name: 'AdminDashboard', component: () => import('@/pages/admin/AdminDashboardPage.vue'), meta: { requiresStaff: true } },
      { path: 'admin/infrastructure', name: 'AdminInfrastructure', component: () => import('@/pages/admin/AdminInfrastructurePage.vue'), meta: { requiresAdmin: true } },
      { path: 'admin/users', name: 'AdminUsers', component: () => import('@/pages/admin/AdminUsersPage.vue'), meta: { requiresAdmin: true } },
      { path: 'admin/operations', name: 'AdminOperations', component: () => import('@/pages/admin/AdminOperationsPage.vue'), meta: { requiresAdmin: true } },
      { path: 'admin/videos', name: 'AdminVideos', component: () => import('@/pages/admin/AdminVideosPage.vue'), meta: { requiresStaff: true } },
      { path: 'admin/danmaku', name: 'AdminDanmaku', component: () => import('@/pages/admin/AdminDanmakuPage.vue'), meta: { requiresStaff: true } },
    ],
  },
  { path: '/login', name: 'Login', component: () => import('@/pages/home/LoginPage.vue') },
  { path: '/register', name: 'Register', component: () => import('@/pages/home/RegisterPage.vue') },
  { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('@/pages/home/NotFoundPage.vue') },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

router.beforeEach((to) => {
  const authStore = useAuthStore()
  if (to.matched.some((record) => record.meta.requiresAdmin)) {
    if (authStore.isAdmin) return true
    return authStore.isLoggedIn ? '/' : { path: '/login', query: { redirect: to.fullPath } }
  }
  if (to.matched.some((record) => record.meta.requiresStaff)) {
    if (authStore.isStaff) return true
    return authStore.isLoggedIn ? '/' : { path: '/login', query: { redirect: to.fullPath } }
  }
  if (!to.matched.some((record) => record.meta.requiresAuth)) return true
  if (authStore.isLoggedIn) return true
  return {
    path: '/login',
    query: { redirect: to.fullPath },
  }
})

export default router
