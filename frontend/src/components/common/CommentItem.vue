<template>
  <div class="comment-item">
    <div class="comment-avatar">
      <img :src="comment.avatar || defaultAvatar" alt="avatar" />
    </div>
    <div class="comment-content">
      <div class="comment-user">{{ comment.username || '匿名用户' }}</div>
      <div class="comment-text">{{ comment.content }}</div>
      <div class="comment-actions">
        <span @click="$emit('like')">👍 {{ comment.likeCount || 0 }}</span>
        <span @click="$emit('reply')">💬 回复</span>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps(['comment'])
defineEmits(['like', 'reply'])

const defaultAvatar = 'https://via.placeholder.com/40'
</script>

<style scoped>
.comment-item {
  display: flex;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #eee;
}
.comment-avatar img {
  width: 40px;
  height: 40px;
  border-radius: 50%;
}
.comment-content {
  flex: 1;
}
.comment-user {
  font-weight: bold;
  margin-bottom: 4px;
}
.comment-text {
  margin-bottom: 8px;
}
.comment-actions span {
  margin-right: 12px;
  cursor: pointer;
  color: #666;
}
</style>