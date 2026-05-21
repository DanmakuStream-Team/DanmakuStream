<template>
  <div class="max-w-3xl mx-auto">
    <el-card class="rounded-2xl" shadow="never">
      <template #header>
        <span class="text-lg font-semibold">上传视频</span>
      </template>

      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" @submit.prevent="() => formRef?.validate(handleValidated)">

        <!-- Video upload -->
        <el-form-item label="视频文件" prop="videoFile">
          <div
            class="w-full min-h-44 border-2 border-dashed rounded-xl flex flex-col items-center justify-center gap-2 cursor-pointer transition-colors"
            :class="isVideoDragover ? 'border-blue-500 bg-blue-50' : 'border-gray-300 bg-gray-50 hover:border-blue-400 hover:bg-blue-50'"
            @dragover.prevent="isVideoDragover = true"
            @dragleave.prevent="isVideoDragover = false"
            @drop.prevent="handleVideoDrop"
            @click="videoInputRef?.click()"
          >
            <input ref="videoInputRef" type="file" accept="video/*" class="hidden" @change="handleVideoChange" />
            <template v-if="form.videoFile">
              <el-icon :size="36" class="text-blue-500"><Film /></el-icon>
              <div class="font-semibold text-gray-800 max-w-xs truncate">{{ form.videoFile.name }}</div>
              <div class="text-sm text-gray-500">{{ formatFileSize(form.videoFile.size) }}</div>
              <el-button size="small" @click.stop="clearVideoFile">重新选择</el-button>
            </template>
            <template v-else>
              <el-icon :size="40" class="text-gray-400"><VideoCamera /></el-icon>
              <div class="font-semibold text-gray-700">点击或拖拽视频到这里上传</div>
              <div class="text-sm text-gray-400">支持 mp4 / webm / mov 等常见格式</div>
            </template>
          </div>
        </el-form-item>

        <!-- Cover upload -->
        <el-form-item label="封面图片" prop="coverFile">
          <div class="flex items-center gap-4">
            <div
              class="w-72 h-44 border-2 border-dashed rounded-xl overflow-hidden flex flex-col items-center justify-center cursor-pointer transition-colors flex-shrink-0"
              :class="isCoverDragover ? 'border-blue-500 bg-blue-50' : 'border-gray-300 bg-gray-50 hover:border-blue-400'"
              @dragover.prevent="isCoverDragover = true"
              @dragleave.prevent="isCoverDragover = false"
              @drop.prevent="handleCoverDrop"
              @click="coverInputRef?.click()"
            >
              <input ref="coverInputRef" type="file" accept="image/*" class="hidden" @change="handleCoverChange" />
              <img v-if="coverPreview" :src="coverPreview" alt="cover" class="w-full h-full object-cover" />
              <template v-else>
                <el-icon :size="36" class="text-gray-400"><Picture /></el-icon>
                <div class="font-semibold text-gray-700 mt-1">点击或拖拽封面图片</div>
                <div class="text-sm text-gray-400">建议 16:9 比例</div>
              </template>
            </div>
            <div v-if="form.coverFile" class="text-sm text-gray-600 space-y-1">
              <div class="font-medium">{{ form.coverFile.name }}</div>
              <div class="text-gray-400">{{ formatFileSize(form.coverFile.size) }}</div>
              <el-button size="small" @click="clearCoverFile">移除封面</el-button>
            </div>
          </div>
        </el-form-item>

        <!-- Title -->
        <el-form-item label="视频标题" prop="title">
          <el-input v-model="form.title" placeholder="请输入视频标题" :maxlength="80" show-word-limit clearable />
        </el-form-item>

        <!-- Description -->
        <el-form-item label="视频简介" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            placeholder="请输入视频简介"
            :maxlength="1000"
            show-word-limit
            :autosize="{ minRows: 4, maxRows: 8 }"
          />
        </el-form-item>

        <!-- Tags -->
        <el-form-item label="视频标签" prop="tags">
          <div class="w-full space-y-2">
            <div class="flex flex-wrap gap-2">
              <el-tag
                v-for="tag in form.tags"
                :key="tag"
                closable
                @close="removeTag(tag)"
              >{{ tag }}</el-tag>
              <el-input
                v-if="form.tags.length < 6"
                v-model="tagInput"
                size="small"
                class="w-24"
                placeholder="+ 添加"
                @keydown.enter.prevent="addTag"
                @blur="addTag"
              />
            </div>
            <div class="text-xs text-gray-400">最多 6 个标签，每个不超过 12 字符，回车确认</div>
          </div>
        </el-form-item>

        <!-- Progress -->
        <div v-if="uploading || uploadProgress > 0" class="mb-4 p-4 bg-gray-50 rounded-xl">
          <div class="flex justify-between text-sm text-gray-600 mb-2">
            <span>上传进度</span>
            <span>{{ uploadProgress }}%</span>
          </div>
          <el-progress :percentage="uploadProgress" :show-text="false" />
        </div>

        <!-- Actions -->
        <div class="flex justify-end gap-3 pt-2">
          <el-button :disabled="uploading" @click="handleReset">重置</el-button>
          <el-button type="primary" native-type="submit" :loading="uploading">
            {{ uploading ? '上传中...' : '提交上传' }}
          </el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onUnmounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
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

