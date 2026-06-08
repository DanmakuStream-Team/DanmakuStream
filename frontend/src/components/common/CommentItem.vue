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
        <el-button text size="small" @click="toggleReplyBox">回复</el-button>
        <el-button text size="small" @click="$emit('like', comment)">
          {{ comment.liked ? '已赞' : '点赞' }} {{ comment.likeCount || '' }}
        </el-button>
      </div>
      <div v-if="replying" class="inline-reply">
        <el-input
          v-model="replyText"
          type="textarea"
          :rows="2"
          :placeholder="`回复 ${comment.author?.nickname || '用户'}`"
        />
        <div class="inline-reply-actions">
          <el-button size="small" @click="cancelReply">取消</el-button>
          <el-button type="primary" size="small" :loading="submitting" @click="submitReply">发送回复</el-button>
        </div>
      </div>
      <div v-if="comment.replies?.length" class="replies">
        <CommentItem
          v-for="reply in visibleReplies"
          :key="reply.id"
          :comment="reply"
          @reply="(target, content) => $emit('reply', target, content)"
          @like="$emit('like', $event)"
        />
        <el-button v-if="hasHiddenReplies" class="reply-toggle" text size="small" @click="repliesExpanded = true">
          展开 {{ comment.replies.length - defaultReplyCount }} 条回复
        </el-button>
        <el-button v-else-if="comment.replies.length > defaultReplyCount" class="reply-toggle" text size="small" @click="repliesExpanded = false">
          收起回复
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import type { Comment } from '@/types'
import { formatTime, mediaUrl } from '@/utils/format'

const props = defineProps<{ comment: Comment }>()
const emit = defineEmits<{ reply: [comment: Comment, content: string]; like: [comment: Comment] }>()

const defaultReplyCount = 2
const repliesExpanded = ref(false)
const replying = ref(false)
const submitting = ref(false)
const replyText = ref('')
const visibleReplies = computed(() => {
  const replies = props.comment.replies || []
  return repliesExpanded.value ? replies : replies.slice(0, defaultReplyCount)
})
const hasHiddenReplies = computed(() => {
  return !repliesExpanded.value && (props.comment.replies?.length || 0) > defaultReplyCount
})

function toggleReplyBox() {
  replying.value = !replying.value
}

function cancelReply() {
  replying.value = false
  replyText.value = ''
}

async function submitReply() {
  const content = replyText.value.trim()
  if (!content) {
    ElMessage.warning('请输入回复内容')
    return
  }
  submitting.value = true
  try {
    emit('reply', props.comment, content)
    cancelReply()
  } finally {
    submitting.value = false
  }
}
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

.inline-reply {
  display: grid;
  gap: 8px;
  margin: 8px 0 10px;
}

.inline-reply-actions {
  display: flex;
  justify-content: flex-end;
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

.reply-toggle {
  justify-self: start;
}
</style>
