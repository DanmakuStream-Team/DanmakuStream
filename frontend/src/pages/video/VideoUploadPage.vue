<template>
  <main class="page-shell upload-page">
    <div class="section-head">
      <div>
        <h1>上传视频</h1>
        <p class="muted">提交后进入 pending 状态，管理员审核通过后会在首页展示。</p>
      </div>
    </div>

    <section class="upload-grid">
      <div class="soft-panel upload-panel">
        <el-upload
          drag
          :auto-upload="false"
          :limit="1"
          accept="video/*"
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
          :on-change="selectCover"
          :on-remove="removeCover"
        >
          <el-button>选择封面图</el-button>
        </el-upload>
      </div>

      <el-form class="soft-panel form-panel" label-position="top">
        <el-form-item label="标题">
          <el-input v-model="form.title" placeholder="请输入视频标题" />
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="form.description" type="textarea" :rows="5" placeholder="介绍一下视频内容" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="form.tags" placeholder="多个标签用英文逗号分隔，如 Go,WebSocket" />
        </el-form-item>
        <el-progress v-if="progress > 0" :percentage="progress" />
        <el-button type="primary" size="large" :loading="loading" @click="submit">提交上传</el-button>
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
const form = reactive({ title: '', description: '', tags: '' })

function selectVideo(file: UploadFile) {
  videoFile.value = file.raw || null
}

function removeVideo() {
  videoFile.value = null
}

function selectCover(file: UploadFile) {
  coverFile.value = file.raw || null
}

function removeCover() {
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
  try {
    const res = await videoApi.upload(data, (value) => (progress.value = value))
    ElMessage.success('上传成功，等待审核')
    router.push(`/video/${res.data.id}`)
  } finally {
    loading.value = false
  }
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

@media (max-width: 820px) {
  .upload-grid {
    grid-template-columns: 1fr;
  }
}
</style>
