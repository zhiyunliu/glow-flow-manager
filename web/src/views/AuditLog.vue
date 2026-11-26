<template>
  <div class="page">
    <div class="page-card">
      <h2 class="page-title">操作审计</h2>
      <p class="page-desc">记录流程定义、版本、节点注册的所有变更操作(含失败操作)，可查看操作前后的数据。</p>
    </div>

    <div class="page-card">
      <el-form :model="query" inline class="filter-form" @submit.prevent="search">
        <el-form-item label="模块">
          <el-select v-model="query.module" clearable placeholder="全部" style="width: 120px">
            <el-option v-for="(label, key) in AUDIT_MODULES" :key="key" :label="label" :value="key" />
          </el-select>
        </el-form-item>
        <el-form-item label="操作">
          <el-select v-model="query.operate_type" clearable placeholder="全部" style="width: 120px">
            <el-option v-for="(label, key) in AUDIT_OPS" :key="key" :label="label" :value="key" />
          </el-select>
        </el-form-item>
        <el-form-item label="流程编号"><el-input v-model="query.chain_no" clearable @keyup.enter="search" /></el-form-item>
        <el-form-item label="操作对象"><el-input v-model="query.target_id" clearable placeholder="精确匹配" @keyup.enter="search" /></el-form-item>
        <el-form-item label="操作人"><el-input v-model="query.operator" clearable @keyup.enter="search" /></el-form-item>
        <el-form-item label="结果">
          <el-select v-model="query.result" clearable placeholder="全部" style="width: 100px">
            <el-option label="成功" value="1" />
            <el-option label="失败" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间">
          <el-date-picker v-model="range" type="datetimerange" value-format="YYYY-MM-DD HH:mm:ss" start-placeholder="开始" end-placeholder="结束" style="width: 340px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="'Search'" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="list" border>
        <el-table-column prop="create_time" label="时间" width="170" />
        <el-table-column label="模块" width="100">
          <template #default="{ row }">{{ AUDIT_MODULES[row.module] ?? row.module }}</template>
        </el-table-column>
        <el-table-column label="操作" width="90">
          <template #default="{ row }">
            <el-tag size="small" effect="plain">{{ AUDIT_OPS[row.operate_type] ?? row.operate_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="summary" label="摘要" min-width="220" show-overflow-tooltip />
        <el-table-column label="对象" width="140" show-overflow-tooltip>
          <template #default="{ row }"><span class="mono">{{ row.target_id }}</span></template>
        </el-table-column>
        <el-table-column label="结果" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.result === 1 ? 'success' : 'danger'">{{ row.result === 1 ? '成功' : '失败' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="error_msg" label="失败原因" min-width="200" show-overflow-tooltip>
          <template #default="{ row }"><span class="text-danger">{{ row.error_msg }}</span></template>
        </el-table-column>
        <el-table-column prop="operator" label="操作人" width="110" />
        <el-table-column prop="client_ip" label="IP" width="120" />
        <el-table-column label="操作" width="70" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pager">
        <el-pagination
          v-model:current-page="query.page_num"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="load"
          @size-change="search"
        />
      </div>
    </div>

    <el-drawer v-model="detailOpen" title="审计详情" size="72%">
      <div v-loading="detailLoading">
        <template v-if="detail">
          <el-descriptions :column="3" border size="small">
            <el-descriptions-item label="时间">{{ detail.create_time }}</el-descriptions-item>
            <el-descriptions-item label="模块">{{ AUDIT_MODULES[detail.module] ?? detail.module }}</el-descriptions-item>
            <el-descriptions-item label="操作">{{ AUDIT_OPS[detail.operate_type] ?? detail.operate_type }}</el-descriptions-item>
            <el-descriptions-item label="摘要" :span="3">{{ detail.summary }}</el-descriptions-item>
            <el-descriptions-item label="对象"><span class="mono">{{ detail.target_id }}</span></el-descriptions-item>
            <el-descriptions-item label="流程编号">{{ detail.chain_no || '-' }}</el-descriptions-item>
            <el-descriptions-item label="版本ID">{{ detail.version_id || '-' }}</el-descriptions-item>
            <el-descriptions-item label="结果">
              <el-tag size="small" :type="detail.result === 1 ? 'success' : 'danger'">{{ detail.result === 1 ? '成功' : '失败' }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="操作人">{{ detail.operator }}</el-descriptions-item>
            <el-descriptions-item label="IP">{{ detail.client_ip }}</el-descriptions-item>
            <el-descriptions-item label="链路ID" :span="3"><span class="mono">{{ detail.trace_id }}</span></el-descriptions-item>
            <el-descriptions-item v-if="detail.error_msg" label="失败原因" :span="3"><span class="text-danger">{{ detail.error_msg }}</span></el-descriptions-item>
          </el-descriptions>
          <div class="diff">
            <div>
              <div class="diff__title">操作前</div>
              <JsonView :value="detail.before_data" max-height="calc(100vh - 360px)" />
            </div>
            <div>
              <div class="diff__title">操作后 / 请求参数</div>
              <JsonView :value="detail.after_data" max-height="calc(100vh - 360px)" />
            </div>
          </div>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { auditApi } from '@/api'
import JsonView from '@/components/JsonView.vue'
import { AUDIT_MODULES, AUDIT_OPS } from '@/constants'
import type { OperateLog } from '@/types'
import { rangeParams } from '@/utils'

const route = useRoute()
const loading = ref(false)
const list = ref<OperateLog[]>([])
const total = ref(0)
const range = ref<[string, string] | null>(null)
const query = reactive({ page_num: 1, page_size: 20, module: '', operate_type: '', chain_no: (route.query.chain_no as string) || '', target_id: '', operator: '', result: '' })

async function load() {
  loading.value = true
  try {
    const res = await auditApi.page({ ...query, ...rangeParams(range.value) })
    list.value = res.list
    total.value = res.total
  } finally {
    loading.value = false
  }
}

function search() {
  query.page_num = 1
  load()
}

function reset() {
  Object.assign(query, { page_num: 1, module: '', operate_type: '', chain_no: '', target_id: '', operator: '', result: '' })
  range.value = null
  load()
}

const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<OperateLog | null>(null)

async function openDetail(row: OperateLog) {
  detailOpen.value = true
  detailLoading.value = true
  detail.value = null
  try {
    detail.value = await auditApi.detail(row.log_id)
  } finally {
    detailLoading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.diff {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-top: 12px;
}

.diff__title {
  margin-bottom: 6px;
  color: #606266;
  font-size: 13px;
  font-weight: 600;
}
</style>
