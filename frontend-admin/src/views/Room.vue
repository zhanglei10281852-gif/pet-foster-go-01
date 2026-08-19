<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-info">
        <h2>🏠 房间管理</h2>
        <p>管理寄养房间和定价</p>
      </div>
    </div>

    <div class="search-bar">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="房间号"><el-input v-model="searchForm.roomNumber" placeholder="请输入" clearable /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchForm.roomType" placeholder="全部" clearable>
            <el-option label="标准间" value="STANDARD" /><el-option label="豪华间" value="DELUXE" /><el-option label="VIP套房" value="VIP" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable>
            <el-option label="空闲" value="AVAILABLE" /><el-option label="已预订" value="RESERVED" /><el-option label="占用中" value="OCCUPIED" /><el-option label="打扫中" value="CLEANING" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
          <el-button type="success" @click="handleAdd">添加房间</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="table-container">
      <el-table :data="tableData" v-loading="loading" style="width:100%">
        <el-table-column prop="roomId" label="ID" width="70" />
        <el-table-column prop="roomNumber" label="房间号" width="100" align="center">
          <template #default="{ row }">
            <span style="font-family:monospace;font-weight:600;color:#6366f1">{{ row.roomNumber }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="roomType" label="房间类型" width="110" align="center">
          <template #default="{ row }"><el-tag :type="typeTag[row.roomType]" size="small">{{ typeMap[row.roomType] }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }"><el-tag :type="statusTag[row.status]" size="small">{{ statusMap[row.status] }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="pricePerDay" label="日价格" width="100" align="center">
          <template #default="{ row }"><span style="font-weight:600;color:#ef4444">¥{{ row.pricePerDay }}</span></template>
        </el-table-column>
        <el-table-column prop="capacity" label="容量" width="100" align="center">
          <template #default="{ row }">
            <span>{{ row.currentOccupancy || 0 }} / {{ row.capacity }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="180" fixed="right" align="center">
          <template #default="{ row }">
            <div style="display:flex;align-items:center;justify-content:center;gap:4px">
              <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
              <el-dropdown @command="(cmd) => handleStatusChange(row.roomId, cmd)">
                <el-button type="warning" link size="small">状态<el-icon><ArrowDown /></el-icon></el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="AVAILABLE">空闲</el-dropdown-item>
                    <el-dropdown-item command="CLEANING">打扫中</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-popconfirm title="确定删除？" @confirm="handleDelete(row.roomId)">
                <template #reference><el-button type="danger" link size="small">删除</el-button></template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-container">
        <el-pagination v-model:current-page="pagination.pageNum" v-model:page-size="pagination.pageSize" :total="pagination.total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="loadData" @current-change="loadData" />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="480px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="房间号" prop="roomNumber"><el-input v-model="form.roomNumber" /></el-form-item>
        <el-form-item label="类型" prop="roomType">
          <el-select v-model="form.roomType" style="width:100%">
            <el-option label="标准间" value="STANDARD" /><el-option label="豪华间" value="DELUXE" /><el-option label="VIP套房" value="VIP" />
          </el-select>
        </el-form-item>
        <el-form-item label="日价格" prop="pricePerDay"><el-input-number v-model="form.pricePerDay" :min="0" :precision="2" style="width:100%" /></el-form-item>
        <el-form-item label="容量"><el-input-number v-model="form.capacity" :min="1" style="width:100%" /></el-form-item>
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
import { getRoomList, addRoom, updateRoom, deleteRoom, updateRoomStatus } from '@/api/room'

const loading = ref(false), submitLoading = ref(false), dialogVisible = ref(false), dialogTitle = ref(''), isEdit = ref(false)
const tableData = ref([]), formRef = ref()
const typeMap = { STANDARD: '标准间', DELUXE: '豪华间', VIP: 'VIP套房' }
const typeTag = { STANDARD: 'info', DELUXE: 'warning', VIP: 'danger' }
const statusMap = { AVAILABLE: '空闲', RESERVED: '已预订', OCCUPIED: '占用中', CLEANING: '打扫中' }
const statusTag = { AVAILABLE: 'success', RESERVED: 'warning', OCCUPIED: 'danger', CLEANING: 'info' }
const searchForm = reactive({ roomNumber: '', roomType: '', status: '' })
const pagination = reactive({ pageNum: 1, pageSize: 10, total: 0 })
const initForm = () => ({ roomId: null, roomNumber: '', roomType: 'STANDARD', pricePerDay: 80, capacity: 1, description: '', status: 'AVAILABLE' })
const form = reactive(initForm())
const rules = { roomNumber: [{ required: true, message: '请输入', trigger: 'blur' }], roomType: [{ required: true, trigger: 'change' }], pricePerDay: [{ required: true, trigger: 'blur' }] }

const loadData = async () => {
  loading.value = true
  try { const r = await getRoomList({ ...searchForm, pageNum: pagination.pageNum, pageSize: pagination.pageSize }); tableData.value = r.data.list; pagination.total = r.data.total } finally { loading.value = false }
}
const resetSearch = () => { Object.assign(searchForm, { roomNumber: '', roomType: '', status: '' }); pagination.pageNum = 1; loadData() }
const handleAdd = () => { dialogTitle.value = '添加房间'; isEdit.value = false; Object.assign(form, initForm()); dialogVisible.value = true }
const handleEdit = (row) => { dialogTitle.value = '编辑房间'; isEdit.value = true; Object.assign(form, row); dialogVisible.value = true }
const handleSubmit = async () => {
  if (!(await formRef.value.validate().catch(() => false))) return
  submitLoading.value = true
  try {
    isEdit.value ? await updateRoom(form) : await addRoom(form)
    ElMessage.success(isEdit.value ? '更新成功' : '添加成功')
    dialogVisible.value = false
    loadData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || e.message || '操作失败')
  } finally { submitLoading.value = false }
}
const handleStatusChange = async (id, s) => {
  try {
    await updateRoomStatus(id, s)
    ElMessage.success('状态更新成功')
    loadData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || e.message || '状态更新失败')
  }
}
const handleDelete = async (id) => {
  try {
    await deleteRoom(id)
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
.room-number { font-family: 'SF Mono', monospace; font-weight: 600; color: #6366f1; background: #eef2ff; padding: 3px 10px; border-radius: 6px; font-size: 13px; }
.price { font-size: 15px; font-weight: 700; color: #ef4444; }
.price-unit { font-size: 12px; color: #94a3b8; margin-left: 1px; }
</style>
