<template>
  <main class="page-shell subscription-page">
    <section class="section-head subscription-head">
      <div>
        <h1>关注管理</h1>
        <p>整理关注分组、付费特别关注与黑名单，并查看对应创作者的最新内容。</p>
      </div>
      <div class="head-actions">
        <el-button :icon="Plus" @click="createGroup">新建分组</el-button>
        <el-button :loading="loading" @click="loadRelationships">刷新</el-button>
      </div>
    </section>

    <section class="relationship-layout">
      <aside class="filter-panel">
        <button :class="{ active: activeFilter === 'all' }" type="button" @click="activeFilter = 'all'">
          <span>全部关注</span><em>{{ followees.length }}</em>
        </button>
        <button :class="{ active: activeFilter === 'special' }" type="button" @click="activeFilter = 'special'">
          <span>付费特别关注</span><em>{{ specialCount }}</em>
        </button>
        <button :class="{ active: activeFilter === 'ungrouped' }" type="button" @click="activeFilter = 'ungrouped'">
          <span>未分组</span><em>{{ ungroupedCount }}</em>
        </button>

        <div class="filter-divider" />
        <div v-for="group in groups" :key="group.id" class="group-row">
          <button :class="{ active: activeFilter === `group-${group.id}` }" type="button" @click="activeFilter = `group-${group.id}`">
            <span>{{ group.name }}</span><em>{{ group.count }}</em>
          </button>
          <el-dropdown trigger="click">
            <button class="group-more" type="button" :aria-label="`${group.name}分组操作`">
              <el-icon><MoreFilled /></el-icon>
            </button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="renameGroup(group)">重命名</el-dropdown-item>
                <el-dropdown-item divided @click="deleteGroup(group)">删除分组</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>

        <div class="filter-divider" />
        <button :class="{ active: activeFilter === 'blocked' }" type="button" @click="activeFilter = 'blocked'">
          <span>黑名单</span><em>{{ blockedUsers.length }}</em>
        </button>
      </aside>

      <section v-loading="loading" class="relationship-content">
        <template v-if="activeFilter === 'blocked'">
          <div class="content-title">
            <div><h2>黑名单</h2><p>被拉黑的用户无法与你互相关注。</p></div>
          </div>
          <div v-if="blockedUsers.length" class="creator-list">
            <article v-for="creator in blockedUsers" :key="creator.id" class="creator-row">
              <button class="creator-main" type="button" @click="router.push(`/user/${creator.id}`)">
                <el-avatar :size="44" :src="mediaUrl(creator.avatar)">{{ creator.nickname.slice(0, 1) }}</el-avatar>
                <span><strong>{{ creator.nickname }}</strong><small>拉黑于 {{ creator.blockedAt }}</small></span>
              </button>
              <el-button @click="unblock(creator)">解除拉黑</el-button>
            </article>
          </div>
          <el-empty v-else description="黑名单为空" />
        </template>

        <template v-else>
          <div class="content-title">
            <div><h2>{{ currentTitle }}</h2><p>共 {{ filteredFollowees.length }} 位创作者</p></div>
          </div>
          <div v-if="filteredFollowees.length" class="creator-list">
            <article v-for="creator in filteredFollowees" :key="creator.id" class="creator-row">
              <button class="creator-main" type="button" @click="router.push(`/user/${creator.id}`)">
                <el-avatar :size="44" :src="mediaUrl(creator.avatar)">{{ creator.nickname.slice(0, 1) }}</el-avatar>
                <span>
                  <strong>{{ creator.nickname }}</strong>
                  <small>{{ creatorMeta(creator) }}</small>
                </span>
              </button>
              <div class="creator-actions">
                <el-tooltip v-if="activeSubscription(creator.id)" content="已开通付费特别关注">
                  <button class="member-badge" type="button" aria-label="已付费特别关注" @click="router.push(`/user/${creator.id}`)">
                    <el-icon><Star /></el-icon>
                  </button>
                </el-tooltip>
                <el-button v-if="activeSubscription(creator.id)" plain @click="router.push(`/user/${creator.id}`)">续费</el-button>
                <el-select :model-value="creator.groupId || 0" class="group-select" @change="(value: number) => changeGroup(creator, value)">
                  <el-option label="未分组" :value="0" />
                  <el-option v-for="group in groups" :key="group.id" :label="group.name" :value="group.id" />
                </el-select>
              </div>
            </article>
          </div>
          <el-empty v-else description="这个分组还没有关注用户" />
        </template>
      </section>
    </section>

    <section v-if="activeFilter !== 'blocked'" class="latest-section">
      <div class="content-title"><div><h2>最新投稿</h2><p>来自当前列表中的创作者</p></div></div>
      <div v-if="videos.length" class="video-grid">
        <VideoCard v-for="video in videos" :key="video.id" :video="video" @open="router.push(`/video/${video.id}`)" />
      </div>
      <div v-else class="soft-panel empty-panel"><el-empty :description="emptyVideoText" /></div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { MoreFilled, Plus, Star } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import VideoCard from '@/components/common/VideoCard.vue'
