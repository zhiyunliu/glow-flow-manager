<template>
  <div class="props">
    <!-- 节点 -->
    <template v-if="node">
      <div class="props__head">
        <span class="props__bar" :style="{ background: nodeMeta?.color || DEFAULT_NODE_COLOR }" />
        <div class="props__title">
          <div>{{ nodeMeta?.name ?? node.type }}</div>
          <div class="text-muted props__sub">{{ node.type }} · {{ categoryLabel(nodeMeta?.category) }}</div>
        </div>
        <el-button v-if="editable" :icon="'Delete'" link type="danger" title="删除节点" @click="$emit('remove', node.id)" />
      </div>
      <el-scrollbar class="props__body">
        <el-alert v-if="!nodeMeta" type="error" :closable="false" show-icon :title="`节点类型 ${node.type} 未注册`" class="props__alert" />
        <el-alert v-else-if="!nodeMeta.engine_supported && nodeMeta.source === 'builtin'" type="warning" :closable="false" show-icon title="glow-flow 引擎暂未实现该节点，包含该节点的流程不能发布" class="props__alert" />
        <el-form label-position="top" size="small">
          <el-form-item label="节点编号">
            <el-input :model-value="node.id" readonly>
              <template #append>
                <el-button :icon="'CopyDocument'" @click="copy(node.id)" />
              </template>
            </el-input>
          </el-form-item>
          <el-form-item label="节点名称" required>
            <el-input :model-value="node.name" :disabled="!editable" maxlength="64" show-word-limit @update:model-value="(v: string) => patchNode({ name: v })" />
          </el-form-item>
          <p v-if="nodeMeta?.desc" class="props__desc">{{ nodeMeta.desc }}</p>
          <template v-for="group in fieldGroups" :key="group.name">
            <div v-if="group.name" class="props__group">{{ group.name }}</div>
            <ConfigField
              v-for="field in group.fields"
              :key="`${node.id}-${field.key}`"
              :field="field"
              :model-value="node.extparams[field.key]"
              :nodes="allNodes"
              :self-id="node.id"
              :disabled="!editable"
              @update:model-value="(v: any) => patchConfig(field.key, v)"
            />
          </template>
          <el-collapse class="props__advanced">
            <el-collapse-item title="高级：配置 JSON" name="json">
              <el-input v-model="extText" type="textarea" class="mono" :autosize="{ minRows: 4, maxRows: 18 }" :disabled="!editable" spellcheck="false" />
              <div v-if="extError" class="text-danger props__error">{{ extError }}</div>
              <el-button v-if="editable" size="small" class="props__apply" @click="applyExtText">应用</el-button>
            </el-collapse-item>
          </el-collapse>
        </el-form>
      </el-scrollbar>
    </template>

    <!-- 连线 -->
    <template v-else-if="edge">
      <div class="props__head">
        <el-icon :size="18"><Right /></el-icon>
        <div class="props__title">
          <div>连接</div>
          <div class="text-muted props__sub">{{ edgeFromName }} → {{ edgeToName }}</div>
        </div>
        <el-button v-if="editable" :icon="'Delete'" link type="danger" title="删除连接" @click="$emit('remove', edge.id)" />
      </div>
      <el-scrollbar class="props__body">
        <el-form label-position="top" size="small">
          <el-form-item label="连接类型(relation)">
            <el-select
              :model-value="edge.type"
              filterable
              allow-create
              default-first-option
              :disabled="!editable"
              class="w-full"
              @update:model-value="(v: string) => patchEdge({ type: v || 'default' })"
            >
              <el-option v-for="r in edgeRelations" :key="r" :label="r" :value="r" />
            </el-select>
            <div class="props__hint">节点执行结果的关系类型与连接类型匹配时走该连接；没有匹配的连接时走 default。</div>
          </el-form-item>
          <el-form-item label="备注">
            <el-input :model-value="edge.remark" :disabled="!editable" maxlength="120" @update:model-value="(v: string) => patchEdge({ remark: v })" />
          </el-form-item>
        </el-form>
      </el-scrollbar>
    </template>

    <!-- 流程 -->
    <template v-else>
      <div class="props__head">
        <el-icon :size="18"><Setting /></el-icon>
        <div class="props__title">
          <div>流程属性</div>
          <div class="text-muted props__sub">未选中节点或连线</div>
        </div>
      </div>
      <el-scrollbar class="props__body">
        <el-form label-position="top" size="small">
          <el-form-item label="流程编号"><el-input :model-value="chain?.chain_no" readonly /></el-form-item>
          <el-form-item label="流程名称"><el-input :model-value="chain?.name" readonly /></el-form-item>
          <el-form-item label="版本说明">
            <el-input
              :model-value="versionDesc"
              type="textarea"
              :autosize="{ minRows: 2, maxRows: 6 }"
              maxlength="500"
              :disabled="!editable"
              @update:model-value="(v: string) => $emit('update:versionDesc', v)"
            />
          </el-form-item>
          <el-form-item label="流程扩展参数(extparams)">
            <el-input v-model="chainExtText" type="textarea" class="mono" :autosize="{ minRows: 4, maxRows: 14 }" :disabled="!editable" spellcheck="false" @blur="applyChainExt" />
            <div v-if="chainExtError" class="text-danger props__error">{{ chainExtError }}</div>
          </el-form-item>
          <el-descriptions :column="1" size="small" border class="props__stats">
            <el-descriptions-item label="节点数">{{ allNodes.length }}</el-descriptions-item>
            <el-descriptions-item label="连接数">{{ edgeCount }}</el-descriptions-item>
          </el-descriptions>
        </el-form>
      </el-scrollbar>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import { computed, ref, watch } from 'vue'
