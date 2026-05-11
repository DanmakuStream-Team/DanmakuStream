<template>
  <div class="login-page">
    <div class="login-card">
      <h2>登录</h2>
      <a-form :model="form" layout="vertical" @submit="handleLogin">
        <a-form-item field="nickname" label="昵称" :rules="[{ required: true }]">
          <a-input v-model="form.nickname" placeholder="请输入昵称" />
        </a-form-item>
        <a-form-item field="password" label="密码" :rules="[{ required: true }]">
          <a-input-password v-model="form.password" placeholder="请输入密码" />
        </a-form-item>
        <a-button type="primary" html-type="submit" long :loading="loading">登录</a-button>
      </a-form>
      <div class="footer-link">
        没有账号？<a-link @click="router.push('/register')">立即注册</a-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { Message } from '@arco-design/web-vue'

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
    Message.error('昵称或密码错误')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f6fa;
}
.login-card {
  background: #fff;
  padding: 40px;
  border-radius: 12px;
  width: 360px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
}
h2 { margin-bottom: 24px; text-align: center; }
.footer-link { text-align: center; margin-top: 16px; color: #86909c; }
</style>
