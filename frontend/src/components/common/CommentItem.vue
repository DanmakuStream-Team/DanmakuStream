<template>
  <div class="comment-item" :class="{ reply: depth > 0 }">
    <button class="avatar-button" type="button" :disabled="!comment.author?.id" @click="openAuthor">
      <el-avatar :size="depth > 0 ? 30 : 42" :src="mediaUrl(comment.author?.avatar)">
        {{ comment.author?.nickname?.slice(0, 1) || 'U' }}
      </el-avatar>
    </button>
    <div class="content">
      <div class="line">
        <button class="name-button" type="button" :disabled="!comment.author?.id" @click="openAuthor">
          <strong>{{ comment.author?.nickname || '匿名用户' }}</strong>
        </button>
        <span>{{ formatTime(comment.createdAt) }}</span>
      </div>
      <p>
        <template v-if="depth > 0 && replyTo">
          <span class="reply-prefix">回复</span>
          <button class="reply-target" type="button" @click="openReplyTarget">@{{ replyTo.nickname || replyTo.username }}</button>
          <span class="reply-prefix">：</span>
        </template>
        {{ comment.content }}
      </p>
      <div class="ops">
        <button type="button" @click="$emit('like', comment)">
          {{ comment.liked ? '已赞' : '点赞' }}{{ comment.likeCount ? ` ${comment.likeCount}` : '' }}
        </button>
        <button type="button" @click="$emit('reply', comment)">回复</button>
      </div>
      <div v-if="depth === 0 && flattenedReplies.length" class="replies">
        <CommentItem
          v-for="reply in flattenedReplies"
          :key="reply.comment.id"
          :comment="reply.comment"
          :depth="1"
          :reply-to="reply.replyTo"
          @reply="$emit('reply', $event)"
          @like="$emit('like', $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { Comment, UserInfo } from '@/types'
import { formatTime, mediaUrl } from '@/utils/format'

interface FlatReply {
  comment: Comment
  replyTo?: UserInfo
}

const props = withDefaults(defineProps<{ comment: Comment; depth?: number; replyTo?: UserInfo }>(), {
  depth: 0,
})
defineEmits<{ reply: [comment: Comment]; like: [comment: Comment] }>()
const router = useRouter()
const flattenedReplies = computed(() => flattenReplies(props.comment.replies || []))

function flattenReplies(items: Comment[], parentAuthor = props.comment.author): FlatReply[] {
  return items.flatMap((item) => [
    { comment: item, replyTo: parentAuthor },
    ...flattenReplies(item.replies || [], item.author),
  ])
}

function openAuthor() {
  if (!props.comment.author?.id) return
  router.push(`/user/${props.comment.author.id}`)
}

function openReplyTarget() {
  if (!props.replyTo?.id) return
  router.push(`/user/${props.replyTo.id}`)
}
</script>

<style scoped>
.comment-item {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 18px 0;
  border-bottom: 1px solid #f1f2f3;
}

.comment-item.reply {
  gap: 10px;
  padding: 10px 0;
  border-bottom: 0;
}

.avatar-button,
.name-button {
  display: inline-flex;
  flex-shrink: 0;
  padding: 0;
  border: 0;
  background: transparent;
  cursor: pointer;
}

.avatar-button:disabled,
.name-button:disabled {
  cursor: default;
}

.content {
  min-width: 0;
  flex: 1;
  padding-top: 0;
}

.line {
  display: flex;
  align-items: center;
  gap: 9px;
  min-height: 30px;
}

.comment-item:not(.reply) .line {
  min-height: 42px;
}

.line strong {
  color: #61666d;
  font-size: 14px;
  font-weight: 600;
}

.name-button:not(:disabled):hover strong {
  color: #00aeec;
}

.line span {
  color: #9499a0;
  font-size: 13px;
}

p {
  margin: 6px 0 8px;
  color: #18191c;
  font-size: 15px;
  line-height: 1.65;
  word-break: break-word;
}

.reply p {
  margin: 4px 0 7px;
  font-size: 14px;
}

.reply-prefix {
  color: #18191c;
  font-weight: 700;
}

.reply-target {
  padding: 0;
  border: 0;
  background: transparent;
  color: #00aeec;
  cursor: pointer;
  font: inherit;
  font-weight: 700;
}

.reply-target:hover {
  color: #fb7299;
}

.ops {
  display: flex;
  align-items: center;
  gap: 22px;
}

.ops button {
  padding: 0;
  border: 0;
  background: transparent;
  color: #9499a0;
  cursor: pointer;
  font-size: 13px;
}

.ops button:hover {
  color: #00aeec;
}

.replies {
  display: grid;
  gap: 2px;
  margin-top: 8px;
  padding-left: 0;
}
</style>
