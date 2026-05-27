<template>
  <main class="auth-page">
    <section class="auth-card soft-panel">
      <el-tag type="primary">欢迎回来</el-tag>
      <h1>登录账号</h1>
      <p>后端当前使用昵称和密码登录。</p>
      <el-form label-position="top" @submit.prevent>
        <el-form-item label="昵称">
          <el-input v-model="form.nickname" placeholder="请输入昵称" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" />
        </el-form-item>
        <el-button type="primary" size="large" :loading="loading" @click="submit" class="wide">登录</el-button>
      </el-form>
      <div class="switch">
        没有账号？
        <el-link type="primary" @click="router.push('/register')">立即注册</el-link>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const loading = ref(false)
const form = reactive({ nickname: '', password: '' })

async function submit() {
  if (!form.nickname.trim() || !form.password) {
    ElMessage.warning('请填写昵称和密码')
    return
  }
  loading.value = true
  try {
    await authStore.login(form.nickname.trim(), form.password)
    ElMessage.success('登录成功')
    router.push(String(route.query.redirect || '/'))
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  display: grid;
  min-height: 100vh;
  place-items: center;
  padding: 24px;
}

.auth-card {
  width: min(420px, 100%);
  padding: 28px;
}

h1 {
  margin: 14px 0 8px;
  color: #142033;
  font-size: 30px;
}

p {
  margin: 0 0 24px;
  color: #667085;
}

.wide {
  width: 100%;
}

.switch {
  margin-top: 18px;
  text-align: center;
  color: #667085;
}
</style>
