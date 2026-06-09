<template>
  <main
    class="page-shell detail-page"
    @touchstart="handleTouchStart"
    @touchend="handleTouchEnd"
  >
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
          <h1>{{ video.title }}</h1>
          <div class="stats">
            <span>{{ formatCount(video.viewCount) }} 播放</span>
            <span>{{ formatCount(video.danmakuCount) }} 弹幕</span>
            <span>{{ formatTime(video.createdAt) }}</span>
          </div>
          <div class="actions">
            <el-button @click="toggleLike">点赞 {{ formatCount(video.likeCount) }}</el-button>
            <el-button @click="toggleCollect">收藏 {{ formatCount(video.collectCount) }}</el-button>
            <el-button :loading="downloading" @click="downloadVideo">下载</el-button>
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
          </div>
          <el-button type="primary" @click="router.push(`/user/${video.author.id}`)">查看主页</el-button>
        </div>

        <div class="soft-panel danmaku-box">
          <div class="danmaku-head">
            <h3>发送弹幕</h3>
            <el-radio-group v-model="danmakuMode" size="small">
              <el-radio-button label="normal">普通</el-radio-button>
              <el-radio-button label="advanced">高级</el-radio-button>
            </el-radio-group>
          </div>

          <template v-if="danmakuMode === 'normal'">
            <el-input
              v-model="danmakuText"
              class="normal-danmaku-input"
              placeholder="此刻想说什么"
              @keyup.enter="sendDanmaku"
            />
            <div class="danmaku-options">
              <el-select v-model="danmakuType" size="small">
                <el-option label="滚动" value="scroll" />
                <el-option label="顶部" value="top" />
                <el-option label="底部" value="bottom" />
              </el-select>
              <el-select v-model="danmakuFontSize" size="small">
                <el-option label="小" value="small" />
                <el-option label="中" value="medium" />
                <el-option label="大" value="large" />
              </el-select>
            </div>
          </template>

          <template v-else>
            <el-input
              v-model="danmakuText"
              type="textarea"
              :rows="4"
              placeholder="@adv x=10 y=20 tx=80 ty=20 dur=4 size=24 color=#FFFFFF | hello"
            />
            <div class="advanced-help">
              <span>x/y/tx/ty 使用 0-100 的播放器百分比，dur 控制持续秒数。</span>
              <el-button text size="small" @click="fillAdvancedTemplate">填入示例</el-button>
            </div>
            <div class="advanced-upload">
              <input
                ref="danmakuFileInput"
                class="file-input"
                type="file"
                accept=".danmaku"
                @change="uploadAdvancedFile"
              >
              <el-button :loading="uploadingAdvanced" @click="chooseDanmakuFile">
                上传 .danmaku 文件
              </el-button>
              <span>文件内部使用{}包裹单条语句,使用,隔开</span>
            </div>
          </template>

          <div class="danmaku-colors">
            <span
              v-for="c in DANMAKU_COLORS"
              :key="c"
              class="color-dot"
              :class="{ active: danmakuColor === c }"
              :style="{ background: c }"
              @click="danmakuColor = c"
            />
          </div>
          <div class="danmaku-actions">
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
import { removeUserLibraryRecord, upsertUserLibraryRecord } from '@/utils/userLibrary'

let touchStartY = 0
const SWIPE_THRESHOLD = 50
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const videoStore = useVideoStore()
const commentStore = useCommentStore()
const playerRef = ref<InstanceType<typeof VideoPlayer>>()
const danmakuFileInput = ref<HTMLInputElement>()
const loading = ref(false)
const currentTime = ref(0)
const danmakus = ref<Danmaku[]>([])
const danmakuText = ref('')
const danmakuMode = ref<'normal' | 'advanced'>('normal')
const danmakuType = ref<'scroll' | 'top' | 'bottom'>('scroll')
const danmakuFontSize = ref<'small' | 'medium' | 'large'>('medium')
const DANMAKU_COLORS = [
  '#FFFFFF', '#000000', '#FF5555', '#55FF55', '#5555FF', '#FFFF55',
  '#FF55FF', '#55FFFF', '#FF8C00', '#FF69B4', '#00CED1', '#FFD700', '#FF6347',
]
const danmakuColor = ref(DANMAKU_COLORS[0])
const commentText = ref('')
const replyTarget = ref<Comment | null>(null)
const downloading = ref(false)
const uploadingAdvanced = ref(false)
const video = computed(() => videoStore.currentVideo)

function handleTouchStart(e: TouchEvent) {
  touchStartY = e.touches[0].clientY
}

function handleTouchEnd(e: TouchEvent) {
  const touchEndY = e.changedTouches[0].clientY
  const diff = touchStartY - touchEndY

  if (diff > SWIPE_THRESHOLD) {
    goToNextVideo()
  }
}

function goToNextVideo() {
  if (!video.value) return
  const nextId = video.value.id + 1
  ElMessage.success('上滑切换到下一个视频')
  router.push(`/video/${nextId}`)
}

onMounted(load)
onUnmounted(() => {
  saveHistory()
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
    saveHistory()
  } finally {
    loading.value = false
  }
}

async function toggleLike() {
  if (!ensureLogin()) return
  if (!video.value) return
  const current = video.value
  const res = await videoApi.like(video.value.id)
  if (res.data.liked) {
    current.likeCount += 1
    upsertUserLibraryRecord('liked', current)
  } else {
    current.likeCount = Math.max(0, current.likeCount - 1)
    removeUserLibraryRecord('liked', current.id)
  }
}

