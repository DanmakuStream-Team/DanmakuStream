import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { authApi } from '@/api/auth'
import type { UserInfo } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(readUser())

  const isLoggedIn = computed(() => Boolean(token.value))
  const isAdmin = computed(() => userInfo.value?.role === 'admin')
  const isCreator = computed(() => userInfo.value?.role === 'creator' || isAdmin.value)

  async function login(nickname: string, password: string) {
    const res = await authApi.login({ nickname, password })
    setSession(res.data.token, res.data.userInfo)
  }

  async function register(nickname: string, password: string) {
    const res = await authApi.register({ nickname, password })
    setSession(res.data.token, res.data.userInfo)
  }

  async function fetchUserInfo() {
    if (!token.value) return
    const res = await authApi.me()
    userInfo.value = res.data
    localStorage.setItem('userInfo', JSON.stringify(res.data))
  }

  function setSession(nextToken: string, nextUser: UserInfo) {
    token.value = nextToken
    userInfo.value = nextUser
    localStorage.setItem('token', nextToken)
    localStorage.setItem('userInfo', JSON.stringify(nextUser))
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
  }

  return { token, userInfo, isLoggedIn, isAdmin, isCreator, login, register, fetchUserInfo, logout }
})

function readUser() {
  const raw = localStorage.getItem('userInfo')
  if (!raw) return null
  try {
    return JSON.parse(raw) as UserInfo
  } catch {
    return null
  }
}
