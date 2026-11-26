<template>
  <div v-loading="loading" class="designer">
    <header class="designer__toolbar">
      <el-button :icon="'ArrowLeft'" link @click="goBack">返回</el-button>
      <el-divider direction="vertical" />
      <template v-if="design">
        <div class="designer__title">
          <span class="designer__name">{{ design.chain.name }}</span>
          <span class="text-muted">{{ design.chain.chain_no }}</span>
          <el-tag size="small" effect="plain">v{{ design.version.version_no }}</el-tag>
          <el-tag size="small" :type="VERSION_STATES[design.version.publish_state]?.type">{{ VERSION_STATES[design.version.publish_state]?.label }}</el-tag>
          <el-tag v-if="design.version.is_effective" size="small" type="success" effect="dark">生效中</el-tag>
          <el-tag v-if="dirty" size="small" type="warning" effect="plain">未保存</el-tag>
        </div>
        <el-radio-group v-model="flowType" size="small" :disabled="!modeSwitchable" class="designer__mode" @change="onModeChange">
          <el-tooltip v-for="t in FLOW_TYPES" :key="t.value" :content="t.desc" placement="bottom">
            <el-radio-button :value="t.value">{{ t.label }}</el-radio-button>
          </el-tooltip>
        </el-radio-group>
        <div class="designer__spacer" />
        <template v-if="editable">
          <el-button-group size="small">
            <el-button :icon="'RefreshLeft'" title="撤销 Ctrl+Z" @click="canvas?.undo()" />
            <el-button :icon="'RefreshRight'" title="重做 Ctrl+Shift+Z" @click="canvas?.redo()" />
          </el-button-group>
          <el-button size="small" :icon="'Operation'" @click="canvas?.layout()">整理布局</el-button>
        </template>
        <el-button size="small" :icon="'CircleCheck'" :loading="validating" @click="serverValidate">校验</el-button>
        <el-button size="small" :icon="'VideoPlay'" @click="debugOpen = true">调试</el-button>
        <el-dropdown size="small" @command="onTransfer">
          <el-button size="small" :icon="'Sort'">导入/导出</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="import" :disabled="!editable" :icon="'Upload'">导入 JSON 到画布</el-dropdown-item>
              <el-dropdown-item command="export" :icon="'Download'">导出当前画布</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <template v-if="editable">
          <el-button size="small" type="primary" plain :icon="'DocumentChecked'" :loading="saving" @click="save()">保存</el-button>
          <el-button size="small" type="primary" :icon="'Promotion'" :loading="publishing" @click="publish">发布</el-button>
        </template>
        <el-button v-else size="small" type="primary" :icon="'DocumentAdd'" @click="newDraftFromThis">基于此版本新建草稿</el-button>
      </template>
    </header>

    <div v-if="design" class="designer__body">
      <NodePalette :types="allTypes" :flow-type="flowType" :disabled="!editable" @drag="(meta, evt) => canvas?.startDrag(meta, evt)" />
      <div class="designer__canvas">
        <el-alert
          v-if="!editable"
          type="info"
          :closable="false"
          show-icon
          class="designer__readonly"
          :title="`v${design.version.version_no} ${VERSION_STATES[design.version.publish_state]?.label}，不能修改。如需调整请基于此版本新建草稿。`"
        />
        <FlowCanvas :key="versionId" ref="canvas" :readonly="!editable" :types="typeMap" @select="onSelect" @change="onChange" />
        <ValidationPanel :result="shownValidation" :source="validationSource" @locate="locateIssue" />
      </div>
      <PropertyPanel
        v-model:version-desc="versionDesc"
        v-model:chain-ext="chainExt"
        :node="selectedNode"
        :edge="selectedEdge"
        :types="typeMap"
        :all-nodes="model.nodes"
        :edge-count="model.connections.length"
        :editable="editable"
        :chain="design.chain"
        @update-node="(id, patch) => canvas?.updateNodeData(id, patch)"
        @update-edge="(id, patch) => canvas?.updateEdgeData(id, patch)"
        @remove="(id) => canvas?.removeCell(id)"
        @update:version-desc="markDirty"
        @update:chain-ext="markDirty"
      />
    </div>

    <DebugPanel v-model="debugOpen" :result="debugResult" :running="debugging" @run="runDebug" @clear="clearDebug" @locate="(id) => canvas?.focusNode(id)" />

    <el-dialog v-model="importOpen" title="导入 JSON 到画布" width="640px">
      <el-alert type="info" :closable="false" show-icon title="支持本系统导出的文件或 glow-flow ChainDefinition JSON，导入后替换当前画布内容(保存前可撤销)" class="designer__import-tip" />
      <el-upload :auto-upload="false" :show-file-list="false" accept=".json,application/json" :on-change="onImportFile">
        <el-button size="small" :icon="'Upload'">选择文件</el-button>
      </el-upload>
      <el-input v-model="importText" type="textarea" class="mono designer__import-text" :rows="12" placeholder="或粘贴 JSON 内容" spellcheck="false" />
      <el-checkbox v-model="importRegenerate">重新生成节点编号(从其他流程导入时需要)</el-checkbox>
      <template #footer>
        <el-button @click="importOpen = false">取消</el-button>
        <el-button type="primary" @click="applyImport">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ElMessage, ElMessageBox, type UploadFile } from 'element-plus'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { nodeTypeApi, versionApi } from '@/api'
