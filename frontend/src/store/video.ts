import { ref } from 'vue'
import { defineStore } from 'pinia'
import { videoApi } from '@/api/video'
import type { VideoInfo } from '@/types'

export const useVideoStore = defineStore('video', () => {
  const videoList = ref<VideoInfo[]>([])
  const currentVideo = ref<VideoInfo | null>(null)
  const total = ref(0)
  const loading = ref(false)

  async function fetchVideoList(
    params: { 
      page: number; 
      pageSize: number; 
      keyword?: string; 
      tag?: string 
      category?: string
    }
  ) {
    loading.value = true
    try {
      const res = await videoApi.list(params)
      videoList.value = res.data.list
      total.value = res.data.total
    } finally {
      loading.value = false
    }
  }

  async function fetchVideoDetail(id: number) {
    loading.value = true
    try {
      const res = await videoApi.detail(id)
      currentVideo.value = res.data
    } finally {
      loading.value = false
    }
  }

  function clearCurrent() {
    currentVideo.value = null
  }

  return { videoList, currentVideo, total, loading, fetchVideoList, fetchVideoDetail, clearCurrent }
})
