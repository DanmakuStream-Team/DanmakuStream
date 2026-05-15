<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
    <div class="bg-white rounded-2xl shadow-lg w-96 p-10">
      <div class="text-center mb-8">
        <div class="text-3xl font-bold text-blue-600 mb-1">Danmaku</div>
        <div class="text-gray-500 text-sm">创建你的账号</div>
      </div>

      <el-form :model="form" label-position="top" @submit.prevent="handleRegister">
        <el-form-item label="昵称">
          <el-input v-model="form.nickname" placeholder="请输入昵称" size="large" clearable />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="至少6位" size="large" show-password />
        </el-form-item>
        <el-button
          type="primary"
          native-type="submit"
          class="w-full mt-2"
          size="large"
          :loading="loading"
        >
          注册
        </el-button>
      </el-form>

      <div class="text-center mt-5 text-sm text-gray-500">
        已有账号？
        <el-link type="primary" @click="router.push('/login')">立即登录</el-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const form = reactive({ nickname: '', password: '' })

async function handleRegister() {
  try {
    loading.value = true
    await authStore.register(form.nickname, form.password)
    ElMessage.success('注册成功')
    router.push('/')
  } catch (error) {
    ElMessage.error(getErrorMessage(error, '注册失败，请重试'))
  } finally {
    loading.value = false
  }
}

function getErrorMessage(error: unknown, fallback: string) {
  if (error instanceof Error && error.message) {
    return error.message
  }
  return fallback
}
</script>
