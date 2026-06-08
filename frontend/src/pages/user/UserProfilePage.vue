<template>
  <main class="page-shell profile-page" v-loading="loading">
    <section v-if="user" class="profile-hero soft-panel">
      <button
        class="avatar-editor"
        type="button"
        :disabled="!isSelf"
        @click="triggerAvatarInput"
      >
        <img v-if="avatarSrc" :src="avatarSrc" :alt="user.nickname" />
        <strong v-else>{{ user.nickname?.slice(0, 1) || 'U' }}</strong>
        <span v-if="isSelf">更换头像</span>
      </button>
      <input
        ref="avatarInputRef"
        class="hidden-file"
        type="file"
        accept="image/*"
        @change="uploadAvatar"
      />
      <div>
        <el-tag>{{ user.role }}</el-tag>
        <h1>{{ user.nickname }}</h1>
        <div v-if="isEditingBio" class="bio-editor">
          <el-input
            v-model="bioDraft"
            type="textarea"
            :rows="3"
            maxlength="160"
            show-word-limit
            placeholder="写一段个人简介"
          />
          <div class="bio-actions">
            <el-button size="small" @click="cancelBioEdit">取消</el-button>
            <el-button size="small" type="primary" :loading="savingBio" @click="saveBio">保存</el-button>
          </div>
        </div>
        <button
          v-else
          class="bio-display"
          type="button"
          :disabled="!isSelf"
          @click="startBioEdit"
        >
          <span>{{ user.bio || '这个用户还没有填写简介。' }}</span>
          <em v-if="isSelf">编辑简介</em>
        </button>
        <div class="stats">
          <span>{{ user.followCount }} 关注</span>
          <span>{{ user.fanCount }} 粉丝</span>
          <span>{{ user.videoCount || 0 }} 视频</span>
        </div>
      </div>
      <el-button v-if="!isSelf" type="primary" @click="follow">{{ user.followed ? '已关注' : '关注' }}</el-button>
    </section>

    <section>
      <div class="section-head">
        <h2>公开视频</h2>
      </div>
      <div v-if="videos.length" class="video-grid">
        <VideoCard v-for="video in videos" :key="video.id" :video="video" @open="router.push(`/video/${video.id}`)" />
      </div>
      <div v-else class="soft-panel empty-panel">
        <el-empty description="暂无公开视频" />
      </div>
    </section>

    <el-dialog v-model="cropVisible" title="选择头像显示区域" width="520px" @closed="resetCrop">
      <div class="cropper">
        <div class="crop-stage" @mousedown="startCropDrag" @touchstart.prevent="startCropTouchDrag">
          <img
            v-if="cropImageUrl"
            :src="cropImageUrl"
            alt="头像裁剪预览"
            :style="cropImageStyle"
            draggable="false"
          />
          <div class="crop-frame" />
        </div>
        <div class="crop-controls">
          <span>缩放</span>
          <el-slider v-model="cropScale" :min="1" :max="3" :step="0.01" />
        </div>
      </div>
      <template #footer>
        <el-button @click="cropVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploadingAvatar" @click="confirmAvatarCrop">保存头像</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import VideoCard from '@/components/common/VideoCard.vue'
import { userApi } from '@/api/user'
import { authApi } from '@/api/auth'
import { videoApi } from '@/api/video'
import { useAuthStore } from '@/store/auth'
import type { UserInfo, VideoInfo } from '@/types'
import { mediaUrl } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const savingBio = ref(false)
const uploadingAvatar = ref(false)
const isEditingBio = ref(false)
const bioDraft = ref('')
const avatarInputRef = ref<HTMLInputElement>()
const avatarVersion = ref(Date.now())
const cropVisible = ref(false)
const cropImageUrl = ref('')
const cropFileName = ref('avatar.png')
const cropScale = ref(1)
const cropOffsetX = ref(0)
const cropOffsetY = ref(0)
let cropDragStart: { x: number; y: number; offsetX: number; offsetY: number } | null = null
const user = ref<UserInfo | null>(null)
const videos = ref<VideoInfo[]>([])
const isSelf = computed(() => Boolean(user.value && authStore.userInfo?.id === user.value.id))
const avatarSrc = computed(() => {
  const url = mediaUrl(user.value?.avatar)
  if (!url) return ''
  return `${url}${url.includes('?') ? '&' : '?'}v=${avatarVersion.value}`
})
const cropImageStyle = computed(() => ({
  transform: `translate(${cropOffsetX.value}px, ${cropOffsetY.value}px) scale(${cropScale.value})`,
}))

