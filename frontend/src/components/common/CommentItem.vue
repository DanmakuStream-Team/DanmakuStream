<template>
  <div class="w-full" :class="{ 'pt-3': props.level > 0 }">
    <div class="flex gap-3">
      <el-avatar :size="props.level > 0 ? 32 : 40">
        <img v-if="props.comment.author?.avatar" :src="props.comment.author.avatar" alt="avatar" />
        <span v-else>{{ nicknameFirstChar }}</span>
      </el-avatar>

      <div class="flex-1 min-w-0">
        <div class="text-sm font-semibold text-gray-800 mb-1">{{ authorName }}</div>
        <div class="text-gray-700 leading-relaxed whitespace-pre-wrap break-words">{{ props.comment.content }}</div>
        <div class="flex items-center gap-3 text-xs text-gray-400 mt-1.5">
          <span>{{ formattedTime }}</span>
          <span v-if="props.comment.likeCount > 0">{{ props.comment.likeCount }} 赞</span>
          <el-button
            v-if="props.showReply"
            type="text"
            size="small"
            class="!p-0 !text-xs !text-gray-400 hover:!text-blue-500"
            @click="handleReply"
          >
            回复
          </el-button>
        </div>

        <div v-if="safeReplies.length > 0" class="mt-3 bg-gray-50 rounded-lg p-3 space-y-3">
          <CommentItem
            v-for="reply in safeReplies"
            :key="reply.id"
            :comment="reply"
            :level="props.level + 1"
            :max-level="props.maxLevel"
            :show-reply="props.showReply"
            @reply="handleChildReply"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Comment } from '@/types'

defineOptions({ name: 'CommentItem' })

const props = withDefaults(
  defineProps<{
    comment: Comment
    level?: number
    maxLevel?: number
    showReply?: boolean
  }>(),
  { level: 0, maxLevel: 4, showReply: true },
)

const emit = defineEmits<{ reply: [comment: Comment] }>()

const authorName = computed(() =>
  props.comment.author?.nickname || props.comment.author?.username || '匿名用户'
)
const nicknameFirstChar = computed(() => authorName.value.slice(0, 1))
const safeReplies = computed(() => {
  if (props.level >= props.maxLevel) return []
  return props.comment.replies || []
})
const formattedTime = computed(() => {
  if (!props.comment.createdAt) return ''
  const date = new Date(props.comment.createdAt)
  if (Number.isNaN(date.getTime())) return props.comment.createdAt
  return date.toLocaleString()
})

function handleReply() { emit('reply', props.comment) }
function handleChildReply(comment: Comment) { emit('reply', comment) }
</script>
