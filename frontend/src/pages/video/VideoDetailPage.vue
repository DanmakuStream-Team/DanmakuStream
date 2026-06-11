<template>
  <main class="page-shell detail-page" @touchstart="handleTouchStart" @touchend="handleTouchEnd">
    <section v-loading="loading" class="watch-grid">
      <div v-if="video" class="main-col">
        <VideoPlayer
          ref="playerRef"
          v-model:danmaku-visible="danmakuVisible"
          :url="video.videoUrl"
          :poster="video.coverUrl"
          :danmakus="danmakus"
          :danmaku-opacity="danmakuOpacity"
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
            <el-tag v-if="video.category" type="success">{{ categoryLabel(video.category) }}</el-tag>
            <el-tag v-for="tag in normalizeTags(video.tags)" :key="tag">{{ tag }}</el-tag>
          </div>
          <p>{{ video.description || '这个视频暂无简介。' }}</p>
        </div>

        <section class="soft-panel comments">
          <div class="comment-head">
            <h2>评论 {{ commentStore.countComments() }}</h2>
            <el-button text @click="commentsCollapsed = !commentsCollapsed">
              {{ commentsCollapsed ? '展开评论' : '收起评论' }}
            </el-button>
          </div>
          <div v-show="!commentsCollapsed" class="comment-body">
            <div class="comment-input">
              <el-input v-model="commentText" type="textarea" :rows="3" placeholder="写下你的看法" />
              <el-button type="primary" :loading="commentStore.submitting" @click="submitComment">
                发表评论
              </el-button>
            </div>
            <div v-loading="commentStore.loading" class="comment-list">
              <CommentItem
                v-for="comment in commentStore.comments"
                :key="comment.id"
                :comment="comment"
                @reply="submitReply"
                @like="likeComment"
              />
              <el-empty v-if="!commentStore.comments.length" description="暂无评论" />
            </div>
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
            <span>{{ video.author?.username }}</span>
          </div>
          <el-button type="primary" @click="router.push(`/user/${video.author.id}`)">查看主页</el-button>
        </div>

        <div v-if="canManageVideo" class="soft-panel collaborator-panel">
          <h3>共创管理</h3>
          <div class="collaborator-search">
            <el-input v-model="collaboratorKeyword" placeholder="搜索用户昵称或用户名" @keyup.enter="searchCollaborators" />
            <el-button @click="searchCollaborators">搜索</el-button>
          </div>
          <div class="collaborator-options">
            <button
              v-for="user in collaboratorOptions"
              :key="user.id"
              type="button"
              class="collaborator-item"
              @click="inviteCollaborator(user.id)"
            >
              <span>{{ user.nickname }}</span>
              <small>@{{ user.username }}</small>
            </button>
          </div>
        </div>

        <div class="soft-panel side-panel">
          <el-radio-group v-model="sideMode" size="small" class="side-tabs">
            <el-radio-button label="normal">普通弹幕</el-radio-button>
            <el-radio-button label="advanced">高级弹幕</el-radio-button>
            <el-radio-button label="collection">视频合集</el-radio-button>
          </el-radio-group>

          <template v-if="sideMode === 'normal'">
            <el-input
              v-model="normalDanmakuText"
              class="normal-danmaku-input"
              placeholder="此刻想说什么"
              @keyup.enter="sendNormalDanmaku"
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
            <div class="danmaku-control-row">
              <span>显示弹幕</span>
              <el-switch v-model="danmakuVisible" />
            </div>
            <div class="danmaku-control-row">
              <span>透明度</span>
              <el-slider v-model="danmakuOpacity" :min="10" :max="100" :step="5" />
            </div>
            <el-button class="send-button" type="primary" @click="sendNormalDanmaku">发送</el-button>
          </template>

          <template v-else-if="sideMode === 'advanced'">
            <el-input
              v-model="advancedDanmakuText"
              type="textarea"
              :rows="5"
              placeholder="@adv x=10 y=20 tx=80 ty=20 dur=4 size=24 color=#FFFFFF | hello"
            />
            <div class="advanced-help">
              <span>x/y/tx/ty 使用 0-100 的播放器百分比，dur 控制持续秒数。</span>
              <el-button text size="small" @click="fillAdvancedTemplate">填入示例</el-button>
            </div>
            <div class="advanced-upload">
              <input ref="danmakuFileInput" class="file-input" type="file" accept=".danmaku" @change="uploadAdvancedFile">
              <el-button :loading="uploadingAdvanced" @click="chooseDanmakuFile">上传 .danmaku 文件</el-button>
              <span>文件内可用花括号包住单条语句，用逗号分隔。</span>
            </div>
            <el-button class="send-button" type="primary" @click="sendAdvancedDanmaku">发送高级弹幕</el-button>
          </template>

          <template v-else>
            <div v-if="canManageVideo" class="collection-editor">
              <el-input v-model="newCollectionTitle" placeholder="新合集标题" />
              <el-button type="primary" @click="createCollectionAndAddVideo">创建并加入当前视频</el-button>
              <el-select v-model="selectedCollectionId" clearable placeholder="选择我的合集">
                <el-option v-for="item in myCollections" :key="item.id" :label="item.title" :value="item.id" />
              </el-select>
              <el-button @click="addVideoToSelectedCollection">加入所选合集</el-button>
            </div>

            <div class="collection-list">
              <article v-for="collection in videoCollections" :key="collection.id" class="collection-card">
                <div>
                  <strong>{{ collection.title }}</strong>
                  <span>{{ collection.owner?.nickname }}</span>
                </div>
                <el-button size="small" @click="openCollection(collection.id)">查看</el-button>
                <el-button
                  v-if="canManageCollection(collection)"
                  size="small"
                  type="danger"
                  plain
                  @click="removeVideoFromCollection(collection.id)"
                >
                  移出
                </el-button>
              </article>
              <el-empty v-if="!videoCollections.length" description="当前视频还没有加入合集" />
            </div>

            <div v-if="selectedCollectionDetail" class="collection-detail">
              <strong>{{ selectedCollectionDetail.title }}</strong>
              <button
                v-for="item in selectedCollectionDetail.videos || []"
                :key="item.id"
                type="button"
                class="collection-video"
                @click="router.push(`/video/${item.id}`)"
              >
                <span>{{ item.title }}</span>
                <small>{{ formatCount(item.viewCount) }} 播放</small>
              </button>
            </div>
          </template>
        </div>

        <div class="soft-panel recommend-panel">
          <h3>相关推荐</h3>
          <div class="recommend-list">
            <article
              v-for="item in recommendedVideos"
              :key="item.id"
              class="recommend-item"
              @click="openRecommendedVideo(item)"
            >
              <div class="recommend-cover">
                <img v-if="item.coverUrl" :src="mediaUrl(item.coverUrl)" :alt="item.title">
                <span v-else>Danmaku</span>
              </div>
              <div class="recommend-body">
                <strong>{{ item.title }}</strong>
                <span>{{ formatCount(item.viewCount) }} 播放 · {{ formatCount(item.danmakuCount) }} 弹幕</span>
              </div>
            </article>
            <el-empty v-if="!recommendedVideos.length" description="暂无推荐" />
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
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import VideoPlayer from '@/components/common/VideoPlayer.vue'
import CommentItem from '@/components/common/CommentItem.vue'
import { collectionApi } from '@/api/collection'
import { commentApi } from '@/api/comment'
import { danmakuApi } from '@/api/danmaku'
import { userApi } from '@/api/user'
import { videoApi } from '@/api/video'
import { useAuthStore } from '@/store/auth'
import { useCommentStore } from '@/store/comment'
import { useVideoStore } from '@/store/video'
import type { Comment, Danmaku, UserInfo, VideoCollectionInfo, VideoInfo } from '@/types'
import { formatCount, formatTime, mediaUrl, normalizeTags } from '@/utils/format'
import { removeUserLibraryRecord, upsertUserLibraryRecord } from '@/utils/userLibrary'

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
const normalDanmakuText = ref('')
const advancedDanmakuText = ref('')
const sideMode = ref<'normal' | 'advanced' | 'collection'>('normal')
const danmakuType = ref<'scroll' | 'top' | 'bottom'>('scroll')
const danmakuFontSize = ref<'small' | 'medium' | 'large'>('medium')
const danmakuVisible = ref(true)
const danmakuOpacity = ref(85)
const DANMAKU_COLORS = [
  '#000000', '#FFFFFF', '#FF5555', '#55FF55', '#5555FF', '#FFFF55',
  '#FF55FF', '#55FFFF', '#FF8C00', '#FF69B4', '#00CED1', '#FFD700', '#FF6347',
]
const danmakuColor = ref(DANMAKU_COLORS[0])
const commentText = ref('')
const commentsCollapsed = ref(false)
const downloading = ref(false)
const uploadingAdvanced = ref(false)
const recommendedVideos = ref<VideoInfo[]>([])
const videoCollections = ref<VideoCollectionInfo[]>([])
const myCollections = ref<VideoCollectionInfo[]>([])
const selectedCollectionDetail = ref<VideoCollectionInfo | null>(null)
const selectedCollectionId = ref<number>()
const newCollectionTitle = ref('')
const collaboratorKeyword = ref('')
const collaboratorOptions = ref<UserInfo[]>([])
const video = computed(() => videoStore.currentVideo)
const canManageVideo = computed(() => Boolean(
  authStore.userInfo &&
  video.value &&
  authStore.userInfo.id === video.value.author?.id
))
let touchStartY = 0
const SWIPE_THRESHOLD = 50

