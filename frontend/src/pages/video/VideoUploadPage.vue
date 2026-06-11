<template>
  <main class="page-shell upload-page">
    <div class="section-head">
      <div>
        <h1>上传视频</h1>
        <p class="muted">大文件上传完成后会进入后台转码，提交成功后可以离开当前页面。</p>
      </div>
    </div>

    <section class="upload-grid">
      <div class="soft-panel upload-panel">
        <el-upload
          drag
          :auto-upload="false"
          :limit="1"
          accept="video/*"
          :disabled="loading"
          :on-change="selectVideo"
          :on-remove="removeVideo"
        >
          <el-icon class="upload-icon"><UploadFilled /></el-icon>
          <div class="el-upload__text">拖拽视频到这里，或点击选择</div>
        </el-upload>

        <el-upload
          class="cover-upload"
          :auto-upload="false"
          :limit="1"
          accept="image/*"
          :disabled="loading"
          :on-change="selectCover"
          :on-remove="removeCover"
        >
          <el-button :disabled="loading">选择封面图</el-button>
        </el-upload>
      </div>

      <el-form class="soft-panel form-panel" label-position="top">
        <el-form-item label="标题">
          <el-input v-model="form.title" :disabled="loading" placeholder="请输入视频标题" />
        </el-form-item>

        <el-form-item label="简介">
          <el-input
            v-model="form.description"
            :disabled="loading"
            type="textarea"
            :rows="5"
            placeholder="介绍一下视频内容"
          />
        </el-form-item>

        <el-form-item label="分区">
          <el-select v-model="form.category" :disabled="loading" clearable placeholder="选择视频分区">
            <el-option v-for="item in categoryOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>

        <el-form-item label="标签">
          <el-select
            v-model="tagValues"
            :disabled="loading"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="选择或输入标签"
          >
            <el-option v-for="tag in existingTags" :key="tag" :label="tag" :value="tag" />
          </el-select>
        </el-form-item>

        <div class="collection-publish">
          <div class="collection-title">发布到合集</div>
          <el-select
            v-model="selectedCollectionId"
            :disabled="loading || !!newCollectionTitle.trim()"
            clearable
            placeholder="选择已有合集"
          >
            <el-option v-for="item in myCollections" :key="item.id" :label="item.title" :value="item.id" />
          </el-select>
          <el-input
            v-model="newCollectionTitle"
            :disabled="loading || !!selectedCollectionId"
            placeholder="或创建一个新合集"
          />
        </div>

        <div v-if="loading || progress > 0" class="progress-box">
          <el-progress :percentage="progress" />
          <span>{{ progress < 100 ? '正在上传文件...' : '上传完成，正在创建视频记录' }}</span>
        </div>

        <div class="upload-actions">
          <el-button type="primary" size="large" :loading="loading" :disabled="loading" @click="submit">
            提交上传
          </el-button>
          <el-button v-if="loading" type="danger" plain size="large" @click="cancelUpload">终止上传</el-button>
        </div>
      </el-form>
    </section>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, type UploadFile } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { collectionApi } from '@/api/collection'
import { videoApi } from '@/api/video'
import { useAuthStore } from '@/store/auth'
import type { VideoCollectionInfo } from '@/types'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const progress = ref(0)
const videoFile = ref<File | null>(null)
const coverFile = ref<File | null>(null)
const abortController = ref<AbortController | null>(null)
const form = reactive({ title: '', description: '', category: '' })
const tagValues = ref<string[]>([])
const existingTags = ref<string[]>([])
const myCollections = ref<VideoCollectionInfo[]>([])
const selectedCollectionId = ref<number>()
const newCollectionTitle = ref('')

const categoryOptions = [
  { label: '游戏', value: 'game' },
  { label: '科技', value: 'tech' },
  { label: '生活', value: 'life' },
  { label: '音乐', value: 'music' },
  { label: '动漫', value: 'anime' },
  { label: '知识', value: 'knowledge' },
]

