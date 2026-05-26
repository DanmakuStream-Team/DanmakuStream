<template>
  <div class="comment-item">
    <el-avatar :size="36" :src="mediaUrl(comment.author?.avatar)">
      {{ comment.author?.nickname?.slice(0, 1) || 'U' }}
    </el-avatar>
    <div class="content">
      <div class="line">
        <strong>{{ comment.author?.nickname || '匿名用户' }}</strong>
        <span>{{ formatTime(comment.createdAt) }}</span>
      </div>
      <p>{{ comment.content }}</p>
      <div class="ops">
        <el-button text size="small" @click="$emit('reply', comment)">回复</el-button>
        <el-button text size="small" @click="$emit('like', comment)">
          {{ comment.liked ? '已赞' : '点赞' }} {{ comment.likeCount || '' }}
        </el-button>
      </div>
      <div v-if="comment.replies?.length" class="replies">
        <CommentItem
          v-for="reply in comment.replies"
          :key="reply.id"
          :comment="reply"
          @reply="$emit('reply', $event)"
          @like="$emit('like', $event)"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Comment } from '@/types'
import { formatTime, mediaUrl } from '@/utils/format'

defineProps<{ comment: Comment }>()
defineEmits<{ reply: [comment: Comment]; like: [comment: Comment] }>()
</script>

<style scoped>
.comment-item {
  display: flex;
  gap: 12px;
}

.content {
  min-width: 0;
  flex: 1;
}

.line {
  display: flex;
  align-items: center;
  gap: 10px;
}

.line strong {
  color: #142033;
  font-size: 14px;
}

.line span {
  color: #98a2b3;
  font-size: 12px;
}

p {
  margin: 6px 0;
  color: #475467;
  line-height: 1.7;
}

.ops {
  display: flex;
  gap: 8px;
}

.replies {
  display: grid;
  gap: 14px;
  margin-top: 12px;
  padding: 12px;
  border-radius: 8px;
  background: #f7f9fc;
}
</style>
