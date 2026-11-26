import type { DesignConnection, DesignNode, FlowType, NodeTypeInfo, SchemaField, ValidationIssue, ValidationResult } from '@/types'

const NODE_ID = /^[A-Za-z0-9_-]{1,32}$/
const RELATION = /^[A-Za-z0-9_.-]{1,64}$/

/** 字段在当前配置下是否显示 */
export function fieldVisible(field: SchemaField, ext: Record<string, any>): boolean {
  if (!field.show_when) return true
  return Object.entries(field.show_when).every(([key, want]) => {
    const got = ext?.[key]
    return got !== undefined && got !== null && got !== '' && String(got) === String(want)
  })
}

export function isEmptyValue(val: any): boolean {
  if (val === undefined || val === null) return true
  if (typeof val === 'string') return val.trim() === ''
  if (typeof val === 'boolean' || typeof val === 'number') return false
  if (Array.isArray(val)) return val.length === 0
  if (typeof val === 'object') return Object.keys(val).length === 0
  return false
}

/** 条件分支配置中的 then 值 */
export function caseRelations(raw: any): string[] {
  let items = raw
  if (typeof raw === 'string') {
    try {
      items = JSON.parse(raw)
    } catch {
      items = []
    }
  }
  if (!Array.isArray(items)) return []
  return items.map((item) => (item && typeof item.then === 'string' ? item.then.trim() : '')).filter(Boolean)
}

/** 节点可输出的连接类型 */
export function allowedRelations(node: Pick<DesignNode, 'extparams'>, meta?: NodeTypeInfo): string[] {
  if (!meta) return ['default']
  const hasCases = meta.config_schema?.fields?.some((f) => f.type === 'cases')
  const result = [...(meta.relation_types ?? [])]
  if (hasCases) result.push(...caseRelations(node.extparams?.cases))
  result.push('default')
  return [...new Set(result)]
}

export function nodeRefValues(raw: any): string[] {
  if (typeof raw === 'string') return raw.split(',').map((s) => s.trim()).filter(Boolean)
  if (Array.isArray(raw)) return raw.filter((s) => typeof s === 'string' && s.trim()).map((s) => s.trim())
  return []
}

function label(node: DesignNode) {
  return node.name?.trim() ? `${node.name}(${node.id})` : node.id
}

/**
 * 实时校验(与后端结构校验规则一致)，发布前后端还会进行引擎编译校验。
 * strict 为 true 时必填配置缺失、使用已禁用节点类型视为错误。
 */
