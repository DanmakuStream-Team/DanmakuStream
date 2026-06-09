<template>
  <main class="page-shell live-list-page">
    <section class="section-head">
      <div>
        <h1>直播</h1>
        <p>预约想看的直播，或直接开始推流。</p>
      </div>
      <div class="head-actions">
        <el-button @click="openScheduleDialog">预约直播</el-button>
        <el-button type="primary" @click="openCreateDialog">开始直播</el-button>
      </div>
    </section>

    <section>
      <div class="section-subhead">
        <h2>直播预约</h2>
        <span>{{ schedules.length }} 场待开播</span>
      </div>
      <div v-loading="scheduleLoading" class="schedule-list">
        <article v-for="schedule in schedules" :key="schedule.id" class="schedule-card soft-panel">
          <div class="schedule-cover">
            <img v-if="schedule.coverUrl" :src="mediaUrl(schedule.coverUrl)" :alt="schedule.title" />
            <span v-else>Danmaku Live</span>
          </div>
          <div class="schedule-main">
            <h3>{{ schedule.title }}</h3>
            <p>{{ schedule.scheduledAt }}</p>
            <button class="live-owner" type="button" :disabled="!schedule.owner?.id" @click="openUser(schedule.owner?.id)">
              <el-avatar :size="28" :src="mediaUrl(schedule.owner?.avatar || '')">
                {{ schedule.owner?.nickname?.slice(0, 1) || 'U' }}
              </el-avatar>
              <span>{{ schedule.owner?.nickname || schedule.owner?.username || '匿名主播' }}</span>
            </button>
          </div>
          <div class="schedule-actions">
            <span>{{ formatCount(schedule.reminderCount) }} 人预约</span>
            <el-button
              v-if="schedule.ownerId !== authStore.userInfo?.id"
              :type="schedule.reserved ? 'success' : 'primary'"
              @click="toggleReserve(schedule)"
            >
              {{ schedule.reserved ? '已预约' : '预约提醒' }}
            </el-button>
            <el-button v-else type="danger" plain @click="cancelSchedule(schedule.id)">取消预约</el-button>
          </div>
        </article>
        <div v-if="!schedules.length && !scheduleLoading" class="soft-panel empty-panel">
          <el-empty description="当前暂无直播预约" />
        </div>
      </div>
    </section>

    <section>
      <div class="section-subhead">
        <h2>正在直播</h2>
        <span>{{ rooms.length }} 个直播间</span>
      </div>
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
    </section>

    <el-dialog v-model="createVisible" title="开始直播" width="460px">
      <el-form label-position="top">
        <el-form-item label="直播标题">
          <el-input v-model="createForm.title" maxlength="40" placeholder="输入直播标题" />
        </el-form-item>
        <el-form-item label="直播封面">
          <div class="cover-uploader">
            <button class="cover-preview" type="button" @click="coverInputRef?.click()">
              <img v-if="createForm.coverUrl" :src="mediaUrl(createForm.coverUrl)" alt="直播封面" />
              <span v-else>选择封面</span>
            </button>
            <div class="cover-actions">
              <el-button :loading="uploadingCover" @click="coverInputRef?.click()">
                {{ createForm.coverUrl ? '更换封面' : '上传封面' }}
              </el-button>
              <el-button v-if="createForm.coverUrl" text @click="clearCover">移除</el-button>
              <input
                ref="coverInputRef"
                class="cover-input"
                type="file"
                accept="image/*"
                @change="uploadCover($event, 'create')"
              />
            </div>
          </div>
        </el-form-item>
      </el-form>

      <div v-if="createdRoom" class="stream-info">
        <div class="stream-info-head">
          <strong>推流参数</strong>
          <el-button size="small" @click="copyAllStreamParams">复制全部</el-button>
        </div>
        <div v-for="item in streamParams" :key="item.label" class="stream-param">
          <span class="param-label">{{ item.label }}</span>
          <code>{{ item.value || '-' }}</code>
          <el-button size="small" :disabled="!item.value" @click="copyText(item.value, item.label)">复制</el-button>
        </div>
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

    <el-dialog v-model="scheduleVisible" title="预约直播" width="460px">
      <el-form label-position="top">
        <el-form-item label="预约标题">
          <el-input v-model="scheduleForm.title" maxlength="40" placeholder="输入预约标题" />
        </el-form-item>
        <el-form-item label="开播时间">
          <el-date-picker
            v-model="scheduleForm.scheduledAt"
            class="date-picker"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            placeholder="选择开播时间"
          />
        </el-form-item>
        <el-form-item label="预约封面">
          <div class="cover-uploader">
            <button class="cover-preview" type="button" @click="scheduleCoverInputRef?.click()">
              <img v-if="scheduleForm.coverUrl" :src="mediaUrl(scheduleForm.coverUrl)" alt="预约封面" />
              <span v-else>选择封面</span>
            </button>
            <div class="cover-actions">
              <el-button :loading="uploadingCover" @click="scheduleCoverInputRef?.click()">
                {{ scheduleForm.coverUrl ? '更换封面' : '上传封面' }}
              </el-button>
              <el-button v-if="scheduleForm.coverUrl" text @click="scheduleForm.coverUrl = ''">移除</el-button>
              <input
                ref="scheduleCoverInputRef"
                class="cover-input"
                type="file"
                accept="image/*"
                @change="uploadCover($event, 'schedule')"
              />
            </div>
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="scheduleVisible = false">取消</el-button>
        <el-button type="primary" :loading="creatingSchedule" @click="createLiveSchedule">创建预约</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { liveApi } from '@/api/live'
