<template>
  <main class="page-shell live-list-page">
    <section class="section-head">
      <div>
        <h1>直播</h1>
        <p>正在播出的直播间</p>
      </div>
      <el-button type="primary" @click="openCreateDialog">开始直播</el-button>
    </section>

    <section v-loading="loading" class="live-grid-list">
      <article
        v-for="room in rooms"
        :key="room.id"
        class="live-card"
        @click="router.push(`/live/${room.id}`)"
      >
        <div class="live-cover">
          <img v-if="room.coverUrl" :src="mediaUrl(room.coverUrl)" :alt="room.title" />
          <div v-else class="live-cover-fallback">Danmaku Live</div>
          <span class="live-badge">直播中</span>
          <span class="live-viewers">{{ formatCount(room.viewerCount) }} 人观看</span>
        </div>
        <div class="live-body">
          <h2>{{ room.title }}</h2>
          <button class="live-owner" type="button" :disabled="!room.owner?.id" @click.stop="openUser(room.owner?.id)">
            <el-avatar :size="28" :src="mediaUrl(room.owner?.avatar || '')">
              {{ room.owner?.nickname?.slice(0, 1) || 'U' }}
            </el-avatar>
            <span>{{ room.owner?.nickname || room.owner?.username || '匿名主播' }}</span>
          </button>
        </div>
      </article>

      <div v-if="!rooms.length && !loading" class="soft-panel empty-panel">
        <el-empty description="当前暂无正在播出的直播间" />
      </div>
    </section>

    <el-dialog v-model="createVisible" title="开始直播" width="460px">
      <el-form label-position="top">
        <el-form-item label="直播标题">
          <el-input v-model="createForm.title" maxlength="40" placeholder="输入直播标题" />
        </el-form-item>
        <el-form-item label="封面地址">
          <el-input v-model="createForm.coverUrl" placeholder="可选，填写图片 URL" />
        </el-form-item>
      </el-form>

      <div v-if="createdRoom" class="stream-info">
        <strong>推流地址</strong>
        <span>{{ createdRoom.publishUrl }}</span>
        <strong>播放地址</strong>
        <span>{{ createdRoom.playUrl }}</span>
      </div>

      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="createLiveRoom">
          {{ createdRoom ? '重新生成' : '开始直播' }}
        </el-button>
        <el-button v-if="createdRoom" type="success" @click="router.push(`/live/${createdRoom.id}`)">
          进入直播间
        </el-button>
      </template>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { liveApi } from '@/api/live'
import type { LiveRoom } from '@/types'
import { formatCount, mediaUrl } from '@/utils/format'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const loading = ref(false)
const creating = ref(false)
const createVisible = ref(false)
const rooms = ref<LiveRoom[]>([])
const createdRoom = ref<LiveRoom>()
const createForm = reactive({
  title: '',
  coverUrl: '',
})

onMounted(async () => {
  await loadLiveRooms()
  if (route.query.create === '1') openCreateDialog()
})

async function loadLiveRooms() {
  loading.value = true
  try {
    const res = await liveApi.list({ page: 1, pageSize: 100 })
    rooms.value = res.data.list
  } catch {
    rooms.value = []
    ElMessage.error('直播列表加载失败')
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  if (!authStore.isLoggedIn) {
    ElMessage.warning('请先登录后再开播')
    router.push({ path: '/login', query: { redirect: '/live' } })
    return
  }
  createdRoom.value = undefined
  createForm.title = `${authStore.userInfo?.nickname || '我的'}的直播间`
  createForm.coverUrl = ''
  createVisible.value = true
}

function openUser(userId?: number) {
  if (!userId) return
  router.push(`/user/${userId}`)
}

async function createLiveRoom() {
  const title = createForm.title.trim()
  if (!title) {
    ElMessage.warning('请输入直播标题')
    return
  }

  creating.value = true
  try {
    const res = await liveApi.create({
      title,
      coverUrl: createForm.coverUrl.trim() || undefined,
    })
    createdRoom.value = res.data
    await loadLiveRooms()
    ElMessage.success('直播间已创建')
  } catch {
    ElMessage.error('直播间创建失败')
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.live-list-page {
  display: grid;
  gap: 18px;
  padding-top: 22px;
}

.section-head p {
  margin: 8px 0 0;
  color: #667085;
}

.live-grid-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 18px;
  min-height: 360px;
}

.live-card {
  min-width: 0;
  cursor: pointer;
}

.live-cover {
  position: relative;
  overflow: hidden;
  aspect-ratio: 16 / 9;
  border-radius: 8px;
  background: #f1f2f3;
}

.live-cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.25s ease;
}

.live-card:hover .live-cover img {
  transform: scale(1.04);
}

.live-cover-fallback {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  background:
    linear-gradient(135deg, rgba(251, 114, 153, 0.18), rgba(0, 174, 236, 0.14)),
    #f6f7f8;
  color: #fb7299;
  font-size: 16px;
  font-weight: 900;
}

.live-badge,
.live-viewers {
  position: absolute;
  bottom: 8px;
  padding: 3px 7px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.72);
  color: #fff;
  font-size: 12px;
  font-weight: 800;
}

.live-badge {
  left: 8px;
  background: #fb7299;
}

.live-viewers {
  right: 8px;
}

.live-body {
  display: grid;
  gap: 8px;
  padding: 9px 2px 0;
}

.live-body h2 {
  margin: 0;
  overflow: hidden;
  color: #18191c;
  font-size: 15px;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.live-card:hover h2 {
  color: #fb7299;
}

.live-owner {
  display: flex;
  align-items: center;
  gap: 7px;
  min-width: 0;
  padding: 0;
  border: 0;
  background: transparent;
  color: #9499a0;
  cursor: pointer;
  font-size: 13px;
}

.live-owner:not(:disabled):hover {
  color: #fb7299;
}

.live-owner:disabled {
  cursor: default;
}

.live-owner span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.stream-info {
  display: grid;
  gap: 6px;
  padding: 12px;
  border-radius: 8px;
  background: #f6f7f8;
}

.stream-info strong {
  color: #18191c;
  font-size: 13px;
}

.stream-info span {
  overflow: hidden;
  color: #61666d;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
