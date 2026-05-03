<template>
  <div class="video-detail">
    <a-spin :loading="loading">
      <div v-if="video" class="main-content">
        <!-- 播放器区域 -->
        <VideoPlayer
          ref="videoPlayerRef"
          :url="video.videoUrl"
          :poster="video.coverUrl"
          :danmakus="danmakuList"
          @timeupdate="handleTimeUpdate"
          @error="handlePlayerError"
        />

        <!-- 视频信息 -->
        <div class="video-info">
          <h1>{{ video.title }}</h1>

          <div class="meta">
            <span>{{ video.viewCount }} 次播放</span>
            <span>{{ formatTime(video.createdAt) }}</span>
          </div>

          <div class="actions">
            <a-button :type="liked ? 'primary' : 'outline'" @click="toggleLike">
              <template #icon>
                <icon-thumb-up />
              </template>
              {{ video.likeCount }}
            </a-button>

            <a-button :type="collected ? 'primary' : 'outline'" @click="toggleCollect">
              <template #icon>
                <icon-star />
              </template>
              {{ video.collectCount }}
            </a-button>
          </div>

          <div class="tags">
            <a-tag v-for="tag in normalizedTags" :key="tag">
              {{ tag }}
            </a-tag>
          </div>

          <div class="description">
            {{ video.description || '暂无简介' }}
          </div>
        </div>

        <!-- 评论区 -->
        <div class="comment-section">
          <div class="comment-header">
            <h3>评论</h3>
            <span>{{ commentCount }} 条</span>
          </div>

          <!-- 评论输入框 -->
          <div v-if="authStore.isLoggedIn" class="comment-input">
            <div v-if="replyTarget" class="reply-target">
              <span>
                正在回复：{{ getCommentAuthorName(replyTarget) }}
              </span>

              <a-button type="text" size="mini" @click="cancelReply">
                取消回复
              </a-button>
            </div>

            <a-textarea
              v-if="replyTarget"
              v-model="replyInput"
              :placeholder="`回复 ${getCommentAuthorName(replyTarget)}...`"
              :max-length="500"
              show-word-limit
              allow-clear
            />

            <a-textarea
              v-else
              v-model="commentInput"
              placeholder="发表你的看法..."
              :max-length="500"
              show-word-limit
              allow-clear
            />

            <div class="comment-submit-row">
              <a-button
                type="primary"
                :loading="commentStore.submitting"
                @click="submitComment(replyTarget?.id)"
              >
                {{ replyTarget ? '发送回复' : '发表评论' }}
              </a-button>
            </div>
          </div>

          <a-alert v-else type="info" show-icon class="login-tip">
            登录后可以发表评论
          </a-alert>

          <!-- 评论列表 -->
          <a-spin :loading="commentStore.loading">
            <div class="comment-list">
              <a-empty
                v-if="commentStore.comments.length === 0"
                description="暂无评论"
              />

              <template v-else>
                <CommentItem
                  v-for="comment in commentStore.comments"
                  :key="comment.id"
                  :comment="comment"
                  @reply="startReply"
                />
              </template>
            </div>
          </a-spin>
        </div>
      </div>

      <a-empty v-else-if="!loading" description="视频不存在" />
    </a-spin>

    <!-- 弹幕发送栏 -->
    <div v-if="video" class="danmaku-control">
      <a-input
        v-model="danmakuInput"
        placeholder="发送弹幕..."
        :max-length="100"
        @keydown.enter="sendDanmaku"
      />

      <a-color-picker v-model="danmakuColor" size="small" />

      <a-button type="primary" @click="sendDanmaku">
        发送
      </a-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { useVideoStore } from '@/store/video'
import { useCommentStore } from '@/store/comment'
import { danmakuApi } from '@/api/danmaku'
import { videoApi } from '@/api/video'
import CommentItem from '@/components/common/CommentItem.vue'
import VideoPlayer from '@/components/common/VideoPlayer.vue'
import type { ApiResponse, Comment, Danmaku } from '@/types'

type VideoPlayerExpose = {
  play: () => void
  pause: () => void
  seek: (time: number) => void
  getCurrentTime: () => number
  getDuration: () => number
}

const route = useRoute()
const authStore = useAuthStore()
const videoStore = useVideoStore()
const commentStore = useCommentStore()

const videoPlayerRef = ref<VideoPlayerExpose>()

const loading = ref(true)

const liked = ref(false)
const collected = ref(false)

const commentInput = ref('')
const replyInput = ref('')
const replyTarget = ref<Comment | null>(null)

const danmakuInput = ref('')
const danmakuColor = ref('#FFFFFF')
const danmakuList = ref<Danmaku[]>([])
const currentTime = ref(0)

const video = computed(() => videoStore.currentVideo)

const normalizedTags = computed(() => {
  const tags = video.value?.tags as unknown

  if (!tags) {
    return []
  }

  if (Array.isArray(tags)) {
    return tags
  }

  return String(tags)
    .split(',')
    .map(tag => tag.trim())
    .filter(Boolean)
})

const commentCount = computed(() => {
  return commentStore.countComments()
})