import { mediaApi } from '@/api/media'
import type { LiveRoom, LiveSchedule } from '@/types'
import { formatCount, mediaUrl } from '@/utils/format'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const loading = ref(false)
const scheduleLoading = ref(false)
const creating = ref(false)
const creatingSchedule = ref(false)
const uploadingCover = ref(false)
const createVisible = ref(false)
const scheduleVisible = ref(false)
const coverInputRef = ref<HTMLInputElement>()
const scheduleCoverInputRef = ref<HTMLInputElement>()
const rooms = ref<LiveRoom[]>([])
const schedules = ref<LiveSchedule[]>([])
const createdRoom = ref<LiveRoom>()
const createForm = reactive({
  title: '',
  coverUrl: '',
})
const scheduleForm = reactive({
  title: '',
  coverUrl: '',
  scheduledAt: '',
})
const streamServer = computed(() => {
  const publishUrl = createdRoom.value?.publishUrl || ''
  const streamKey = createdRoom.value?.streamKey || ''
  if (!publishUrl || !streamKey) return ''
  return publishUrl.endsWith(`/${streamKey}`) ? publishUrl.slice(0, -streamKey.length - 1) : publishUrl
})
const streamParams = computed(() => {
  if (!createdRoom.value) return []
  return [
    { label: 'OBS 服务器', value: streamServer.value },
    { label: '串流密钥', value: createdRoom.value.streamKey || '' },
    { label: '完整推流地址', value: createdRoom.value.publishUrl || '' },
    { label: '播放地址', value: createdRoom.value.playUrl || '' },
  ]
})

