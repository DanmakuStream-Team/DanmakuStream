<template>
  <div class="flex min-h-[calc(100vh-120px)]">
    <!-- Sidebar -->
    <aside class="w-48 bg-white border-r border-gray-200 flex-shrink-0">
      <div class="px-4 py-4 font-semibold text-blue-600 border-b border-gray-100">管理后台</div>
      <el-menu
        :default-active="route.name as string"
        @select="(key: string) => router.push({ name: key })"
      >
        <el-menu-item index="AdminDashboard">
          <el-icon><DataBoard /></el-icon>
          <span>概览</span>
        </el-menu-item>
        <el-menu-item index="AdminVideos">
          <el-icon><VideoCamera /></el-icon>
          <span>视频审核</span>
        </el-menu-item>
        <el-menu-item index="AdminDanmaku">
          <el-icon><ChatDotRound /></el-icon>
          <span>弹幕管理</span>
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
