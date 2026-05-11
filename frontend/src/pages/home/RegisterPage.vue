<template>
  <div class="register-page">
    <div class="register-card">
      <h2>注册</h2>
      <a-form :model="form" layout="vertical" @submit="handleRegister">
        <a-form-item field="nickname" label="昵称" :rules="[{ required: true }]">
          <a-input v-model="form.nickname" placeholder="请输入昵称" />
        </a-form-item>
        <a-form-item field="password" label="密码" :rules="[{ required: true, minLength: 6 }]">
          <a-input-password v-model="form.password" placeholder="至少6位" />
        </a-form-item>
        <a-button type="primary" html-type="submit" long :loading="loading">注册</a-button>
      </a-form>
      <div class="footer-link">
        已有账号？<a-link @click="router.push('/login')">立即登录</a-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { Message } from '@arco-design/web-vue'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)
const form = reactive({ nickname: '', password: '' })

async function handleRegister() {
  try {
    loading.value = true
    await authStore.register(form.nickname, form.password)
    Message.success('注册成功')
    router.push('/')
  } catch (error) {
    Message.error(getErrorMessage(error, '注册失败，请重试'))
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

<style scoped>
.register-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f6fa;
}
.register-card {
  background: #fff;
  padding: 40px;
  border-radius: 12px;
  width: 360px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
}
h2 { margin-bottom: 24px; text-align: center; }
.footer-link { text-align: center; margin-top: 16px; color: #86909c; }
</style>