async function toggleCollect() {
  if (!ensureLogin()) return
  if (!video.value) return
  const current = video.value
  const res = await videoApi.collect(video.value.id)
  if (res.data.collected) {
    current.collectCount += 1
    upsertUserLibraryRecord('collections', current)
  } else {
    current.collectCount = Math.max(0, current.collectCount - 1)
    removeUserLibraryRecord('collections', current.id)
  }
}

async function downloadVideo() {
  if (!ensureLogin()) return
  if (!video.value) return
  downloading.value = true
  try {
    const current = video.value
    const res = await videoApi.download(current.id)
    saveBlob(res.data, `${current.title || 'danmaku-video'}.mp4`)
    upsertUserLibraryRecord('downloads', current)
    ElMessage.success('下载已开始')
  } catch (error: any) {
    ElMessage.error(error.message || '下载失败')
  } finally {
    downloading.value = false
  }
}

function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

function fillAdvancedTemplate() {
  danmakuText.value = '@adv x=10 y=20 tx=80 ty=20 dur=4 size=24 color=#FFFFFF | hello'
}

function chooseDanmakuFile() {
  if (!ensureLogin()) return
  danmakuFileInput.value?.click()
}

async function uploadAdvancedFile(event: Event) {
  if (!ensureLogin()) return
  if (!video.value) return

  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return

  if (!file.name.toLowerCase().endsWith('.danmaku')) {
    ElMessage.error('请选择 .danmaku 文件')
    return
  }

  uploadingAdvanced.value = true
  try {
    const res = await danmakuApi.uploadAdvanced(video.value.id, file)
    danmakus.value.push(...res.data.list)
    video.value.danmakuCount += res.data.count
    ElMessage.success(`已上传 ${res.data.count} 条高级弹幕`)
  } catch (error: any) {
    ElMessage.error(error.message || '高级弹幕上传失败')
  } finally {
    uploadingAdvanced.value = false
  }
}

async function sendDanmaku() {
  if (!ensureLogin()) return
  if (!video.value || !danmakuText.value.trim()) return

  const content = danmakuText.value.trim()
  const res = await danmakuApi.send({
    videoId: video.value.id,
    content,
    time: Math.floor(playerRef.value?.getCurrentTime() || currentTime.value),
    color: danmakuColor.value,
    fontSize: danmakuFontSize.value,
    type: danmakuMode.value === 'advanced' ? 'advanced' : danmakuType.value,
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

function saveHistory() {
  if (!authStore.isLoggedIn || !video.value) return
  const duration = video.value.duration || 0
  const progress = duration > 0 ? Math.min(100, Math.round((currentTime.value / duration) * 100)) : 0
  upsertUserLibraryRecord('history', video.value, progress)
}
</script>

<style scoped>
.detail-page {
  display: grid;
  min-height: 100vh;
}

.watch-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 360px;
  gap: 18px;
  align-items: start;
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
  grid-template-rows: auto minmax(0, 1fr);
  gap: 16px;
  min-width: 0;
  min-height: clamp(360px, calc((min(100vw, 1180px) - 420px) * 0.5625), 620px);
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

.author-panel strong {
  display: block;
}

.danmaku-box {
  display: grid;
  align-content: start;
  gap: 16px;
  min-width: 0;
  min-height: 0;
  padding-top: 24px;
}

.danmaku-head,
.danmaku-options {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-width: 0;
}

.danmaku-head {
  flex-wrap: wrap;
}

.danmaku-head :deep(.el-radio-group) {
  max-width: 100%;
}

.danmaku-box :deep(.el-input),
.danmaku-box :deep(.el-textarea),
.danmaku-box :deep(.el-select) {
  min-width: 0;
  max-width: 100%;
}

.normal-danmaku-input {
  margin-top: 8px;
}

.normal-danmaku-input :deep(.el-input__wrapper) {
  min-height: 44px;
}

.danmaku-options .el-select {
  flex: 1;
}

.advanced-help,
.advanced-upload {
  display: grid;
  gap: 10px;
  justify-items: start;
  color: #667085;
  font-size: 12px;
  line-height: 1.5;
}

.advanced-help {
  padding-top: 2px;
}

.advanced-upload {
  padding: 10px;
  border: 1px dashed #d0d5dd;
  border-radius: 8px;
  background: #f8fafc;
}

.advanced-help span,
.advanced-upload span {
  min-width: 0;
}

.advanced-help :deep(.el-button) {
  margin-left: 0;
  padding-left: 0;
}

.file-input {
  display: none;
}

.danmaku-colors {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
  padding-top: 4px;
}

.color-dot {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid transparent;
  transition: border-color 0.15s, transform 0.15s;
}

.color-dot:hover {
  transform: scale(1.15);
}

.color-dot.active {
  border-color: #165dff;
  transform: scale(1.1);
}

.danmaku-box h3,
.comment-head h2 {
  margin: 0;
}

.comments {
  display: grid;
  gap: 14px;
  padding: 22px 26px;
}

.comment-head h2 {
  color: #18191c;
  font-size: 22px;
  font-weight: 700;
}

.comment-input {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
}

.comment-list {
  display: grid;
  gap: 0;
}

@media (max-width: 920px) {
  .watch-grid {
    grid-template-columns: 1fr;
  }
}
</style>
