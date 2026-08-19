<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-info">
        <h2>✨ 服务项目</h2>
        <p>管理寄养附加服务和定价</p>
      </div>
    </div>

    <div class="search-bar">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="服务名称"><el-input v-model="searchForm.serviceName" placeholder="请输入" clearable /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable>
            <el-option label="上架" :value="1" /><el-option label="下架" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
          <el-button type="success" @click="handleAdd">添加服务</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="table-container">
      <el-table :data="tableData" v-loading="loading" style="width:100%">
        <el-table-column prop="serviceId" label="ID" width="70" />
        <el-table-column prop="serviceName" label="服务名称" min-width="140" />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column prop="price" label="价格" width="120" align="center">
          <template #default="{ row }"><span style="font-weight:600;color:#22c55e">¥{{ row.price }}/天</span></template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" align="center">
          <template #default="{ row }"><el-tag :type="row.status === 1 ? 'success' : 'info'" size="small">{{ row.status === 1 ? '上架' : '下架' }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180">
          <template #default="{ row }">{{ fmt(row.createTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDelete(row.serviceId)">
              <template #reference><el-button type="danger" link size="small">删除</el-button></template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-container">
        <el-pagination v-model:current-page="pagination.pageNum" v-model:page-size="pagination.pageSize" :total="pagination.total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="loadData" @current-change="loadData" />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="服务名称" prop="serviceName"><el-input v-model="form.serviceName" /></el-form-item>
        <el-form-item label="价格(元/天)" prop="price"><el-input-number v-model="form.price" :min="0" :precision="2" style="width:100%" /></el-form-item>
        <el-form-item label="状态"><el-switch v-model="form.status" :active-value="1" :inactive-value="0" active-text="上架" inactive-text="下架" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getServiceList, addService, updateService, deleteService } from '@/api/service'

const loading = ref(false), submitLoading = ref(false), dialogVisible = ref(false), dialogTitle = ref(''), isEdit = ref(false)
const tableData = ref([]), formRef = ref()
const searchForm = reactive({ serviceName: '', status: null })
const pagination = reactive({ pageNum: 1, pageSize: 10, total: 0 })
const initForm = () => ({ serviceId: null, serviceName: '', description: '', price: 0, status: 1 })
const form = reactive(initForm())
const rules = { serviceName: [{ required: true, message: '请输入', trigger: 'blur' }], price: [{ required: true, trigger: 'blur' }] }
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
  try { const r = await getServiceList({ ...searchForm, pageNum: pagination.pageNum, pageSize: pagination.pageSize }); tableData.value = r.data.list; pagination.total = r.data.total } finally { loading.value = false }
}
const resetSearch = () => { searchForm.serviceName = ''; searchForm.status = null; pagination.pageNum = 1; loadData() }
const handleAdd = () => { dialogTitle.value = '添加服务'; isEdit.value = false; Object.assign(form, initForm()); dialogVisible.value = true }
const handleEdit = (row) => { dialogTitle.value = '编辑服务'; isEdit.value = true; Object.assign(form, row); dialogVisible.value = true }
const handleSubmit = async () => {
  if (!(await formRef.value.validate().catch(() => false))) return
  submitLoading.value = true
  try {
    isEdit.value ? await updateService(form) : await addService(form)
    ElMessage.success(isEdit.value ? '更新成功' : '添加成功')
    dialogVisible.value = false
    loadData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || e.message || '操作失败')
  } finally { submitLoading.value = false }
}
const handleDelete = async (id) => {
  try {
    await deleteService(id)
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
.service-cell { display: flex; align-items: center; gap: 8px; .service-icon { font-size: 18px; } .service-name { font-weight: 500; color: #1e293b; } }
.price { font-size: 15px; font-weight: 700; color: #10b981; }
.price-unit { font-size: 12px; color: #94a3b8; margin-left: 1px; }
</style>
