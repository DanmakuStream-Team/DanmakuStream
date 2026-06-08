<template>
  <main class="page-shell operations-page">
    <div class="section-head">
      <div>
        <h1>运营工具</h1>
        <p class="muted">维护首页轮播图和全站系统公告。</p>
      </div>
      <el-button :loading="loading" @click="load">刷新</el-button>
    </div>

    <section class="ops-grid">
      <div class="soft-panel ops-panel">
        <div class="panel-head">
          <h2>首页轮播图</h2>
          <el-button type="primary" @click="openBanner()">新增横幅</el-button>
        </div>
        <div v-if="banners.length" class="rows">
          <article v-for="item in banners" :key="item.id" class="banner-row">
            <img v-if="item.imageUrl" :src="mediaUrl(item.imageUrl)" :alt="item.title" />
            <div v-else class="thumb">Banner</div>
            <div>
              <strong>{{ item.title }}</strong>
              <span>{{ item.link || '未设置跳转链接' }}</span>
            </div>
            <el-tag :type="item.enabled ? 'success' : 'info'">{{ item.enabled ? '启用' : '停用' }}</el-tag>
            <el-button text @click="openBanner(item)">编辑</el-button>
            <el-button text type="danger" @click="deleteBanner(item.id)">删除</el-button>
          </article>
        </div>
        <el-empty v-else description="暂无横幅" />
      </div>

      <div class="soft-panel ops-panel">
        <div class="panel-head">
          <h2>系统公告</h2>
          <el-button type="primary" @click="openAnnouncement()">新增公告</el-button>
        </div>
        <div v-if="announcements.length" class="rows">
          <article v-for="item in announcements" :key="item.id" class="announcement-row">
            <div>
              <strong>{{ item.content }}</strong>
              <span>{{ item.startedAt || '立即生效' }} - {{ item.endedAt || '长期' }}</span>
            </div>
            <el-tag :type="item.enabled ? 'success' : 'info'">{{ item.enabled ? '启用' : '停用' }}</el-tag>
            <el-button text @click="openAnnouncement(item)">编辑</el-button>
            <el-button text type="danger" @click="deleteAnnouncement(item.id)">删除</el-button>
          </article>
        </div>
        <el-empty v-else description="暂无公告" />
      </div>
    </section>

    <el-dialog v-model="bannerVisible" title="轮播图" width="520px">
      <el-form label-position="top">
        <el-form-item label="标题">
          <el-input v-model="bannerForm.title" />
        </el-form-item>
        <el-form-item label="封面地址">
          <el-input v-model="bannerForm.imageUrl" />
        </el-form-item>
        <el-form-item label="跳转链接">
          <el-input v-model="bannerForm.link" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="bannerForm.sort" :min="0" />
        </el-form-item>
        <el-form-item>
          <el-switch v-model="bannerForm.enabled" active-text="启用" inactive-text="停用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bannerVisible = false">取消</el-button>
        <el-button type="primary" @click="saveBanner">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="announcementVisible" title="系统公告" width="520px">
      <el-form label-position="top">
        <el-form-item label="公告内容">
          <el-input v-model="announcementForm.content" type="textarea" :rows="3" maxlength="500" show-word-limit />
        </el-form-item>
        <el-form-item label="开始时间">
          <el-input v-model="announcementForm.startedAt" placeholder="可选，如 2026-06-08 12:00:00" />
        </el-form-item>
        <el-form-item label="结束时间">
          <el-input v-model="announcementForm.endedAt" placeholder="可选，如 2026-06-08 13:00:00" />
        </el-form-item>
        <el-form-item>
          <el-switch v-model="announcementForm.enabled" active-text="启用" inactive-text="停用" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="announcementVisible = false">取消</el-button>
        <el-button type="primary" @click="saveAnnouncement">保存</el-button>
      </template>
    </el-dialog>
  </main>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminApi, type SiteAnnouncement, type SiteBanner } from '@/api/admin'
