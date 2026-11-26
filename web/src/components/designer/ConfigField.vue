<template>
  <el-form-item :required="field.required" class="config-field">
    <template #label>
      <span>{{ field.label || field.key }}</span>
      <el-tooltip v-if="field.help" :content="field.help" placement="top">
        <el-icon class="config-field__help"><QuestionFilled /></el-icon>
      </el-tooltip>
    </template>

    <el-input v-if="field.type === 'string'" :model-value="str" :placeholder="field.placeholder" :disabled="disabled" clearable @update:model-value="emitValue" />

    <el-input
      v-else-if="field.type === 'text'"
      :model-value="str"
      type="textarea"
      :autosize="{ minRows: 3, maxRows: 10 }"
      :placeholder="field.placeholder"
      :disabled="disabled"
      @update:model-value="emitValue"
    />

    <div v-else-if="field.type === 'code'" class="config-field__code">
      <el-input
        :model-value="str"
        type="textarea"
        class="mono"
        :autosize="{ minRows: field.lang === 'expr' ? 2 : 5, maxRows: 16 }"
        :placeholder="field.placeholder"
        :disabled="disabled"
        spellcheck="false"
        @update:model-value="emitValue"
      />
      <span v-if="field.lang" class="config-field__lang">{{ field.lang }}</span>
    </div>

    <el-input-number
      v-else-if="field.type === 'number'"
      :model-value="num"
      :placeholder="field.placeholder"
      :disabled="disabled"
      controls-position="right"
      class="config-field__number"
      @update:model-value="(v: number | undefined | null) => emitValue(v ?? undefined)"
    />

    <el-switch v-else-if="field.type === 'bool'" :model-value="!!modelValue" :disabled="disabled" @update:model-value="emitValue" />

    <div v-else-if="field.type === 'select' && field.options_source" class="w-full">
      <el-select
        :model-value="str || undefined"
        :placeholder="field.placeholder || '请选择'"
        :disabled="disabled"
        :loading="dyn.loading"
        no-data-text="未配置可选项"
        filterable
        clearable
        class="w-full"
        @visible-change="(open: boolean) => open && fetchOptions()"
        @update:model-value="emitValue"
      >
        <el-option v-for="opt in dyn.items" :key="opt.value" :label="opt.label" :value="opt.value">
          <span>{{ opt.label }}</span>
          <span v-if="opt.desc" class="config-field__opt-desc">{{ opt.desc }}</span>
        </el-option>
      </el-select>
      <div v-if="dyn.error" class="text-danger config-field__error">
        加载可选项失败：{{ dyn.error }}
        <el-button link type="primary" size="small" @click="fetchOptions(true)">重试</el-button>
      </div>
      <div v-else-if="dyn.loaded && str && !dyn.items.some((o) => o.value === str)" class="text-warning config-field__error">
        「{{ str }}」不在可选列表中
      </div>
    </div>

    <el-select
      v-else-if="field.type === 'select'"
      :model-value="modelValue"
      :placeholder="field.placeholder || '请选择'"
      :disabled="disabled"
      clearable
      class="w-full"
      @update:model-value="emitValue"
    >
      <el-option v-for="opt in field.options ?? []" :key="String(opt.value)" :label="opt.label" :value="opt.value" />
    </el-select>

    <el-select
      v-else-if="field.type === 'list'"
      :model-value="listValue"
      multiple
      filterable
      allow-create
      default-first-option
      :reserve-keyword="false"
      :placeholder="field.placeholder || '输入后回车添加'"
      :disabled="disabled"
      class="w-full"
      @update:model-value="(v: string[]) => emitValue(v.length ? v : undefined)"
    />

    <el-select
      v-else-if="field.type === 'nodeRef'"
      :model-value="str || undefined"
      filterable
      clearable
      placeholder="选择流程中的节点"
      :disabled="disabled"
      class="w-full"
      @update:model-value="emitValue"
    >
      <el-option v-for="n in refOptions" :key="n.id" :label="`${n.name} (${n.id})`" :value="n.id" />
    </el-select>

    <el-select
      v-else-if="field.type === 'nodeRefs'"
      :model-value="refsValue"
      multiple
      filterable
      placeholder="选择流程中的节点"
      :disabled="disabled"
      class="w-full"
      @update:model-value="(v: string[]) => emitValue(v.length ? v : undefined)"
    >
      <el-option v-for="n in refOptions" :key="n.id" :label="`${n.name} (${n.id})`" :value="n.id" />
    </el-select>

    <div v-else-if="field.type === 'kv'" class="config-field__rows">
      <div v-for="(row, i) in kvRows" :key="i" class="config-field__row">
        <el-input v-model="row.key" placeholder="键" :disabled="disabled" @change="emitKV" />
        <el-input v-model="row.value" placeholder="值" :disabled="disabled" class="mono" @change="emitKV" />
        <el-button :icon="'Delete'" link type="danger" :disabled="disabled" @click="removeKV(i)" />
      </div>
      <el-button size="small" :icon="'Plus'" :disabled="disabled" @click="kvRows.push({ key: '', value: '' })">添加</el-button>
    </div>

    <div v-else-if="field.type === 'cases'" class="config-field__rows">
      <div v-for="(row, i) in caseRows" :key="i" class="config-field__case">
        <div class="config-field__case-head">
          <span class="text-muted">分支 {{ i + 1 }}</span>
          <el-button :icon="'Delete'" link type="danger" :disabled="disabled" @click="removeCase(i)" />
        </div>
        <el-input v-model="row.case" placeholder="条件：msg.temperature > 50" class="mono" :disabled="disabled" @change="emitCases" />
        <el-input v-model="row.then" placeholder="连接类型：hot" :disabled="disabled" @change="emitCases">
          <template #prepend>then</template>
        </el-input>
      </div>
      <el-button size="small" :icon="'Plus'" :disabled="disabled" @click="caseRows.push({ case: '', then: '' })">添加分支</el-button>
    </div>

    <div v-else class="config-field__json">
      <el-input
        v-model="jsonText"
        type="textarea"
        class="mono"
        :autosize="{ minRows: 3, maxRows: 14 }"
        :placeholder="field.placeholder || 'JSON'"
        :disabled="disabled"
        spellcheck="false"
        @blur="emitJson"
      />
      <div v-if="jsonError" class="text-danger config-field__error">{{ jsonError }}</div>
    </div>
  </el-form-item>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import type { DesignNode, OptionItem, SchemaField } from '@/types'
