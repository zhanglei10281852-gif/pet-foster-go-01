<template>
  <el-container class="layout">
    <!-- 侧边栏 -->
    <el-aside :width="collapsed ? '68px' : '240px'" class="sidebar">
      <div class="sidebar-logo">
        <span class="logo-emoji">🐾</span>
        <transition name="fade">
          <span v-show="!collapsed" class="logo-text">宠物寄养</span>
        </transition>
      </div>

      <el-menu
        :default-active="$route.path"
        :collapse="collapsed"
        router
        class="sidebar-nav"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataLine /></el-icon>
          <span>首页概览</span>
        </el-menu-item>
        <el-menu-item v-if="userStore.isAdmin" index="/user">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/pet">
          <el-icon><Chicken /></el-icon>
          <span>宠物管理</span>
        </el-menu-item>
        <el-menu-item v-if="userStore.isAdmin" index="/room">
          <el-icon><OfficeBuilding /></el-icon>
          <span>房间管理</span>
        </el-menu-item>
        <el-menu-item v-if="userStore.isAdmin" index="/service">
          <el-icon><Service /></el-icon>
          <span>服务项目</span>
        </el-menu-item>
        <el-menu-item index="/order">
          <el-icon><Document /></el-icon>
          <span>寄养订单</span>
        </el-menu-item>
        <el-menu-item index="/record">
          <el-icon><Notebook /></el-icon>
          <span>日常记录</span>
        </el-menu-item>
      </el-menu>

      <div class="sidebar-bottom">
        <div class="collapse-toggle" @click="collapsed = !collapsed">
          <el-icon><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
        </div>
      </div>
    </el-aside>

    <!-- 主区域 -->
    <el-container class="main-area">
      <el-header class="topbar">
        <div class="topbar-left">
          <h3 class="page-title">{{ $route.meta.title || '首页' }}</h3>
        </div>
        <div class="topbar-right">
          <el-dropdown @command="handleCommand" trigger="click">
            <div class="user-pill">
              <el-avatar :size="32" class="avatar">
                {{ userStore.username?.charAt(0)?.toUpperCase() }}
              </el-avatar>
              <span class="name">{{ userStore.username }}</span>
              <el-icon class="caret"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item disabled>
                  <span style="color: #94a3b8; font-size: 12px">{{ userStore.isAdmin ? '管理员' : '普通用户' }}</span>
                </el-dropdown-item>
                <el-dropdown-item command="password">
                  <el-icon><Lock /></el-icon>
                  修改密码
                </el-dropdown-item>
                <el-dropdown-item command="logout" divided>
                  <el-icon><SwitchButton /></el-icon>
                  退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="content">
        <router-view v-slot="{ Component }">
          <transition name="view" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
    
    <!-- 修改密码弹窗 -->
    <el-dialog v-model="pwdVisible" title="修改密码" width="400px">
      <el-form ref="pwdFormRef" :model="pwdForm" :rules="pwdRules" label-width="80px">
        <el-form-item label="旧密码" prop="oldPassword">
          <el-input v-model="pwdForm.oldPassword" type="password" show-password placeholder="请输入旧密码" />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input v-model="pwdForm.newPassword" type="password" show-password placeholder="请输入新密码" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="pwdForm.confirmPassword" type="password" show-password placeholder="请再次输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="pwdVisible = false">取消</el-button>
        <el-button type="primary" :loading="pwdLoading" @click="handleChangePwd">确定</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/store/user'
import { changePassword } from '@/api/user'
import { SwitchButton, Lock } from '@element-plus/icons-vue'

const router = useRouter()
const userStore = useUserStore()
const collapsed = ref(false)
const pwdVisible = ref(false)
const pwdLoading = ref(false)
const pwdFormRef = ref()
const pwdForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