import { ApiError } from '@/api/http'
import DebugPanel from '@/components/designer/DebugPanel.vue'
import FlowCanvas, { type CanvasSelection } from '@/components/designer/FlowCanvas.vue'
import type { EdgeData, NodeData } from '@/components/designer/graph'
import { buildExport, parseImportContent, regenerateIds } from '@/components/designer/model'
import NodePalette from '@/components/designer/NodePalette.vue'
import PropertyPanel, { type SelectedEdge, type SelectedNode } from '@/components/designer/PropertyPanel.vue'
import ValidationPanel from '@/components/designer/ValidationPanel.vue'
import { validateModel } from '@/components/designer/validator'
import { FLOW_TYPES, VERSION_STATES } from '@/constants'
import type { ChainDesign, DebugResult, DesignConnection, DesignNode, FlowType, NodeTypeInfo, ValidationIssue, ValidationResult } from '@/types'
import { downloadJson, readFileText } from '@/utils'

const route = useRoute()
const router = useRouter()
const versionId = computed(() => Number(route.params.versionId))

const canvas = ref<InstanceType<typeof FlowCanvas>>()
const loading = ref(false)
const design = shallowRef<ChainDesign | null>(null)
const allTypes = ref<NodeTypeInfo[]>([])
const typeMap = computed(() => new Map(allTypes.value.map((t) => [t.node_type, t])))
const flowType = ref<FlowType>('RULE_CHAIN')
const versionDesc = ref('')
const chainExt = ref<Record<string, any>>({})
const dirty = ref(false)

const editable = computed(() => !!design.value?.editable)
// 流程从未发布过时允许切换流程类型
const modeSwitchable = computed(() => editable.value && (design.value?.chain.published_count ?? 0) === 0)

// ---------- 画布模型与选中 ----------
const model = shallowRef<{ nodes: DesignNode[]; connections: DesignConnection[] }>({ nodes: [], connections: [] })
const selection = ref<CanvasSelection>({ kind: 'none' })
const tick = ref(0)

const selectedNode = computed<SelectedNode | null>(() => {
  void tick.value
  const sel = selection.value
  if (sel.kind !== 'node') return null
  const node = canvas.value?.getNode(sel.id)
  if (!node) return null
  const data = (node.getData() ?? {}) as NodeData
  return { id: node.id, type: data.type, name: data.name, extparams: data.extparams ?? {} }
})

const selectedEdge = computed<SelectedEdge | null>(() => {
  void tick.value
  const sel = selection.value
  if (sel.kind !== 'edge') return null
  const edge = canvas.value?.getEdge(sel.id)
  if (!edge) return null
  const data = (edge.getData() ?? {}) as EdgeData
  return { id: edge.id, from_id: edge.getSourceCellId(), to_id: edge.getTargetCellId(), type: data.type || 'default', remark: data.remark ?? '' }
})

function onSelect(sel: CanvasSelection) {
  selection.value = sel
}

// ---------- 实时校验 ----------
const realtime = ref<ValidationResult | null>(null)
const serverResult = ref<ValidationResult | null>(null)
const validationSource = ref<'realtime' | 'server'>('realtime')
const shownValidation = computed(() => (validationSource.value === 'server' ? serverResult.value : realtime.value))
let validateTimer: number | undefined
let debugHighlight = false

