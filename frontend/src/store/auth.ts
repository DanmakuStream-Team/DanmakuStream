import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserInfo } from '@/types'
import { authApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => userInfo.value?.role === 'admin')
  const isCreator = computed(() => userInfo.value?.role === 'creator' || isAdmin.value)

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('token', t)
  }

  function setUserInfo(info: UserInfo) {
    userInfo.value = info
  }

  async function login(nickname: string, password: string) {
    const res = await authApi.login({ nickname, password })
    setToken(res.data.token)
    setUserInfo(res.data.userInfo)
  }

  async function register(nickname: string, password: string) {
    const res = await authApi.register({ nickname, password })
    setToken(res.data.token)
    setUserInfo(res.data.userInfo)
  }

  async function fetchUserInfo() {
    const res = await authApi.getUserInfo()
    setUserInfo(res.data)
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
  }

  return { token, userInfo, isLoggedIn, isAdmin, isCreator, login, register, logout, fetchUserInfo }
})
