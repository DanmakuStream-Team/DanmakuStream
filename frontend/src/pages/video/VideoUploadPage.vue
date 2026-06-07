<template>
  <main class="page-shell upload-page">
    <div class="section-head">
      <div>
        <h1>上传视频</h1>
        <p class="muted">大文件上传可能需要几分钟，上传过程中请保持页面打开。</p>
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
          <el-input v-model="form.description" :disabled="loading" type="textarea" :rows="5" placeholder="介绍一下视频内容" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="form.tags" :disabled="loading" placeholder="多个标签用英文逗号分隔，如 Go,WebSocket" />
        </el-form-item>

        <div v-if="loading || progress > 0" class="progress-box">
          <el-progress :percentage="progress" />
          <span>{{ progress < 100 ? '正在上传文件...' : '上传完成，后端正在转码，可离开当前页面' }}</span>
        </div>

        <div class="upload-actions">
          <el-button type="primary" size="large" :loading="loading" :disabled="loading" @click="submit">提交上传</el-button>
          <el-button v-if="loading" type="danger" plain size="large" @click="cancelUpload">终止上传</el-button>
        </div>
      </el-form>
    </section>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage, type UploadFile } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { videoApi } from '@/api/video'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const progress = ref(0)
const videoFile = ref<File | null>(null)
const coverFile = ref<File | null>(null)
const abortController = ref<AbortController | null>(null)
const form = reactive({ title: '', description: '', tags: '' })

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
  data.append('tags', form.tags.trim())
  data.append('video', videoFile.value)
  if (coverFile.value) data.append('cover', coverFile.value)

  loading.value = true
  progress.value = 0
  abortController.value = new AbortController()

  try {
    await videoApi.upload(
      data,
      (value) => {
        progress.value = Math.min(value, 100)
      },
      abortController.value.signal
    )
    ElMessage.success('上传成功，后端正在转码')
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

function cancelUpload() {
  abortController.value?.abort()
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