export function validateModel(
  flowType: FlowType,
  nodes: DesignNode[],
  conns: DesignConnection[],
  types: Map<string, NodeTypeInfo>,
  strict = false,
): ValidationResult {
  const errors: ValidationIssue[] = []
  const warnings: ValidationIssue[] = []
  const add = (level: 'error' | 'warning', code: string, message: string, extra: Partial<ValidationIssue> = {}) =>
    (level === 'error' ? errors : warnings).push({ level, code, message, ...extra })
  const strictLevel = strict ? 'error' : 'warning'

  if (!nodes.length) {
    add(strictLevel, 'empty', '流程中没有节点')
    return { valid: errors.length === 0, errors, warnings }
  }

  const byId = new Map<string, DesignNode>()
  for (const node of nodes) {
    if (!NODE_ID.test(node.id)) {
      add('error', 'node_id_invalid', `节点编号 ${node.id} 无效`, { node_id: node.id })
      continue
    }
    if (byId.has(node.id)) {
      add('error', 'node_id_duplicate', `节点编号 ${node.id} 重复`, { node_id: node.id })
      continue
    }
    byId.set(node.id, node)
    if (!node.name?.trim()) add('error', 'node_name_empty', `节点 ${node.id} 名称为空`, { node_id: node.id })
    const meta = types.get(node.type)
    if (!meta) {
      add('error', 'node_type_unregistered', `节点 ${label(node)} 的类型 ${node.type} 未注册`, { node_id: node.id })
      continue
    }
    if (meta.status !== 1) add(strictLevel, 'node_type_disabled', `节点 ${label(node)} 的类型 ${node.type} 已禁用`, { node_id: node.id })
    if (!meta.flow_modes.includes(flowType)) {
      add('error', 'node_type_mode', `节点 ${label(node)} 的类型 ${meta.name} 不能用于${flowType === 'WORKFLOW' ? '工作流' : '规则链'}`, { node_id: node.id })
    }
    for (const field of meta.config_schema?.fields ?? []) {
      if (field.required && fieldVisible(field, node.extparams) && isEmptyValue(node.extparams?.[field.key])) {
        add(strictLevel, 'config_required', `节点 ${label(node)} 缺少必填配置：${field.label || field.key}`, { node_id: node.id })
      }
    }
  }

  const incoming = new Map<string, number>()
  const outgoing = new Map<string, number>()
  const seen = new Set<string>()
  for (const conn of conns) {
    const from = byId.get(conn.from_id)
    const to = byId.get(conn.to_id)
    const type = conn.type?.trim() || 'default'
    if (!from || !to) {
      add('error', 'conn_node_missing', `连接 ${conn.from_id} → ${conn.to_id} 引用了不存在的节点`, { from_id: conn.from_id, to_id: conn.to_id })
      continue
    }
    if (!RELATION.test(type)) {
      add('error', 'conn_type_invalid', `连接 ${label(from)} → ${label(to)} 的类型 ${type} 无效`, { from_id: from.id, to_id: to.id })
      continue
    }
    const key = `${from.id}|${to.id}|${type}`
    if (seen.has(key)) {
      add('error', 'conn_duplicate', `连接 ${label(from)} → ${label(to)}(${type}) 重复`, { from_id: from.id, to_id: to.id })
      continue
    }
    seen.add(key)
    incoming.set(to.id, (incoming.get(to.id) ?? 0) + 1)
    outgoing.set(from.id, (outgoing.get(from.id) ?? 0) + 1)
    if (from.type === 'end') add('warning', 'end_outgoing', `结束节点 ${label(from)} 的后续连接不会执行`, { node_id: from.id })
    const meta = types.get(from.type)
    const allowed = allowedRelations(from, meta)
    if (meta && allowed.length > 1 && !allowed.includes(type)) {
      add('warning', 'conn_type_unknown', `连接 ${label(from)} → ${label(to)} 的类型 ${type} 不在节点输出类型 [${allowed.join(', ')}] 中`, {
        from_id: from.id,
        to_id: to.id,
      })
    }
  }

  const referenced = new Set<string>()
  for (const node of byId.values()) {
    const meta = types.get(node.type)
    for (const field of meta?.config_schema?.fields ?? []) {
      if (field.type !== 'nodeRef' && field.type !== 'nodeRefs') continue
      for (const ref of nodeRefValues(node.extparams?.[field.key])) {
        if (ref === node.id) add('error', 'node_ref_self', `节点 ${label(node)} 的配置 ${field.label} 引用了自身`, { node_id: node.id })
        else if (!byId.has(ref)) add('error', 'node_ref_missing', `节点 ${label(node)} 的配置 ${field.label} 引用了不存在的节点 ${ref}`, { node_id: node.id })
        else referenced.add(ref)
      }
    }
  }

  if (flowType === 'RULE_CHAIN') {
    const indegree = new Map<string, number>([...byId.keys()].map((id) => [id, 0]))
    const adjacency = new Map<string, string[]>()
    const edgeSeen = new Set<string>()
    for (const conn of conns) {
      if (!byId.has(conn.from_id) || !byId.has(conn.to_id)) continue
      const key = `${conn.from_id}|${conn.to_id}`
      if (edgeSeen.has(key)) continue
      edgeSeen.add(key)
      adjacency.set(conn.from_id, [...(adjacency.get(conn.from_id) ?? []), conn.to_id])
      indegree.set(conn.to_id, (indegree.get(conn.to_id) ?? 0) + 1)
    }
    const queue = [...indegree.entries()].filter(([, d]) => d === 0).map(([id]) => id)
    let visited = 0
    while (queue.length) {
      const id = queue.shift()!
      visited++
      for (const next of adjacency.get(id) ?? []) {
        indegree.set(next, indegree.get(next)! - 1)
        if (indegree.get(next) === 0) queue.push(next)
      }
    }
    if (visited < indegree.size) {
      const cycle = [...indegree.entries()].filter(([, d]) => d > 0).map(([id]) => id)
      cycle.forEach((id, i) =>
        add('error', 'cycle', i === 0 ? `规则链必须是有向无环图，存在循环连接：${cycle.map((c) => label(byId.get(c)!)).join('、')}` : `节点 ${label(byId.get(id)!)} 处于循环连接中`, {
          node_id: id,
        }),
      )
    }
  }

  let starts = 0
  for (const node of byId.values()) {
    if (flowType === 'WORKFLOW' && node.type === 'start') {
      starts++
      continue
    }
    if ((incoming.get(node.id) ?? 0) > 0 || referenced.has(node.id)) continue
    starts++
    if ((outgoing.get(node.id) ?? 0) === 0 && byId.size > 1) {
      add('warning', 'isolated', `节点 ${label(node)} 没有任何连接，将作为独立的起始节点执行`, { node_id: node.id })
    }
  }
  if (starts === 0) add('error', 'no_start', '没有起始节点：每个节点都有入边或被其他节点引用')
  if (flowType === 'WORKFLOW' && ![...byId.values()].some((n) => n.type === 'end')) add('warning', 'no_end', '工作流中没有结束节点')

  return { valid: errors.length === 0, errors, warnings }
}

/** 新建节点的初始配置：默认配置 + 字段默认值 */
export function initialConfig(meta: NodeTypeInfo): Record<string, any> {
  const ext: Record<string, any> = JSON.parse(JSON.stringify(meta.default_config ?? {}))
  for (const field of meta.config_schema?.fields ?? []) {
    if (field.default !== undefined && ext[field.key] === undefined) ext[field.key] = field.default
  }
  return ext
}

/** 新建连线默认的连接类型：取节点尚未使用的第一个输出类型 */
export function defaultRelation(node: DesignNode, meta: NodeTypeInfo | undefined, used: string[]): string {
  const allowed = allowedRelations(node, meta).filter((r) => r !== 'failure')
  return allowed.find((r) => !used.includes(r)) ?? allowed[0] ?? 'default'
}
