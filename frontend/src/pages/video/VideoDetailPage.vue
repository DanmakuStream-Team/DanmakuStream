<template>
  <main class="page-shell detail-page">
    <section v-loading="loading" class="watch-grid">
      <div v-if="video" class="main-col">
        <VideoPlayer
          ref="playerRef"
          :url="video.videoUrl"
          :poster="video.coverUrl"
          :danmakus="danmakus"
          @timeupdate="currentTime = $event"
          @error="ElMessage.error($event)"
        />

        <div class="video-info">
          <el-tag type="primary" effect="light">{{ video.status }}</el-tag>
          <h1>{{ video.title }}</h1>
          <div class="stats">
            <span>{{ formatCount(video.viewCount) }} 播放</span>
            <span>{{ formatCount(video.danmakuCount) }} 弹幕</span>
            <span>{{ formatTime(video.createdAt) }}</span>
          </div>
          <div class="actions">
            <el-button @click="toggleLike">点赞 {{ formatCount(video.likeCount) }}</el-button>
            <el-button @click="toggleCollect">收藏 {{ formatCount(video.collectCount) }}</el-button>
          </div>
          <div class="tags">
            <el-tag v-for="tag in normalizeTags(video.tags)" :key="tag">{{ tag }}</el-tag>
          </div>
          <p>{{ video.description || '这个视频暂无简介。' }}</p>
        </div>

        <section class="soft-panel comments">
          <div class="comment-head">
            <h2>评论 {{ commentStore.countComments() }}</h2>
          </div>
          <div class="comment-input">
            <el-input
              v-model="commentText"
              type="textarea"
              :rows="3"
              :placeholder="replyTarget ? `回复 ${replyTarget.author?.nickname || '用户'}` : '写下你的看法'"
            />
            <el-button type="primary" :loading="commentStore.submitting" @click="submitComment">
              {{ replyTarget ? '发送回复' : '发表评论' }}
            </el-button>
            <el-button v-if="replyTarget" text @click="replyTarget = null">取消回复</el-button>
          </div>
          <div v-loading="commentStore.loading" class="comment-list">
            <CommentItem
              v-for="comment in commentStore.comments"
              :key="comment.id"
              :comment="comment"
              @reply="replyTarget = $event"
              @like="likeComment"
            />
            <el-empty v-if="!commentStore.comments.length" description="暂无评论" />
          </div>
        </section>
      </div>

      <aside v-if="video" class="side-col">
        <div class="soft-panel author-panel">
          <el-avatar :size="52" :src="mediaUrl(video.author?.avatar)">
            {{ video.author?.nickname?.slice(0, 1) || 'U' }}
          </el-avatar>
          <div>
            <strong>{{ video.author?.nickname || '匿名用户' }}</strong>
            <span>{{ video.author?.role }}</span>
          </div>
          <el-button type="primary" @click="router.push(`/user/${video.author.id}`)">查看主页</el-button>
        </div>

        <div class="soft-panel danmaku-box">
          <h3>发送弹幕</h3>
          <el-input v-model="danmakuText" placeholder="此刻想说什么" @keyup.enter="sendDanmaku" />
          <div class="danmaku-actions">
            <el-color-picker v-model="danmakuColor" />
            <el-button type="primary" @click="sendDanmaku">发送</el-button>
          </div>
        </div>
      </aside>

      <div v-if="!loading && !video" class="soft-panel empty-panel">
        <el-empty description="视频不存在或尚未通过审核" />
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import VideoPlayer from '@/components/common/VideoPlayer.vue'
import CommentItem from '@/components/common/CommentItem.vue'
import { danmakuApi } from '@/api/danmaku'
import { videoApi } from '@/api/video'
import { commentApi } from '@/api/comment'
import { useAuthStore } from '@/store/auth'
import { useCommentStore } from '@/store/comment'
import { useVideoStore } from '@/store/video'
import type { Comment, Danmaku } from '@/types'
import { formatCount, formatTime, mediaUrl, normalizeTags } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const videoStore = useVideoStore()
const commentStore = useCommentStore()
const playerRef = ref<InstanceType<typeof VideoPlayer>>()
const loading = ref(false)
const currentTime = ref(0)
const danmakus = ref<Danmaku[]>([])
const danmakuText = ref('')
const danmakuColor = ref('#FFFFFF')
const commentText = ref('')
const replyTarget = ref<Comment | null>(null)
const video = computed(() => videoStore.currentVideo)

onMounted(load)
onUnmounted(() => {
  videoStore.clearCurrent()
  commentStore.clearComments()
})

async function load() {
  const id = Number(route.params.id)
  loading.value = true
  try {
    await videoStore.fetchVideoDetail(id)
    const [danmakuRes] = await Promise.all([
      danmakuApi.list(id),
      commentStore.fetchComments(id),
    ])
    danmakus.value = danmakuRes.data
  } finally {
    loading.value = false
  }
}

async function toggleLike() {
  if (!ensureLogin()) return
  if (!video.value) return
  await videoApi.like(video.value.id)
  await videoStore.fetchVideoDetail(video.value.id)
}

async function toggleCollect() {
  if (!ensureLogin()) return
  if (!video.value) return
  await videoApi.collect(video.value.id)
  await videoStore.fetchVideoDetail(video.value.id)
}

async function sendDanmaku() {
  if (!ensureLogin()) return
  if (!video.value || !danmakuText.value.trim()) return
  const res = await danmakuApi.send({
    videoId: video.value.id,
    content: danmakuText.value.trim(),
    time: Math.floor(playerRef.value?.getCurrentTime() || currentTime.value),
    color: danmakuColor.value,
    fontSize: 'medium',
    type: 'scroll',
  })
  danmakus.value.push(res.data)
  danmakuText.value = ''
}

async function submitComment() {
  if (!ensureLogin()) return
  if (!video.value || !commentText.value.trim()) return
  await commentStore.createComment(video.value.id, commentText.value.trim(), replyTarget.value?.id)
  commentText.value = ''
  replyTarget.value = null
}

async function likeComment(comment: Comment) {
  if (!ensureLogin()) return
  const res = await commentApi.like(comment.id)
  comment.liked = res.data.liked
  comment.likeCount = res.data.likeCount
}

function ensureLogin() {
  if (authStore.isLoggedIn) return true
  ElMessage.warning('请先登录')
  router.push('/login')
  return false
}
</script>

<style scoped>
.detail-page {
  display: grid;
}

.watch-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 18px;
}

.main-col {
  display: grid;
  gap: 18px;
}

.video-info {
  display: grid;
  gap: 12px;
}

.video-info h1 {
  margin: 0;
  color: #142033;
  font-size: 30px;
  line-height: 1.3;
}

.stats,
.actions,
.tags,
.danmaku-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.stats {
  color: #667085;
  font-size: 14px;
}

.video-info p {
  margin: 0;
  color: #475467;
  line-height: 1.8;
}

.side-col {
  display: grid;
  align-content: start;
  gap: 16px;
}

.author-panel,
.danmaku-box,
.comments {
  padding: 18px;
}

.author-panel {
  display: grid;
  gap: 12px;
}

.author-panel strong,
.author-panel span {
  display: block;
}

.author-panel span {
  color: #667085;
  margin-top: 4px;
}

.danmaku-box {
  display: grid;
  gap: 12px;
}

.danmaku-box h3,
.comment-head h2 {
  margin: 0;
}

.comments {
  display: grid;
  gap: 16px;
}

.comment-input {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
}

.comment-list {
  display: grid;
  gap: 18px;
}

@media (max-width: 920px) {
  .watch-grid {
    grid-template-columns: 1fr;
  }
}
</style>
