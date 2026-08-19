<template>
  <div class="login-page">
    <!-- 左侧品牌区 -->
    <div class="brand-side">
      <div class="brand-content">
        <div class="brand-logo">🐾</div>
        <h1>宠物寄养<br/>管理系统</h1>
        <p>为每一只毛孩子提供温暖的家</p>
        <div class="brand-features">
          <div class="feature"><span class="dot"></span>智能房间管理</div>
          <div class="feature"><span class="dot"></span>订单全流程追踪</div>
          <div class="feature"><span class="dot"></span>每日健康记录</div>
        </div>
      </div>
      <div class="brand-footer">Pet Foster Management System v1.0</div>
    </div>

    <!-- 右侧表单区 -->
    <div class="form-side">
      <div class="form-wrapper">
        <div class="form-header">
          <h2>{{ activeTab === 'login' ? '欢迎回来' : '创建账号' }}</h2>
          <p>{{ activeTab === 'login' ? '登录您的账号以继续' : '注册一个新账号' }}</p>
        </div>

        <!-- 测试账号快捷入口 -->
        <div class="quick-accounts">
          <div class="quick-item" @click="fillAccount('admin', 'admin123')">
            <span class="badge admin">管理员</span>
            <span class="cred">admin / admin123</span>
          </div>
          <div class="quick-item" @click="fillAccount('testuser', 'user123')">
            <span class="badge user">用户</span>
            <span class="cred">testuser / user123</span>
          </div>
        </div>

        <el-tabs v-model="activeTab" class="auth-tabs">
          <el-tab-pane label="登录" name="login">
            <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" @keyup.enter="handleLogin">
              <el-form-item prop="username">
                <el-input v-model="loginForm.username" placeholder="用户名" prefix-icon="User" size="large" />
              </el-form-item>
              <el-form-item prop="password">
                <el-input v-model="loginForm.password" type="password" placeholder="密码" prefix-icon="Lock" size="large" show-password />
              </el-form-item>
              <el-button type="primary" size="large" :loading="loading" @click="handleLogin" class="auth-btn">
                {{ loading ? '登录中...' : '登 录' }}
              </el-button>
            </el-form>
          </el-tab-pane>

          <el-tab-pane label="注册" name="register">
            <el-form ref="registerFormRef" :model="registerForm" :rules="registerRules">
              <el-form-item prop="username">
                <el-input v-model="registerForm.username" placeholder="用户名" prefix-icon="User" size="large" />
              </el-form-item>
              <el-form-item prop="password">
                <el-input v-model="registerForm.password" type="password" placeholder="密码" prefix-icon="Lock" size="large" show-password />
              </el-form-item>
              <el-form-item prop="phone">
                <el-input v-model="registerForm.phone" placeholder="手机号" prefix-icon="Phone" size="large" />
              </el-form-item>
              <el-form-item prop="email">
                <el-input v-model="registerForm.email" placeholder="邮箱" prefix-icon="Message" size="large" />
              </el-form-item>
              <el-button type="primary" size="large" :loading="loading" @click="handleRegister" class="auth-btn">
                {{ loading ? '注册中...' : '注 册' }}
              </el-button>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/store/user'
import { register } from '@/api/user'

const router = useRouter()
const userStore = useUserStore()
const activeTab = ref('login')
const loading = ref(false)
const loginFormRef = ref()
const registerFormRef = ref()

const loginForm = reactive({ username: '', password: '' })
const registerForm = reactive({ username: '', password: '', phone: '', email: '' })

const loginRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}
const registerRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '长度在3-20之间', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '长度在6-20之间', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  if (!(await loginFormRef.value.validate().catch(() => false))) return
  loading.value = true
  try {
    await userStore.login(loginForm)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } finally { loading.value = false }
}

const handleRegister = async () => {
  if (!(await registerFormRef.value.validate().catch(() => false))) return
  loading.value = true
  try {
    await register(registerForm)
    ElMessage.success('注册成功，请登录')
    activeTab.value = 'login'
    loginForm.username = registerForm.username
    loginForm.password = ''
  } finally { loading.value = false }
}

const fillAccount = (u, p) => {
  loginForm.username = u
  loginForm.password = p
  activeTab.value = 'login'
  ElMessage.success('已填入，点击登录即可')
}
</script>

<style lang="scss" scoped>
$primary: #6366f1;
$accent: #8b5cf6;
$gray-50: #f8fafc;
$gray-100: #f1f5f9;
$gray-200: #e2e8f0;
$gray-300: #cbd5e1;
$gray-400: #94a3b8;
$gray-500: #64748b;
$gray-800: #1e293b;
$gray-900: #0f172a;
$radius: 12px;

.login-page {
  display: flex;
  min-height: 100vh;
}

