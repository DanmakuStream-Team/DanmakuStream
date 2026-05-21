import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/pages/home/HomePage.vue'),
        meta: { title: '首页' },
      },
      {
        path: 'video/:id',
        name: 'VideoDetail',
        component: () => import('@/pages/video/VideoDetailPage.vue'),
        meta: { title: '视频播放' },
      },
      {
        path: 'live/:id',
        name: 'LiveRoom',
        component: () => import('@/pages/live/LiveRoomPage.vue'),
        meta: { title: '直播间' },
      },
      {
        path: 'user/:id',
        name: 'UserProfile',
        component: () => import('@/pages/user/UserProfilePage.vue'),
        meta: { title: '用户主页' },
      },
      {
        path: 'creator',
        component: () => import('@/layouts/CreatorLayout.vue'),
        meta: { requiresAuth: true },
        children: [
          {
            path: '',
            name: 'CreatorDashboard',
            component: () => import('@/pages/user/CreatorDashboardPage.vue'),
            meta: { title: '创作者中心' },
          },
          {
            path: 'upload',
            name: 'VideoUpload',
            component: () => import('@/pages/video/VideoUploadPage.vue'),
            meta: { title: '上传视频' },
          },
        ],
      },
      {
        path: 'admin',
        component: () => import('@/layouts/AdminLayout.vue'),
        meta: { requiresAuth: true, role: 'admin' },
        children: [
          {
            path: '',
            name: 'AdminDashboard',
            component: () => import('@/pages/admin/AdminDashboardPage.vue'),
            meta: { title: '管理后台' },
          },
          {
            path: 'videos',
            name: 'AdminVideos',
            component: () => import('@/pages/admin/AdminVideosPage.vue'),
            meta: { title: '视频审核' },
          },
          {
            path: 'danmaku',
            name: 'AdminDanmaku',
            component: () => import('@/pages/admin/AdminDanmakuPage.vue'),
            meta: { title: '弹幕管理' },
          },
        ],
      },
    ],
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/pages/home/LoginPage.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/pages/home/RegisterPage.vue'),
    meta: { title: '注册' },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/pages/home/NotFoundPage.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation guard
router.beforeEach((to) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    return { name: 'Login', query: { redirect: to.fullPath } }
  }
})

export default router