onMounted(load)
onUnmounted(() => {
  saveHistory()
  videoStore.clearCurrent()
  commentStore.clearComments()
})

watch(() => route.params.id, () => {
  videoStore.clearCurrent()
  commentStore.clearComments()
  danmakus.value = []
  recommendedVideos.value = []
  videoCollections.value = []
  selectedCollectionDetail.value = null
  load()
})

function handleTouchStart(e: TouchEvent) {
  touchStartY = e.touches[0].clientY
}

function handleTouchEnd(e: TouchEvent) {
  const diff = touchStartY - e.changedTouches[0].clientY
  if (diff > SWIPE_THRESHOLD) goToNextVideo()
}

function goToNextVideo() {
  if (!video.value) return
  router.push(`/video/${video.value.id + 1}`)
}

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
    await Promise.all([loadRecommendations(), loadVideoCollections(), loadMyCollections()])
    saveHistory()
  } finally {
    loading.value = false
  }
}

async function loadRecommendations() {
  if (!video.value) return
  try {
    const tags = normalizeTags(video.value.tags)
    const res = await videoApi.list({
      page: 1,
      pageSize: 8,
      tag: tags[0] || undefined,
      category: video.value.category || undefined,
      sort: 'hot',
    })
    recommendedVideos.value = res.data.list.filter((item) => item.id !== video.value?.id).slice(0, 6)
  } catch {
    recommendedVideos.value = []
  }
}