import {
  userApi,
  type BlockedUserInfo,
  type FolloweeInfo,
  type FollowGroupInfo,
} from '@/api/user'
import { videoApi } from '@/api/video'
import { membershipApi, type CreatorSubscriptionInfo } from '@/api/membership'
import type { VideoInfo } from '@/types'
import { mediaUrl } from '@/utils/format'

const router = useRouter()
const loading = ref(false)
const followees = ref<FolloweeInfo[]>([])
const groups = ref<FollowGroupInfo[]>([])
const blockedUsers = ref<BlockedUserInfo[]>([])
const paidSubscriptions = ref<CreatorSubscriptionInfo[]>([])
const videos = ref<VideoInfo[]>([])
const activeFilter = ref('all')
let videoRequestVersion = 0

const specialCount = computed(() => followees.value.filter(item => activeSubscription(item.id)).length)
const ungroupedCount = computed(() => followees.value.filter(item => !item.groupId).length)
const filteredFollowees = computed(() => {
  if (activeFilter.value === 'special') return followees.value.filter(item => activeSubscription(item.id))
  if (activeFilter.value === 'ungrouped') return followees.value.filter(item => !item.groupId)
  if (activeFilter.value.startsWith('group-')) {
    const groupId = Number(activeFilter.value.slice(6))
    return followees.value.filter(item => item.groupId === groupId)
  }
  return followees.value
})
const currentTitle = computed(() => {
  if (activeFilter.value === 'special') return '付费特别关注'
  if (activeFilter.value === 'ungrouped') return '未分组'
  if (activeFilter.value.startsWith('group-')) {
    return groups.value.find(item => `group-${item.id}` === activeFilter.value)?.name || '关注分组'
  }
  return '全部关注'
})
const emptyVideoText = computed(() => filteredFollowees.value.length ? '这些创作者暂时没有公开投稿' : '当前没有可展示的订阅内容')

onMounted(loadRelationships)
watch(activeFilter, () => void loadSubscriptionVideos())

async function loadRelationships() {
  loading.value = true
  try {
    const [followingRes, groupsRes, blockedRes, subscriptionsRes] = await Promise.all([
      userApi.following(),
      userApi.followGroups(),
      userApi.blocked(),
      membershipApi.mine(),
    ])
    followees.value = followingRes.data.list
    groups.value = groupsRes.data.list
    blockedUsers.value = blockedRes.data.list
    paidSubscriptions.value = subscriptionsRes.data.list
    await loadSubscriptionVideos()
  } catch {
    ElMessage.error('关注关系加载失败')
  } finally {
    loading.value = false
  }
}

async function loadCreatorVideos(creatorId: number) {
  const res = await videoApi.userVideos(creatorId, { page: 1, pageSize: 20 })
  return res.data.list
}

async function loadSubscriptionVideos() {
  const version = ++videoRequestVersion
  if (activeFilter.value === 'blocked' || !filteredFollowees.value.length) {
    videos.value = []
    return
  }
  try {
    const videoGroups = await Promise.all(filteredFollowees.value.map(item => loadCreatorVideos(item.id)))
    if (version !== videoRequestVersion) return
    videos.value = videoGroups.flat().sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt))
  } catch {
    if (version === videoRequestVersion) videos.value = []
  }
}

async function createGroup() {
  try {
    const result = await ElMessageBox.prompt('输入关注分组名称', '新建分组', {
      inputPlaceholder: '例如：学习、朋友、常看',
      inputValidator: value => Boolean(value.trim()) && [...value.trim()].length <= 20 || '请输入 1 到 20 个字符',
    })
    await userApi.createFollowGroup(result.value.trim())
    await loadRelationships()
    ElMessage.success('分组已创建')
  } catch (error: any) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error(error.message || '创建分组失败')
  }
}

async function renameGroup(group: FollowGroupInfo) {
  try {
    const result = await ElMessageBox.prompt('输入新的分组名称', '重命名分组', {
      inputValue: group.name,
      inputValidator: value => Boolean(value.trim()) && [...value.trim()].length <= 20 || '请输入 1 到 20 个字符',
    })
    await userApi.updateFollowGroup(group.id, result.value.trim())
    await loadRelationships()
  } catch (error: any) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error(error.message || '重命名失败')
  }
}