type UploadResponse = ApiResponse<UploadVideoResponse> | UploadVideoResponse

const router = useRouter()

const formRef = ref<FormInstance>()
const videoInputRef = ref<HTMLInputElement>()
const coverInputRef = ref<HTMLInputElement>()

const uploading = ref(false)
const uploadProgress = ref(0)
const isVideoDragover = ref(false)
const isCoverDragover = ref(false)
const coverPreview = ref('')
const tagInput = ref('')

const form = reactive<UploadForm>({
  videoFile: null,
  coverFile: null,
  title: '',
  description: '',
  tags: [],
})

const rules: FormRules = {
  title: [
    { required: true, message: '请输入视频标题', trigger: 'blur' },
    { min: 2, message: '标题至少需要 2 个字符', trigger: 'blur' },
    { max: 80, message: '标题不能超过 80 个字符', trigger: 'blur' },
  ],
  description: [
    { required: true, message: '请输入视频简介', trigger: 'blur' },
    { max: 1000, message: '简介不能超过 1000 个字符', trigger: 'blur' },
  ],
}

const normalizedTags = computed(() =>
  form.tags.map(tag => tag.trim()).filter(Boolean).slice(0, 6)
)

onUnmounted(() => { revokeCoverPreview() })

function addTag() {
  const tag = tagInput.value.trim()
  if (!tag) return
  if (form.tags.length >= 6) { ElMessage.warning('最多 6 个标签'); tagInput.value = ''; return }
  if (tag.length > 12) { ElMessage.warning('标签不能超过 12 个字符'); return }
  if (!form.tags.includes(tag)) form.tags.push(tag)
  tagInput.value = ''
}

function removeTag(tag: string) {
  form.tags = form.tags.filter(t => t !== tag)
}

function handleVideoChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) setVideoFile(file)
  target.value = ''
}

function handleCoverChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) setCoverFile(file)
  target.value = ''
}

function handleVideoDrop(event: DragEvent) {
  isVideoDragover.value = false
  const file = event.dataTransfer?.files?.[0]
  if (file) setVideoFile(file)
}

function handleCoverDrop(event: DragEvent) {
  isCoverDragover.value = false
  const file = event.dataTransfer?.files?.[0]
  if (file) setCoverFile(file)
}

function setVideoFile(file: File) {
  if (!file.type.startsWith('video/')) { ElMessage.error('请选择视频文件'); return }
  form.videoFile = file
  if (!form.title) form.title = getFileNameWithoutExt(file.name)
}

function setCoverFile(file: File) {
  if (!file.type.startsWith('image/')) { ElMessage.error('请选择图片文件'); return }
  form.coverFile = file
  revokeCoverPreview()
  coverPreview.value = URL.createObjectURL(file)
}

function clearVideoFile() { form.videoFile = null }
function clearCoverFile() { form.coverFile = null; revokeCoverPreview() }
function revokeCoverPreview() {
  if (coverPreview.value) { URL.revokeObjectURL(coverPreview.value); coverPreview.value = '' }
}

function getFileNameWithoutExt(fileName: string) {
  const index = fileName.lastIndexOf('.')
  return index > 0 ? fileName.slice(0, index) : fileName
}

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`
  return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`
}

async function handleValidated(valid: boolean) {
  if (!valid) { ElMessage.warning('请先完善上传信息'); return }
  if (!form.videoFile) { ElMessage.warning('请选择视频文件'); return }
  if (!form.coverFile) { ElMessage.warning('请选择封面图片'); return }
  if (form.tags.length === 0) { ElMessage.warning('请至少填写一个标签'); return }

  const formData = new FormData()
  formData.append('video', form.videoFile)
  formData.append('cover', form.coverFile)
  formData.append('title', form.title.trim())
  formData.append('description', form.description.trim())
  formData.append('tags', normalizedTags.value.join(','))

  uploading.value = true
  uploadProgress.value = 5

  const progressTimer = window.setInterval(() => {
    if (uploadProgress.value < 90) uploadProgress.value += 5
  }, 300)

  try {
    const res = await videoApi.uploadVideo(formData)
    window.clearInterval(progressTimer)
    uploadProgress.value = 100
    ElMessage.success('视频上传成功，等待审核')
    const data = normalizeUploadResponse(res.data)
    if (data?.videoId) await router.push(`/video/${data.videoId}`)
    else await router.push('/')
  } catch {
    window.clearInterval(progressTimer)
    uploadProgress.value = 0
    ElMessage.error('视频上传失败，请稍后重试')
  } finally {
    uploading.value = false
  }
}

function normalizeUploadResponse(res: UploadResponse) {
  if ('data' in res) return res.data
  return res
}

function handleReset() {
  if (uploading.value) return
  form.videoFile = null
  form.coverFile = null
  form.title = ''
  form.description = ''
  form.tags = []
  uploadProgress.value = 0
  revokeCoverPreview()
}
</script>
