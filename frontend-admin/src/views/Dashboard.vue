<template>
  <div class="dashboard">
    <!-- 欢迎横幅 -->
    <div class="hero">
      <div class="hero-text">
        <h1>Hi, {{ userStore.username }} 👋</h1>
        <p>欢迎回来，今天也要好好照顾毛孩子们哦！</p>
      </div>
      <div class="hero-deco">
        <div class="deco-circle c1"></div>
        <div class="deco-circle c2"></div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats">
      <div class="stat-card" v-for="s in statList" :key="s.key">
        <div class="stat-icon" :class="s.color">{{ s.emoji }}</div>
        <div class="stat-body">
          <span class="stat-num">{{ stats[s.key] || 0 }}</span>
          <span class="stat-label">{{ s.label }}</span>
        </div>
      </div>
    </div>

    <div class="grid-2">
      <!-- 最近订单 -->
      <div class="panel main-panel">
        <div class="panel-head">
          <h3>最近订单</h3>
          <el-button type="primary" link @click="$router.push('/order')">查看全部 →</el-button>
        </div>
        <div class="panel-body">
          <el-table :data="recentOrders" size="small" class="clean-table" style="width:100%">
            <el-table-column prop="orderNo" label="订单号" min-width="160">
              <template #default="{ row }"><span class="mono">{{ row.orderNo }}</span></template>
            </el-table-column>
            <el-table-column prop="petName" label="宠物" min-width="90">
              <template #default="{ row }">🐾 {{ row.petName }}</template>
            </el-table-column>
            <el-table-column prop="roomNumber" label="房间" min-width="70" />
            <el-table-column prop="startTime" label="开始时间" min-width="140">
              <template #default="{ row }">{{ fmt(row.startTime) }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态" min-width="90">
              <template #default="{ row }">
                <el-tag :type="statusTag[row.status]" size="small">{{ statusMap[row.status] }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="totalAmount" label="金额" min-width="90">
              <template #default="{ row }"><span class="amount">¥{{ row.totalAmount }}</span></template>
            </el-table-column>
          </el-table>
        </div>
      </div>

      <!-- 右侧面板 -->
      <div class="side-panels">
        <div class="panel">
          <div class="panel-head"><h3>快捷操作</h3></div>
          <div class="shortcuts">
            <div class="shortcut" @click="$router.push('/order')">
              <span class="sc-icon purple">📋</span><span>新建订单</span>
            </div>
            <div class="shortcut" @click="$router.push('/pet')">
              <span class="sc-icon green">🐾</span><span>管理宠物</span>
            </div>
            <div class="shortcut" v-if="userStore.isAdmin" @click="$router.push('/room')">
              <span class="sc-icon amber">🏠</span><span>房间管理</span>
            </div>
            <div class="shortcut" @click="$router.push('/record')">
              <span class="sc-icon blue">📝</span><span>日常记录</span>
            </div>
          </div>
        </div>

        <div class="panel">
          <div class="panel-head"><h3>系统信息</h3></div>
          <div class="info-rows">
            <div class="info-row">
              <span class="info-label">当前用户</span>
              <span class="info-value">{{ userStore.username }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">角色</span>
              <el-tag :type="userStore.isAdmin ? 'danger' : 'primary'" size="small">
                {{ userStore.isAdmin ? '管理员' : '普通用户' }}
              </el-tag>
            </div>
            <div class="info-row">
              <span class="info-label">版本</span>
              <span class="info-value">v1.0.0</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getStatistics, getOrderList } from '@/api/order'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()
const stats = ref({})
const recentOrders = ref([])

const statList = [
  { key: 'total',      label: '总订单',   emoji: '📊', color: 'purple' },
  { key: 'pending',   label: '待确认',   emoji: '⏳', color: 'amber' },
  { key: 'inProgress', label: '进行中',  emoji: '▶️', color: 'green' },
  { key: 'completed', label: '已完成',   emoji: '✅', color: 'slate' },
]

const statusMap = { PENDING:'待确认', CONFIRMED:'已确认', IN_PROGRESS:'进行中', COMPLETED:'已完成', CANCELLED:'已取消' }
const statusTag = { PENDING:'warning', CONFIRMED:'primary', IN_PROGRESS:'success', COMPLETED:'info', CANCELLED:'danger' }
const fmt = (d) => {
  if (!d) return ''
  if (Array.isArray(d)) {
    const [y, m, day, h = 0, min = 0] = d
    return `${y}-${String(m).padStart(2,'0')}-${String(day).padStart(2,'0')} ${String(h).padStart(2,'0')}:${String(min).padStart(2,'0')}`
  }
  return typeof d === 'string' ? d.replace('T',' ').substring(0,16) : ''
}

onMounted(async () => {
  try {
    const [s, o] = await Promise.all([getStatistics(), getOrderList({ pageNum:1, pageSize:5 })])
    stats.value = s.data; recentOrders.value = o.data.list
  } catch(e) { console.error(e) }
})
</script>

<style lang="scss" scoped>
.dashboard { animation: pageIn .35s cubic-bezier(.4,0,.2,1) both; }
@keyframes pageIn { from { opacity:0; transform:translateY(8px); } to { opacity:1; transform:translateY(0); } }

.hero {
  background: linear-gradient(135deg, #6366f1 0%, #a855f7 50%, #ec4899 100%);
  border-radius: 16px; padding: 32px 36px; margin-bottom: 20px;
  position: relative; overflow: hidden;
  .hero-text { position: relative; z-index: 2; }
  h1 { font-size: 26px; font-weight: 700; color: white; margin: 0 0 6px; }
  p  { font-size: 14px; color: rgba(255,255,255,.85); margin: 0; }
  .hero-deco { position: absolute; inset: 0; pointer-events: none; }
  .deco-circle { position: absolute; border-radius: 50%; background: rgba(255,255,255,.08); }
  .c1 { width: 260px; height: 260px; top: -80px; right: -40px; }
  .c2 { width: 140px; height: 140px; bottom: -40px; right: 120px; }
}

.stats {
  display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 20px;
}
.stat-card {
  background: white; border-radius: 14px; padding: 20px;
  display: flex; align-items: center; gap: 16px;
  box-shadow: 0 1px 3px rgba(0,0,0,.04); border: 1px solid #f1f5f9;
  transition: all .25s;
  &:hover { transform: translateY(-3px); box-shadow: 0 8px 20px rgba(0,0,0,.06); }
}
.stat-icon {
  width: 48px; height: 48px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  font-size: 22px; flex-shrink: 0;
  &.purple { background: linear-gradient(135deg, #ede9fe, #ddd6fe); }
  &.amber  { background: linear-gradient(135deg, #fef3c7, #fde68a); }
  &.green  { background: linear-gradient(135deg, #d1fae5, #a7f3d0); }
  &.slate  { background: linear-gradient(135deg, #f1f5f9, #e2e8f0); }
}
.stat-body { display: flex; flex-direction: column; }
.stat-num  { font-size: 28px; font-weight: 700; color: #1e293b; line-height: 1.1; }
.stat-label { font-size: 13px; color: #94a3b8; margin-top: 2px; }

.grid-2 { display: grid; grid-template-columns: 1fr 320px; gap: 20px; }
.side-panels { display: flex; flex-direction: column; gap: 16px; }

.panel {
  background: white; border-radius: 14px;
  box-shadow: 0 1px 3px rgba(0,0,0,.04); border: 1px solid #f1f5f9;
  overflow: hidden;
}
.panel-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 16px 20px; border-bottom: 1px solid #f1f5f9;
  h3 { font-size: 15px; font-weight: 600; color: #1e293b; margin: 0; }
}
.panel-body { padding: 0; }

.clean-table {
  :deep(th.el-table__cell) {
    background: #f8fafc !important; font-weight: 600; color: #64748b;
    font-size: 12px; text-transform: uppercase; letter-spacing: .5px;
  }
  .mono   { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 12px; color: #6366f1; }
  .amount { font-weight: 600; color: #ef4444; }
}

.shortcuts { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; padding: 16px; }
.shortcut {
  display: flex; flex-direction: column; align-items: center; gap: 8px;
  padding: 16px 12px; border-radius: 10px; background: #f8fafc;
  cursor: pointer; transition: all .25s; font-size: 13px; font-weight: 500; color: #475569;
  &:hover { background: #f1f5f9; transform: translateY(-2px); }
}
.sc-icon {
  width: 40px; height: 40px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center; font-size: 20px;
  &.purple { background: linear-gradient(135deg, #ede9fe, #ddd6fe); }
  &.green  { background: linear-gradient(135deg, #d1fae5, #a7f3d0); }
  &.amber  { background: linear-gradient(135deg, #fef3c7, #fde68a); }
  &.blue   { background: linear-gradient(135deg, #dbeafe, #bfdbfe); }
}

.info-rows { padding: 8px 20px 16px; }
.info-row {
  display: flex; justify-content: space-between; align-items: center;
  padding: 10px 0; border-bottom: 1px solid #f8fafc;
  &:last-child { border-bottom: none; }
}
.info-label { font-size: 13px; color: #94a3b8; }
.info-value { font-size: 13px; font-weight: 500; color: #334155; }

@media (max-width: 1024px) {
  .stats { grid-template-columns: repeat(2, 1fr); }
  .grid-2 { grid-template-columns: 1fr; }
}
</style>
