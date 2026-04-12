<template>
  <div class="home-page">
    <div class="banner">
      <h2>发现优质视频内容</h2>
    </div>
    <div class="tag-filter">
      <a-tag
        v-for="tag in popularTags"
        :key="tag"
        :color="selectedTag === tag ? 'arcoblue' : ''"
        checkable
        :checked="selectedTag === tag"
        @check="selectTag(tag)"
      >{{ tag }}</a-tag>
    </div>
    <a-spin :loading="videoStore.loading">
      <div class="video-grid">
        <VideoCard
          v-for="video in videoStore.videoList"
          :key="video.id"
          :video="video"
          @click="router.push(`/video/${video.id}`)"
        />
      </div>
    </a-spin>
    <div class="pagination">
      <a-pagination
        v-model:current="page"
        :total="videoStore.total"
        :page-size="pageSize"
        @change="fetchVideos"
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
const pageSize = 24
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

<style scoped>
.home-page { max-width: 1400px; margin: 0 auto; }
.banner { padding: 32px 0 16px; text-align: center; }
.tag-filter { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 24px; }
.video-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
}
.pagination { display: flex; justify-content: center; margin-top: 32px; }
</style>