import { parseJson, prettyJson } from '@/utils'
import { loadOptions } from './options'
import { nodeRefValues } from './validator'

const props = defineProps<{ field: SchemaField; modelValue: any; nodes: Pick<DesignNode, 'id' | 'name'>[]; selfId: string; disabled?: boolean }>()
const emit = defineEmits<{ (e: 'update:modelValue', value: any): void }>()

const str = computed(() => (props.modelValue === undefined || props.modelValue === null ? '' : typeof props.modelValue === 'string' ? props.modelValue : JSON.stringify(props.modelValue)))
const num = computed(() => {
  const n = Number(props.modelValue)
  return props.modelValue === undefined || props.modelValue === '' || Number.isNaN(n) ? undefined : n
})
const listValue = computed<string[]>(() => {
  if (Array.isArray(props.modelValue)) return props.modelValue.map(String)
  if (typeof props.modelValue === 'string' && props.modelValue.trim()) return props.modelValue.split(',').map((s) => s.trim()).filter(Boolean)
  return []
})
const refsValue = computed(() => nodeRefValues(props.modelValue))
const refOptions = computed(() => props.nodes.filter((n) => n.id !== props.selfId))

function emitValue(value: any) {
  emit('update:modelValue', value === '' ? undefined : value)
}

// ---------- 动态下拉选项 ----------
const dyn = reactive({ items: [] as OptionItem[], loading: false, loaded: false, error: '' })
let fetchSeq = 0