async function loadVideoCollections() {
  if (!video.value) return
  try {
    const res = await collectionApi.videoCollections(video.value.id)
    videoCollections.value = res.data
  } catch {
    videoCollections.value = []
  }
}

async function loadMyCollections() {
  if (!authStore.isLoggedIn) return
  try {
    const res = await collectionApi.mine()
    myCollections.value = res.data
  } catch {
    myCollections.value = []
  }
}

function openRecommendedVideo(item: VideoInfo) {
  router.push(`/video/${item.id}`)
}

async function openCollection(id: number) {
  const res = await collectionApi.detail(id)
  selectedCollectionDetail.value = res.data
}

async function createCollectionAndAddVideo() {
  if (!ensureLogin() || !video.value) return
  const title = newCollectionTitle.value.trim()
  if (!title) {
    ElMessage.warning('请先输入合集标题')
    return
  }
  const created = await collectionApi.create({ title })
  await collectionApi.addVideo(created.data.id, video.value.id)
  newCollectionTitle.value = ''
  ElMessage.success('已创建合集并加入当前视频')
  await Promise.all([loadVideoCollections(), loadMyCollections()])
}

async function addVideoToSelectedCollection() {
  if (!ensureLogin() || !video.value || !selectedCollectionId.value) return
  await collectionApi.addVideo(selectedCollectionId.value, video.value.id)
  ElMessage.success('已加入合集')
  await loadVideoCollections()
}

