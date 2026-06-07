import axios from 'axios'
import type { AxiosInstance, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/auth'
import type { ApiResponse } from '@/types'

const request: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 20000,
})

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
        ElMessage.error(body.message || '请求失败')
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
      ElMessage.warning('请先登录')
    } else {
      ElMessage.error(message || '请求失败')
    }
    return Promise.reject(new Error(message))
  }
)

function getErrorMessage(error: any) {
  const data = error.response?.data
  if (data?.message) return data.message
  if (typeof data === 'string') return data
  return error.message || '网络错误'
}

export default request