const validateConfirm = (rule, value, callback) => {
  if (value !== pwdForm.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}
const pwdRules = {
  oldPassword: [{ required: true, message: '请输入旧密码', trigger: 'blur' }],
  newPassword: [{ required: true, message: '请输入新密码', trigger: 'blur' }, { min: 6, max: 20, message: '密码长度6-20位', trigger: 'blur' }],
  confirmPassword: [{ required: true, message: '请确认新密码', trigger: 'blur' }, { validator: validateConfirm, trigger: 'blur' }]
}

const handleCommand = (cmd) => {
  if (cmd === 'logout') {
    userStore.logout()
    ElMessage.success('已退出登录')
    router.push('/login')
  } else if (cmd === 'password') {
    Object.assign(pwdForm, { oldPassword: '', newPassword: '', confirmPassword: '' })
    pwdVisible.value = true
  }
}

const handleChangePwd = async () => {
  if (!(await pwdFormRef.value.validate().catch(() => false))) return
  pwdLoading.value = true
  try {
    await changePassword({ oldPassword: pwdForm.oldPassword, newPassword: pwdForm.newPassword })
    ElMessage.success('密码修改成功')
    pwdVisible.value = false
  } catch (e) {
    ElMessage.error(e.response?.data?.message || e.message || '修改失败')
  } finally {
    pwdLoading.value = false
  }
}
</script>

<style lang="scss" scoped>
$primary: #6366f1;
$gray-50: #f8fafc;
$gray-100: #f1f5f9;
$gray-200: #e2e8f0;
$gray-400: #94a3b8;
$gray-800: #1e293b;
$gray-900: #0f172a;
$radius: 12px;
$ease: cubic-bezier(0.16, 1, 0.3, 1);

.layout { height: 100vh; }

// --- 侧边栏 ---
.sidebar {
  background: linear-gradient(180deg, $gray-900 0%, #0c0f1a 100%);
  display: flex;
  flex-direction: column;
  transition: width 0.3s $ease;
  overflow: hidden;
  border-right: 1px solid rgba(255,255,255,0.06);

  .sidebar-logo {
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    padding: 0 16px;
    border-bottom: 1px solid rgba(255,255,255,0.06);

    .logo-emoji { font-size: 26px; flex-shrink: 0; }
    .logo-text {
      font-size: 16px;
      font-weight: 700;
      color: white;
      white-space: nowrap;
    }
  }

  .sidebar-nav {
    flex: 1;
    background: transparent;
    border-right: none;
    padding: 12px 8px;

    :deep(.el-menu-item) {
      height: 44px;
      line-height: 44px;
      margin: 2px 0;
      border-radius: 10px;
      color: rgba(255,255,255,0.55);
      font-size: 13px;
      font-weight: 500;
      transition: all 0.2s $ease;

      .el-icon { font-size: 18px; margin-right: 10px; }

      &:hover {
        background: rgba(255,255,255,0.06);
        color: rgba(255,255,255,0.9);
      }
      &.is-active {
        background: $primary;
        color: white;
        box-shadow: 0 2px 12px rgba($primary, 0.4);
      }
    }

    &.el-menu--collapse :deep(.el-menu-item) {
      padding: 0 !important;
      justify-content: center;
      .el-icon { margin-right: 0; }
    }
  }

  .sidebar-bottom {
    padding: 12px 8px;
    border-top: 1px solid rgba(255,255,255,0.06);

    .collapse-toggle {
      height: 36px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 8px;
      color: rgba(255,255,255,0.4);
      cursor: pointer;
      transition: all 0.2s;
      &:hover { background: rgba(255,255,255,0.06); color: white; }
      .el-icon { font-size: 16px; }
    }
  }
}

// --- 顶栏 ---
.topbar {
  height: 64px;
  background: white;
  border-bottom: 1px solid $gray-200;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28px;

  .page-title {
    font-size: 16px;
    font-weight: 600;
    color: $gray-800;
    margin: 0;
  }

  .user-pill {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 4px 10px 4px 4px;
    border-radius: 10px;
    cursor: pointer;
    transition: background 0.2s;
    &:hover { background: $gray-100; }

    .avatar {
      background: $primary;
      color: white;
      font-weight: 600;
      font-size: 13px;
    }
    .name {
      font-size: 13px;
      font-weight: 500;
      color: $gray-800;
    }
    .caret {
      font-size: 12px;
      color: $gray-400;
    }
  }
}

// --- 内容区 ---
.content {
  background: $gray-50;
  padding: 24px;
  overflow-y: auto;
}

// --- 动画 ---
.view-enter-active, .view-leave-active { transition: all 0.25s $ease; }
.view-enter-from { opacity: 0; transform: translateY(6px); }
.view-leave-to   { opacity: 0; transform: translateY(-6px); }

.fade-enter-active, .fade-leave-active { transition: opacity 0.2s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
