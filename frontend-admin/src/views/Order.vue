<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-info">
        <h2>📋 寄养订单</h2>
        <p>管理所有寄养订单</p>
      </div>
    </div>

    <div class="search-bar">
      <el-form :inline="true" :model="searchForm">
        <el-form-item label="订单号"><el-input v-model="searchForm.orderNo" placeholder="请输入" clearable /></el-form-item>
        <el-form-item label="宠物"><el-input v-model="searchForm.petName" placeholder="请输入" clearable /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable>
            <el-option label="待确认" value="PENDING" /><el-option label="已确认" value="CONFIRMED" />
            <el-option label="进行中" value="IN_PROGRESS" /><el-option label="已完成" value="COMPLETED" /><el-option label="已取消" value="CANCELLED" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadData">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
          <el-button type="success" @click="handleAdd">创建订单</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="table-container">
      <el-table :data="tableData" v-loading="loading" style="width:100%">
        <el-table-column prop="orderNo" label="订单号" width="200">
          <template #default="{ row }"><span style="font-family:monospace;font-size:12px;color:#6366f1">{{ row.orderNo }}</span></template>
        </el-table-column>
        <el-table-column prop="petName" label="宠物" min-width="100">
          <template #default="{ row }">🐾 {{ row.petName }}</template>
        </el-table-column>
        <el-table-column prop="username" label="客户" min-width="100" />
        <el-table-column prop="roomNumber" label="房间" width="80" align="center" />
        <el-table-column prop="startTime" label="开始时间" width="160">
          <template #default="{ row }">{{ fmt(row.startTime) }}</template>
        </el-table-column>
        <el-table-column prop="endTime" label="结束时间" width="160">
          <template #default="{ row }">{{ fmt(row.endTime) }}</template>
        </el-table-column>
        <el-table-column prop="totalAmount" label="金额" width="100" align="center">
          <template #default="{ row }"><span style="font-weight:600;color:#ef4444">¥{{ row.totalAmount }}</span></template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }"><el-tag :type="statusTag[row.status]" size="small">{{ statusMap[row.status] }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right" align="center">
          <template #default="{ row }">
            <div style="display:flex;align-items:center;justify-content:center;gap:4px">
              <el-button type="primary" link size="small" @click="handleView(row)">详情</el-button>
              <el-dropdown v-if="userStore.isAdmin && !['COMPLETED','CANCELLED'].includes(row.status)" @command="(c) => handleStatusChange(row.orderId, c)">
                <el-button type="warning" link size="small">状态<el-icon><ArrowDown /></el-icon></el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item v-if="row.status==='PENDING'" command="CONFIRMED">确认</el-dropdown-item>
                    <el-dropdown-item v-if="row.status==='CONFIRMED'" command="IN_PROGRESS">开始</el-dropdown-item>
                    <el-dropdown-item v-if="row.status==='IN_PROGRESS'" command="COMPLETED">完成</el-dropdown-item>
                    <el-dropdown-item command="CANCELLED" divided>取消</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-container">
        <el-pagination v-model:current-page="pagination.pageNum" v-model:page-size="pagination.pageSize" :total="pagination.total" :page-sizes="[10,20,50]" layout="total, sizes, prev, pager, next" @size-change="loadData" @current-change="loadData" />
      </div>
    </div>

    <!-- 创建订单 -->
    <el-dialog v-model="createVisible" title="创建寄养订单" width="580px">
      <el-form ref="createFormRef" :model="createForm" :rules="createRules" label-width="100px">
        <el-form-item label="选择宠物" prop="petId">
          <el-select v-model="createForm.petId" placeholder="请选择" style="width:100%">
            <el-option v-for="p in myPets" :key="p.petId" :label="`${p.petName} (${ptMap[p.petType]})`" :value="p.petId" />
          </el-select>
        </el-form-item>
        <el-form-item label="选择房间" prop="roomId">
          <el-select v-model="createForm.roomId" placeholder="请选择" style="width:100%">
            <el-option v-for="r in rooms" :key="r.roomId" :label="`${r.roomNumber} - ${rtMap[r.roomType]} (¥${r.pricePerDay}/天) [${r.currentOccupancy || 0}/${r.capacity}]`" :value="r.roomId" />
          </el-select>
        </el-form-item>
        <el-form-item label="寄养时间" prop="dateRange">
          <el-date-picker v-model="createForm.dateRange" type="datetimerange" range-separator="至" start-placeholder="开始" end-placeholder="结束" style="width:100%" />
        </el-form-item>
        <el-form-item label="附加服务">
          <el-checkbox-group v-model="createForm.selectedServices">
            <el-checkbox v-for="s in services" :key="s.serviceId" :label="s.serviceId">{{ s.serviceName }} (¥{{ s.price }}/天)</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="createForm.remarks" type="textarea" :rows="2" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleCreateSubmit">提交</el-button>
      </template>
    </el-dialog>

    <!-- 订单详情 -->
    <el-dialog v-model="detailVisible" title="订单详情" width="600px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="订单号">{{ cur.orderNo }}</el-descriptions-item>
        <el-descriptions-item label="状态"><el-tag :type="statusTag[cur.status]" size="small">{{ statusMap[cur.status] }}</el-tag></el-descriptions-item>
        <el-descriptions-item label="宠物">{{ cur.petName }}</el-descriptions-item>
        <el-descriptions-item label="客户">{{ cur.username }}</el-descriptions-item>
        <el-descriptions-item label="房间">{{ cur.roomNumber }}</el-descriptions-item>
        <el-descriptions-item label="房间类型">{{ rtMap[cur.roomType] }}</el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ fmt(cur.startTime) }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ fmt(cur.endTime) }}</el-descriptions-item>
        <el-descriptions-item label="总金额" :span="2"><span style="font-size:18px;font-weight:700;color:#ef4444">¥{{ cur.totalAmount }}</span></el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ cur.remarks || '无' }}</el-descriptions-item>
      </el-descriptions>
      <div v-if="cur.services?.length" style="margin-top:16px">
        <h4 style="margin:0 0 8px;font-size:14px">附加服务</h4>
        <el-table :data="cur.services" size="small">
          <el-table-column prop="serviceName" label="服务" />
          <el-table-column prop="quantity" label="数量" width="80" />
          <el-table-column prop="subtotal" label="小计" width="100">
            <template #default="{ row }">¥{{ row.subtotal }}</template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getOrderList, createOrder, updateOrderStatus, getOrderById } from '@/api/order'
