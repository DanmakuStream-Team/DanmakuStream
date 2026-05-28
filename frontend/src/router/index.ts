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
      { path: 'live/:id', name: 'LiveRoom', component: () => import('@/pages/live/LiveRoomPage.vue') },
      { path: 'user/:id', name: 'UserProfile', component: () => import('@/pages/user/UserProfilePage.vue') },
      { path: 'me/:kind(history|liked|downloads)', name: 'UserLibrary', component: () => import('@/pages/user/UserLibraryPage.vue'), meta: { requiresAuth: true } },
      { path: 'creator', name: 'CreatorDashboard', component: () => import('@/pages/user/CreatorDashboardPage.vue'), meta: { requiresAuth: true } },
      { path: 'creator/upload', name: 'VideoUpload', component: () => import('@/pages/video/VideoUploadPage.vue'), meta: { requiresAuth: true } },
      { path: 'admin', name: 'AdminDashboard', component: () => import('@/pages/admin/AdminDashboardPage.vue') },
      { path: 'admin/videos', name: 'AdminVideos', component: () => import('@/pages/admin/AdminVideosPage.vue') },
      { path: 'admin/danmaku', name: 'AdminDanmaku', component: () => import('@/pages/admin/AdminDanmakuPage.vue') },
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
  if (!to.matched.some((record) => record.meta.requiresAuth)) return true
  const authStore = useAuthStore()
  if (authStore.isLoggedIn) return true
  return {
    path: '/login',
    query: { redirect: to.fullPath },
  }
})

export default router