onMounted(async () => {
  const videoId = Number(route.params.id)

  if (!Number.isFinite(videoId) || videoId <= 0) {
    Message.error('视频 ID 不合法')
    loading.value = false
    return
  }

  try {
    loading.value = true

    await videoStore.fetchVideoDetail(videoId)

    await Promise.allSettled([
      commentStore.fetchComments(videoId),
      fetchDanmakuList(videoId),
    ])
  } catch {
    Message.error('视频详情加载失败')
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  commentStore.clearComments()
})

async function fetchDanmakuList(videoId: number) {
  try {
    const res = await danmakuApi.getDanmakuList(videoId)
    danmakuList.value = normalizeResponseData<Danmaku[]>(res, [])
  } catch {
    danmakuList.value = []
  }
}

async function submitComment(parentId?: number) {
  if (!authStore.isLoggedIn) {
    Message.warning('请先登录')
    return
  }

  if (!video.value) {
    return
  }

  const content = parentId ? replyInput.value.trim() : commentInput.value.trim()

  if (!content) {
    Message.warning('评论内容不能为空')
    return
  }

  try {
    await commentStore.createComment(video.value.id, content, parentId)

    Message.success(parentId ? '回复成功' : '评论成功')

    if (parentId) {
      replyInput.value = ''
      replyTarget.value = null
    } else {
      commentInput.value = ''
    }
  } catch {
    Message.error(parentId ? '回复失败' : '评论失败')
  }
}

function startReply(comment: Comment) {
  if (!authStore.isLoggedIn) {
    Message.warning('请先登录')
    return
  }

  replyTarget.value = comment
  replyInput.value = ''
}

function cancelReply() {
  replyTarget.value = null
  replyInput.value = ''
}

async function toggleLike() {
  if (!authStore.isLoggedIn) {
    Message.warning('请先登录')
    return
  }

  if (!video.value) {
    return
  }

  try {
    await videoApi.likeVideo(video.value.id)
    liked.value = !liked.value
  } catch {
    Message.error('操作失败')
  }
}

async function toggleCollect() {
  if (!authStore.isLoggedIn) {
    Message.warning('请先登录')
    return
  }

  if (!video.value) {
    return
  }

  try {
    await videoApi.collectVideo(video.value.id)
    collected.value = !collected.value
  } catch {
    Message.error('操作失败')
  }
}

async function sendDanmaku() {
  const content = danmakuInput.value.trim()

  if (!content) {
    return
  }

  if (!authStore.isLoggedIn) {
    Message.warning('请先登录')
    return
  }

  if (!video.value) {
    return
  }

  const sendTime = videoPlayerRef.value?.getCurrentTime() || currentTime.value || 0

  try {
    await danmakuApi.sendDanmaku({
      videoId: video.value.id,
      content,
      time: sendTime,
      color: danmakuColor.value,
      fontSize: 'medium',
      type: 'scroll',
    })

    danmakuList.value.push({
      id: Date.now(),
      videoId: video.value.id,
      userId: authStore.userInfo?.id || 0,
      content,
      time: sendTime,
      color: danmakuColor.value,
      fontSize: 'medium',
      type: 'scroll',
      createdAt: new Date().toISOString(),
    })

    danmakuInput.value = ''
  } catch {
    Message.error('弹幕发送失败')
  }
}

function handleTimeUpdate(time: number) {
  currentTime.value = time
}

function handlePlayerError(message: string) {
  Message.error(message)
}

function getCommentAuthorName(comment: Comment) {
  return comment.author?.nickname || comment.author?.username || '匿名用户'
}

function formatTime(time?: string) {
  if (!time) {
    return ''
  }

  const date = new Date(time)

  if (Number.isNaN(date.getTime())) {
    return time
  }

  return date.toLocaleString()
}

function normalizeResponseData<T>(res: unknown, fallback: T): T {
  const response = res as {
    data?: ApiResponse<T> | T
  }

  if (
    response.data &&
    typeof response.data === 'object' &&
    'data' in response.data
  ) {
    return response.data.data ?? fallback
  }

  return (response.data as T) ?? fallback
}
</script>

<style scoped>
.video-detail {
  max-width: 1200px;
  margin: 0 auto;
  padding-bottom: 72px;
}

.main-content {
  width: 100%;
}

.video-info {
  padding: 16px 0;
}

.video-info h1 {
  margin: 0;
  font-size: 24px;
  line-height: 1.4;
}

.meta {
  color: #86909c;
  margin: 8px 0;
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}

.actions {
  display: flex;
  gap: 12px;
  margin: 12px 0;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 12px 0;
}

.description {
  color: #4e5969;
  line-height: 1.8;
  white-space: pre-wrap;
}

.comment-section {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #e5e6eb;
}

.comment-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.comment-header h3 {
  margin: 0;
  font-size: 20px;
}

.comment-header span {
  color: #86909c;
  font-size: 14px;
}

.comment-input {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 20px;
}

.reply-target {
  display: flex;
  align-items: center;
  justify-content: space-between;
  color: #4e5969;
  background: #f2f3f5;
  border-radius: 6px;
  padding: 6px 10px;
}

.comment-submit-row {
  display: flex;
  justify-content: flex-end;
}

.login-tip {
  margin-bottom: 20px;
}

.comment-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.danmaku-control {
  position: sticky;
  bottom: 0;
  z-index: 50;
  background: #fff;
  padding: 12px;
  display: flex;
  gap: 8px;
  align-items: center;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.1);
}
</style>