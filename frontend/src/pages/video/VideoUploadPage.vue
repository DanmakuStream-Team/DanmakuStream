<template>
  <div class="upload-page">
    <a-card class="upload-card" title="上传视频">
      <a-form
        ref="formRef"
        :model="form"
        :rules="rules"
        layout="vertical"
        @submit="handleSubmit"
      >
        <!-- 视频上传 -->
        <a-form-item field="videoFile" label="视频文件">
          <div
            class="upload-drop-area"
            :class="{ 'is-dragover': isVideoDragover }"
            @dragover.prevent="isVideoDragover = true"
            @dragleave.prevent="isVideoDragover = false"
            @drop.prevent="handleVideoDrop"
            @click="videoInputRef?.click()"
          >
            <input
              ref="videoInputRef"
              class="hidden-input"
              type="file"
              accept="video/*"
              @change="handleVideoChange"
            />

            <template v-if="form.videoFile">
              <div class="file-name">
                {{ form.videoFile.name }}
              </div>
              <div class="file-tip">
                {{ formatFileSize(form.videoFile.size) }}
              </div>
              <a-button size="small" @click.stop="clearVideoFile">
                重新选择
              </a-button>
            </template>

            <template v-else>
              <div class="upload-icon">🎬</div>
              <div class="upload-title">点击或拖拽视频到这里上传</div>
              <div class="upload-tip">
                支持常见视频格式，例如 mp4 / webm / mov
              </div>
            </template>
          </div>
        </a-form-item>

        <!-- 封面上传 -->
        <a-form-item field="coverFile" label="封面图片">
          <div class="cover-upload-wrapper">
            <div
              class="cover-drop-area"
              :class="{ 'is-dragover': isCoverDragover }"
              @dragover.prevent="isCoverDragover = true"
              @dragleave.prevent="isCoverDragover = false"
              @drop.prevent="handleCoverDrop"
              @click="coverInputRef?.click()"
            >
              <input
                ref="coverInputRef"
                class="hidden-input"
                type="file"
                accept="image/*"
                @change="handleCoverChange"
              />

              <template v-if="coverPreview">
                <img
                  class="cover-preview"
                  :src="coverPreview"
                  alt="cover preview"
                />
              </template>

              <template v-else>
                <div class="upload-icon">🖼️</div>
                <div class="upload-title">点击或拖拽封面图片</div>
                <div class="upload-tip">建议使用 16:9 图片</div>
              </template>
            </div>

            <div v-if="form.coverFile" class="cover-info">
              <div>{{ form.coverFile.name }}</div>
              <div>{{ formatFileSize(form.coverFile.size) }}</div>
              <a-button size="small" @click="clearCoverFile">
                移除封面
              </a-button>
            </div>
          </div>
        </a-form-item>

        <!-- 标题 -->
        <a-form-item field="title" label="视频标题">
          <a-input
            v-model="form.title"
            placeholder="请输入视频标题"
            :max-length="80"
            show-word-limit
            allow-clear
          />
        </a-form-item>

        <!-- 简介 -->
        <a-form-item field="description" label="视频简介">
          <a-textarea
            v-model="form.description"
            placeholder="请输入视频简介"
            :max-length="1000"
            show-word-limit
            allow-clear
            :auto-size="{ minRows: 4, maxRows: 8 }"
          />
        </a-form-item>

        <!-- 标签 -->
        <a-form-item field="tags" label="视频标签">
          <a-input-tag
            v-model="form.tags"
            placeholder="输入标签后回车，例如：游戏、科技、生活"
            allow-clear
          />
          <div class="form-tip">
            最多填写 6 个标签，每个标签不超过 12 个字符。
          </div>
        </a-form-item>

        <!-- 上传进度 -->
        <div v-if="uploading || uploadProgress > 0" class="progress-box">
          <div class="progress-header">
            <span>上传进度</span>
            <span>{{ uploadProgress }}%</span>
          </div>
          <a-progress :percent="uploadProgress / 100" />
        </div>

        <!-- 操作按钮 -->
        <div class="actions">
          <a-button @click="handleReset" :disabled="uploading">
            重置
          </a-button>

          <a-button
            type="primary"
            html-type="submit"
            :loading="uploading"
          >
            {{ uploading ? '上传中...' : '提交上传' }}
          </a-button>
        </div>
      </a-form>
    </a-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onUnmounted, reactive, ref } from 'vue'