function refreshModel() {
  if (!canvas.value) return
  model.value = canvas.value.getModel()
  realtime.value = validateModel(flowType.value, model.value.nodes, model.value.connections, typeMap.value, false)
  validationSource.value = 'realtime'
  if (!debugHighlight) canvas.value.showIssues(realtime.value)
}

function onChange() {
  tick.value++
  markDirty()
  window.clearTimeout(validateTimer)
  validateTimer = window.setTimeout(refreshModel, 300)
}

function markDirty() {
  if (editable.value) dirty.value = true
}

function onModeChange() {
  markDirty()
  refreshModel()
}

function locateIssue(issue: ValidationIssue) {
  if (issue.node_id) canvas.value?.focusNode(issue.node_id)
  else if (issue.from_id && issue.to_id) canvas.value?.focusEdge(issue.from_id, issue.to_id)
}

// ---------- 加载 ----------
async function load() {
  loading.value = true
  try {
    const [detail, types] = await Promise.all([versionApi.detail(versionId.value), nodeTypeApi.list()])
    allTypes.value = types
    design.value = null
    await nextTick()
    design.value = detail
    flowType.value = detail.chain.flow_type
    versionDesc.value = detail.version.desc
    chainExt.value = detail.extparams ?? {}
    selection.value = { kind: 'none' }
    serverResult.value = null
    debugResult.value = null
    debugHighlight = false
    await nextTick()
    canvas.value?.load(detail.nodes, detail.connections)
    await nextTick()
    refreshModel()
    dirty.value = false
  } finally {
    loading.value = false
  }
}

function currentPayload() {
  const m = canvas.value!.getModel()
  return {
    chain_no: design.value!.chain.chain_no,
    version_id: versionId.value,
    flow_type: flowType.value,
    extparams: chainExt.value,
    nodes: m.nodes,
    connections: m.connections,
  }
}

// ---------- 校验 / 保存 / 发布 ----------
const validating = ref(false)
const saving = ref(false)
const publishing = ref(false)

async function serverValidate() {
  validating.value = true
  try {
    serverResult.value = await versionApi.validate(currentPayload())
    validationSource.value = 'server'
    canvas.value?.showIssues(serverResult.value)
    debugHighlight = false
    if (serverResult.value.valid) ElMessage.success(serverResult.value.warnings.length ? `校验通过，有 ${serverResult.value.warnings.length} 个警告` : '校验通过，可以发布')
    else ElMessage.warning(`校验未通过：${serverResult.value.errors.length} 个错误`)
  } finally {
    validating.value = false
  }
}

async function save(silent = false): Promise<boolean> {
  if (!editable.value) return false
  saving.value = true
  try {
    const result = await versionApi.saveDraft({ ...currentPayload(), desc: versionDesc.value, layout: design.value!.layout }, { silent: true })
    dirty.value = false
    serverResult.value = result
    validationSource.value = 'server'
    if (!silent) {
      if (result.valid) ElMessage.success('已保存')
      else ElMessage.warning(`已保存，模型存在 ${result.errors.length} 个问题，发布前需要处理`)
    }
    return true
  } catch (err) {
    handleModelError(err, '保存失败')
    return false
  } finally {
    saving.value = false
  }
}

function handleModelError(err: unknown, title: string) {
  if (err instanceof ApiError && err.data && Array.isArray(err.data.errors)) {
    serverResult.value = err.data as ValidationResult
    validationSource.value = 'server'
    canvas.value?.showIssues(serverResult.value)
  }
  ElMessage.error(`${title}：${(err as Error).message}`)
}

async function publish() {
  if (dirty.value && !(await save(true))) return
  let desc = ''
  try {
    const res = await ElMessageBox.prompt('发布后版本不可修改，引擎将加载该版本。请填写发布说明：', `发布 v${design.value!.version.version_no}`, {
      confirmButtonText: '发布',
      cancelButtonText: '取消',
      inputValue: versionDesc.value,
      inputPlaceholder: '发布说明(可选)',
    })
    desc = res.value ?? ''
  } catch {
    return
  }
  publishing.value = true
  try {
    const result = await versionApi.publish(versionId.value, desc, { silent: true })
    ElMessage.success(result?.warnings?.length ? `发布成功，有 ${result.warnings.length} 个警告` : '发布成功')
    await load()
  } catch (err) {
    handleModelError(err, '发布失败')
  } finally {
    publishing.value = false
  }
}

