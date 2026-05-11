<template>
  <a-layout class="layout">
    <a-layout-sider class="sider" :width="200">
      <div class="sider-title">创作者中心</div>
      <a-menu :selected-keys="[route.name as string]" @menu-item-click="(key: string) => router.push({ name: key })">
        <a-menu-item key="CreatorDashboard"><icon-home /> 概览</a-menu-item>
        <a-menu-item key="VideoUpload"><icon-upload /> 上传视频</a-menu-item>
      </a-menu>
    </a-layout-sider>
    <a-layout-content class="content">
      <router-view />
    </a-layout-content>
  </a-layout>
</template>

<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
const router = useRouter()
const route = useRoute()
</script>

<style scoped>
.layout { min-height: calc(100vh - 120px); }
.sider { background: #fff; border-right: 1px solid #e5e6e7; }
.sider-title { padding: 16px; font-weight: 600; color: #165dff; }
.content { padding: 24px; background: #f5f6fa; }
</style>