onMounted(() => {
  loadExistingTags()
  loadMyCollections()
})

function selectVideo(file: UploadFile) {
  videoFile.value = file.raw || null
}

function removeVideo() {
  if (loading.value) return
  videoFile.value = null
}

function selectCover(file: UploadFile) {
  coverFile.value = file.raw || null
}

function removeCover() {
  if (loading.value) return
  coverFile.value = null
}

async function loadExistingTags() {
  try {
    const res = await videoApi.list({ page: 1, pageSize: 100, sort: 'date' })
    const tags = new Set<string>()
    res.data.list.forEach((video) => {
      normalizeTags(video.tags).forEach((tag) => tags.add(tag))
    })
    existingTags.value = Array.from(tags)
  } catch {
    existingTags.value = []
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

async function submit() {
  if (!authStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  if (!form.title.trim() || !videoFile.value) {
    ElMessage.warning('请填写标题并选择视频文件')
    return
  }

  const data = new FormData()
  data.append('title', form.title.trim())
  data.append('description', form.description.trim())
  data.append('category', form.category)
  data.append('tags', tagValues.value.map((tag) => tag.trim()).filter(Boolean).join(','))
  data.append('video', videoFile.value)
  if (coverFile.value) data.append('cover', coverFile.value)

  loading.value = true
  progress.value = 0
  abortController.value = new AbortController()

  try {
    const res = await videoApi.upload(
      data,
      (value) => {
        progress.value = Math.min(value, 100)
      },
      abortController.value.signal
    )
    await publishToCollection(res.data.id)
    ElMessage.success('上传成功，视频已进入审核和后台转码')
    router.push('/creator')
  } catch (error: any) {
    if (error?.code === 'ERR_CANCELED' || error?.name === 'CanceledError') {
      ElMessage.info('已终止上传')
      progress.value = 0
      return
    }
    ElMessage.error(error?.message || '上传失败')
  } finally {
    loading.value = false
    abortController.value = null
  }
}

async function publishToCollection(videoId: number) {
  const collectionTitle = newCollectionTitle.value.trim()
  if (collectionTitle) {
    const created = await collectionApi.create({ title: collectionTitle })
    await collectionApi.addVideo(created.data.id, videoId)
    return
  }
  if (selectedCollectionId.value) {
    await collectionApi.addVideo(selectedCollectionId.value, videoId)
  }
}

function cancelUpload() {
  abortController.value?.abort()
}

function normalizeTags(tags: string | string[]) {
  if (Array.isArray(tags)) return tags.map((tag) => tag.trim()).filter(Boolean)
  return tags.split(',').map((tag) => tag.trim()).filter(Boolean)
}
</script>

<style scoped>
.upload-page {
  display: grid;
  gap: 18px;
}

.section-head p {
  margin: 8px 0 0;
}

.upload-grid {
  display: grid;
  grid-template-columns: minmax(0, 0.9fr) minmax(360px, 1.1fr);
  gap: 18px;
}

.upload-panel,
.form-panel {
  padding: 22px;
}

.upload-panel {
  display: grid;
  align-content: start;
  gap: 16px;
}

.upload-icon {
  color: #165dff;
  font-size: 52px;
}

.cover-upload {
  display: flex;
  justify-content: center;
}

.form-panel {
  display: grid;
}

.collection-publish {
  display: grid;
  gap: 10px;
  padding: 14px;
  margin-bottom: 18px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #fafafa;
}

.collection-title {
  color: #101828;
  font-size: 14px;
  font-weight: 700;
}

.progress-box {
  display: grid;
  gap: 8px;
  margin-bottom: 12px;
}

.progress-box span {
  color: #667085;
  font-size: 13px;
}

.upload-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

@media (max-width: 820px) {
  .upload-grid {
    grid-template-columns: 1fr;
  }
}
</style>