async function removeVideoFromCollection(collectionId: number) {
  if (!ensureLogin() || !video.value) return
  await collectionApi.removeVideo(collectionId, video.value.id)
  ElMessage.success('已从合集移出')
  await loadVideoCollections()
}

function canManageCollection(collection: VideoCollectionInfo) {
  return Boolean(authStore.userInfo && collection.owner?.id === authStore.userInfo.id)
}

async function searchCollaborators() {
  const q = collaboratorKeyword.value.trim()
  if (!q) return
  const res = await userApi.search({ q, page: 1, pageSize: 8 })
  collaboratorOptions.value = res.data.list.filter((user) => user.id !== authStore.userInfo?.id)
}

async function inviteCollaborator(userId: number) {
  if (!ensureLogin() || !video.value) return
  await collectionApi.addCollaborator(video.value.id, userId)
  ElMessage.success('已添加共创')
}

async function toggleLike() {
  if (!ensureLogin() || !video.value) return
  const current = video.value
  const res = await videoApi.like(current.id)
  if (res.data.liked) {
    current.likeCount += 1
    upsertUserLibraryRecord('liked', current)
  } else {
    current.likeCount = Math.max(0, current.likeCount - 1)
    removeUserLibraryRecord('liked', current.id)
  }
}

async function toggleCollect() {
  if (!ensureLogin() || !video.value) return
  const current = video.value
  const res = await videoApi.collect(current.id)
  if (res.data.collected) {
    current.collectCount += 1
    upsertUserLibraryRecord('collections', current)
  } else {
    current.collectCount = Math.max(0, current.collectCount - 1)
    removeUserLibraryRecord('collections', current.id)
  }
}

async function downloadVideo() {
  if (!ensureLogin() || !video.value) return
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
  advancedDanmakuText.value = '@adv x=10 y=20 tx=80 ty=20 dur=4 size=24 color=#FFFFFF | hello'
}

function chooseDanmakuFile() {
  if (!ensureLogin()) return
  danmakuFileInput.value?.click()
}

