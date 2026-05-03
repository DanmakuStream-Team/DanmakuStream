<template>
  <div
    class="comment-item"
    :class="{ 'is-child': props.level > 0 }"
  >
    <div class="comment-main">
      <a-avatar :size="avatarSize">
        <img
          v-if="props.comment.author?.avatar"
          :src="props.comment.author.avatar"
          alt="avatar"
        />
        <span v-else>
          {{ nicknameFirstChar }}
        </span>
      </a-avatar>

      <div class="comment-body">
        <div class="comment-user">
          {{ authorName }}
        </div>

        <div class="comment-content">
          {{ props.comment.content }}
        </div>

        <div class="comment-meta">
          <span>{{ formattedTime }}</span>

          <span v-if="props.comment.likeCount > 0">
            {{ props.comment.likeCount }} 赞
          </span>

          <a-button
            v-if="props.showReply"
            type="text"
            size="mini"
            @click="handleReply"
          >
            回复
          </a-button>
        </div>

        <div
          v-if="safeReplies.length > 0"
          class="comment-replies"
        >
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

defineOptions({
  name: 'CommentItem',
})

const props = withDefaults(
  defineProps<{
    comment: Comment
    level?: number
    maxLevel?: number
    showReply?: boolean
  }>(),
  {
    level: 0,
    maxLevel: 4,
    showReply: true,
  },
)

const emit = defineEmits<{
  reply: [comment: Comment]
}>()

const authorName = computed(() => {
  return props.comment.author?.nickname || props.comment.author?.username || '匿名用户'
})

const nicknameFirstChar = computed(() => {
  return authorName.value.slice(0, 1)
})

const avatarSize = computed(() => {
  return props.level > 0 ? 32 : 40
})

const safeReplies = computed(() => {
  if (props.level >= props.maxLevel) {
    return []
  }

  return props.comment.replies || []
})

const formattedTime = computed(() => {
  if (!props.comment.createdAt) {
    return ''
  }

  const date = new Date(props.comment.createdAt)

  if (Number.isNaN(date.getTime())) {
    return props.comment.createdAt
  }

  return date.toLocaleString()
})

function handleReply() {
  emit('reply', props.comment)
}

function handleChildReply(comment: Comment) {
  emit('reply', comment)
}
</script>

<style scoped>
.comment-item {
  width: 100%;
}

.comment-item.is-child {
  padding-top: 12px;
}

.comment-main {
  display: flex;
  gap: 12px;
  width: 100%;
}

.comment-body {
  flex: 1;
  min-width: 0;
}

.comment-user {
  color: #1d2129;
  font-weight: 600;
  margin-bottom: 4px;
}

.comment-content {
  color: #1d2129;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.comment-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #86909c;
  font-size: 13px;
  margin-top: 6px;
}

.comment-replies {
  margin-top: 14px;
  padding: 12px;
  background: #f7f8fa;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>