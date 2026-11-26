<template>
  <div class="palette">
    <div class="palette__head">
      <span class="palette__title">节点工具箱</span>
      <el-tag size="small" effect="plain">{{ flowTypeLabel(flowType) }}</el-tag>
    </div>
    <el-input v-model="keyword" size="small" placeholder="搜索节点" clearable :prefix-icon="'Search'" class="palette__search" />
    <div class="palette__body">
      <el-empty v-if="!groups.length" :image-size="60" description="没有可用的节点类型" />
      <el-collapse v-else v-model="opened">
        <el-collapse-item v-for="group in groups" :key="group.category" :name="group.category">
          <template #title>
            <span class="palette__group">{{ group.label }}</span>
            <span class="palette__count">{{ group.items.length }}</span>
          </template>
          <div
            v-for="item in group.items"
            :key="item.node_type"
            class="palette__item"
            :class="{ 'is-disabled': disabled }"
            :title="item.desc"
            @mousedown="onMouseDown(item, $event)"
          >
            <span class="palette__bar" :style="{ background: item.color || DEFAULT_NODE_COLOR }" />
            <el-icon class="palette__icon" :style="{ color: item.color || DEFAULT_NODE_COLOR }">
              <component :is="iconOf(item)" />
            </el-icon>
            <div class="palette__text">
              <div class="palette__name">{{ item.name }}</div>
              <div class="palette__type">{{ item.node_type }}</div>
            </div>
            <el-tooltip v-if="!item.engine_supported" :content="item.source === 'service' ? '独立服务节点：需要引擎宿主注册' : 'glow-flow 引擎暂未实现，不能发布'" placement="right">
              <el-tag size="small" :type="item.source === 'service' ? 'warning' : 'info'" effect="plain" class="palette__flag">
                {{ item.source === 'service' ? '服务' : '未实现' }}
              </el-tag>
            </el-tooltip>
          </div>
        </el-collapse-item>
      </el-collapse>
    </div>
    <div class="palette__foot text-muted">拖拽节点到画布</div>
  </div>
</template>

<script setup lang="ts">
import * as Icons from '@element-plus/icons-vue'
import { computed, ref, watch } from 'vue'
import { DEFAULT_NODE_COLOR, flowTypeLabel, NODE_CATEGORIES } from '@/constants'
import type { FlowType, NodeTypeInfo } from '@/types'

const props = defineProps<{ types: NodeTypeInfo[]; flowType: FlowType; disabled?: boolean }>()
const emit = defineEmits<{ (e: 'drag', meta: NodeTypeInfo, evt: MouseEvent): void }>()

const keyword = ref('')
const opened = ref<string[]>([])

/** 根据当前流程类型加载节点，避免跨模式混用 */
const groups = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  const visible = props.types.filter(
    (t) => t.status === 1 && t.flow_modes.includes(props.flowType) && (!kw || t.name.toLowerCase().includes(kw) || t.node_type.toLowerCase().includes(kw)),
  )
  const order = NODE_CATEGORIES.map((c) => c.value)
  const result = NODE_CATEGORIES.map((c) => ({ category: c.value, label: c.label, items: visible.filter((t) => t.category === c.value) }))
  const others = visible.filter((t) => !order.includes(t.category))
  if (others.length) result.push({ category: 'other', label: '其他', items: others })
  return result.filter((g) => g.items.length)
})

watch(
  groups,
  (val) => {
    // 默认展开所有分组；搜索时也展开匹配的分组
    opened.value = val.map((g) => g.category)
  },
  { immediate: true },
)

function iconOf(item: NodeTypeInfo) {
  return item.icon && item.icon in Icons ? item.icon : 'Box'
}

function onMouseDown(item: NodeTypeInfo, evt: MouseEvent) {
  if (props.disabled || evt.button !== 0) return
  emit('drag', item, evt)
}
</script>

<style scoped>
.palette {
  display: flex;
  flex-direction: column;
  width: 232px;
  height: 100%;
  background: #fff;
  border-right: 1px solid #e4e7ed;
}

.palette__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 12px 8px;
}

.palette__title {
  font-weight: 600;
}

.palette__search {
  padding: 0 12px 8px;
}

.palette__body {
  flex: 1;
  overflow: auto;
  padding: 0 8px;
}

.palette__body :deep(.el-collapse) {
  border: none;
}

.palette__body :deep(.el-collapse-item__header) {
  height: 36px;
  border: none;
  padding-left: 4px;
}

.palette__body :deep(.el-collapse-item__wrap) {
  border: none;
}

.palette__body :deep(.el-collapse-item__content) {
  padding-bottom: 8px;
}

.palette__group {
  font-weight: 500;
}

.palette__count {
  margin-left: 6px;
  color: #a8abb2;
  font-size: 12px;
}

.palette__item {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 4px 0;
  padding: 6px 8px 6px 12px;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  background: #fff;
  cursor: grab;
  user-select: none;
  overflow: hidden;
}

.palette__item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 6px rgba(64, 158, 255, 0.15);
}

.palette__item.is-disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.palette__bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
}

.palette__icon {
  font-size: 18px;
}

.palette__text {
  flex: 1;
  min-width: 0;
}

.palette__name {
  font-size: 13px;
  line-height: 18px;
}

.palette__type {
  color: #a8abb2;
  font-size: 11px;
  line-height: 16px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.palette__flag {
  flex: none;
}

.palette__foot {
  padding: 8px 12px;
  border-top: 1px solid #f0f2f5;
  font-size: 12px;
}
</style>
