<template>
  <div class="profile-page max-w-7xl mx-auto px-4 md:px-6 py-8">
    <!-- 个人信息头部卡片 -->
    <div class="profile-header rounded-2xl py-10 px-8 mb-8 flex flex-col md:flex-row items-center gap-6">
      <!-- 头像区域 -->
      <el-avatar :size="120" class="shadow-md">
        <img src="https://picsum.photos/200/200" alt="用户头像" />
      </el-avatar>
      <!-- 昵称简介 -->
      <div class="user-info text-center md:text-left">
        <h2 class="text-2xl font-bold text-white">匿名用户</h2>
        <p class="text-white/80 mt-2">热爱视频 · 喜欢弹幕互动</p>
        <div class="flex gap-4 mt-4 justify-center md:justify-start">
          <div class="text-white text-center">
            <div class="text-xl font-semibold">0</div>
            <div class="text-sm opacity-80">作品</div>
          </div>
          <div class="text-white text-center">
            <div class="text-xl font-semibold">0</div>
            <div class="text-sm opacity-80">粉丝</div>
          </div>
          <div class="text-white text-center">
            <div class="text-xl font-semibold">0</div>
            <div class="text-sm opacity-80">关注</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 个人功能选项卡 -->
    <div class="tab-wrap bg-white rounded-xl shadow-sm p-6 mb-8">
      <el-tabs v-model="activeTab">
        <el-tab-pane label="我的作品" name="video"></el-tab-pane>
        <el-tab-pane label="我的收藏" name="collect"></el-tab-pane>
        <el-tab-pane label="浏览记录" name="history"></el-tab-pane>
        <el-tab-pane label="个人设置" name="setting"></el-tab-pane>
      </el-tabs>
    </div>

    <!-- 空状态提示 -->
    <div class="empty-content py-16">
      <el-result
        icon="info"
        title="用户主页功能开发中"
        sub-title="作品展示、收藏管理、个人资料编辑等功能即将上线"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
// 仅定义切换标签变量，无任何接口与业务逻辑
const activeTab = ref('video')
</script>

<style scoped>
/* 顶部渐变背景 统一项目色系 */
.profile-header {
  background: linear-gradient(135deg, #36d1dc 0%, #5b86e5 100%);
  box-shadow: 0 4px 16px rgba(54, 209, 220, 0.25);
}
.tab-wrap {
  border: 1px solid #f5f5f5;
}
.empty-content {
  background: #fafbfc;
  border-radius: 12px;
}
</style>