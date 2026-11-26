import type { DesignConnection, DesignNode, FlowType } from '@/types'
import { newNodeId } from '@/utils'

/** 导出文件格式(与后端 /chain/export 一致) */
export const EXPORT_FORMAT = 'manager-glow/chain@1'

/** 解析导入内容：支持本系统导出文件与 glow-flow ChainDefinition JSON */
export function parseImportContent(text: string): { nodes: DesignNode[]; connections: DesignConnection[]; flowType?: FlowType } {
  const data = JSON.parse(text)
  if (!data || typeof data !== 'object') throw new Error('导入内容必须是 JSON 对象')
  const def = data.definition ?? data
  if (!Array.isArray(def.nodes)) throw new Error('无法识别的导入内容：缺少 nodes')
  const nodes: DesignNode[] = def.nodes.map((n: any) => ({
    id: String(n.id ?? ''),
    type: String(n.type ?? ''),
    name: String(n.name ?? n.id ?? ''),
    layout: n.layout ?? {},
    extparams: n.extparams ?? {},
  }))
  const extras = new Map<string, any>()
  for (const edge of data.edges ?? []) extras.set(`${edge.from_id}|${edge.to_id}|${edge.type || 'default'}`, edge.extparams)
  const connections: DesignConnection[] = (def.connections ?? []).map((c: any) => {
    const type = c.type || 'default'
    const conn: DesignConnection = { from_id: String(c.from_id), to_id: String(c.to_id), type, remark: c.remark ?? '' }
    const ext = extras.get(`${conn.from_id}|${conn.to_id}|${type}`)
    if (ext) conn.extparams = ext
    return conn
  })
  return { nodes, connections, flowType: data.flow_type }
}

/** 生成导出文件内容 */
export function buildExport(chainNo: string, name: string, flowType: FlowType, desc: string, versionNo: number, ext: Record<string, any>, nodes: DesignNode[], conns: DesignConnection[]) {
  return {
    format: EXPORT_FORMAT,
    export_time: new Date().toLocaleString('zh-CN', { hour12: false }),
    flow_type: flowType,
    desc,
    version_no: versionNo,
    definition: {
      id: chainNo,
      version: String(versionNo),
      metadata: { id: chainNo, name, status: 1, extparams: ext },
      endpoints: [],
      nodes: nodes.map((n) => ({ id: n.id, type: n.type, name: n.name, layout: n.layout, extparams: n.extparams })),
      connections: conns.map((c) => ({ from_id: c.from_id, to_id: c.to_id, type: c.type || 'default', remark: c.remark || undefined })),
    },
    edges: conns.filter((c) => c.extparams && Object.keys(c.extparams).length).map((c) => ({ from_id: c.from_id, to_id: c.to_id, type: c.type, extparams: c.extparams })),
  }
}

/** 重新生成全部节点编号，并替换连接与配置中引用的节点编号(与后端 RemapNodeIDs 一致) */
export function regenerateIds(nodes: DesignNode[], conns: DesignConnection[]) {
  const mapping = new Map(nodes.map((n) => [n.id, newNodeId()]))
  const remap = (val: any): any => {
    if (typeof val === 'string') {
      if (mapping.has(val)) return mapping.get(val)
      if (val.includes(',')) {
        const parts = val.split(',').map((p) => p.trim())
        if (parts.every((p) => mapping.has(p))) return parts.map((p) => mapping.get(p)).join(',')
      }
      return val
    }
    if (Array.isArray(val)) return val.map(remap)
    if (val && typeof val === 'object') return Object.fromEntries(Object.entries(val).map(([k, v]) => [k, remap(v)]))
    return val
  }
  return {
    nodes: nodes.map((n) => ({ ...n, id: mapping.get(n.id)!, extparams: remap(n.extparams ?? {}) })),
    connections: conns.map((c) => ({ ...c, from_id: mapping.get(c.from_id) ?? c.from_id, to_id: mapping.get(c.to_id) ?? c.to_id })),
  }
}
