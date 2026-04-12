<template>
  <div class="video-detail">
    <a-spin :loading="loading">
      <div class="main-content" v-if="video">
        <!-- 播放器区域 -->
        <div class="player-wrapper">
          <div ref="playerContainer" class="player-container" />
        </div>

        <!-- 视频信息 -->
        <div class="video-info">
          <h1>{{ video.title }}</h1>
          <div class="meta">
            <span>{{ video.viewCount }} 次播放</span>
            <span>{{ video.createdAt }}</span>
          </div>
          <div class="actions">
            <a-button :type="liked ? 'primary' : 'outline'" @click="toggleLike">
              <template #icon><icon-thumb-up /></template>
              {{ video.likeCount }}
            </a-button>
            <a-button :type="collected ? 'primary' : 'outline'" @click="toggleCollect">
              <template #icon><icon-star /></template>
              {{ video.collectCount }}
            </a-button>
          </div>
          <div class="tags">
            <a-tag v-for="tag in video.tags" :key="tag">{{ tag }}</a-tag>
          </div>
          <div class="description">{{ video.description }}</div>
        </div>

        <!-- 评论区 -->
        <div class="comment-section">
          <h3>评论 ({{ comments.length }})</h3>
          <div class="comment-input" v-if="authStore.isLoggedIn">
            <a-textarea v-model="commentInput" placeholder="发表你的看法..." :max-length="500" show-word-limit />
            <a-button type="primary" @click="submitComment">发表评论</a-button>
          </div>
          <div class="comment-list">
            <CommentItem v-for="c in comments" :key="c.id" :comment="c" />
          </div>
        </div>
      </div>
    </a-spin>

    <!-- 弹幕控制栏 -->
    <div class="danmaku-control" v-if="video">
      <a-input
        v-model="danmakuInput"
        placeholder="发送弹幕..."
        :max-length="100"
        @keydown.enter="sendDanmaku"
      />
      <a-color-picker v-model="danmakuColor" size="small" />
      <a-button type="primary" @click="sendDanmaku">发送</a-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { useVideoStore } from '@/store/video'
import { danmakuApi } from '@/api/danmaku'
import { videoApi } from '@/api/video'
import CommentItem from '@/components/common/CommentItem.vue'
import type { Comment, Danmaku } from '@/types'

const route = useRoute()
const authStore = useAuthStore()
const videoStore = useVideoStore()

const playerContainer = ref<HTMLElement>()
const loading = ref(true)
const liked = ref(false)
const collected = ref(false)
const comments = ref<Comment[]>([])
const commentInput = ref('')
const danmakuInput = ref('')
const danmakuColor = ref('#FFFFFF')
const danmakuList = ref<Danmaku[]>([])

const video = ref(videoStore.currentVideo)

onMounted(async () => {
  const id = Number(route.params.id)
  await videoStore.fetchVideoDetail(id)
  video.value = videoStore.currentVideo
  const res = await danmakuApi.getDanmakuList(id)
  danmakuList.value = res.data.data
  loading.value = false
  // 初始化播放器（xgplayer）
  // initPlayer()
})

async function toggleLike() {
  if (!authStore.isLoggedIn) return
  await videoApi.likeVideo(video.value!.id)
  liked.value = !liked.value
}

async function toggleCollect() {
  if (!authStore.isLoggedIn) return
  await videoApi.collectVideo(video.value!.id)
  collected.value = !collected.value
}

async function submitComment() {
  if (!commentInput.value.trim()) return
  // TODO: call comment API
  commentInput.value = ''
}

async function sendDanmaku() {
  if (!danmakuInput.value.trim() || !authStore.isLoggedIn) return
  await danmakuApi.sendDanmaku({
    videoId: video.value!.id,
    content: danmakuInput.value,
    time: 0, // 当前播放时间
    color: danmakuColor.value,
    fontSize: 'medium',
    type: 'scroll',
  })
  danmakuInput.value = ''
}
</script>

<style scoped>
.video-detail { max-width: 1200px; margin: 0 auto; }
.player-wrapper { background: #000; border-radius: 8px; overflow: hidden; aspect-ratio: 16/9; }
.player-container { width: 100%; height: 100%; }
.video-info { padding: 16px 0; }
.meta { color: #86909c; margin: 8px 0; display: flex; gap: 16px; }
.actions { display: flex; gap: 12px; margin: 12px 0; }
.tags { display: flex; gap: 8px; margin: 12px 0; }
.comment-section { margin-top: 24px; }
.comment-input { display: flex; flex-direction: column; gap: 8px; margin-bottom: 16px; }
.danmaku-control {
  position: sticky;
  bottom: 0;
  background: #fff;
  padding: 12px;
  display: flex;
  gap: 8px;
  align-items: center;
  box-shadow: 0 -2px 8px rgba(0,0,0,.1);
}
</style>