async function newDraftFromThis() {
  const draft = await versionApi.createDraft(design.value!.chain.chain_no, versionId.value)
  if (draft.base_version_id !== versionId.value) {
    ElMessage.info(`流程已有草稿 v${draft.version_no}，已打开该草稿`)
  }
  router.push(`/designer/${draft.version_id}`)
}

// ---------- 调试 ----------
const debugOpen = ref(false)
const debugging = ref(false)
const debugResult = ref<DebugResult | null>(null)

async function runDebug(payload: { input: any; msg_id: string; timeout: number; mock_actions: boolean }) {
  debugging.value = true
  try {
    const result = await versionApi.debug({ ...currentPayload(), ...payload })
    debugResult.value = result
    if (result.validation && !result.validation.valid) {
      serverResult.value = result.validation
      validationSource.value = 'server'
      canvas.value?.showIssues(result.validation)
      debugHighlight = false
    } else {
      canvas.value?.showDebug(result)
      debugHighlight = true
    }
  } finally {
    debugging.value = false
  }
}

function clearDebug() {
  debugHighlight = false
  canvas.value?.clearHighlight()
  canvas.value?.showIssues(shownValidation.value ?? undefined)
}

// ---------- 导入导出 ----------
const importOpen = ref(false)
const importText = ref('')
const importRegenerate = ref(true)

function onTransfer(command: string) {
  if (command === 'import') {
    importText.value = ''
    importOpen.value = true
    return
  }
  const m = canvas.value!.getModel()
  const d = design.value!
  downloadJson(`${d.chain.chain_no}_v${d.version.version_no}_canvas.json`, buildExport(d.chain.chain_no, d.chain.name, flowType.value, d.chain.desc, d.version.version_no, chainExt.value, m.nodes, m.connections))
}

async function onImportFile(file: UploadFile) {
  if (file.raw) importText.value = await readFileText(file.raw)
}

function applyImport() {
  try {
    let { nodes, connections, flowType: importedType } = parseImportContent(importText.value)
    if (importRegenerate.value) ({ nodes, connections } = regenerateIds(nodes, connections))
    canvas.value!.load(nodes, connections)
    if (importedType && importedType !== flowType.value && modeSwitchable.value) flowType.value = importedType
    importOpen.value = false
    markDirty()
    nextTick(refreshModel)
    ElMessage.success(`已导入 ${nodes.length} 个节点、${connections.length} 个连接，保存后生效`)
  } catch (err) {
    ElMessage.error(`导入失败：${(err as Error).message}`)
  }
}

// ---------- 离开保护 ----------
function goBack() {
  router.push('/chains')
}

onBeforeRouteLeave(async () => {
  if (!dirty.value) return true
  try {
    await ElMessageBox.confirm('当前修改尚未保存，确定离开吗？', '提示', { type: 'warning', confirmButtonText: '离开', cancelButtonText: '留下' })
    return true
  } catch {
    return false
  }
})

function onBeforeUnload(e: BeforeUnloadEvent) {
  if (dirty.value) {
    e.preventDefault()
    e.returnValue = ''
  }
}

function onKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
    e.preventDefault()
    if (editable.value && !saving.value) save()
  }
}

watch(versionId, (val, old) => {
  if (val && val !== old) load()
})

onMounted(() => {
  load()
  window.addEventListener('beforeunload', onBeforeUnload)
  window.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  window.clearTimeout(validateTimer)
  window.removeEventListener('beforeunload', onBeforeUnload)
  window.removeEventListener('keydown', onKeydown)
})
</script>

<style scoped>
.designer {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #f2f3f5;
}

.designer__toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 52px;
  padding: 0 14px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  flex: none;
}

.designer__title {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.designer__name {
  font-size: 15px;
  font-weight: 600;
  white-space: nowrap;
}

.designer__mode {
  margin-left: 12px;
}

.designer__spacer {
  flex: 1;
}

.designer__body {
  display: flex;
  flex: 1;
  min-height: 0;
}

.designer__canvas {
  position: relative;
  flex: 1;
  min-width: 0;
}

.designer__readonly {
  position: absolute;
  top: 8px;
  left: 50%;
  z-index: 4;
  width: auto;
  transform: translateX(-50%);
}

.designer__import-tip {
  margin-bottom: 10px;
}

.designer__import-text {
  margin: 10px 0;
}
</style>
