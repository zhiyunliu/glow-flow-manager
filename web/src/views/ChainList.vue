<template>
  <div class="page">
    <div class="page-card">
      <h2 class="page-title">流程定义</h2>
      <p class="page-desc">维护规则链/工作流定义及其版本。发布后的版本写入 glow-flow 引擎的流程表，由引擎直接加载执行。</p>
    </div>

    <div class="page-card">
      <el-form :model="query" inline class="filter-form" @submit.prevent="search">
        <el-form-item label="流程编号"><el-input v-model="query.chain_no" clearable placeholder="模糊匹配" @keyup.enter="search" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="query.name" clearable placeholder="模糊匹配" @keyup.enter="search" /></el-form-item>
        <el-form-item label="类型">
          <el-select v-model="query.flow_type" clearable placeholder="全部" style="width: 120px">
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
          <el-button type="primary" :icon="'Plus'" @click="openCreate">新建流程</el-button>
          <el-button :icon="'Upload'" @click="openImport">导入</el-button>
        </div>
        <el-button :icon="'Refresh'" circle @click="load" />
      </div>

      <el-table v-loading="loading" :data="list" border>
        <el-table-column prop="chain_no" label="流程编号" min-width="140">
          <template #default="{ row }">
            <span class="mono">{{ row.chain_no }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="160" show-overflow-tooltip />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="row.flow_type === 'WORKFLOW' ? 'warning' : 'primary'" effect="plain">{{ flowTypeLabel(row.flow_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="启用" width="80">
          <template #default="{ row }">
            <el-switch :model-value="row.status === 1" @change="(v: string | number | boolean) => changeStatus(row, !!v)" />
          </template>
        </el-table-column>
        <el-table-column label="生效版本" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.effective_version_id" type="success" size="small">v{{ row.effective_version_no }}</el-tag>
            <span v-else class="text-muted">未发布</span>
          </template>
        </el-table-column>
        <el-table-column label="草稿" width="80">
          <template #default="{ row }">
            <el-tag v-if="row.draft_version_id" type="info" size="small">v{{ row.draft_version_no }}</el-tag>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="version_count" label="版本数" width="70" />
        <el-table-column prop="desc" label="描述" min-width="160" show-overflow-tooltip />
        <el-table-column label="更新" width="170">
          <template #default="{ row }">
            <div>{{ row.update_time }}</div>
            <div class="text-muted">{{ row.operator }}</div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="290" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="design(row)">设计</el-button>
            <el-button link type="primary" @click="openVersions(row)">版本</el-button>
            <el-button link @click="openEdit(row)">编辑</el-button>
            <el-button link @click="openCopy(row)">复制</el-button>
            <el-button link :disabled="!row.version_count" @click="exportChain(row)">导出</el-button>
            <el-tooltip :disabled="row.published_count === 0" content="已发布过的流程不能删除，可禁用" placement="top">
              <span class="ml">
                <el-button link type="danger" :disabled="row.published_count > 0" @click="remove(row)">删除</el-button>
              </span>
            </el-tooltip>
          </template>
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

    <!-- 新建/编辑 -->
    <el-dialog v-model="editOpen" :title="editing ? '编辑流程' : '新建流程'" width="520px">
      <el-form ref="editFormRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="流程编号" prop="chain_no">
          <el-input v-model="form.chain_no" :disabled="editing" placeholder="字母开头，字母数字下划线中划线，最长 32 位" maxlength="32" />
        </el-form-item>
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" maxlength="128" /></el-form-item>
        <el-form-item label="流程类型" prop="flow_type">
          <el-radio-group v-model="form.flow_type" :disabled="editing && (editingRow?.published_count ?? 0) > 0">
            <el-radio v-for="t in FLOW_TYPES" :key="t.value" :value="t.value">{{ t.label }}</el-radio>
          </el-radio-group>
          <div class="text-muted form-hint">{{ FLOW_TYPES.find((t) => t.value === form.flow_type)?.desc }}</div>
        </el-form-item>
        <el-form-item label="描述"><el-input v-model="form.desc" type="textarea" :rows="3" maxlength="500" show-word-limit /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editOpen = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitEdit">{{ editing ? '保存' : '创建并设计' }}</el-button>
      </template>
    </el-dialog>

    <!-- 复制 -->
    <el-dialog v-model="copyOpen" title="复制流程" width="480px">
      <el-form :model="copyForm" label-width="100px">
        <el-form-item label="来源">{{ copyRow?.name }}({{ copyRow?.chain_no }})</el-form-item>
        <el-form-item label="来源版本">
          <span class="text-muted">{{ copyRow?.effective_version_id ? `生效版本 v${copyRow.effective_version_no}` : '最新版本' }}</span>
        </el-form-item>
        <el-form-item label="新流程编号" required><el-input v-model="copyForm.new_chain_no" maxlength="32" /></el-form-item>
        <el-form-item label="新流程名称" required><el-input v-model="copyForm.new_name" maxlength="128" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="copyOpen = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitCopy">复制</el-button>
      </template>
    </el-dialog>

    <!-- 导入 -->
    <el-dialog v-model="importOpen" title="导入流程" width="640px">
      <el-form :model="importForm" label-width="90px">
        <el-form-item label="导入方式">
          <el-radio-group v-model="importForm.mode">
            <el-radio value="new_chain">创建新流程</el-radio>
            <el-radio value="new_version">作为已有流程的草稿</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="流程编号">
          <el-input v-model="importForm.chain_no" :placeholder="importForm.mode === 'new_chain' ? '为空时使用文件中的编号' : '已有流程的编号，为空时使用文件中的编号'" />
        </el-form-item>
        <el-form-item v-if="importForm.mode === 'new_chain'" label="名称">
          <el-input v-model="importForm.name" placeholder="为空时使用文件中的名称" />
        </el-form-item>
        <el-form-item label="内容">
          <div class="import-box">
            <el-upload :auto-upload="false" :show-file-list="false" accept=".json,application/json" :on-change="onImportFile">
              <el-button size="small" :icon="'Upload'">选择文件</el-button>
            </el-upload>
            <el-input v-model="importForm.content" type="textarea" :rows="10" class="mono" placeholder="或粘贴导出的 JSON / glow-flow ChainDefinition JSON" spellcheck="false" />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importOpen = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitImport">导入</el-button>
      </template>
    </el-dialog>

    <VersionDrawer v-model="versionOpen" :chain="versionChain" @changed="load" />
  </div>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox, type FormInstance, type FormRules, type UploadFile } from 'element-plus'
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { chainApi, versionApi } from '@/api'
import { ApiError } from '@/api/http'
import VersionDrawer from '@/components/VersionDrawer.vue'
import { FLOW_TYPES, flowTypeLabel } from '@/constants'
import type { ChainInfo, FlowType } from '@/types'
import { downloadJson, readFileText } from '@/utils'

const router = useRouter()
const loading = ref(false)
const submitting = ref(false)
const list = ref<ChainInfo[]>([])
const total = ref(0)
const query = reactive({ page_num: 1, page_size: 20, chain_no: '', name: '', flow_type: '', status: '' })

async function load() {
  loading.value = true
  try {
    const res = await chainApi.page(query)
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
  Object.assign(query, { page_num: 1, chain_no: '', name: '', flow_type: '', status: '' })
  load()
}

// ---------- 新建/编辑 ----------
const editOpen = ref(false)
const editing = ref(false)
const editingRow = ref<ChainInfo | null>(null)
const editFormRef = ref<FormInstance>()
const form = reactive({ chain_no: '', name: '', flow_type: 'RULE_CHAIN' as FlowType, desc: '' })
const rules: FormRules = {
  chain_no: [
    { required: true, message: '请输入流程编号' },
    { pattern: /^[A-Za-z][A-Za-z0-9_-]{0,31}$/, message: '字母开头，只能包含字母、数字、下划线和中划线' },
  ],
  name: [{ required: true, message: '请输入名称' }],
  flow_type: [{ required: true }],
}

function openCreate() {
  editing.value = false
  editingRow.value = null
  Object.assign(form, { chain_no: '', name: '', flow_type: 'RULE_CHAIN', desc: '' })
  editOpen.value = true
  editFormRef.value?.clearValidate()
}

function openEdit(row: ChainInfo) {
  editing.value = true
  editingRow.value = row
  Object.assign(form, { chain_no: row.chain_no, name: row.name, flow_type: row.flow_type, desc: row.desc })
  editOpen.value = true
}

async function submitEdit() {
  await editFormRef.value?.validate()
  submitting.value = true
  try {
    if (editing.value) {
      await chainApi.update({ ...form })
      ElMessage.success('已保存')
      editOpen.value = false
      load()
    } else {
      const chain = await chainApi.create({ ...form })
      editOpen.value = false
      router.push(`/designer/${chain.draft_version_id}`)
    }
  } finally {
    submitting.value = false
  }
}

async function changeStatus(row: ChainInfo, enabled: boolean) {
  if (!enabled) {
    await ElMessageBox.confirm(`禁用后引擎将不再加载流程「${row.name}」，确定禁用吗？`, '禁用流程', { type: 'warning' })
  }
  await chainApi.changeStatus(row.chain_no, enabled ? 1 : 0)
  ElMessage.success(enabled ? '已启用' : '已禁用')
  load()
}

/** 设计：打开草稿；没有草稿时基于生效版本(或最新版本)创建草稿 */
async function design(row: ChainInfo) {
  if (row.draft_version_id) {
    router.push(`/designer/${row.draft_version_id}`)
    return
  }
  const draft = await versionApi.createDraft(row.chain_no)
  ElMessage.success(`已创建草稿 v${draft.version_no}`)
  router.push(`/designer/${draft.version_id}`)
}

async function remove(row: ChainInfo) {
  await ElMessageBox.confirm(`确定删除流程「${row.name}」及其所有草稿吗？`, '删除流程', { type: 'warning' })
  await chainApi.remove(row.chain_no)
  ElMessage.success('已删除')
  load()
}

async function exportChain(row: ChainInfo) {
  const res = await chainApi.export(row.chain_no)
  downloadJson(res.file_name, res.content)
}

// ---------- 版本 ----------
const versionOpen = ref(false)
const versionChain = ref<ChainInfo | null>(null)
function openVersions(row: ChainInfo) {
  versionChain.value = row
  versionOpen.value = true
}

// ---------- 复制 ----------
const copyOpen = ref(false)
const copyRow = ref<ChainInfo | null>(null)
const copyForm = reactive({ new_chain_no: '', new_name: '' })
function openCopy(row: ChainInfo) {
  copyRow.value = row
  Object.assign(copyForm, { new_chain_no: `${row.chain_no}_copy`.slice(0, 32), new_name: `${row.name}(副本)` })
  copyOpen.value = true
}

async function submitCopy() {
  if (!copyRow.value) return
  submitting.value = true
  try {
    const res = await chainApi.copy({ chain_no: copyRow.value.chain_no, ...copyForm })
    copyOpen.value = false
    ElMessage.success('已复制')
    router.push(`/designer/${res.version_id}`)
  } finally {
    submitting.value = false
  }
}

// ---------- 导入 ----------
const importOpen = ref(false)
const importForm = reactive({ mode: 'new_chain', chain_no: '', name: '', content: '' })
function openImport() {
  Object.assign(importForm, { mode: 'new_chain', chain_no: '', name: '', content: '' })
  importOpen.value = true
}

async function onImportFile(file: UploadFile) {
  if (file.raw) importForm.content = await readFileText(file.raw)
}

async function submitImport() {
  if (!importForm.content.trim()) {
    ElMessage.warning('请选择文件或粘贴 JSON 内容')
    return
  }
  submitting.value = true
  try {
    const res = await chainApi.import({ ...importForm }, { silent: true })
    importOpen.value = false
    const remapped = Object.keys(res.remapped ?? {}).length
    ElMessage.success(remapped ? `导入成功，${remapped} 个节点编号与其他流程冲突已重新生成` : '导入成功')
    router.push(`/designer/${res.version_id}`)
  } catch (err) {
    const detail = err instanceof ApiError && err.data?.errors?.length ? `：${err.data.errors[0].message}` : ''
    ElMessage.error(`${(err as Error).message}${detail && !(err as Error).message.includes(detail.slice(1)) ? detail : ''}`)
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.ml {
  margin-left: 12px;
}

.form-hint {
  width: 100%;
  font-size: 12px;
  line-height: 18px;
}

.import-box {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}
</style>
