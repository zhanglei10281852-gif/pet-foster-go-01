<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-info">
        <h2>用户管理</h2>
        <p>管理系统中的所有用户账户</p>
      </div>
    </div>

    <div class="search-bar">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="用户名">
          <el-input v-model="searchForm.username" placeholder="请输入" clearable />
        </el-form-item>
        <el-form-item label="手机号">
          <el-input v-model="searchForm.phone" placeholder="请输入" clearable />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="searchForm.role" placeholder="全部" clearable>
            <el-option label="管理员" value="ADMIN" />
            <el-option label="普通用户" value="USER" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="table-container">
      <el-table :data="tableData" v-loading="loading" style="width:100%">
        <el-table-column prop="userId" label="ID" width="70" />
        <el-table-column prop="username" label="用户名" min-width="140">
          <template #default="{ row }">
            <div style="display:flex;align-items:center;gap:8px">
              <el-avatar :size="28" style="background:#6366f1;color:#fff;font-size:12px;font-weight:600;flex-shrink:0">{{ row.username?.charAt(0)?.toUpperCase() }}</el-avatar>
              <span>{{ row.username }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column prop="email" label="邮箱" min-width="180" show-overflow-tooltip />
        <el-table-column prop="address" label="地址" min-width="160" show-overflow-tooltip />
        <el-table-column prop="role" label="角色" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.role === 'ADMIN' ? 'danger' : 'primary'" size="small">{{ row.role === 'ADMIN' ? '管理员' : '用户' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="注册时间" width="170">
          <template #default="{ row }">{{ fmt(row.createTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="warning" link size="small" @click="handleResetPwd(row)">重置密码</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDelete(row.userId)">
              <template #reference><el-button type="danger" link size="small" :disabled="row.userId === userStore.userId || row.username === 'admin'">删除</el-button></template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-container">
        <el-pagination v-model:current-page="pagination.pageNum" v-model:page-size="pagination.pageSize" :total="pagination.total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="loadData" @current-change="loadData" />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="用户名"><el-input v-model="form.username" disabled /></el-form-item>
        <el-form-item label="手机号" prop="phone"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="邮箱" prop="email"><el-input v-model="form.email" /></el-form-item>
        <el-form-item label="地址"><el-input v-model="form.address" /></el-form-item>
        <el-form-item label="角色">
          <el-select v-model="form.role" style="width:100%">
            <el-option label="管理员" value="ADMIN" />
            <el-option label="普通用户" value="USER" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="form.status" :active-value="1" :inactive-value="0" :disabled="form.userId === userStore.userId || form.username === 'admin'" />
          <span v-if="form.userId === userStore.userId" style="margin-left:8px;color:#94a3b8;font-size:12px">不能禁用自己</span>
          <span v-else-if="form.username === 'admin'" style="margin-left:8px;color:#94a3b8;font-size:12px">admin不能被禁用</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
    
    <!-- 重置密码弹窗 -->
    <el-dialog v-model="resetPwdVisible" title="重置密码" width="400px">
      <el-form label-width="80px">
        <el-form-item label="用户名">
          <el-input v-model="resetPwdForm.username" disabled />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="resetPwdForm.newPassword" type="password" show-password placeholder="请输入新密码(至少6位)" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetPwdVisible = false">取消</el-button>
        <el-button type="primary" :loading="resetPwdLoading" @click="handleResetPwdSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getUserList, updateUser, deleteUser, resetPassword } from '@/api/user'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()
const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const tableData = ref([])
const formRef = ref()
const resetPwdVisible = ref(false)
const resetPwdForm = reactive({ userId: null, username: '', newPassword: '' })
const resetPwdLoading = ref(false)

const searchForm = reactive({ username: '', phone: '', role: '' })
const pagination = reactive({ pageNum: 1, pageSize: 10, total: 0 })
const form = reactive({ userId: null, username: '', phone: '', email: '', address: '', role: '', status: 1 })

const rules = {
  phone: [{ pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' }],
  email: [{ type: 'email', message: '邮箱格式不正确', trigger: 'blur' }]
}

const fmt = (d) => {
  if (!d) return ''
  if (Array.isArray(d)) {
    const [y, m, day, h = 0, min = 0, s = 0] = d
    return `${y}-${String(m).padStart(2,'0')}-${String(day).padStart(2,'0')} ${String(h).padStart(2,'0')}:${String(min).padStart(2,'0')}:${String(s).padStart(2,'0')}`
  }
  return typeof d === 'string' ? d.replace('T', ' ').substring(0, 19) : String(d)
}

const loadData = async () => {
  loading.value = true
  try {
    const r = await getUserList({ ...searchForm, pageNum: pagination.pageNum, pageSize: pagination.pageSize })
    tableData.value = r.data.list
    pagination.total = r.data.total
  } finally { loading.value = false }
}

const resetSearch = () => {
  Object.assign(searchForm, { username: '', phone: '', role: '' })
  pagination.pageNum = 1
  loadData()
}

const handleEdit = (row) => {
  dialogTitle.value = '编辑用户'
  Object.assign(form, row)
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!(await formRef.value.validate().catch(() => false))) return
  submitLoading.value = true
  try {
    await updateUser(form)
    ElMessage.success('更新成功')
    dialogVisible.value = false
    loadData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || e.message || '更新失败')
  } finally { submitLoading.value = false }
}

const handleResetPwd = (row) => {
  resetPwdForm.userId = row.userId
  resetPwdForm.username = row.username
  resetPwdForm.newPassword = ''
  resetPwdVisible.value = true
}

const handleResetPwdSubmit = async () => {
  if (!resetPwdForm.newPassword || resetPwdForm.newPassword.length < 6) {
    ElMessage.error('密码长度至少6位')
    return
  }
  resetPwdLoading.value = true
  try {
    await resetPassword(resetPwdForm.userId, resetPwdForm.newPassword)
    ElMessage.success('密码重置成功')
    resetPwdVisible.value = false
  } catch (e) {
    ElMessage.error(e.response?.data?.message || e.message || '重置失败')
  } finally { resetPwdLoading.value = false }
}

const handleDelete = async (id) => {
  try {
    await deleteUser(id)
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || e.message || '删除失败')
  }
}

onMounted(loadData)
</script>

<style lang="scss" scoped>
.page-header {
  margin-bottom: 20px;
  h1 { font-size: 22px; font-weight: 700; color: #1e293b; margin: 0 0 4px; }
  p  { font-size: 13px; color: #94a3b8; margin: 0; }
}
.user-cell {
  display: flex; align-items: center; gap: 10px;
  .user-avatar { background: linear-gradient(135deg, #6366f1, #8b5cf6); color: white; font-weight: 600; font-size: 12px; }
}
</style>