onMounted(load)
onUnmounted(resetCrop)
watch(() => route.params.id, () => load())

async function load() {
  const id = Number(route.params.id)
  loading.value = true
  try {
    const [profileRes, videosRes] = await Promise.all([
      userApi.profile(id),
      videoApi.userVideos(id, { page: 1, pageSize: 20 }),
    ])
    user.value = profileRes.data
    videos.value = videosRes.data.list
  } finally {
    loading.value = false
  }
}

async function follow() {
  if (!authStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  if (!user.value) return
  const res = await userApi.follow(user.value.id)
  user.value.followed = res.data.followed
  user.value.fanCount += res.data.followed ? 1 : -1
}

function triggerAvatarInput() {
  if (!isSelf.value || uploadingAvatar.value) return
  avatarInputRef.value?.click()
}

function uploadAvatar(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file || !user.value) return

  resetCrop()
  cropFileName.value = file.name || 'avatar.png'
  cropImageUrl.value = URL.createObjectURL(file)
  cropVisible.value = true
}

async function confirmAvatarCrop() {
  if (!cropImageUrl.value || !user.value) return
  const blob = await cropImageToBlob()
  if (!blob) {
    ElMessage.error('头像裁剪失败')
    return
  }

  const formData = new FormData()
  formData.append('avatar', blob, cropFileName.value.replace(/\.[^.]+$/, '.png'))
  uploadingAvatar.value = true
  try {
    const res = await authApi.uploadAvatar(formData)
    user.value.avatar = res.data.avatar
    avatarVersion.value = Date.now()
    syncAuthUser({ avatar: res.data.avatar })
    cropVisible.value = false
    ElMessage.success('头像已更新')
  } catch {
    ElMessage.error('头像上传失败')
  } finally {
    uploadingAvatar.value = false
  }
}

function startCropDrag(event: MouseEvent) {
  cropDragStart = {
    x: event.clientX,
    y: event.clientY,
    offsetX: cropOffsetX.value,
    offsetY: cropOffsetY.value,
  }
  window.addEventListener('mousemove', moveCropDrag)
  window.addEventListener('mouseup', stopCropDrag)
}

function startCropTouchDrag(event: TouchEvent) {
  const touch = event.touches[0]
  cropDragStart = {
    x: touch.clientX,
    y: touch.clientY,
    offsetX: cropOffsetX.value,
    offsetY: cropOffsetY.value,
  }
  window.addEventListener('touchmove', moveCropTouchDrag, { passive: false })
  window.addEventListener('touchend', stopCropDrag)
}

function moveCropDrag(event: MouseEvent) {
  if (!cropDragStart) return
  cropOffsetX.value = cropDragStart.offsetX + event.clientX - cropDragStart.x
  cropOffsetY.value = cropDragStart.offsetY + event.clientY - cropDragStart.y
}

function moveCropTouchDrag(event: TouchEvent) {
  if (!cropDragStart) return
  event.preventDefault()
  const touch = event.touches[0]
  cropOffsetX.value = cropDragStart.offsetX + touch.clientX - cropDragStart.x
  cropOffsetY.value = cropDragStart.offsetY + touch.clientY - cropDragStart.y
}

function stopCropDrag() {
  cropDragStart = null
  window.removeEventListener('mousemove', moveCropDrag)
  window.removeEventListener('mouseup', stopCropDrag)
  window.removeEventListener('touchmove', moveCropTouchDrag)
  window.removeEventListener('touchend', stopCropDrag)
}

function cropImageToBlob() {
  return new Promise<Blob | null>((resolve) => {
    const image = new Image()
    image.onload = () => {
      const size = 512
      const canvas = document.createElement('canvas')
      canvas.width = size
      canvas.height = size
      const ctx = canvas.getContext('2d')
      if (!ctx) {
        resolve(null)
        return
      }

      const stageSize = 320
      const scale = Math.max(stageSize / image.width, stageSize / image.height) * cropScale.value
      const drawWidth = image.width * scale
      const drawHeight = image.height * scale
      const drawX = (stageSize - drawWidth) / 2 + cropOffsetX.value
      const drawY = (stageSize - drawHeight) / 2 + cropOffsetY.value
      const outputScale = size / stageSize

      ctx.fillStyle = '#fff'
      ctx.fillRect(0, 0, size, size)
      ctx.drawImage(
        image,
        drawX * outputScale,
        drawY * outputScale,
        drawWidth * outputScale,
        drawHeight * outputScale,
      )
      canvas.toBlob(resolve, 'image/png', 0.92)
    }
    image.onerror = () => resolve(null)
    image.src = cropImageUrl.value
  })
}

