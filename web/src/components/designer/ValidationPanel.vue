<template>
  <div class="vpanel" :class="{ 'is-open': open }">
    <div class="vpanel__bar" @click="open = !open">
      <el-icon><component :is="open ? 'ArrowDown' : 'ArrowUp'" /></el-icon>
      <span class="vpanel__title">校验结果</span>
      <span class="vpanel__source text-muted">{{ sourceLabel }}</span>
      <el-tag v-if="result && result.errors.length" type="danger" size="small" effect="dark" round>{{ result.errors.length }} 错误</el-tag>
      <el-tag v-if="result && result.warnings.length" type="warning" size="small" effect="dark" round>{{ result.warnings.length }} 警告</el-tag>
      <el-tag v-if="result && !result.errors.length && !result.warnings.length" type="success" size="small" round>通过</el-tag>
    </div>
    <div v-show="open" class="vpanel__body">
      <el-empty v-if="!issues.length" :image-size="48" description="没有问题" />
      <div v-for="(issue, i) in issues" :key="i" class="vpanel__item" @click="$emit('locate', issue)">
        <el-icon :class="issue.level === 'error' ? 'text-danger' : 'vpanel__warn'">
          <component :is="issue.level === 'error' ? 'CircleCloseFilled' : 'WarningFilled'" />
        </el-icon>
        <span class="vpanel__msg">{{ issue.message }}</span>
        <span class="vpanel__code text-muted">{{ issue.code }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { ValidationIssue, ValidationResult } from '@/types'

const props = defineProps<{ result: ValidationResult | null; source: 'realtime' | 'server' }>()
defineEmits<{ (e: 'locate', issue: ValidationIssue): void }>()

const open = ref(false)
const issues = computed(() => [...(props.result?.errors ?? []), ...(props.result?.warnings ?? [])])
const sourceLabel = computed(() => (props.source === 'server' ? '服务端校验(含引擎编译)' : '实时校验'))

// 服务端校验返回错误时自动展开
watch(
  () => props.result,
  (val) => {
    if (props.source === 'server' && val && (val.errors.length || val.warnings.length)) open.value = true
  },
)

defineExpose({ open: () => (open.value = true) })
</script>

<style scoped>
.vpanel {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 5;
  background: #fff;
  border-top: 1px solid #e4e7ed;
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.04);
}

.vpanel__bar {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 34px;
  padding: 0 12px;
  cursor: pointer;
  user-select: none;
}

.vpanel__title {
  font-weight: 600;
  font-size: 13px;
}

.vpanel__source {
  flex: 1;
  font-size: 12px;
}

.vpanel__body {
  max-height: 220px;
  overflow: auto;
  padding: 0 12px 8px;
}

.vpanel__item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 6px;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
}

.vpanel__item:hover {
  background: #f5f7fa;
}

.vpanel__msg {
  flex: 1;
}

.vpanel__code {
  font-size: 11px;
}

.vpanel__warn {
  color: #e6a23c;
}
</style>