import { Message } from '@arco-design/web-vue'
import type { FieldRule, ValidatedError } from '@arco-design/web-vue'
import { useRouter } from 'vue-router'
import { videoApi } from '@/api/video'
import type { ApiResponse } from '@/types'

interface UploadForm {
  videoFile: File | null
  coverFile: File | null
  title: string
  description: string
  tags: string[]
}

interface UploadVideoResponse {
  videoId: number
  videoURL?: string
  videoUrl?: string
  coverURL?: string
  coverUrl?: string
}

type UploadResponse =
  | ApiResponse<UploadVideoResponse>
  | UploadVideoResponse

const router = useRouter()

const formRef = ref()
const videoInputRef = ref<HTMLInputElement>()
const coverInputRef = ref<HTMLInputElement>()

const uploading = ref(false)
const uploadProgress = ref(0)
const isVideoDragover = ref(false)
const isCoverDragover = ref(false)
const coverPreview = ref('')

const form = reactive<UploadForm>({
  videoFile: null,
  coverFile: null,
  title: '',
  description: '',
  tags: [],
})

const rules: Record<string, FieldRule[]> = {
  videoFile: [
    {
      validator: (_value, callback) => {
        if (!form.videoFile) {
          callback('请选择视频文件')
          return
        }

        if (!form.videoFile.type.startsWith('video/')) {
          callback('请选择正确的视频文件')
          return
        }

        callback()
      },
    },
  ],
  coverFile: [
    {
      validator: (_value, callback) => {
        if (!form.coverFile) {
          callback('请选择封面图片')
          return
        }

        if (!form.coverFile.type.startsWith('image/')) {
          callback('请选择正确的图片文件')
          return
        }

        callback()
      },
    },
  ],
  title: [
    { required: true, message: '请输入视频标题' },
    { minLength: 2, message: '标题至少需要 2 个字符' },
    { maxLength: 80, message: '标题不能超过 80 个字符' },
  ],
  description: [
    { required: true, message: '请输入视频简介' },
    { maxLength: 1000, message: '简介不能超过 1000 个字符' },
  ],
  tags: [
    {
      validator: (_value, callback) => {
        if (form.tags.length === 0) {
          callback('请至少填写一个标签')
          return
        }

        if (form.tags.length > 6) {
          callback('最多填写 6 个标签')
          return
        }

        const invalidTag = form.tags.find(tag => tag.trim().length > 12)

        if (invalidTag) {
          callback('每个标签不能超过 12 个字符')
          return
        }

        callback()
      },
    },
  ],
}

const normalizedTags = computed(() => {
  return form.tags
    .map(tag => tag.trim())
    .filter(Boolean)
    .slice(0, 6)
})

onUnmounted(() => {
  revokeCoverPreview()
})

function handleVideoChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  if (file) {
    setVideoFile(file)
  }

  target.value = ''
}

function handleCoverChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  if (file) {
    setCoverFile(file)
  }

  target.value = ''
}

function handleVideoDrop(event: DragEvent) {
  isVideoDragover.value = false

  const file = event.dataTransfer?.files?.[0]

  if (file) {
    setVideoFile(file)
  }
}

function handleCoverDrop(event: DragEvent) {
  isCoverDragover.value = false

  const file = event.dataTransfer?.files?.[0]

  if (file) {
    setCoverFile(file)
  }
}

function setVideoFile(file: File) {
  if (!file.type.startsWith('video/')) {
    Message.error('请选择视频文件')
    return
  }

  form.videoFile = file

  if (!form.title) {
    form.title = getFileNameWithoutExt(file.name)
  }
}

function setCoverFile(file: File) {
  if (!file.type.startsWith('image/')) {
    Message.error('请选择图片文件')
    return
  }

  form.coverFile = file
  revokeCoverPreview()
  coverPreview.value = URL.createObjectURL(file)
}

