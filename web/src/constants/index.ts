import type { FlowType } from '@/types'

type TagType = 'primary' | 'success' | 'info' | 'warning' | 'danger'

export const FLOW_TYPES: { value: FlowType; label: string; desc: string }[] = [
  { value: 'RULE_CHAIN', label: '规则链', desc: '无状态、事件驱动，必须为有向无环图' },
  { value: 'WORKFLOW', label: '工作流', desc: '长周期、人工参与，允许循环与回退' },
]

export function flowTypeLabel(value?: string) {
  return FLOW_TYPES.find((item) => item.value === value)?.label ?? value ?? ''
}

export const VERSION_STATES: Record<number, { label: string; type: TagType }> = {
  0: { label: '草稿', type: 'info' },
  1: { label: '已发布', type: 'success' },
  2: { label: '已下线', type: 'warning' },
}

export const NODE_CATEGORIES: { value: string; label: string }[] = [
  { value: 'filter', label: '过滤' },
  { value: 'transform', label: '转换' },
  { value: 'action', label: '动作' },
  { value: 'logical', label: '逻辑控制' },
  { value: 'workflow', label: '工作流' },
  { value: 'integration', label: '集成服务' },
]

export function categoryLabel(value?: string) {
  return NODE_CATEGORIES.find((item) => item.value === value)?.label ?? value ?? ''
}

export const NODE_SOURCES: Record<string, { label: string; type: TagType }> = {
  builtin: { label: '内置', type: 'primary' },
  service: { label: '独立服务', type: 'warning' },
}

export const INSTANCE_STATUS: Record<string, { label: string; type: TagType }> = {
  pending: { label: '等待中', type: 'info' },
  running: { label: '运行中', type: 'primary' },
  paused: { label: '已暂停', type: 'warning' },
  completed: { label: '已完成', type: 'success' },
  failed: { label: '失败', type: 'danger' },
  canceled: { label: '已取消', type: 'info' },
  skipped: { label: '已跳过', type: 'info' },
}

export const AUDIT_MODULES: Record<string, string> = {
  chain: '流程定义',
  version: '流程版本',
  node_type: '节点注册',
}

export const AUDIT_OPS: Record<string, string> = {
  create: '创建',
  update: '修改',
  delete: '删除',
  enable: '启用',
  disable: '禁用',
  copy: '复制',
  import: '导入',
  save_draft: '保存草稿',
  publish: '发布',
  offline: '下线',
  online: '上线',
  sync: '同步',
}

export const FIELD_TYPES: { value: string; label: string }[] = [
  { value: 'string', label: '单行文本' },
  { value: 'text', label: '多行文本' },
  { value: 'number', label: '数字' },
  { value: 'bool', label: '开关' },
  { value: 'select', label: '下拉选择' },
  { value: 'code', label: '代码/表达式' },
  { value: 'kv', label: '键值对' },
  { value: 'list', label: '字符串列表' },
  { value: 'cases', label: '条件分支' },
  { value: 'nodeRef', label: '引用节点' },
  { value: 'nodeRefs', label: '引用多个节点' },
  { value: 'json', label: 'JSON' },
]

/** 连接类型颜色 */
export function relationColor(type?: string) {
  switch (type) {
    case 'success':
    case 'true':
    case 'approve':
      return '#52a35b'
    case 'false':
      return '#8d939e'
    case 'failure':
    case 'reject':
      return '#e05252'
    case 'default':
    case '':
    case undefined:
      return '#5b6270'
    default:
      return '#3d7fe0'
  }
}

export const DEFAULT_NODE_COLOR = '#909399'
