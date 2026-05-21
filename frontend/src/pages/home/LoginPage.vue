<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
    <div class="bg-white rounded-2xl shadow-lg w-96 p-10">
      <div class="text-center mb-8">
        <div class="text-3xl font-bold text-blue-600 mb-1">Danmaku</div>
        <div class="text-gray-500 text-sm">登录你的账号</div>
      </div>

      <el-form :model="form" label-position="top" @submit.prevent="handleLogin">
        <el-form-item label="昵称">
          <el-input v-model="form.nickname" placeholder="请输入昵称" size="large" clearable />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" size="large" show-password />
        </el-form-item>
        <el-button
          type="primary"
          native-type="submit"
          class="w-full mt-2"
          size="large"
          :loading="loading"
        >
          登录
        </el-button>
      </el-form>

      <div class="text-center mt-5 text-sm text-gray-500">
        没有账号？
        <el-link type="primary" @click="router.push('/register')">立即注册</el-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loading = ref(false)
const form = reactive({ nickname: '', password: '' })

async function handleLogin() {
  try {
    loading.value = true
    await authStore.login(form.nickname, form.password)
    const redirect = route.query.redirect as string
    router.push(redirect || '/')
  } catch {
    ElMessage.error('昵称或密码错误')
  } finally {
    loading.value = false
  }
}
</script>
