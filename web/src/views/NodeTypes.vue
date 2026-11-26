<template>
  <div class="page">
    <div class="page-card">
      <h2 class="page-title">节点注册</h2>
      <p class="page-desc">
        维护设计器工具箱中的节点类型。内置节点来自 glow-flow 引擎；独立服务节点由单独部署的服务提供，需注册到注册中心(如 nacos)，并由引擎宿主注册对应的节点类型。
      </p>
    </div>

    <div class="page-card">
      <el-form :model="query" inline class="filter-form" @submit.prevent="search">
        <el-form-item label="关键字"><el-input v-model="query.keyword" clearable placeholder="类型或名称" @keyup.enter="search" /></el-form-item>
        <el-form-item label="分类">
          <el-select v-model="query.category" clearable placeholder="全部" style="width: 120px">
            <el-option v-for="c in NODE_CATEGORIES" :key="c.value" :label="c.label" :value="c.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="来源">
          <el-select v-model="query.source" clearable placeholder="全部" style="width: 110px">
            <el-option v-for="(s, key) in NODE_SOURCES" :key="key" :label="s.label" :value="key" />
          </el-select>
        </el-form-item>
        <el-form-item label="流程类型">
          <el-select v-model="query.flow_mode" clearable placeholder="全部" style="width: 110px">
            <el-option v-for="t in FLOW_TYPES" :key="t.value" :label="t.label" :value="t.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" clearable placeholder="全部" style="width: 100px">
            <el-option label="启用" value="1" />
            <el-option label="禁用" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="'Search'" @click="search">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>

      <div class="toolbar">
        <div>
          <el-button type="primary" :icon="'Plus'" @click="openCreate">注册服务节点</el-button>
          <el-button :icon="'Refresh'" :loading="syncing" @click="syncBuiltin">同步内置节点</el-button>
        </div>
        <el-button :icon="'Refresh'" circle @click="load" />
      </div>

      <el-table v-loading="loading" :data="list" border>
        <el-table-column label="节点类型" min-width="170">
          <template #default="{ row }">
            <span class="dot" :style="{ background: row.color || DEFAULT_NODE_COLOR }" />
            <span class="mono">{{ row.node_type }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column label="分类" width="90">
          <template #default="{ row }">{{ categoryLabel(row.category) }}</template>
        </el-table-column>
        <el-table-column label="适用流程" width="150">
          <template #default="{ row }">
            <el-tag v-for="m in row.flow_modes" :key="m" size="small" effect="plain" class="tag-gap">{{ flowTypeLabel(m) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="来源" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="NODE_SOURCES[row.source]?.type">{{ NODE_SOURCES[row.source]?.label ?? row.source }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="服务" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <template v-if="row.service_name">
              <div class="mono">{{ row.service_name }}</div>
              <div class="text-muted mono">{{ row.protocol }} {{ row.service_path }}</div>
            </template>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="输出类型" min-width="150">
          <template #default="{ row }">
            <el-tag v-for="r in row.relation_types" :key="r" size="small" :color="relationColor(r)" effect="dark" class="tag-gap rel-tag">{{ r }}</el-tag>
            <span v-if="!row.relation_types.length" class="text-muted">按配置</span>
          </template>
        </el-table-column>
        <el-table-column label="引擎" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.engine_supported" size="small" type="success" effect="plain">已实现</el-tag>
            <el-tag v-else-if="row.source === 'service'" size="small" type="warning" effect="plain">宿主注册</el-tag>
            <el-tag v-else size="small" type="info" effect="plain">未实现</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="usage_count" label="使用" width="70" />
        <el-table-column label="启用" width="70">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" size="small" @change="(v: string | number | boolean) => changeStatus(row, !!v)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.source === 'service'" link type="primary" @click="openInstances(row)">实例</el-button>
            <el-button v-if="row.source === 'service'" link type="danger" :disabled="row.usage_count > 0" @click="remove(row)">删除</el-button>
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

    <!-- 编辑 -->
    <el-dialog v-model="editOpen" :title="editing ? `编辑节点类型 - ${form.node_type}` : '注册服务节点'" width="760px" top="5vh">
      <el-alert v-if="isBuiltin" type="info" :closable="false" show-icon title="内置节点的分类、输出类型与属性描述由引擎决定，只能修改名称、适用流程类型与展示信息" class="mb" />
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="节点类型" prop="node_type"><el-input v-model="form.node_type" :disabled="editing" placeholder="如 risk.check" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="分类" prop="category">
              <el-select v-model="form.category" :disabled="isBuiltin" style="width: 100%">
                <el-option v-for="c in NODE_CATEGORIES" :key="c.value" :label="c.label" :value="c.value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="适用流程" prop="flow_modes">
              <el-checkbox-group v-model="form.flow_modes">
                <el-checkbox v-for="t in FLOW_TYPES" :key="t.value" :value="t.value">{{ t.label }}</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </el-col>
          <template v-if="!isBuiltin">
            <el-col :span="12">
              <el-form-item label="服务名称"><el-input v-model="form.service_name" placeholder="注册中心中的服务名" /></el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="协议" label-width="60px">
                <el-select v-model="form.protocol" clearable style="width: 100%">
                  <el-option label="grpc" value="grpc" />
                  <el-option label="http" value="http" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="6">
              <el-form-item label="路径" label-width="60px"><el-input v-model="form.service_path" placeholder="/node/execute" /></el-form-item>
            </el-col>
          </template>
          <el-col :span="24">
            <el-form-item label="输出类型">
              <el-select v-model="form.relation_types" multiple filterable allow-create default-first-option :disabled="isBuiltin" placeholder="节点执行结果的关系类型，如 success / failure" style="width: 100%">
                <el-option v-for="r in ['success', 'failure', 'true', 'false']" :key="r" :label="r" :value="r" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="图标">
              <el-select v-model="form.icon" filterable clearable placeholder="Element Plus 图标" style="width: 100%">
                <template #prefix>
                  <el-icon v-if="form.icon && form.icon in Icons"><component :is="form.icon" /></el-icon>
                </template>
                <el-option v-for="name in iconNames" :key="name" :label="name" :value="name">
                  <el-icon><component :is="name" /></el-icon>
                  <span class="icon-name">{{ name }}</span>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="颜色"><el-color-picker v-model="form.color" /></el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="排序"><el-input-number v-model="form.sort_no" controls-position="right" style="width: 100%" /></el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="描述"><el-input v-model="form.desc" type="textarea" :rows="2" /></el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="属性描述">
              <div class="schema-box">
                <el-input v-model="schemaText" type="textarea" class="mono" :rows="10" :disabled="isBuiltin" spellcheck="false" />
                <div class="schema-help text-muted">
                  JSON：<code>{"fields":[{"key":"level","label":"等级","type":"select","required":true,"options":[{"label":"高","value":"high"}]}]}</code>，
                  type 可选 {{ FIELD_TYPES.map((t) => t.value).join(' / ') }}；可用 group 分组、show_when 按其他字段取值显示；select 可用 options_source 引用动态下拉数据源(如 email_servers)。
                </div>
                <div v-if="schemaError" class="text-danger">{{ schemaError }}</div>
              </div>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="默认配置">
              <el-input v-model="defaultText" type="textarea" class="mono" :rows="3" :disabled="isBuiltin" placeholder="{}" spellcheck="false" />
              <div v-if="defaultError" class="text-danger">{{ defaultError }}</div>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="editOpen = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
      </template>
    </el-dialog>

    <!-- 实例 -->
    <el-dialog v-model="instanceOpen" :title="`服务实例 - ${instanceRow?.service_name ?? ''}`" width="760px">
      <div v-loading="instanceLoading">
        <el-alert v-if="instanceError" type="warning" :closable="false" show-icon :title="instanceError" />
        <template v-else-if="instances">
          <div class="text-muted mb">注册中心：{{ instances.registry }}，共 {{ instances.instances.length }} 个实例</div>
          <el-table :data="instances.instances" border size="small">
            <el-table-column prop="id" label="实例ID" min-width="160" show-overflow-tooltip />
            <el-table-column label="健康" width="70">
              <template #default="{ row }">
                <el-tag size="small" :type="row.healthy ? 'success' : 'danger'">{{ row.healthy ? '健康' : '异常' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="version" label="版本" width="100" />
            <el-table-column prop="weight" label="权重" width="70" />
            <el-table-column label="地址" min-width="200">
              <template #default="{ row }">
                <div v-for="ep in row.endpoints" :key="ep" class="mono">{{ ep }}</div>
              </template>
            </el-table-column>
          </el-table>
        </template>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import * as Icons from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { computed, onMounted, reactive, ref } from 'vue'
import { nodeTypeApi } from '@/api'
import { ApiError } from '@/api/http'
import { categoryLabel, DEFAULT_NODE_COLOR, FIELD_TYPES, FLOW_TYPES, flowTypeLabel, NODE_CATEGORIES, NODE_SOURCES, relationColor } from '@/constants'
import type { FlowType, NodeTypeInfo, NodeTypeInstances } from '@/types'
import { parseJson, prettyJson } from '@/utils'

const iconNames = Object.keys(Icons)
const loading = ref(false)
const syncing = ref(false)
const submitting = ref(false)
const list = ref<NodeTypeInfo[]>([])
const total = ref(0)
const query = reactive({ page_num: 1, page_size: 50, keyword: '', category: '', source: '', flow_mode: '', status: '' })

async function load() {
  loading.value = true
  try {
    const res = await nodeTypeApi.page(query)
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
  Object.assign(query, { page_num: 1, keyword: '', category: '', source: '', flow_mode: '', status: '' })
  load()
}

async function syncBuiltin() {
  syncing.value = true
  try {
    const res = await nodeTypeApi.syncBuiltin()
    ElMessage.success(`同步完成：新增 ${res.inserted.length} 个，更新 ${res.updated.length} 个`)
    load()
  } finally {
    syncing.value = false
  }
}

async function changeStatus(row: NodeTypeInfo, enabled: boolean) {
  await nodeTypeApi.changeStatus(row.node_type, enabled ? 1 : 0)
  ElMessage.success(enabled ? '已启用' : '已禁用(设计器工具箱中不再显示)')
  load()
}

async function remove(row: NodeTypeInfo) {
  await ElMessageBox.confirm(`确定删除节点类型 ${row.node_type} 吗？`, '删除', { type: 'warning' })
  await nodeTypeApi.remove(row.node_type)
  ElMessage.success('已删除')
  load()
}

// ---------- 编辑 ----------
const editOpen = ref(false)
const editing = ref(false)
const editingSource = ref<'builtin' | 'service'>('service')
const isBuiltin = computed(() => editing.value && editingSource.value === 'builtin')
const formRef = ref<FormInstance>()
const emptyForm = () => ({
  node_type: '',
  name: '',
  category: 'integration',
  flow_modes: ['RULE_CHAIN'] as FlowType[],
  service_name: '',
  service_path: '',
  protocol: 'grpc',
  relation_types: ['success', 'failure'] as string[],
  icon: 'Connection',
  color: '',
  sort_no: 900,
  desc: '',
})
const form = reactive(emptyForm())
const schemaText = ref('')
const defaultText = ref('')
const schemaError = ref('')
const defaultError = ref('')
const rules: FormRules = {
  node_type: [
    { required: true, message: '请输入节点类型' },
    { pattern: /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/, message: '以字母开头，只能包含字母、数字、下划线、中划线和点' },
  ],
  name: [{ required: true, message: '请输入名称' }],
  category: [{ required: true }],
  flow_modes: [{ type: 'array', required: true, min: 1, message: '至少选择一种流程类型' }],
}

function openCreate() {
  editing.value = false
  editingSource.value = 'service'
  Object.assign(form, emptyForm())
  schemaText.value = prettyJson({ fields: [] })
  defaultText.value = '{}'
  schemaError.value = defaultError.value = ''
  editOpen.value = true
  formRef.value?.clearValidate()
}

function openEdit(row: NodeTypeInfo) {
  editing.value = true
  editingSource.value = row.source
  Object.assign(form, {
    node_type: row.node_type,
    name: row.name,
    category: row.category,
    flow_modes: [...row.flow_modes],
    service_name: row.service_name,
    service_path: row.service_path,
    protocol: row.protocol,
    relation_types: [...row.relation_types],
    icon: row.icon,
    color: row.color,
    sort_no: row.sort_no,
    desc: row.desc,
  })
  schemaText.value = prettyJson(row.config_schema ?? { fields: [] })
  defaultText.value = prettyJson(row.default_config ?? {})
  schemaError.value = defaultError.value = ''
  editOpen.value = true
}

async function submit() {
  await formRef.value?.validate()
  const [schema, schemaErr] = parseJson(schemaText.value)
  const [defaults, defaultErr] = parseJson(defaultText.value || '{}')
  schemaError.value = schemaErr
  defaultError.value = defaultErr || (defaults !== undefined && (typeof defaults !== 'object' || Array.isArray(defaults)) ? '默认配置必须是 JSON 对象' : '')
  if (schemaError.value || defaultError.value) return
  submitting.value = true
  try {
    const payload = { ...form, color: form.color || '', config_schema: schema ?? { fields: [] }, default_config: defaults ?? {} }
    if (editing.value) await nodeTypeApi.update(payload)
    else await nodeTypeApi.create(payload)
    ElMessage.success('已保存')
    editOpen.value = false
    load()
  } finally {
    submitting.value = false
  }
}

// ---------- 实例 ----------
const instanceOpen = ref(false)
const instanceLoading = ref(false)
const instanceRow = ref<NodeTypeInfo | null>(null)
const instances = ref<NodeTypeInstances | null>(null)
const instanceError = ref('')

async function openInstances(row: NodeTypeInfo) {
  instanceRow.value = row
  instances.value = null
  instanceError.value = ''
  instanceOpen.value = true
  instanceLoading.value = true
  try {
    instances.value = await nodeTypeApi.instances(row.node_type)
  } catch (err) {
    instanceError.value = err instanceof ApiError ? err.message : String(err)
  } finally {
    instanceLoading.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.tag-gap {
  margin: 2px 4px 2px 0;
}

.rel-tag {
  border: none;
}

.mb {
  margin-bottom: 12px;
}

.schema-box {
  display: flex;
  flex-direction: column;
  gap: 4px;
  width: 100%;
}

.schema-help {
  font-size: 12px;
  line-height: 18px;
}

.schema-help code {
  padding: 0 4px;
  background: #f5f7fa;
  border-radius: 3px;
  font-size: 11px;
}

.icon-name {
  margin-left: 8px;
}
</style>