import { mediaUrl } from '@/utils/format'

const loading = ref(false)
const banners = ref<SiteBanner[]>([])
const announcements = ref<SiteAnnouncement[]>([])
const bannerVisible = ref(false)
const announcementVisible = ref(false)
const editingBannerId = ref<number>()
const editingAnnouncementId = ref<number>()
const bannerForm = reactive({
  title: '',
  imageUrl: '',
  link: '',
  enabled: true,
  sort: 0,
})
const announcementForm = reactive({
  content: '',
  enabled: true,
  startedAt: '',
  endedAt: '',
})

onMounted(load)

async function load() {
  loading.value = true
  try {
    const [bannerRes, announcementRes] = await Promise.all([
      adminApi.banners(),
      adminApi.announcements(),
    ])
    banners.value = bannerRes.data
    announcements.value = announcementRes.data
  } finally {
    loading.value = false
  }
}

function openBanner(item?: SiteBanner) {
  editingBannerId.value = item?.id
  bannerForm.title = item?.title || ''
  bannerForm.imageUrl = item?.imageUrl || ''
  bannerForm.link = item?.link || ''
  bannerForm.enabled = item?.enabled ?? true
  bannerForm.sort = item?.sort || 0
  bannerVisible.value = true
}

async function saveBanner() {
  const payload = { ...bannerForm }
  if (editingBannerId.value) await adminApi.updateBanner(editingBannerId.value, payload)
  else await adminApi.createBanner(payload)
  bannerVisible.value = false
  ElMessage.success('横幅已保存')
  load()
}

async function deleteBanner(id: number) {
  await adminApi.deleteBanner(id)
  ElMessage.success('横幅已删除')
  load()
}

function openAnnouncement(item?: SiteAnnouncement) {
  editingAnnouncementId.value = item?.id
  announcementForm.content = item?.content || ''
  announcementForm.enabled = item?.enabled ?? true
  announcementForm.startedAt = item?.startedAt || ''
  announcementForm.endedAt = item?.endedAt || ''
  announcementVisible.value = true
}

async function saveAnnouncement() {
  const payload = { ...announcementForm }
  if (editingAnnouncementId.value) await adminApi.updateAnnouncement(editingAnnouncementId.value, payload)
  else await adminApi.createAnnouncement(payload)
  announcementVisible.value = false
  ElMessage.success('公告已保存')
  load()
}

async function deleteAnnouncement(id: number) {
  await adminApi.deleteAnnouncement(id)
  ElMessage.success('公告已删除')
  load()
}
</script>

<style scoped>
.operations-page {
  display: grid;
  gap: 18px;
}

.section-head p {
  margin: 8px 0 0;
}

.ops-grid {
  display: grid;
  gap: 18px;
}

.ops-panel {
  display: grid;
  gap: 16px;
  padding: 18px;
}

.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.panel-head h2 {
  margin: 0;
  font-size: 20px;
}

.rows {
  display: grid;
  gap: 12px;
}

.banner-row,
.announcement-row {
  display: grid;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  background: #f7f9fc;
}

.banner-row {
  grid-template-columns: 140px minmax(0, 1fr) auto auto auto;
}

.announcement-row {
  grid-template-columns: minmax(0, 1fr) auto auto auto;
}

.banner-row img,
.thumb {
  width: 140px;
  height: 72px;
  border-radius: 8px;
  object-fit: cover;
}

.thumb {
  display: grid;
  place-items: center;
  background: rgba(251, 114, 153, 0.12);
  color: #fb7299;
  font-weight: 900;
}

.banner-row strong,
.banner-row span,
.announcement-row strong,
.announcement-row span {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.banner-row span,
.announcement-row span {
  margin-top: 6px;
  color: #667085;
  font-size: 13px;
}

@media (max-width: 920px) {
  .banner-row,
  .announcement-row {
    grid-template-columns: 1fr;
  }
}
</style>