async function uploadAdvancedFile(event: Event) {
  if (!ensureLogin() || !video.value) return
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

async function sendNormalDanmaku() {
  await sendDanmaku(normalDanmakuText.value, danmakuType.value, () => {
    normalDanmakuText.value = ''
  })
}

async function sendAdvancedDanmaku() {
  await sendDanmaku(advancedDanmakuText.value, 'advanced', () => {
    advancedDanmakuText.value = ''
  })
}

async function sendDanmaku(contentValue: string, type: 'scroll' | 'top' | 'bottom' | 'advanced', afterSend: () => void) {
  if (!ensureLogin() || !video.value) return
  const content = contentValue.trim()
  if (!content) return

  const res = await danmakuApi.send({
    videoId: video.value.id,
    content,
    time: Math.floor(playerRef.value?.getCurrentTime() || currentTime.value),
    color: type === 'advanced' ? '#FFFFFF' : danmakuColor.value,
    fontSize: danmakuFontSize.value,
    type,
  })
  danmakus.value.push(res.data)
  video.value.danmakuCount += 1
  afterSend()
}

async function submitComment() {
  if (!ensureLogin() || !video.value || !commentText.value.trim()) return
  await commentStore.createComment(video.value.id, commentText.value.trim())
  commentText.value = ''
}

async function submitReply(target: Comment, content: string) {
  if (!ensureLogin() || !video.value || !content.trim()) return
  await commentStore.createComment(video.value.id, content.trim(), target.id)
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

function categoryLabel(value: string) {
  const map: Record<string, string> = {
    game: '游戏',
    tech: '科技',
    life: '生活',
    music: '音乐',
    anime: '动漫',
    knowledge: '知识',
  }
  return map[value] || value
}
</script>

<style scoped>
.detail-page {
  display: grid;
  min-height: 100vh;
}

.watch-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 380px;
  gap: 18px;
  align-items: start;
}

.main-col,
.side-col,
.comment-body,
.recommend-panel,
.side-panel,
.collection-list,
.collection-detail,
.collection-editor {
  display: grid;
  gap: 16px;
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
.danmaku-options,
.danmaku-colors,
.comment-head,
.collaborator-search {
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

.author-panel,
.collaborator-panel,
.side-panel,
.recommend-panel,
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
  font-size: 13px;
}

.side-tabs {
  width: 100%;
}

.side-tabs :deep(.el-radio-button) {
  flex: 1;
}

.side-tabs :deep(.el-radio-button__inner) {
  width: 100%;
}

.normal-danmaku-input :deep(.el-input__wrapper) {
  min-height: 42px;
}

.danmaku-options .el-select {
  flex: 1;
  min-width: 120px;
}

.danmaku-control-row {
  display: grid;
  grid-template-columns: 68px minmax(0, 1fr);
  align-items: center;
  gap: 12px;
  color: #667085;
  font-size: 13px;
}

.send-button {
  width: 96px;
  justify-self: end;
}

.color-dot {
  width: 22px;
  height: 22px;
  border: 2px solid transparent;
  border-radius: 50%;
  cursor: pointer;
}

.color-dot.active {
  border-color: #165dff;
}

.advanced-help,
.advanced-upload {
  display: grid;
  gap: 8px;
  color: #667085;
  font-size: 12px;
  line-height: 1.5;
}

.advanced-upload {
  padding: 10px;
  border: 1px dashed #d0d5dd;
  border-radius: 8px;
  background: #f8fafc;
}

.file-input {
  display: none;
}

.collection-card,
.collaborator-item,
.recommend-item {
  cursor: pointer;
}

.collection-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  gap: 8px;
  align-items: center;
  padding: 10px;
  border: 1px solid #edf0f3;
  border-radius: 8px;
}

.collection-card strong,
.collection-card span {
  display: block;
}

.collection-card span {
  color: #667085;
  font-size: 12px;
}

.collection-detail {
  gap: 8px;
  padding-top: 4px;
}

.collection-video {
  display: grid;
  gap: 4px;
  padding: 9px 10px;
  border: 1px solid #edf0f3;
  border-radius: 8px;
  background: #fff;
  color: #18191c;
  text-align: left;
  cursor: pointer;
}

.collection-video small {
  color: #667085;
}

.collaborator-panel {
  display: grid;
  gap: 12px;
}

.collaborator-panel h3,
.side-panel h3,
.recommend-panel h3,
.comment-head h2 {
  margin: 0;
}

.collaborator-options {
  display: grid;
  gap: 8px;
}

.collaborator-item {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  padding: 9px 10px;
  border: 1px solid #edf0f3;
  border-radius: 8px;
  background: #fff;
}

.collaborator-item small {
  color: #667085;
}

.recommend-list {
  display: grid;
  gap: 12px;
}

.recommend-item {
  display: grid;
  grid-template-columns: 132px minmax(0, 1fr);
  gap: 10px;
}

.recommend-cover {
  display: grid;
  overflow: hidden;
  aspect-ratio: 16 / 9;
  place-items: center;
  border-radius: 8px;
  background: #f1f2f3;
  color: #00aeec;
  font-size: 12px;
  font-weight: 900;
}

.recommend-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.recommend-body {
  display: grid;
  align-content: start;
  gap: 6px;
  min-width: 0;
}

.recommend-body strong {
  overflow: hidden;
  color: #18191c;
  font-size: 14px;
  line-height: 1.4;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recommend-body span {
  color: #9499a0;
  font-size: 12px;
}

.comment-head {
  align-items: center;
  justify-content: space-between;
}

.comment-input {
  display: grid;
  justify-items: start;
  gap: 10px;
}

@media (max-width: 920px) {
  .watch-grid {
    grid-template-columns: 1fr;
  }
}
</style>
