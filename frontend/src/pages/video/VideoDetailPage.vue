<template>
  <div class="max-w-5xl mx-auto pb-20">
    <div v-loading="loading">
      <div v-if="video">
        <!-- Player -->
        <VideoPlayer
          ref="videoPlayerRef"
          :url="video.videoUrl"
          :poster="video.coverUrl"
          :danmakus="danmakuList"
          @timeupdate="handleTimeUpdate"
          @error="handlePlayerError"
        />

        <!-- Video info -->
        <div class="mt-4">
          <h1 class="text-2xl font-bold text-gray-900">{{ video.title }}</h1>
          <div class="flex flex-wrap gap-4 text-sm text-gray-500 mt-2">
            <span>{{ video.viewCount }} 次播放</span>
            <span>{{ formatTime(video.createdAt) }}</span>
          </div>

          <div class="flex gap-3 mt-3">
            <el-button
              :type="liked ? 'primary' : 'default'"
              :icon="liked ? 'StarFilled' : 'Star'"
              @click="toggleLike"
            >
              <el-icon><Pointer /></el-icon>
              {{ video.likeCount }}
            </el-button>
            <el-button
              :type="collected ? 'primary' : 'default'"
              @click="toggleCollect"
            >
              <el-icon><Star /></el-icon>
              {{ video.collectCount }}
            </el-button>
          </div>

          <div class="flex flex-wrap gap-2 mt-3">
            <el-tag v-for="tag in normalizedTags" :key="tag" size="small">{{ tag }}</el-tag>
          </div>

          <p class="text-gray-600 leading-relaxed mt-3 whitespace-pre-wrap">
            {{ video.description || '暂无简介' }}
          </p>
        </div>

        <!-- Comments -->
        <div class="mt-8 pt-6 border-t border-gray-200">
          <div class="flex items-center gap-3 mb-5">
            <h3 class="text-xl font-bold text-gray-800">评论</h3>
            <span class="text-sm text-gray-500">{{ commentCount }} 条</span>
          </div>

          <!-- Input -->
          <div v-if="authStore.isLoggedIn" class="mb-6 space-y-2">
            <div v-if="replyTarget" class="flex items-center justify-between bg-gray-100 rounded-lg px-3 py-2 text-sm text-gray-600">
              <span>正在回复：{{ getCommentAuthorName(replyTarget) }}</span>
              <el-button type="text" size="small" @click="cancelReply">取消回复</el-button>
            </div>
            <el-input
              v-if="replyTarget"
              v-model="replyInput"
              :placeholder="`回复 ${getCommentAuthorName(replyTarget)}...`"
              type="textarea"
              :maxlength="500"
              show-word-limit
              :autosize="{ minRows: 2, maxRows: 5 }"
            />
            <el-input
              v-else
              v-model="commentInput"
              placeholder="发表你的看法..."
              type="textarea"
              :maxlength="500"
              show-word-limit
              :autosize="{ minRows: 2, maxRows: 5 }"
            />
            <div class="flex justify-end">
              <el-button
                type="primary"
                :loading="commentStore.submitting"
                @click="submitComment(replyTarget?.id)"
              >
                {{ replyTarget ? '发送回复' : '发表评论' }}
              </el-button>
            </div>
          </div>

          <el-alert v-else title="登录后可以发表评论" type="info" show-icon :closable="false" class="mb-6" />

          <!-- List -->
          <div v-loading="commentStore.loading" class="space-y-5">
            <el-empty v-if="commentStore.comments.length === 0" description="暂无评论" />
            <CommentItem
              v-for="comment in commentStore.comments"
              :key="comment.id"
              :comment="comment"
              @reply="startReply"
            />
          </div>
        </div>
      </div>

      <el-empty v-else-if="!loading" description="视频不存在" class="py-20" />
    </div>

    <!-- Danmaku bar -->
    <div v-if="video" class="fixed bottom-0 left-0 right-0 z-50 bg-white border-t border-gray-200 px-4 py-3 flex gap-2 items-center">
      <el-input
        v-model="danmakuInput"
        placeholder="发送弹幕..."
        :maxlength="100"
        class="flex-1"
        @keydown.enter="sendDanmaku"
      />
      <el-color-picker v-model="danmakuColor" size="small" />
      <el-button type="primary" @click="sendDanmaku">发送</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
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
  if (!tags) return []
  if (Array.isArray(tags)) return tags
  return String(tags).split(',').map(tag => tag.trim()).filter(Boolean)
})

const commentCount = computed(() => commentStore.countComments())

onMounted(async () => {
  const videoId = Number(route.params.id)
  if (!Number.isFinite(videoId) || videoId <= 0) {
    ElMessage.error('视频 ID 不合法')
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
    ElMessage.error('视频详情加载失败')
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
  if (!authStore.isLoggedIn) { ElMessage.warning('请先登录'); return }
  if (!video.value) return
  const content = parentId ? replyInput.value.trim() : commentInput.value.trim()
  if (!content) { ElMessage.warning('评论内容不能为空'); return }
  try {
    await commentStore.createComment(video.value.id, content, parentId)
    ElMessage.success(parentId ? '回复成功' : '评论成功')
    if (parentId) { replyInput.value = ''; replyTarget.value = null }
    else { commentInput.value = '' }
  } catch {
    ElMessage.error(parentId ? '回复失败' : '评论失败')
  }
}

function startReply(comment: Comment) {
  if (!authStore.isLoggedIn) { ElMessage.warning('请先登录'); return }
  replyTarget.value = comment
  replyInput.value = ''
}

function cancelReply() {
  replyTarget.value = null
  replyInput.value = ''
}

async function toggleLike() {
  if (!authStore.isLoggedIn) { ElMessage.warning('请先登录'); return }
  if (!video.value) return
  try {
    await videoApi.likeVideo(video.value.id)
    liked.value = !liked.value
  } catch { ElMessage.error('操作失败') }
}

async function toggleCollect() {
  if (!authStore.isLoggedIn) { ElMessage.warning('请先登录'); return }
  if (!video.value) return
  try {
    await videoApi.collectVideo(video.value.id)
    collected.value = !collected.value
  } catch { ElMessage.error('操作失败') }
}

async function sendDanmaku() {
  const content = danmakuInput.value.trim()
  if (!content) return
  if (!authStore.isLoggedIn) { ElMessage.warning('请先登录'); return }
  if (!video.value) return
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
  } catch { ElMessage.error('弹幕发送失败') }
}

function handleTimeUpdate(time: number) { currentTime.value = time }
function handlePlayerError(message: string) { ElMessage.error(message) }
function getCommentAuthorName(comment: Comment) {
  return comment.author?.nickname || comment.author?.username || '匿名用户'
}
function formatTime(time?: string) {
  if (!time) return ''
  const date = new Date(time)
  if (Number.isNaN(date.getTime())) return time
  return date.toLocaleString()
}
function normalizeResponseData<T>(res: unknown, fallback: T): T {
  const response = res as { data?: ApiResponse<T> | T }
  if (response.data && typeof response.data === 'object' && 'data' in response.data) {
    return response.data.data ?? fallback
  }
  return (response.data as T) ?? fallback
}
</script>