import { categoryLabel, DEFAULT_NODE_COLOR } from '@/constants'
import type { ChainInfo, DesignNode, NodeTypeInfo, SchemaField } from '@/types'
import { copyText, parseJson, prettyJson } from '@/utils'
import ConfigField from './ConfigField.vue'
import { allowedRelations, fieldVisible } from './validator'

export interface SelectedNode {
  id: string
  type: string
  name: string
  extparams: Record<string, any>
}

export interface SelectedEdge {
  id: string
  from_id: string
  to_id: string
  type: string
  remark: string
}

const props = defineProps<{
  node: SelectedNode | null
  edge: SelectedEdge | null
  types: Map<string, NodeTypeInfo>
  allNodes: DesignNode[]
  edgeCount: number
  editable: boolean
  chain?: ChainInfo
  versionDesc: string
  chainExt: Record<string, any>
}>()

const emit = defineEmits<{
  (e: 'update-node', id: string, patch: Partial<SelectedNode>): void
  (e: 'update-edge', id: string, patch: Partial<SelectedEdge>): void
  (e: 'remove', id: string): void
  (e: 'update:versionDesc', value: string): void
  (e: 'update:chainExt', value: Record<string, any>): void
}>()

const nodeMeta = computed(() => (props.node ? props.types.get(props.node.type) : undefined))

/** 按 show_when 过滤并按 group 分组 */
const fieldGroups = computed(() => {
  const visible = (nodeMeta.value?.config_schema?.fields ?? []).filter((f) => fieldVisible(f, props.node?.extparams ?? {}))
  const groups: { name: string; fields: SchemaField[] }[] = []
  for (const field of visible) {
    const name = field.group ?? ''
    let group = groups.find((g) => g.name === name)
    if (!group) {
      group = { name, fields: [] }
      groups.push(group)
    }
    group.fields.push(field)
  }
  return groups
})

function patchNode(patch: Partial<SelectedNode>) {
  if (props.node) emit('update-node', props.node.id, patch)
}

function patchConfig(key: string, value: any) {
  if (!props.node) return
  const ext = { ...props.node.extparams }
  if (value === undefined) delete ext[key]
  else ext[key] = value
  patchNode({ extparams: ext })
}

// ---------- 配置 JSON ----------
const extText = ref('')
const extError = ref('')
watch(
  () => props.node?.extparams,
  (val) => {
    extText.value = prettyJson(val ?? {})
    extError.value = ''
  },
  { immediate: true, deep: true },
)

function applyExtText() {
  const [parsed, err] = parseJson(extText.value || '{}')
  if (err || typeof parsed !== 'object' || Array.isArray(parsed) || parsed === null) {
    extError.value = err || '配置必须是 JSON 对象'
    return
  }
  extError.value = ''
  patchNode({ extparams: parsed })
}

// ---------- 连线 ----------
const edgeFrom = computed(() => props.allNodes.find((n) => n.id === props.edge?.from_id))
const edgeFromName = computed(() => edgeFrom.value?.name ?? props.edge?.from_id)
const edgeToName = computed(() => props.allNodes.find((n) => n.id === props.edge?.to_id)?.name ?? props.edge?.to_id)
const edgeRelations = computed(() => {
  const from = edgeFrom.value
  const list = from ? allowedRelations(from, props.types.get(from.type)) : ['default']
  if (props.edge?.type && !list.includes(props.edge.type)) list.push(props.edge.type)
  return list
})

function patchEdge(patch: Partial<SelectedEdge>) {
  if (props.edge) emit('update-edge', props.edge.id, patch)
}

// ---------- 流程扩展参数 ----------
const chainExtText = ref('')
const chainExtError = ref('')
watch(
  () => props.chainExt,
  (val) => {
    chainExtText.value = val && Object.keys(val).length ? prettyJson(val) : ''
    chainExtError.value = ''
  },
  { immediate: true },
)

function applyChainExt() {
  const [parsed, err] = parseJson(chainExtText.value)
  if (err || (parsed !== undefined && (typeof parsed !== 'object' || Array.isArray(parsed)))) {
    chainExtError.value = err || '扩展参数必须是 JSON 对象'
    return
  }
  chainExtError.value = ''
  emit('update:chainExt', parsed ?? {})
}

async function copy(text: string) {
  if (await copyText(text)) ElMessage.success('已复制')
}
</script>

<style scoped>
.props {
  display: flex;
  flex-direction: column;
  width: 340px;
  height: 100%;
  background: #fff;
  border-left: 1px solid #e4e7ed;
}

.props__head {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-bottom: 1px solid #f0f2f5;
}

.props__bar {
  width: 4px;
  height: 32px;
  border-radius: 2px;
}

.props__title {
  flex: 1;
  min-width: 0;
  font-weight: 600;
}

.props__sub {
  font-size: 12px;
  font-weight: normal;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.props__body {
  flex: 1;
  padding: 12px 14px;
}

.props__body :deep(.el-scrollbar__view) {
  padding: 12px 14px;
}

.props__alert {
  margin-bottom: 12px;
}

.props__desc {
  margin: -4px 0 12px;
  color: #909399;
  font-size: 12px;
  line-height: 18px;
}

.props__group {
  margin: 4px 0 10px;
  padding-bottom: 4px;
  border-bottom: 1px dashed #e4e7ed;
  color: #606266;
  font-size: 12px;
  font-weight: 600;
}

.props__advanced {
  margin-top: 8px;
}

.props__error {
  font-size: 12px;
}

.props__apply {
  margin-top: 6px;
}

.props__hint {
  color: #a8abb2;
  font-size: 12px;
  line-height: 18px;
}

.props__stats {
  margin-top: 8px;
}

.w-full {
  width: 100%;
}
</style>