// --- 左侧品牌 ---
.brand-side {
  flex: 0 0 480px;
  background: linear-gradient(160deg, $gray-900 0%, #1a1040 50%, #3730a3 100%);
  color: white;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 60px;
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    width: 500px; height: 500px;
    border-radius: 50%;
    background: radial-gradient(circle, rgba($primary, 0.15) 0%, transparent 70%);
    top: -100px; right: -150px;
  }
  &::after {
    content: '';
    position: absolute;
    width: 300px; height: 300px;
    border-radius: 50%;
    background: radial-gradient(circle, rgba($accent, 0.1) 0%, transparent 70%);
    bottom: -50px; left: -80px;
  }

  .brand-content {
    position: relative;
    z-index: 1;

    .brand-logo {
      font-size: 48px;
      margin-bottom: 24px;
      width: 80px; height: 80px;
      background: rgba(255,255,255,0.1);
      border-radius: 20px;
      display: flex;
      align-items: center;
      justify-content: center;
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255,255,255,0.1);
    }

    h1 {
      font-size: 36px;
      font-weight: 800;
      line-height: 1.2;
      margin-bottom: 12px;
      letter-spacing: -0.5px;
    }

    > p {
      font-size: 16px;
      color: rgba(255,255,255,0.6);
      margin-bottom: 40px;
    }
  }

  .brand-features {
    display: flex;
    flex-direction: column;
    gap: 14px;

    .feature {
      display: flex;
      align-items: center;
      gap: 12px;
      font-size: 14px;
      color: rgba(255,255,255,0.7);

      .dot {
        width: 8px; height: 8px;
        border-radius: 50%;
        background: $primary;
        box-shadow: 0 0 10px rgba($primary, 0.5);
      }
    }
  }

  .brand-footer {
    position: absolute;
    bottom: 32px;
    left: 60px;
    font-size: 12px;
    color: rgba(255,255,255,0.3);
  }
}

// --- 右侧表单 ---
.form-side {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: $gray-50;
  padding: 40px;
}

.form-wrapper {
  width: 100%;
  max-width: 420px;
  animation: slideIn 0.5s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes slideIn {
  from { opacity: 0; transform: translateX(20px); }
  to   { opacity: 1; transform: translateX(0); }
}

.form-header {
  margin-bottom: 28px;

  h2 {
    font-size: 28px;
    font-weight: 700;
    color: $gray-900;
    margin: 0 0 6px;
    letter-spacing: -0.5px;
  }
  p {
    font-size: 15px;
    color: $gray-400;
    margin: 0;
  }
}

// --- 测试账号快捷入口 ---
.quick-accounts {
  display: flex;
  gap: 10px;
  margin-bottom: 24px;

  .quick-item {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 14px;
    background: white;
    border: 1.5px solid $gray-200;
    border-radius: $radius;
    cursor: pointer;
    transition: all 0.2s;

    &:hover {
      border-color: $primary;
      background: rgba($primary, 0.03);
      transform: translateY(-1px);
      box-shadow: 0 4px 12px rgba($primary, 0.1);
    }

    .badge {
      font-size: 11px;
      font-weight: 600;
      padding: 2px 8px;
      border-radius: 6px;
      white-space: nowrap;

      &.admin { background: #fef3c7; color: #92400e; }
      &.user  { background: #ede9fe; color: #5b21b6; }
    }

    .cred {
      font-size: 12px;
      font-family: 'JetBrains Mono', 'SF Mono', monospace;
      color: $gray-500;
    }
  }
}

// --- Tabs ---
.auth-tabs {
  :deep(.el-tabs__header) { margin-bottom: 24px; }
  :deep(.el-tabs__nav-wrap::after) { display: none; }
  :deep(.el-tabs__nav) { width: 100%; display: flex; }
  :deep(.el-tabs__item) {
    flex: 1;
    text-align: center;
    font-size: 14px;
    font-weight: 600;
    color: $gray-400;
    padding: 0 0 10px;
    transition: color 0.2s;
    &.is-active { color: $primary; }
    &:hover { color: $primary; }
  }
  :deep(.el-tabs__active-bar) {
    background: $primary;
    height: 2.5px;
    border-radius: 2px;
  }

  :deep(.el-form-item) { margin-bottom: 18px; }

  :deep(.el-input__wrapper) {
    border-radius: 10px;
    padding: 4px 14px;
    box-shadow: none;
    border: 1.5px solid $gray-200;
    transition: all 0.2s;
    &:hover { border-color: $gray-300; }
    &.is-focus {
      border-color: $primary;
      box-shadow: 0 0 0 3px rgba($primary, 0.08);
    }
  }
  :deep(.el-input__inner) { height: 42px; font-size: 14px; }
  :deep(.el-input__prefix) { color: $gray-400; }
}

.auth-btn {
  width: 100%;
  height: 48px;
  font-size: 15px;
  font-weight: 600;
  border-radius: 10px;
  background: $primary;
  border: none;
  margin-top: 4px;
  transition: all 0.2s;

  &:hover {
    background: #4f46e5;
    box-shadow: 0 0 20px rgba($primary, 0.35);
  }
}

// --- 响应式 ---
@media (max-width: 960px) {
  .login-page { flex-direction: column; }
  .brand-side {
    flex: none;
    padding: 40px 32px;
    h1 { font-size: 28px; br { display: none; } }
    .brand-features { display: none; }
    .brand-footer { display: none; }
  }
}
</style>
