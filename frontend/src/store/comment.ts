import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Comment } from '@/types'
import { commentApi } from '@/api/comment'

export const useCommentStore = defineStore('comment', () => {
  const comments = ref<Comment[]>([])
  const loading = ref(false)
  const submitting = ref(false)

  async function fetchComments(videoId: number) {
    loading.value = true
    try {
      const res = await commentApi.getComments(videoId)
      comments.value = res.data ?? []
    } finally {
      loading.value = false
    }
  }

  async function createComment(videoId: number, content: string, parentId?: number) {
    submitting.value = true
    try {
      const res = await commentApi.createComment({ videoId, content, parentId })
      const comment = res.data

      if (parentId) {
        const parent = findComment(comments.value, parentId)
        if (parent) {
          parent.replies = [...(parent.replies ?? []), comment]
        }
      } else {
        comments.value = [comment, ...comments.value]
      }

      return comment
    } finally {
      submitting.value = false
    }
  }

  function countComments() {
    return comments.value.reduce((total, comment) => {
      return total + 1 + countReplies(comment.replies ?? [])
    }, 0)
  }

  function clearComments() {
    comments.value = []
  }

  return { comments, loading, submitting, fetchComments, createComment, countComments, clearComments }
})

function findComment(comments: Comment[], id: number): Comment | undefined {
  for (const comment of comments) {
    if (comment.id === id) {
      return comment
    }

    const reply = findComment(comment.replies ?? [], id)
    if (reply) {
      return reply
    }
  }
}

function countReplies(replies: Comment[]): number {
  return replies.reduce((total, reply) => total + 1 + countReplies(reply.replies ?? []), 0)
}