async function fetchOptions(force = false) {
  const source = props.field.options_source
  if (!source) return
  const seq = ++fetchSeq
  dyn.loading = true
  try {
    const items = await loadOptions(source, force)
    if (seq !== fetchSeq) return
    dyn.items = items
    dyn.error = ''
    dyn.loaded = true
  } catch (e: any) {
    if (seq !== fetchSeq) return
    dyn.error = e?.message || String(e)
  } finally {
    if (seq === fetchSeq) dyn.loading = false
  }
}

watch(
  () => props.field.options_source,
  (source) => {
    dyn.items = []
    dyn.loaded = false
    dyn.error = ''
    if (source) fetchOptions()
  },
  { immediate: true },
)

// ---------- 键值对 ----------
const kvRows = ref<{ key: string; value: string }[]>([])
// ---------- 条件分支 ----------
const caseRows = ref<{ case: string; then: string }[]>([])
// ---------- JSON ----------
const jsonText = ref('')
const jsonError = ref('')

watch(
  () => [props.field.type, props.modelValue, props.selfId] as const,
  () => {
    if (props.field.type === 'kv') {
      const obj = parseObject(props.modelValue)
      kvRows.value = Object.entries(obj).map(([key, value]) => ({ key, value: typeof value === 'string' ? value : JSON.stringify(value) }))
    } else if (props.field.type === 'cases') {
      const list = Array.isArray(props.modelValue) ? props.modelValue : parseArray(props.modelValue)
      caseRows.value = list.map((item: any) => ({ case: String(item?.case ?? ''), then: String(item?.then ?? '') }))
    } else if (props.field.type === 'json') {
      jsonText.value = props.modelValue === undefined ? '' : prettyJson(props.modelValue)
      jsonError.value = ''
    }
  },
  { immediate: true },
)

function parseObject(val: any): Record<string, any> {
  if (val && typeof val === 'object' && !Array.isArray(val)) return val
  if (typeof val === 'string') {
    const [parsed] = parseJson(val)
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) return parsed
  }
  return {}
}

function parseArray(val: any): any[] {
  if (typeof val === 'string') {
    const [parsed] = parseJson(val)
    if (Array.isArray(parsed)) return parsed
  }
  return []
}

function emitKV() {
  const obj: Record<string, string> = {}
  for (const row of kvRows.value) {
    if (row.key.trim()) obj[row.key.trim()] = row.value
  }
  emitValue(Object.keys(obj).length ? obj : undefined)
}

function removeKV(i: number) {
  kvRows.value.splice(i, 1)
  emitKV()
}

function emitCases() {
  const list = caseRows.value.filter((row) => row.case.trim() || row.then.trim()).map((row) => ({ case: row.case.trim(), then: row.then.trim() }))
  emitValue(list.length ? list : undefined)
}

function removeCase(i: number) {
  caseRows.value.splice(i, 1)
  emitCases()
}

function emitJson() {
  const [parsed, err] = parseJson(jsonText.value)
  jsonError.value = err
  if (!err) emitValue(parsed)
}
</script>

<style scoped>
.config-field :deep(.el-form-item__label) {
  display: inline-flex;
  align-items: center;
}

.config-field__help {
  margin-left: 4px;
  color: #a8abb2;
  cursor: help;
}

.config-field__code {
  position: relative;
  width: 100%;
}

.config-field__lang {
  position: absolute;
  right: 8px;
  bottom: 4px;
  color: #c0c4cc;
  font-size: 11px;
  pointer-events: none;
}

.config-field__number {
  width: 100%;
}

.config-field__rows {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
}

.config-field__row {
  display: grid;
  grid-template-columns: 1fr 1.4fr auto;
  gap: 6px;
}

.config-field__case {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 6px 8px;
  background: #f7f8fa;
  border-radius: 4px;
}

.config-field__case-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12px;
}

.config-field__json {
  width: 100%;
}

.config-field__opt-desc {
  float: right;
  margin-left: 16px;
  color: #a8abb2;
  font-size: 12px;
}

.config-field__error {
  font-size: 12px;
  line-height: 18px;
}

.w-full {
  width: 100%;
}
</style>