function clearVideoFile() {
  form.videoFile = null
}

function clearCoverFile() {
  form.coverFile = null
  revokeCoverPreview()
}

function revokeCoverPreview() {
  if (coverPreview.value) {
    URL.revokeObjectURL(coverPreview.value)
    coverPreview.value = ''
  }
}

function getFileNameWithoutExt(fileName: string) {
  const index = fileName.lastIndexOf('.')
  return index > 0 ? fileName.slice(0, index) : fileName
}

function formatFileSize(size: number) {
  if (size < 1024) {
    return `${size} B`
  }

  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`
  }

  if (size < 1024 * 1024 * 1024) {
    return `${(size / 1024 / 1024).toFixed(1)} MB`
  }

  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`
}

async function handleSubmit({
  errors,
}: {
  errors: Record<string, ValidatedError> | undefined
}) {
  if (errors) {
    Message.warning('请先完善上传信息')
    return
  }

  if (!form.videoFile || !form.coverFile) {
    Message.warning('请选择视频和封面')
    return
  }

  const formData = new FormData()
  formData.append('video', form.videoFile)
  formData.append('cover', form.coverFile)
  formData.append('title', form.title.trim())
  formData.append('description', form.description.trim())
  formData.append('tags', normalizedTags.value.join(','))

  uploading.value = true
  uploadProgress.value = 5

  const progressTimer = window.setInterval(() => {
    if (uploadProgress.value < 90) {
      uploadProgress.value += 5
    }
  }, 300)

  try {
    const res = await videoApi.uploadVideo(formData)

    window.clearInterval(progressTimer)
    uploadProgress.value = 100

    Message.success('视频上传成功，等待审核')

    const data = normalizeUploadResponse(res.data)

    if (data?.videoId) {
      await router.push(`/video/${data.videoId}`)
    } else {
      await router.push('/')
    }
  } catch {
    window.clearInterval(progressTimer)
    uploadProgress.value = 0
    Message.error('视频上传失败，请稍后重试')
  } finally {
    uploading.value = false
  }
}

function normalizeUploadResponse(res: UploadResponse) {
  if ('data' in res) {
    return res.data
  }

  return res
}

function handleReset() {
  if (uploading.value) {
    return
  }

  form.videoFile = null
  form.coverFile = null
  form.title = ''
  form.description = ''
  form.tags = []
  uploadProgress.value = 0
  revokeCoverPreview()
}
</script>

<style scoped>
.upload-page {
  max-width: 900px;
  margin: 0 auto;
}

.upload-card {
  border-radius: 12px;
}

.upload-drop-area,
.cover-drop-area {
  border: 1px dashed #c9cdd4;
  border-radius: 10px;
  background: #f7f8fa;
  cursor: pointer;
  transition: all 0.2s ease;
}

.upload-drop-area {
  min-height: 180px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.cover-drop-area {
  width: 320px;
  height: 180px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.upload-drop-area:hover,
.cover-drop-area:hover,
.upload-drop-area.is-dragover,
.cover-drop-area.is-dragover {
  border-color: #165dff;
  background: #eef4ff;
}

.hidden-input {
  display: none;
}

.upload-icon {
  font-size: 36px;
}

.upload-title {
  color: #1d2129;
  font-weight: 600;
}

.upload-tip,
.file-tip,
.form-tip {
  color: #86909c;
  font-size: 13px;
}

.file-name {
  color: #1d2129;
  font-weight: 600;
  max-width: 90%;
  word-break: break-all;
}

.cover-upload-wrapper {
  display: flex;
  align-items: center;
  gap: 16px;
}

.cover-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-info {
  color: #4e5969;
  font-size: 14px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.progress-box {
  margin: 16px 0;
  padding: 12px;
  background: #f7f8fa;
  border-radius: 8px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  color: #4e5969;
  font-size: 14px;
  margin-bottom: 8px;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
}

@media (max-width: 768px) {
  .cover-upload-wrapper {
    flex-direction: column;
    align-items: flex-start;
  }

  .cover-drop-area {
    width: 100%;
  }
}
</style>