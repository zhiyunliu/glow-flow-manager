<template>
  <el-drawer :model-value="modelValue" title="预览调试" size="520px" :modal="false" class="debug-drawer" @update:model-value="(v: boolean) => $emit('update:modelValue', v)">
    <el-form label-position="top" size="small">
      <el-form-item label="输入消息(JSON)">
        <el-input v-model="input" type="textarea" class="mono" :autosize="{ minRows: 6, maxRows: 14 }" spellcheck="false" placeholder='{"temperature": 61, "device": "D-01"}' />
        <div v-if="inputError" class="text-danger debug__error">{{ inputError }}</div>
      </el-form-item>
      <div class="debug__opts">
        <el-form-item label="消息ID">
          <el-input v-model="msgId" placeholder="默认自动生成" />
        </el-form-item>
        <el-form-item label="超时(秒)">
          <el-input-number v-model="timeout" :min="1" :max="120" controls-position="right" />
        </el-form-item>
        <el-form-item label="模拟动作节点">
          <el-tooltip content="开启时 HTTP/RPC/消息/邮件/钉钉/命令/数据库等动作节点及独立服务节点不真实执行，原样传递消息" placement="top">
            <el-switch v-model="mockActions" />
          </el-tooltip>
        </el-form-item>
      </div>
      <el-alert v-if="!mockActions" type="warning" :closable="false" show-icon title="动作节点将真实执行(发送请求、消息、邮件，执行命令等)，请谨慎操作" class="debug__warn" />
      <div class="debug__actions">
        <el-button type="primary" :icon="'VideoPlay'" :loading="running" @click="run">执行</el-button>
        <el-button :disabled="!result" @click="$emit('clear')">清除高亮</el-button>
      </div>
    </el-form>

    <template v-if="result">
      <el-divider />
      <div class="debug__summary">
        <el-tag :type="result.success ? 'success' : 'danger'" effect="dark">{{ result.success ? '执行成功' : result.timed_out ? '执行超时' : '执行失败' }}</el-tag>
        <span>耗时 {{ formatDuration(result.duration_ms) }}</span>
        <span class="text-muted">执行 {{ result.events.length }} 次 · 经过 {{ result.executed_nodes.length }} 个节点</span>
      </div>
      <el-alert v-if="result.error" type="error" :closable="false" :title="result.error" class="debug__warn" />
      <el-timeline class="debug__timeline">
        <el-timeline-item
          v-for="ev in result.events"
          :key="ev.seq"
          :type="ev.error ? 'danger' : ev.mocked ? 'warning' : 'success'"
          :timestamp="`+${ev.start_offset_ms}ms · ${formatDuration(ev.duration_ms)}`"
          placement="top"
        >
          <div class="debug__event" @click="$emit('locate', ev.node_id)">
            <div class="debug__event-head">
              <span class="debug__event-name">{{ ev.node_name || ev.node_id }}</span>
              <span class="text-muted">{{ ev.node_type }}</span>
              <el-tag v-if="ev.mocked" size="small" type="warning" effect="plain">模拟</el-tag>
              <el-tag v-if="ev.end" size="small" type="info" effect="plain">结束分支</el-tag>
              <span class="debug__relation">→ {{ (ev.relations?.length ? ev.relations : [ev.relation_type]).join(', ') || '-' }}</span>
            </div>
            <div v-if="ev.error" class="text-danger debug__event-error">{{ ev.error }}</div>
          </div>
          <el-collapse class="debug__io">
            <el-collapse-item title="输入" :name="`in-${ev.seq}`"><JsonView :value="ev.input" max-height="220px" /></el-collapse-item>
            <el-collapse-item title="输出" :name="`out-${ev.seq}`"><JsonView :value="ev.output" max-height="220px" /></el-collapse-item>
          </el-collapse>
        </el-timeline-item>
      </el-timeline>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import JsonView from '@/components/JsonView.vue'
import type { DebugResult } from '@/types'
import { formatDuration, parseJson } from '@/utils'

defineProps<{ modelValue: boolean; result: DebugResult | null; running: boolean }>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'run', payload: { input: any; msg_id: string; timeout: number; mock_actions: boolean }): void
  (e: 'clear'): void
  (e: 'locate', nodeId: string): void
}>()

const input = ref('{\n  \n}')
const inputError = ref('')
const msgId = ref('')
const timeout = ref(30)
const mockActions = ref(true)

function run() {
  const [parsed, err] = parseJson(input.value)
  inputError.value = err
  if (err) return
  emit('run', { input: parsed ?? null, msg_id: msgId.value.trim(), timeout: timeout.value, mock_actions: mockActions.value })
}
</script>

<style scoped>
.debug__opts {
  display: grid;
  grid-template-columns: 1.4fr 1fr 1fr;
  gap: 10px;
}

.debug__error {
  font-size: 12px;
}

.debug__warn {
  margin-bottom: 10px;
}

.debug__actions {
  display: flex;
  gap: 8px;
}

.debug__summary {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
  font-size: 13px;
}

.debug__timeline {
  padding-left: 2px;
}

.debug__event {
  cursor: pointer;
}

.debug__event-head {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
  font-size: 13px;
}

.debug__event-name {
  font-weight: 600;
}

.debug__relation {
  margin-left: auto;
  color: #409eff;
  font-size: 12px;
}

.debug__event-error {
  margin-top: 4px;
  font-size: 12px;
}

.debug__io {
  margin-top: 4px;
  border: none;
}

.debug__io :deep(.el-collapse-item__header) {
  height: 28px;
  font-size: 12px;
  border: none;
}

.debug__io :deep(.el-collapse-item__wrap) {
  border: none;
}
</style>
