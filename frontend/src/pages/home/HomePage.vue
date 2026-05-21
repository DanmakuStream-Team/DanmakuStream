<template>
  <div class="max-w-7xl mx-auto">
    <!-- Banner -->
    <div class="py-10 text-center">
      <h2 class="text-3xl font-bold text-gray-800">发现优质视频内容</h2>
      <p class="text-gray-500 mt-2">探索海量弹幕视频，尽情享受</p>
    </div>

    <!-- Tag filter -->
    <div class="flex flex-wrap gap-2 mb-6">
      <el-check-tag
        v-for="tag in popularTags"
        :key="tag"
        :checked="selectedTag === tag"
        @change="selectTag(tag)"
      >
        {{ tag }}
      </el-check-tag>
    </div>

    <!-- Video grid -->
    <div v-loading="videoStore.loading">
      <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-5">
        <VideoCard
          v-for="video in videoStore.videoList"
          :key="video.id"
          :video="video"
          @click="router.push(`/video/${video.id}`)"
        />
      </div>
      <el-empty v-if="!videoStore.loading && videoStore.videoList.length === 0" description="暂无视频" class="py-20" />
    </div>

    <!-- Pagination -->
    <div class="flex justify-center mt-10">
      <el-pagination
        v-model:current-page="page"
        :total="videoStore.total"
        :page-size="pageSize"
        layout="prev, pager, next"
        background
        @current-change="fetchVideos"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useVideoStore } from '@/store/video'
import VideoCard from '@/components/common/VideoCard.vue'

const router = useRouter()
const route = useRoute()
const videoStore = useVideoStore()

const page = ref(1)
const pageSize = 20
const selectedTag = ref('')
const popularTags = ['全部', '游戏', '科技', '生活', '美食', '音乐', '动漫', '知识']

function selectTag(tag: string) {
  selectedTag.value = selectedTag.value === tag ? '' : tag
  page.value = 1
  fetchVideos()
}

function fetchVideos() {
  videoStore.fetchVideoList({
    page: page.value,
    pageSize,
    keyword: route.query.keyword as string,
    tag: selectedTag.value || undefined,
  })
}

onMounted(fetchVideos)
watch(() => route.query.keyword, fetchVideos)
</script>
