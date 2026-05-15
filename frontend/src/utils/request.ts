import axios from 'axios'
import type { AxiosInstance, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/auth'

interface WrappedResponse<T = unknown> {
  code: number
  message?: string
  data?: T
}

const request: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

// Request interceptor: inject JWT token
request.interceptors.request.use((config) => {
  const authStore = useAuthStore()
  if (authStore.token) {
    config.headers.Authorization = `Bearer ${authStore.token}`
  }
  return config
})

// Response interceptor: normalize and handle errors
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const body = response.data

    if (isWrappedResponse(body)) {
      if (body.code !== 0) {
        ElMessage.error(body.message || '请求失败')
        return Promise.reject(new Error(body.message || '请求失败'))
      }

      return { ...response, data: body.data } as any
    }

    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      const authStore = useAuthStore()
      authStore.logout()
      window.location.href = '/login'
    } else {
      ElMessage.error(getErrorMessage(error, '网络错误'))
    }
    return Promise.reject(new Error(getErrorMessage(error, '网络错误')))
  }
)

function isWrappedResponse(value: unknown): value is WrappedResponse {
  return (
    !!value &&
    typeof value === 'object' &&
    'code' in value &&
    typeof (value as WrappedResponse).code === 'number'
  )
}

function getErrorMessage(error: unknown, fallback: string) {
  if (!axios.isAxiosError(error)) {
    return fallback
  }

  const data = error.response?.data
  if (data && typeof data === 'object') {
    const message = (data as { message?: unknown; error?: unknown }).message ?? (data as { error?: unknown }).error
    if (typeof message === 'string' && message) {
      return message
    }
  }

  if (typeof data === 'string' && data) {
    return data
  }

  return fallback
}

export default request