function resetCrop() {
  stopCropDrag()
  if (cropImageUrl.value) URL.revokeObjectURL(cropImageUrl.value)
  cropImageUrl.value = ''
  cropScale.value = 1
  cropOffsetX.value = 0
  cropOffsetY.value = 0
}

function startBioEdit() {
  if (!isSelf.value || !user.value) return
  bioDraft.value = user.value.bio || ''
  isEditingBio.value = true
}

function cancelBioEdit() {
  isEditingBio.value = false
  bioDraft.value = ''
}

async function saveBio() {
  if (!user.value) return
  savingBio.value = true
  try {
    const nextBio = bioDraft.value.trim()
    await authApi.updateMe({ bio: nextBio })
    user.value.bio = nextBio
    syncAuthUser({ bio: nextBio })
    isEditingBio.value = false
    ElMessage.success('简介已更新')
  } catch {
    ElMessage.error('简介保存失败')
  } finally {
    savingBio.value = false
  }
}

function syncAuthUser(patch: Partial<UserInfo>) {
  if (!authStore.userInfo) return
  authStore.userInfo = { ...authStore.userInfo, ...patch }
  localStorage.setItem('userInfo', JSON.stringify(authStore.userInfo))
}
</script>

<style scoped>
.profile-page {
  display: grid;
  gap: 26px;
}

.profile-hero {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 22px;
  padding: 28px;
}

.avatar-editor {
  position: relative;
  display: grid;
  overflow: hidden;
  width: 78px;
  height: 78px;
  place-items: center;
  padding: 0;
  border: 1px solid #f1f2f3;
  border-radius: 50%;
  background:
    linear-gradient(135deg, rgba(251, 114, 153, 0.12), rgba(0, 174, 236, 0.1)),
    #f6f7f8;
  color: #fb7299;
  cursor: pointer;
}

.avatar-editor:disabled {
  cursor: default;
}

.avatar-editor img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-editor strong {
  font-size: 28px;
  font-weight: 900;
}

.avatar-editor span {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.52);
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  opacity: 0;
  transition: opacity 0.16s ease;
}

.avatar-editor:not(:disabled):hover span {
  opacity: 1;
}

.hidden-file {
  display: none;
}

.cropper {
  display: grid;
  gap: 16px;
}

.crop-stage {
  position: relative;
  overflow: hidden;
  width: 320px;
  height: 320px;
  margin: 0 auto;
  border-radius: 10px;
  background: #f1f2f3;
  cursor: move;
  user-select: none;
}

.crop-stage img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  transform-origin: center center;
}

.crop-frame {
  position: absolute;
  inset: 0;
  border: 2px solid #fb7299;
  box-shadow: inset 0 0 0 999px rgba(0, 0, 0, 0.16);
  pointer-events: none;
}

.crop-controls {
  display: grid;
  grid-template-columns: 42px minmax(0, 1fr);
  align-items: center;
  gap: 12px;
}

.crop-controls span {
  color: #61666d;
  font-size: 13px;
  font-weight: 800;
}

h1 {
  margin: 10px 0 8px;
  color: #142033;
}

.bio-display {
  position: relative;
  display: inline-grid;
  max-width: 100%;
  margin: 0 0 12px;
  padding: 0;
  border: 0;
  background: transparent;
  color: #667085;
  cursor: pointer;
  text-align: left;
}

.bio-display:disabled {
  cursor: default;
}

.bio-display span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bio-display em {
  position: absolute;
  right: -66px;
  top: 50%;
  color: #fb7299;
  font-size: 12px;
  font-style: normal;
  font-weight: 800;
  opacity: 0;
  transform: translateY(-50%);
  transition: opacity 0.16s ease;
}

.bio-display:not(:disabled):hover em {
  opacity: 1;
}

.bio-editor {
  display: grid;
  max-width: 560px;
  gap: 8px;
  margin: 0 0 12px;
}

.bio-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.stats {
  display: flex;
  gap: 14px;
  color: #475467;
}

.video-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 18px;
}

@media (max-width: 820px) {
  .profile-hero,
  .video-grid {
    grid-template-columns: 1fr;
  }
}
</style>