onMounted(async () => {
  await Promise.all([loadLiveRooms(), loadLiveSchedules()])
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

async function loadLiveSchedules() {
  scheduleLoading.value = true
  try {
    const res = await liveApi.schedules({ page: 1, pageSize: 100, status: 'pending' })
    schedules.value = res.data.list
  } catch {
    schedules.value = []
    ElMessage.error('直播预约加载失败')
  } finally {
    scheduleLoading.value = false
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

function openScheduleDialog() {
  if (!authStore.isLoggedIn) {
    ElMessage.warning('请先登录后再预约直播')
    router.push({ path: '/login', query: { redirect: '/live' } })
    return
  }
  scheduleForm.title = `${authStore.userInfo?.nickname || '我的'}的直播预告`
  scheduleForm.coverUrl = ''
  scheduleForm.scheduledAt = ''
  scheduleVisible.value = true
}

function openUser(userId?: number) {
  if (!userId) return
  router.push(`/user/${userId}`)
}

async function uploadCover(event: Event, target: 'create' | 'schedule') {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return

  if (!file.type.startsWith('image/')) {
    ElMessage.warning('请选择图片文件')
    return
  }

  uploadingCover.value = true
  try {
    const res = await mediaApi.uploadImage(file, 'live')
    if (target === 'create') createForm.coverUrl = res.data.url
    else scheduleForm.coverUrl = res.data.url
    ElMessage.success('封面上传成功')
  } catch {
    ElMessage.error('封面上传失败')
  } finally {
    uploadingCover.value = false
  }
}

function clearCover() {
  createForm.coverUrl = ''
}

async function createLiveSchedule() {
  const title = scheduleForm.title.trim()
  if (!title) {
    ElMessage.warning('请输入预约标题')
    return
  }
  if (!scheduleForm.scheduledAt) {
    ElMessage.warning('请选择开播时间')
    return
  }

  creatingSchedule.value = true
  try {
    await liveApi.createSchedule({
      title,
      scheduledAt: scheduleForm.scheduledAt,
      coverUrl: scheduleForm.coverUrl.trim() || undefined,
    })
    scheduleVisible.value = false
    await loadLiveSchedules()
    ElMessage.success('直播预约已创建')
  } catch {
    ElMessage.error('直播预约创建失败')
  } finally {
    creatingSchedule.value = false
  }
}

async function toggleReserve(schedule: LiveSchedule) {
  if (!authStore.isLoggedIn) {
    ElMessage.warning('请先登录后再预约')
    router.push({ path: '/login', query: { redirect: '/live' } })
    return
  }
  try {
    const res = await liveApi.reserveSchedule(schedule.id)
    schedule.reserved = res.data.reserved
    schedule.reminderCount = res.data.reminderCount
    ElMessage.success(res.data.reserved ? '已预约开播提醒' : '已取消预约提醒')
  } catch {
    ElMessage.error('预约操作失败')
  }
}

async function cancelSchedule(id: number) {
  try {
    await liveApi.cancelSchedule(id)
    schedules.value = schedules.value.filter((item) => item.id !== id)
    ElMessage.success('直播预约已取消')
  } catch {
    ElMessage.error('直播预约取消失败')
  }
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

async function copyText(text: string, label = '内容') {
  if (!text) return
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text)
    } else {
      fallbackCopy(text)
    }
    ElMessage.success(`${label}已复制`)
  } catch {
    fallbackCopy(text)
    ElMessage.success(`${label}已复制`)
  }
}

function fallbackCopy(text: string) {
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', 'readonly')
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  document.body.appendChild(textarea)
  textarea.select()
  document.execCommand('copy')
  document.body.removeChild(textarea)
}

function copyAllStreamParams() {
  const text = streamParams.value
    .filter((item) => item.value)
    .map((item) => `${item.label}: ${item.value}`)
    .join('\n')
  copyText(text, '推流参数')
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

.head-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.section-subhead {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 14px;
}

.section-subhead h2 {
  margin: 0;
  color: #18191c;
  font-size: 24px;
  font-weight: 900;
}

.section-subhead span {
  color: #9499a0;
  font-size: 14px;
  font-weight: 800;
}

.schedule-list {
  display: grid;
  gap: 12px;
  min-height: 180px;
}

.schedule-card {
  display: grid;
  grid-template-columns: 180px minmax(0, 1fr) auto;
  gap: 16px;
  align-items: center;
  padding: 12px;
}

.schedule-cover {
  display: grid;
  overflow: hidden;
  width: 100%;
  aspect-ratio: 16 / 9;
  place-items: center;
  border-radius: 8px;
  background: #f6f7f8;
  color: #fb7299;
  font-size: 15px;
  font-weight: 900;
}

.schedule-cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.schedule-main {
  display: grid;
  min-width: 0;
  gap: 8px;
}

.schedule-main h3 {
  margin: 0;
  overflow: hidden;
  color: #18191c;
  font-size: 18px;
  font-weight: 900;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.schedule-main p {
  margin: 0;
  color: #61666d;
  font-size: 14px;
  font-weight: 800;
}

.schedule-actions {
  display: grid;
  justify-items: end;
  gap: 10px;
}

.schedule-actions span {
  color: #9499a0;
  font-size: 13px;
  font-weight: 800;
}

.live-grid-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 18px;
  min-height: 360px;
}

.live-grid-list .empty-panel {
  grid-column: 1 / -1;
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
  gap: 10px;
  padding: 12px;
  border-radius: 8px;
  background: #f6f7f8;
}

.stream-info-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.stream-info-head strong {
  color: #18191c;
  font-size: 15px;
}

.stream-param {
  display: grid;
  grid-template-columns: 86px minmax(0, 1fr) auto;
  gap: 8px;
  align-items: center;
}

.param-label {
  color: #61666d;
  font-size: 13px;
  font-weight: 800;
}

.stream-param code {
  overflow: hidden;
  color: #61666d;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cover-uploader {
  display: grid;
  grid-template-columns: 168px minmax(0, 1fr);
  gap: 14px;
  align-items: center;
  width: 100%;
}

.cover-preview {
  display: block;
  overflow: hidden;
  width: 168px;
  aspect-ratio: 16 / 9;
  border: 1px dashed #c9ccd0;
  border-radius: 8px;
  background: #f6f7f8;
  color: #9499a0;
  cursor: pointer;
  font-size: 14px;
  font-weight: 800;
}

.cover-preview img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-preview span {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
}

.cover-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.cover-input {
  display: none;
}

.date-picker {
  width: 100%;
}

@media (max-width: 520px) {
  .schedule-card {
    grid-template-columns: 1fr;
  }

  .schedule-actions {
    justify-items: stretch;
  }

  .cover-uploader {
    grid-template-columns: 1fr;
  }

  .cover-preview {
    width: 100%;
  }

  .stream-param {
    grid-template-columns: 1fr auto;
  }

  .param-label {
    grid-column: 1 / -1;
  }
}
</style>
