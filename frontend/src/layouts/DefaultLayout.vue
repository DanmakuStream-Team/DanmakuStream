<template>
  <div class="min-h-screen flex flex-col bg-gray-50">
    <!-- Header -->
    <header class="bg-white shadow-sm sticky top-0 z-50">
      <div class="max-w-7xl mx-auto px-4 h-16 flex items-center gap-6">
        <div
          class="text-xl font-bold text-blue-600 cursor-pointer whitespace-nowrap select-none"
          @click="router.push('/')"
        >
          Danmaku
        </div>

        <div class="flex-1 max-w-xl">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索视频"
            size="large"
            clearable
            @keydown.enter="handleSearch(searchKeyword)"
          >
            <template #suffix>
              <el-icon class="cursor-pointer" @click="handleSearch(searchKeyword)">
                <Search />
              </el-icon>
            </template>
          </el-input>
        </div>

        <div class="ml-auto flex items-center gap-3">
          <template v-if="authStore.isLoggedIn">
            <el-button @click="router.push('/creator/upload')">投稿</el-button>
            <el-dropdown>
              <el-avatar
                :src="authStore.userInfo?.avatar"
                class="cursor-pointer"
              >
                {{ authStore.userInfo?.nickname?.slice(0, 1) }}
              </el-avatar>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="router.push(`/user/${authStore.userInfo?.id}`)">个人主页</el-dropdown-item>
                  <el-dropdown-item v-if="authStore.isCreator" @click="router.push('/creator')">创作者中心</el-dropdown-item>
                  <el-dropdown-item v-if="authStore.isAdmin" @click="router.push('/admin')">管理后台</el-dropdown-item>
                  <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
          <template v-else>
            <el-button @click="router.push('/login')">登录</el-button>
            <el-button type="primary" @click="router.push('/register')">注册</el-button>
          </template>
        </div>
      </div>
    </header>

    <!-- Content -->
    <main class="flex-1 p-6">
      <router-view />
    </main>

    <!-- Footer -->
    <footer class="text-center text-gray-400 py-4 text-sm">
      Danmaku DanmakuStream © 2026
    </footer>
  </div>
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