import { getMyPets } from '@/api/pet'
import { getAvailableRooms } from '@/api/room'
import { getAvailableServices } from '@/api/service'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()
const loading = ref(false), submitLoading = ref(false), createVisible = ref(false), detailVisible = ref(false)
const tableData = ref([]), createFormRef = ref(), cur = ref({})
const myPets = ref([]), rooms = ref([]), services = ref([])

const ptMap = { DOG:'狗', CAT:'猫', BIRD:'鸟', OTHER:'其他' }
const rtMap = { STANDARD:'标准间', DELUXE:'豪华间', VIP:'VIP套房' }
const statusMap = { PENDING:'待确认', CONFIRMED:'已确认', IN_PROGRESS:'进行中', COMPLETED:'已完成', CANCELLED:'已取消' }
const statusTag = { PENDING:'warning', CONFIRMED:'primary', IN_PROGRESS:'success', COMPLETED:'info', CANCELLED:'danger' }

const searchForm = reactive({ orderNo:'', petName:'', status:'' })
const pagination = reactive({ pageNum:1, pageSize:10, total:0 })
const createForm = reactive({ petId:null, roomId:null, dateRange:null, selectedServices:[], remarks:'' })
const createRules = { petId:[{required:true,message:'请选择',trigger:'change'}], roomId:[{required:true,message:'请选择',trigger:'change'}], dateRange:[{required:true,message:'请选择',trigger:'change'}] }
const fmt = (d) => {
  if (!d) return ''
  if (Array.isArray(d)) {
    const [y, m, day, h = 0, min = 0] = d
    return `${y}-${String(m).padStart(2,'0')}-${String(day).padStart(2,'0')} ${String(h).padStart(2,'0')}:${String(min).padStart(2,'0')}`
  }
  return typeof d === 'string' ? d.replace('T',' ').substring(0,16) : String(d)
}

const loadData = async () => {
  loading.value = true
  try { const r = await getOrderList({ ...searchForm, pageNum:pagination.pageNum, pageSize:pagination.pageSize }); tableData.value = r.data.list; pagination.total = r.data.total } finally { loading.value = false }
}
const resetSearch = () => { Object.assign(searchForm, { orderNo:'', petName:'', status:'' }); pagination.pageNum = 1; loadData() }

const handleAdd = async () => {
  Object.assign(createForm, { petId:null, roomId:null, dateRange:null, selectedServices:[], remarks:'' })
  try {
    const [a,b,c] = await Promise.all([getMyPets(), getAvailableRooms(), getAvailableServices()])
    myPets.value = a.data; rooms.value = b.data; services.value = c.data
    createVisible.value = true
  } catch(e) { console.error(e) }
}

const handleCreateSubmit = async () => {
  if (!(await createFormRef.value.validate().catch(() => false))) return
  submitLoading.value = true
  try {
    await createOrder({ petId:createForm.petId, roomId:createForm.roomId, startTime:createForm.dateRange[0], endTime:createForm.dateRange[1], remarks:createForm.remarks, services:createForm.selectedServices.map(id => ({ serviceId:id, quantity:1 })) })
    ElMessage.success('创建成功'); createVisible.value = false; loadData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || e.message || '创建失败')
  } finally { submitLoading.value = false }
}

const handleView = async (row) => { try { cur.value = (await getOrderById(row.orderId)).data; detailVisible.value = true } catch(e) { ElMessage.error(e.response?.data?.message || e.message || '获取详情失败') } }
const handleStatusChange = async (id, s) => {
  try {
    await updateOrderStatus(id, s)
    ElMessage.success('更新成功')
    loadData()
  } catch (e) {
    ElMessage.error(e.response?.data?.message || e.message || '状态更新失败')
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
.room-badge { background: #eef2ff; color: #6366f1; padding: 3px 10px; border-radius: 6px; font-weight: 600; font-size: 12px; }
.amount { font-weight: 700; color: #ef4444; }
.service-checkbox { display: flex; flex-direction: column; gap: 8px; }
.order-detail {
  .detail-header {
    display: flex; justify-content: space-between; align-items: center;
    padding: 20px; background: #f8fafc; border-radius: 12px; margin-bottom: 20px;
    .order-info { display: flex; align-items: center; gap: 16px;
      .order-no-large { font-family: 'SF Mono', monospace; font-size: 16px; font-weight: 700; color: #1e293b; }
    }
    .order-amount {
      text-align: right;
      .label { display: block; font-size: 12px; color: #94a3b8; margin-bottom: 4px; }
      .value { font-size: 26px; font-weight: 700; color: #ef4444; }
    }
  }
  .services-section {
    h4 { font-size: 14px; font-weight: 600; color: #1e293b; margin: 0 0 12px; }
    .service-amount { font-weight: 600; color: #10b981; }
  }
}
</style>
