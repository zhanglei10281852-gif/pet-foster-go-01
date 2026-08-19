<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-info">
        <h2>🐾 宠物管理</h2>
        <p>管理您的爱宠信息</p>
      </div>
    </div>

    <div class="search-bar">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="宠物名称">
          <el-input v-model="searchForm.petName" placeholder="请输入" clearable />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchForm.petType" placeholder="全部" clearable>
            <el-option label="狗" value="DOG" />
            <el-option label="猫" value="CAT" />
            <el-option label="鸟" value="BIRD" />
            <el-option label="其他" value="OTHER" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
          <el-button type="success" @click="handleAdd">添加宠物</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="table-container">
      <el-table :data="tableData" v-loading="loading" style="width:100%">
        <el-table-column prop="petId" label="ID" width="70" />
        <el-table-column prop="petName" label="名称" min-width="120">
          <template #default="{ row }">
            <span>{{ emojiMap[row.petType] || '🐾' }} {{ row.petName }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="petType" label="类型" width="80" align="center">
          <template #default="{ row }">
            <el-tag size="small">{{ typeMap[row.petType] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="breed" label="品种" min-width="100" show-overflow-tooltip />
        <el-table-column prop="age" label="年龄(月)" width="100" align="center" />
        <el-table-column prop="weight" label="体重(kg)" width="100" align="center" />
        <el-table-column prop="healthStatus" label="健康状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.healthStatus === '健康' ? 'success' : 'warning'" size="small">{{ row.healthStatus }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ownerName" label="主人" min-width="100" />
        <el-table-column prop="specialRequirements" label="特殊要求" min-width="140" show-overflow-tooltip />
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDelete(row.petId)">
              <template #reference><el-button type="danger" link size="small">删除</el-button></template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-container">
        <el-pagination v-model:current-page="pagination.pageNum" v-model:page-size="pagination.pageSize" :total="pagination.total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="loadData" @current-change="loadData" />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="520px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="宠物名称" prop="petName"><el-input v-model="form.petName" /></el-form-item>
        <el-form-item label="类型" prop="petType">
          <el-select v-model="form.petType" style="width:100%">
            <el-option label="狗" value="DOG" /><el-option label="猫" value="CAT" />
            <el-option label="鸟" value="BIRD" /><el-option label="其他" value="OTHER" />
          </el-select>
        </el-form-item>
        <el-form-item label="品种"><el-input v-model="form.breed" /></el-form-item>
        <el-row :gutter="20">
          <el-col :span="12"><el-form-item label="年龄(月)"><el-input-number v-model="form.age" :min="0" style="width:100%" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="体重(kg)"><el-input-number v-model="form.weight" :min="0" :precision="2" style="width:100%" /></el-form-item></el-col>
        </el-row>
        <el-form-item label="健康状况"><el-input v-model="form.healthStatus" /></el-form-item>
        <el-form-item label="特殊要求"><el-input v-model="form.specialRequirements" type="textarea" :rows="3" /></el-form-item>
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
import { getPetList, addPet, updatePet, deletePet } from '@/api/pet'

const loading = ref(false), submitLoading = ref(false), dialogVisible = ref(false), dialogTitle = ref(''), isEdit = ref(false)
const tableData = ref([]), formRef = ref()
const typeMap = { DOG: '狗', CAT: '猫', BIRD: '鸟', OTHER: '其他' }
const emojiMap = { DOG: '🐕', CAT: '🐈', BIRD: '🐦', OTHER: '🐹' }
const searchForm = reactive({ petName: '', petType: '' })
const pagination = reactive({ pageNum: 1, pageSize: 10, total: 0 })
const initForm = () => ({ petId: null, petName: '', petType: 'DOG', breed: '', age: null, weight: null, healthStatus: '健康', specialRequirements: '' })
const form = reactive(initForm())
const rules = { petName: [{ required: true, message: '请输入', trigger: 'blur' }], petType: [{ required: true, message: '请选择', trigger: 'change' }] }

const loadData = async () => {
  loading.value = true
  try { const r = await getPetList({ ...searchForm, pageNum: pagination.pageNum, pageSize: pagination.pageSize }); tableData.value = r.data.list; pagination.total = r.data.total } finally { loading.value = false }
}
const resetSearch = () => { searchForm.petName = ''; searchForm.petType = ''; pagination.pageNum = 1; loadData() }
const handleAdd = () => { dialogTitle.value = '添加宠物'; isEdit.value = false; Object.assign(form, initForm()); dialogVisible.value = true }
const handleEdit = (row) => { dialogTitle.value = '编辑宠物'; isEdit.value = true; Object.assign(form, row); dialogVisible.value = true }
const handleSubmit = async () => {
  if (!(await formRef.value.validate().catch(() => false))) return
  submitLoading.value = true
  try {
    isEdit.value ? await updatePet(form) : await addPet(form)
    ElMessage.success(isEdit.value ? '更新成功' : '添加成功')
    dialogVisible.value = false
    loadData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || e.message || '操作失败')
  } finally { submitLoading.value = false }
}
const handleDelete = async (id) => {
  try {
    await deletePet(id)
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
.pet-cell { display: flex; align-items: center; gap: 8px; .pet-emoji { font-size: 18px; } .pet-name { font-weight: 500; color: #1e293b; } }
</style>
