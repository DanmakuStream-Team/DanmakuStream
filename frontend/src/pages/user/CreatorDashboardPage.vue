<template>
  <div class="creator-page max-w-7xl mx-auto px-4 md:px-6 py-8">
    <!-- 顶部标题 -->
    <div class="creator-header rounded-2xl py-10 px-8 mb-8">
      <h2 class="text-2xl md:text-3xl font-bold text-white">创作者中心</h2>
      <p class="text-white/80 mt-2">管理你的视频、数据与粉丝</p>
    </div>

    <!-- 数据统计卡片 -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-6 mb-10">
      <div class="stat-card rounded-xl p-6 shadow-sm">
        <p class="text-gray-500 text-sm">总视频数</p>
        <p class="text-3xl font-bold mt-2">0</p>
      </div>
      <div class="stat-card rounded-xl p-6 shadow-sm">
        <p class="text-gray-500 text-sm">总播放量</p>
        <p class="text-3xl font-bold mt-2">0</p>
      </div>
      <div class="stat-card rounded-xl p-6 shadow-sm">
        <p class="text-gray-500 text-sm">总获赞数</p>
        <p class="text-3xl font-bold mt-2">0</p>
      </div>
    </div>

    <!-- 功能入口 -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
      <div class="func-card rounded-xl p-6 text-center cursor-pointer hover:scale-105 transition-transform">
        <div class="text-xl font-semibold">视频管理</div>
        <p class="text-gray-400 mt-2 text-sm">发布、编辑、删除视频</p>
      </div>
      <div class="func-card rounded-xl p-6 text-center cursor-pointer hover:scale-105 transition-transform">
        <div class="text-xl font-semibold">数据中心</div>
        <p class="text-gray-400 mt-2 text-sm">播放、点赞、粉丝数据</p>
      </div>
      <div class="func-card rounded-xl p-6 text-center cursor-pointer hover:scale-105 transition-transform">
        <div class="text-xl font-semibold">评论管理</div>
        <p class="text-gray-400 mt-2 text-sm">查看与回复评论</p>
      </div>
      <div class="func-card rounded-xl p-6 text-center cursor-pointer hover:scale-105 transition-transform">
        <div class="text-xl font-semibold">个人设置</div>
        <p class="text-gray-400 mt-2 text-sm">修改头像、简介</p>
      </div>
    </div>

    <!-- 开发中提示（保留原来的功能提示） -->
    <div class="mt-16">
      <el-result
        title="温馨提示"
        sub-title="完整功能正在开发中..."
      />
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
/* 顶部渐变标题 */
.creator-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.25);
}

/* 统计卡片 */
.stat-card {
  background: #fff;
  border: 1px solid #f3f4f6;
}

/* 功能卡片 */
.func-card {
  background: #fff;
  border: 1px solid #f3f4f6;
}
.func-card:hover {
  border-color: #667eea;
}
</style>