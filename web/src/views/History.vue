<template>
  <div class="page">
    <div class="page-card">
      <h2 class="page-title">执行历史</h2>
      <p class="page-desc">展示 glow-flow 引擎写入状态存储(flow_instance_state / flow_node_state)的流程实例及节点执行情况。</p>
    </div>

    <div class="page-card">
      <el-form :model="query" inline class="filter-form" @submit.prevent="search">
        <el-form-item label="流程编号"><el-input v-model="query.chain_no" clearable @keyup.enter="search" /></el-form-item>
        <el-form-item label="实例ID"><el-input v-model="query.instance_id" clearable placeholder="模糊匹配" @keyup.enter="search" /></el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" clearable placeholder="全部" style="width: 110px">
            <el-option v-for="(s, key) in INSTANCE_STATUS" :key="key" :label="s.label" :value="key" />
          </el-select>
        </el-form-item>
        <el-form-item label="开始时间">
          <el-date-picker v-model="range" type="datetimerange" value-format="YYYY-MM-DD HH:mm:ss" start-placeholder="开始" end-placeholder="结束" style="width: 340px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="'Search'" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="list" border>
        <el-table-column label="实例ID" min-width="250">
          <template #default="{ row }">
            <el-link type="primary" class="mono" @click="openDetail(row)">{{ row.instance_id }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="流程" min-width="180">
          <template #default="{ row }">
            <div>{{ row.chain_name || '-' }}</div>
            <div class="text-muted mono">{{ row.chain_no }}</div>
          </template>
        </el-table-column>
        <el-table-column label="版本" width="90">
          <template #default="{ row }">{{ row.version_no ? `v${row.version_no}` : row.version }}</template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="INSTANCE_STATUS[row.status]?.type ?? 'info'">{{ INSTANCE_STATUS[row.status]?.label ?? row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="开始时间" width="170" />
        <el-table-column label="结束时间" width="170">
          <template #default="{ row }">{{ row.finished_at || '-' }}</template>
        </el-table-column>
        <el-table-column label="耗时" width="100">
          <template #default="{ row }">{{ row.finished_at ? formatDuration(row.duration_ms) : '-' }}</template>
        </el-table-column>
        <el-table-column prop="error" label="错误" min-width="200" show-overflow-tooltip>
          <template #default="{ row }"><span class="text-danger">{{ row.error }}</span></template>
        </el-table-column>
      </el-table>
      <div class="pager">
        <el-pagination
          v-model:current-page="query.page_num"
          v-model:page-size="query.page_size"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @current-change="load"
          @size-change="search"
        />
      </div>
    </div>

    <el-drawer v-model="detailOpen" title="执行详情" size="78%" destroy-on-close>
      <div v-loading="detailLoading" class="detail">
        <template v-if="detail">
          <el-descriptions :column="3" border size="small">
            <el-descriptions-item label="实例ID"><span class="mono">{{ detail.instance.instance_id }}</span></el-descriptions-item>
            <el-descriptions-item label="流程">{{ detail.instance.chain_name }}({{ detail.instance.chain_no }})</el-descriptions-item>
            <el-descriptions-item label="版本">{{ detail.instance.version_no ? `v${detail.instance.version_no}` : detail.instance.version }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag size="small" :type="INSTANCE_STATUS[detail.instance.status]?.type ?? 'info'">{{ INSTANCE_STATUS[detail.instance.status]?.label ?? detail.instance.status }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="开始时间">{{ detail.instance.created_at }}</el-descriptions-item>
            <el-descriptions-item label="耗时">{{ detail.instance.finished_at ? formatDuration(detail.instance.duration_ms) : '-' }}</el-descriptions-item>
            <el-descriptions-item v-if="detail.instance.error" label="错误" :span="3"><span class="text-danger">{{ detail.instance.error }}</span></el-descriptions-item>
          </el-descriptions>

          <div class="detail__graph">
            <FlowCanvas v-if="detail.design_nodes.length" ref="canvas" readonly :minimap="false" :types="typeMap" @select="onSelect" />
            <el-empty v-else description="未找到执行版本的流程图" :image-size="60" />
          </div>

          <el-tabs>
            <el-tab-pane :label="`节点执行(${detail.nodes.length})`">
              <el-table :data="detail.nodes" border size="small" row-key="node_id" :row-class-name="rowClass">
                <el-table-column type="expand">
                  <template #default="{ row }">
                    <div class="detail__io">
                      <div><div class="detail__io-title">输入</div><JsonView :value="row.input" max-height="260px" /></div>
                      <div><div class="detail__io-title">输出</div><JsonView :value="row.output" max-height="260px" /></div>
                    </div>
                    <div v-if="row.metadata" class="detail__meta"><div class="detail__io-title">元数据</div><JsonView :value="row.metadata" max-height="160px" /></div>
                  </template>
                </el-table-column>
                <el-table-column label="节点" min-width="180">
                  <template #default="{ row }">
                    <div>{{ row.node_name || row.node_id }}</div>
                    <div class="text-muted mono">{{ row.node_id }} · {{ row.node_type || '-' }}</div>
                  </template>
                </el-table-column>
                <el-table-column label="状态" width="90">
                  <template #default="{ row }">
                    <el-tag size="small" :type="INSTANCE_STATUS[row.status]?.type ?? 'info'">{{ INSTANCE_STATUS[row.status]?.label ?? row.status }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="attempts" label="尝试次数" width="90" />
                <el-table-column prop="started_at" label="开始" width="170" />
                <el-table-column label="耗时" width="100">
                  <template #default="{ row }">{{ row.finished_at ? formatDuration(row.duration_ms) : '-' }}</template>
                </el-table-column>
                <el-table-column prop="error" label="错误" min-width="200" show-overflow-tooltip>
                  <template #default="{ row }"><span class="text-danger">{{ row.error }}</span></template>
                </el-table-column>
              </el-table>
            </el-tab-pane>
            <el-tab-pane label="流程变量"><JsonView :value="detail.instance.variables" /></el-tab-pane>
            <el-tab-pane v-if="detail.instance.current_node_ids?.length" label="当前节点">
              <el-tag v-for="id in detail.instance.current_node_ids" :key="id" class="detail__current">{{ id }}</el-tag>
            </el-tab-pane>
          </el-tabs>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { historyApi, nodeTypeApi } from '@/api'
import FlowCanvas, { type CanvasSelection } from '@/components/designer/FlowCanvas.vue'
import { edgeKey, type NodeRunStatus } from '@/components/designer/graph'
import JsonView from '@/components/JsonView.vue'
import { INSTANCE_STATUS } from '@/constants'
import type { InstanceDetail, InstanceInfo, NodeStateInfo, NodeTypeInfo } from '@/types'
import { formatDuration, rangeParams } from '@/utils'

const loading = ref(false)
const list = ref<InstanceInfo[]>([])
const total = ref(0)
const range = ref<[string, string] | null>(null)
const query = reactive({ page_num: 1, page_size: 20, chain_no: '', instance_id: '', status: '' })

async function load() {
  loading.value = true
  try {
    const res = await historyApi.page({ ...query, ...rangeParams(range.value) })
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
  Object.assign(query, { page_num: 1, chain_no: '', instance_id: '', status: '' })
  range.value = null
  load()
}

// ---------- 详情 ----------
const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<InstanceDetail | null>(null)
const types = ref<NodeTypeInfo[]>([])
const typeMap = computed(() => new Map(types.value.map((t) => [t.node_type, t])))
const canvas = ref<InstanceType<typeof FlowCanvas>>()
const selectedNodeId = ref('')

const statusMap: Record<string, NodeRunStatus> = {
  completed: 'success',
  failed: 'error',
  running: 'running',
  pending: 'pending',
  skipped: 'pending',
}

async function openDetail(row: InstanceInfo) {
  detailOpen.value = true
  detailLoading.value = true
  detail.value = null
  selectedNodeId.value = ''
  try {
    const [res, nodeTypes] = await Promise.all([historyApi.detail(row.instance_id), types.value.length ? Promise.resolve(types.value) : nodeTypeApi.list()])
    types.value = nodeTypes
    detail.value = res
    await nextTick()
    if (canvas.value && res.design_nodes.length) {
      canvas.value.load(res.design_nodes, res.design_connections)
      const status: Record<string, NodeRunStatus> = {}
      for (const node of res.nodes) status[node.node_id] = statusMap[node.status] ?? 'pending'
      const edges = new Set<string>()
      for (const conn of res.design_connections) {
        if (status[conn.from_id] === 'success' && status[conn.to_id]) edges.add(edgeKey(conn.from_id, conn.to_id, conn.type))
      }
      canvas.value.showStatus(status, edges)
    }
  } finally {
    detailLoading.value = false
  }
}

function onSelect(sel: CanvasSelection) {
  selectedNodeId.value = sel.kind === 'node' ? sel.id : ''
}

function rowClass({ row }: { row: NodeStateInfo }) {
  return row.node_id === selectedNodeId.value ? 'is-selected-row' : ''
}

onMounted(load)
</script>

<style scoped>
.detail {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 200px;
}

.detail__graph {
  height: 340px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  overflow: hidden;
}

.detail__io {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  padding: 4px 12px;
}

.detail__meta {
  padding: 4px 12px;
}

.detail__io-title {
  margin-bottom: 4px;
  color: #606266;
  font-size: 12px;
  font-weight: 600;
}

.detail__current {
  margin-right: 6px;
}

:deep(.is-selected-row) {
  --el-table-tr-bg-color: #ecf5ff;
}
</style>