async function deleteGroup(group: FollowGroupInfo) {
  try {
    await ElMessageBox.confirm(`删除“${group.name}”后，组内用户将移至未分组。`, '删除分组', { type: 'warning' })
    await userApi.deleteFollowGroup(group.id)
    if (activeFilter.value === `group-${group.id}`) activeFilter.value = 'all'
    await loadRelationships()
    ElMessage.success('分组已删除')
  } catch (error: any) {
    if (error !== 'cancel' && error !== 'close') ElMessage.error(error.message || '删除分组失败')
  }
}

function activeSubscription(creatorId: number) {
  return paidSubscriptions.value.find(item => item.creator.id === creatorId && item.status === 'active' && item.daysRemaining >= 0)
}

function creatorMeta(creator: FolloweeInfo) {
  const subscription = activeSubscription(creator.id)
  if (subscription) return `${creator.groupName || '未分组'} · 付费特别关注剩余 ${subscription.daysRemaining} 天`
  return creator.groupName || '未分组'
}

async function changeGroup(creator: FolloweeInfo, groupId: number) {
  await userApi.updateFollowSettings(creator.id, { groupId })
  creator.groupId = groupId || null
  creator.groupName = groups.value.find(item => item.id === groupId)?.name || ''
  await refreshGroups()
  await loadSubscriptionVideos()
}

async function refreshGroups() {
  const res = await userApi.followGroups()
  groups.value = res.data.list
}

async function unblock(creator: BlockedUserInfo) {
  const res = await userApi.block(creator.id)
  if (!res.data.blocked) blockedUsers.value = blockedUsers.value.filter(item => item.id !== creator.id)
  ElMessage.success('已解除拉黑')
}
</script>

<style scoped>
.subscription-page { display: grid; gap: 24px; padding-top: 24px; }
.subscription-head { margin: 0; }
.subscription-head p, .content-title p { margin: 6px 0 0; color: #667085; font-size: 13px; }
.head-actions, .creator-actions { display: flex; align-items: center; gap: 8px; }
.relationship-layout { display: grid; grid-template-columns: 210px minmax(0, 1fr); min-height: 420px; border: 1px solid #e7e9ed; border-radius: 8px; }
.filter-panel { padding: 12px; border-right: 1px solid #e7e9ed; }
.filter-panel > button, .group-row > button { display: flex; width: 100%; height: 38px; align-items: center; justify-content: space-between; padding: 0 10px; border: 0; border-radius: 6px; background: transparent; color: #475467; cursor: pointer; text-align: left; }
.filter-panel button:hover, .filter-panel button.active { background: #f4f6f8; color: #18191c; }
.filter-panel button.active { font-weight: 700; }
.filter-panel em { color: #98a2b3; font-size: 12px; font-style: normal; }
.filter-divider { height: 1px; margin: 10px 4px; background: #eceef1; }
.group-row { position: relative; display: grid; grid-template-columns: minmax(0, 1fr) 28px; align-items: center; }
.group-more { display: grid; width: 28px; height: 28px; padding: 0; place-items: center; border: 0; background: transparent; color: #98a2b3; cursor: pointer; }
.relationship-content { min-width: 0; padding: 22px; }
.content-title { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; }
.content-title h2 { margin: 0; color: #18191c; font-size: 19px; }
.creator-list { display: grid; }
.creator-row { display: flex; min-width: 0; min-height: 72px; align-items: center; justify-content: space-between; gap: 16px; border-bottom: 1px solid #eef0f2; }
.creator-main { display: flex; min-width: 0; align-items: center; gap: 12px; padding: 8px 0; border: 0; background: transparent; cursor: pointer; text-align: left; }
.creator-main > span { display: grid; min-width: 0; gap: 4px; }
.creator-main strong, .creator-main small { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.creator-main strong { color: #18191c; font-size: 14px; }
.creator-main small { color: #98a2b3; font-size: 12px; }
.member-badge { display: grid; width: 34px; height: 34px; padding: 0; place-items: center; border: 1px solid #fb7299; border-radius: 6px; background: #fff5f8; color: #fb7299; cursor: pointer; }
.group-select { width: 132px; }
.latest-section { min-width: 0; }
.video-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 24px 18px; }
@media (max-width: 980px) { .video-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
@media (max-width: 760px) { .relationship-layout { grid-template-columns: 1fr; } .filter-panel { display: flex; overflow-x: auto; border-right: 0; border-bottom: 1px solid #e7e9ed; } .filter-panel > button, .group-row { min-width: 126px; } .filter-divider { width: 1px; height: 30px; flex: 0 0 auto; } .video-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
@media (max-width: 520px) { .creator-row, .subscription-head { align-items: flex-start; flex-direction: column; } .creator-actions { width: 100%; } .group-select { flex: 1; } .video-grid { grid-template-columns: 1fr; } }
</style>
