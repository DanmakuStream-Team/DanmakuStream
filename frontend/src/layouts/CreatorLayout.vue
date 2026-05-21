<template>
  <div class="flex min-h-[calc(100vh-120px)]">
    <!-- Sidebar -->
    <aside class="w-48 bg-white border-r border-gray-200 flex-shrink-0">
      <div class="px-4 py-4 font-semibold text-blue-600 border-b border-gray-100">创作者中心</div>
      <el-menu
        :default-active="route.name as string"
        @select="(key: string) => router.push({ name: key })"
      >
        <el-menu-item index="CreatorDashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>概览</span>
        </el-menu-item>
        <el-menu-item index="VideoUpload">
          <el-icon><Upload /></el-icon>
          <span>上传视频</span>
        </el-menu-item>
      </el-menu>
    </aside>

    <!-- Content -->
    <main class="flex-1 bg-gray-50 p-6">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
const router = useRouter()
const route = useRoute()
</script>
