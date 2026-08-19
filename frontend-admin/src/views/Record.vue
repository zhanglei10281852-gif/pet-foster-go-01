<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-info">
        <h2>📝 日常记录</h2>
        <p>记录宠物每日的饮食、活动和健康状况</p>
      </div>
    </div>

    <div class="search-bar">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="订单ID"><el-input v-model="searchForm.orderId" placeholder="请输入" clearable /></el-form-item>
        <el-form-item label="日期">
          <el-date-picker v-model="searchForm.dateRange" type="daterange" range-separator="至" start-placeholder="开始" end-placeholder="结束" value-format="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
          <el-button v-if="userStore.isAdmin" type="success" @click="handleAdd">添加记录</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="table-container">
      <el-table :data="tableData" v-loading="loading" style="width:100%">
        <el-table-column prop="recordId" label="ID" width="70" />
        <el-table-column prop="orderNo" label="订单号" width="200">
          <template #default="{ row }"><span style="font-family:monospace;font-size:12px;color:#6366f1">{{ row.orderNo }}</span></template>
        </el-table-column>
        <el-table-column prop="petName" label="宠物" min-width="100">
          <template #default="{ row }">🐾 {{ row.petName }}</template>
        </el-table-column>
        <el-table-column prop="recordDate" label="记录日期" width="120" align="center" />
        <el-table-column prop="diet" label="饮食" width="100" align="center">
          <template #default="{ row }"><el-tag :type="row.diet==='正常'||row.diet==='食欲旺盛'?'success':row.diet==='拒食'?'danger':'warning'" size="small">{{ row.diet }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="defecation" label="排便" width="90" align="center">
          <template #default="{ row }"><el-tag :type="row.defecation==='正常'?'success':'warning'" size="small">{{ row.defecation }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="activity" label="活动" width="90" align="center">
          <template #default="{ row }"><el-tag :type="row.activity==='活跃'||row.activity==='正常'?'success':'warning'" size="small">{{ row.activity }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="spirit" label="精神" width="90" align="center">
          <template #default="{ row }"><el-tag :type="row.spirit==='良好'?'success':row.spirit==='萎靡'?'danger':'warning'" size="small">{{ row.spirit }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="remarks" label="备注" min-width="140" show-overflow-tooltip />
        <el-table-column v-if="userStore.isAdmin" label="操作" width="120" fixed="right" align="center">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-popconfirm title="确定删除？" @confirm="handleDelete(row.recordId)">
              <template #reference><el-button type="danger" link size="small">删除</el-button></template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-container">
        <el-pagination v-model:current-page="pagination.pageNum" v-model:page-size="pagination.pageSize" :total="pagination.total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="loadData" @current-change="loadData" />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="580px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="选择订单" prop="orderId">
          <el-select v-model="form.orderId" placeholder="请选择进行中的订单" style="width:100%" :disabled="isEdit" filterable>
            <el-option v-for="o in orders" :key="o.orderId" :label="`${o.orderNo} - ${o.petName}`" :value="o.orderId" />
          </el-select>
        </el-form-item>
        <el-form-item label="记录日期" prop="recordDate"><el-date-picker v-model="form.recordDate" type="date" value-format="YYYY-MM-DD" style="width:100%" /></el-form-item>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="饮食">
              <el-select v-model="form.diet" style="width:100%">
                <el-option label="正常" value="正常" /><el-option label="食欲旺盛" value="食欲旺盛" /><el-option label="食欲不振" value="食欲不振" /><el-option label="拒食" value="拒食" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="排便">
              <el-select v-model="form.defecation" style="width:100%">
                <el-option label="正常" value="正常" /><el-option label="稀便" value="稀便" /><el-option label="便秘" value="便秘" /><el-option label="未排便" value="未排便" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="活动">
              <el-select v-model="form.activity" style="width:100%">
                <el-option label="活跃" value="活跃" /><el-option label="正常" value="正常" /><el-option label="安静" value="安静" /><el-option label="嗜睡" value="嗜睡" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="精神">
              <el-select v-model="form.spirit" style="width:100%">
                <el-option label="良好" value="良好" /><el-option label="一般" value="一般" /><el-option label="萎靡" value="萎靡" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="备注"><el-input v-model="form.remarks" type="textarea" :rows="3" /></el-form-item>
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
import { getRecordList, addRecord, updateRecord, deleteRecord } from '@/api/record'
import { getOrderList } from '@/api/order'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()
const loading = ref(false), submitLoading = ref(false), dialogVisible = ref(false), dialogTitle = ref(''), isEdit = ref(false)
const tableData = ref([]), formRef = ref(), orders = ref([])
const searchForm = reactive({ orderId: '', dateRange: null })
const pagination = reactive({ pageNum: 1, pageSize: 10, total: 0 })
const initForm = () => ({ recordId: null, orderId: null, recordDate: new Date().toISOString().split('T')[0], diet: '正常', defecation: '正常', activity: '正常', spirit: '良好', remarks: '' })
const form = reactive(initForm())
const rules = { orderId: [{ required: true, message: '请输入', trigger: 'blur' }], recordDate: [{ required: true, message: '请选择', trigger: 'change' }] }

const loadData = async () => {
  loading.value = true
  try {
    const p = { pageNum: pagination.pageNum, pageSize: pagination.pageSize }
    if (searchForm.orderId) p.orderId = searchForm.orderId
    if (searchForm.dateRange) { p.startDate = searchForm.dateRange[0]; p.endDate = searchForm.dateRange[1] }
    const r = await getRecordList(p); tableData.value = r.data.list; pagination.total = r.data.total
  } finally { loading.value = false }
}
const resetSearch = () => { searchForm.orderId = ''; searchForm.dateRange = null; pagination.pageNum = 1; loadData() }
const handleAdd = async () => {
  dialogTitle.value = '添加记录'
  isEdit.value = false
  Object.assign(form, initForm())
  // 加载进行中的订单
  try {
    const r = await getOrderList({ status: 'IN_PROGRESS', pageNum: 1, pageSize: 100 })
    orders.value = r.data.list
  } catch (e) {
    orders.value = []
  }
  dialogVisible.value = true
}
const handleEdit = async (row) => {
  dialogTitle.value = '编辑记录'
  isEdit.value = true
  Object.assign(form, {
    ...row,
    recordDate: typeof row.recordDate === 'string' ? row.recordDate.slice(0, 10) : row.recordDate,
  })
  // 加载订单列表以显示订单号
  try {
    const r = await getOrderList({ pageNum: 1, pageSize: 100 })
    orders.value = r.data.list
  } catch (e) {
    orders.value = []
  }
  dialogVisible.value = true
}
const handleSubmit = async () => {
  if (!(await formRef.value.validate().catch(() => false))) return
  submitLoading.value = true
  try {
    isEdit.value ? await updateRecord(form) : await addRecord(form)
    ElMessage.success(isEdit.value ? '更新成功' : '添加成功')
    dialogVisible.value = false
    loadData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || e.message || '操作失败')
  } finally { submitLoading.value = false }
}
const handleDelete = async (id) => {
  try {
    await deleteRecord(id)
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
.order-no { font-family: 'SF Mono', monospace; font-size: 12px; color: #6366f1; font-weight: 500; }
.pet-cell { display: flex; align-items: center; gap: 6px; .pet-icon { font-size: 14px; } }
.date-badge { background: #eef2ff; color: #6366f1; padding: 3px 10px; border-radius: 6px; font-weight: 500; font-size: 12px; }
</style>
