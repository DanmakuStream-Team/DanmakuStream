import axios from 'axios'
import type { AxiosInstance, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/auth'
import type { ApiResponse } from '@/types'

declare module 'axios' {
  export interface AxiosRequestConfig {
    skipErrorMessage?: boolean
  }
}

const request: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 20000,
})

let handlingUnauthorized = false
let lastErrorMessage = ''
let lastErrorAt = 0

function showErrorOnce(message: string) {
  const now = Date.now()
  if (message === lastErrorMessage && now - lastErrorAt < 1500) return
  lastErrorMessage = message
  lastErrorAt = now
  ElMessage.error(message)
}

request.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  return config
})

request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const body = response.data
    if (body && typeof body.code === 'number') {
      if (body.code !== 0) {
        if (!response.config.skipErrorMessage) {
          showErrorOnce(body.message || '请求失败')
        }
        return Promise.reject(new Error(body.message || '请求失败'))
      }
      return { ...response, data: body.data } as any
    }
    return response
  },
  (error) => {
    if (axios.isCancel(error) || error.code === 'ERR_CANCELED') {
      return Promise.reject(error)
    }

    const message = getErrorMessage(error)
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      handleUnauthorizedOnce()
    } else if (!error.config?.skipErrorMessage) {
      showErrorOnce(message || '请求失败')
    }
    return Promise.reject(new Error(message))
  }
)

function handleUnauthorizedOnce() {
  if (handlingUnauthorized) return
  handlingUnauthorized = true

  ElMessage.warning('登录状态已失效，请重新登录')
  const redirect = `${window.location.pathname}${window.location.search}`
  void import('@/router').then(({ default: router }) => {
    if (router.currentRoute.value.name === 'Login') return
    void router.replace({
      name: 'Login',
      query: redirect === '/' ? undefined : { redirect },
    })
  })

  window.setTimeout(() => {
    handlingUnauthorized = false
  }, 2000)
}

function getErrorMessage(error: any) {
  const data = error.response?.data
  if (data?.message) return data.message
  if (typeof data === 'string') return data
  return error.message || '网络错误'
}

export default request
