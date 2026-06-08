import { ref } from 'vue'
import { defineStore } from 'pinia'
import { commentApi } from '@/api/comment'
import type { Comment } from '@/types'

export const useCommentStore = defineStore('comment', () => {
  const comments = ref<Comment[]>([])
  const loading = ref(false)
  const submitting = ref(false)

  async function fetchComments(videoId: number, params?: { sort?: 'date' | 'like' }) {
    loading.value = true
    try {
      const res = await commentApi.list(videoId, params)
      comments.value = res.data
    } finally {
      loading.value = false
    }
  }

  async function createComment(videoId: number, content: string, parentId?: number) {
    submitting.value = true
    try {
      const res = await commentApi.create({ videoId, content, parentId })
      if (parentId) {
        const parent = findComment(comments.value, parentId)
        if (parent) parent.replies = [...(parent.replies || []), res.data]
      } else {
        comments.value = [res.data, ...comments.value]
      }
      return res.data
    } finally {
      submitting.value = false
    }
  }

  function countComments() {
    return comments.value.reduce((sum, item) => sum + 1 + countReplies(item.replies || []), 0)
  }

  function clearComments() {
    comments.value = []
  }

  return { comments, loading, submitting, fetchComments, createComment, countComments, clearComments }
})

function findComment(list: Comment[], id: number): Comment | undefined {
  for (const item of list) {
    if (item.id === id) return item
    const child = findComment(item.replies || [], id)
    if (child) return child
  }
}

function countReplies(list: Comment[]): number {
  return list.reduce((sum, item) => sum + 1 + countReplies(item.replies || []), 0)
}
