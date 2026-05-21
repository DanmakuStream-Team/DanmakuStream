import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { VideoInfo } from '@/types'
import { videoApi } from '@/api/video'

export const useVideoStore = defineStore('video', () => {
  const videoList = ref<VideoInfo[]>([])
  const currentVideo = ref<VideoInfo | null>(null)
  const loading = ref(false)
  const total = ref(0)

  async function fetchVideoList(params: { page: number; pageSize: number; keyword?: string; tag?: string }) {
    loading.value = true
    try {
      const res = await videoApi.getVideoList(params)
      videoList.value = res.data.list
      total.value = res.data.total
    } finally {
      loading.value = false
    }
  }

  async function fetchVideoDetail(id: number) {
    const res = await videoApi.getVideoDetail(id)
    currentVideo.value = res.data
  }

  return { videoList, currentVideo, loading, total, fetchVideoList, fetchVideoDetail }
})
