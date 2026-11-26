import { Graph, Shape, type Cell, type Edge, type EdgeMetadata, type Node, type NodeMetadata } from '@antv/x6'
import { DEFAULT_NODE_COLOR, relationColor } from '@/constants'
import type { DesignConnection, DesignNode, NodeTypeInfo } from '@/types'

export const NODE_SHAPE = 'gf-node'
export const NODE_W = 188
export const NODE_H = 48

/** 节点 cell 中保存的业务数据 */
export interface NodeData {
  type: string
  name: string
  extparams: Record<string, any>
}

/** 连线 cell 中保存的业务数据 */
export interface EdgeData {
  type: string
  remark?: string
  extparams?: Record<string, any>
}

const portAttrs = {
  circle: {
    r: 5,
    magnet: true,
    stroke: '#8f959e',
    strokeWidth: 1,
    fill: '#fff',
    style: { visibility: 'hidden' },
  },
}

let registered = false

export function registerShapes() {
  if (registered) return
  registered = true
  Graph.registerNode(
    NODE_SHAPE,
    {
      inherit: 'rect',
      width: NODE_W,
      height: NODE_H,
      markup: [
        { tagName: 'rect', selector: 'body' },
        { tagName: 'rect', selector: 'bar' },
        { tagName: 'text', selector: 'title' },
        { tagName: 'text', selector: 'subtitle' },
        { tagName: 'circle', selector: 'badge' },
      ],
      attrs: {
        body: { refWidth: '100%', refHeight: '100%', rx: 6, ry: 6, fill: '#fff', stroke: '#d4d7de', strokeWidth: 1 },
        bar: { width: 5, refHeight: '100%', rx: 2, ry: 2, fill: DEFAULT_NODE_COLOR },
        title: {
          refX: 16,
          refY: 17,
          fontSize: 13,
          fontWeight: 500,
          fill: '#303133',
          textAnchor: 'start',
          textVerticalAnchor: 'middle',
          textWrap: { width: -28, height: 18, ellipsis: true },
        },
        subtitle: {
          refX: 16,
          refY: 34,
          fontSize: 11,
          fill: '#909399',
          textAnchor: 'start',
          textVerticalAnchor: 'middle',
          textWrap: { width: -28, height: 16, ellipsis: true },
        },
        badge: { refX: '100%', refX2: -10, refY: 10, r: 4, fill: 'transparent' },
      },
      ports: {
        groups: {
          in: { position: 'top', attrs: portAttrs },
          out: { position: 'bottom', attrs: portAttrs },
        },
      },
    },
    true,
  )
}

export interface GraphOptions {
  readonly?: boolean
  /** 连线是否合法(业务层校验) */
  validateEdge?: (sourceId: string, targetId: string) => boolean
}

export function createGraph(container: HTMLElement, opts: GraphOptions = {}) {
  registerShapes()
  const readonly = !!opts.readonly
  return new Graph({
    container,
    autoResize: true,
    background: { color: '#f7f8fa' },
    grid: { visible: true, size: 10, type: 'dot', args: { color: '#cfd3dc', thickness: 1 } },
    panning: readonly ? { enabled: true } : { enabled: true, eventTypes: ['leftMouseDown'], modifiers: [] },
    mousewheel: { enabled: true, modifiers: ['ctrl', 'meta'], minScale: 0.2, maxScale: 2.5 },
    connecting: {
      router: { name: 'manhattan', args: { padding: 16 } },
      connector: { name: 'rounded', args: { radius: 8 } },
      anchor: 'center',
      connectionPoint: 'anchor',
      allowBlank: false,
      allowLoop: false,
      allowNode: false,
      allowEdge: false,
      allowMulti: true,
      highlight: true,
      snap: { radius: 24 },
      createEdge() {
        return new Shape.Edge({
          zIndex: 0,
          attrs: { line: { stroke: relationColor('default'), strokeWidth: 1.5, targetMarker: { name: 'block', width: 8, height: 6 } } },
        })
      },
      validateMagnet({ magnet }) {
        return !readonly && magnet.getAttribute('port-group') === 'out'
      },
      validateConnection({ sourceCell, targetCell, targetMagnet }) {
        if (!sourceCell || !targetCell || sourceCell === targetCell) return false
        if (!targetMagnet || targetMagnet.getAttribute('port-group') !== 'in') return false
        return opts.validateEdge ? opts.validateEdge(sourceCell.id, targetCell.id) : true
      },
    },
    highlighting: {
      magnetAdsorbed: { name: 'stroke', args: { attrs: { fill: '#409eff', stroke: '#409eff' } } },
    },
    interacting: readonly
      ? { nodeMovable: false, edgeMovable: false, edgeLabelMovable: false, vertexMovable: false, vertexAddable: false, vertexDeletable: false, magnetConnectable: false }
      : { edgeLabelMovable: false },
  })
}

/** 节点端口：结束节点没有输出端口 */
function portItems(type: string) {
  const items = [{ id: 'in', group: 'in' }]
  if (type !== 'end') items.push({ id: 'out', group: 'out' })
  return items
}

export function nodeMetadata(node: DesignNode, meta: NodeTypeInfo | undefined, fallbackIndex = 0): NodeMetadata {
  const x = node.layout?.X ?? 80 + (fallbackIndex % 4) * 230
  const y = node.layout?.Y ?? 60 + Math.floor(fallbackIndex / 4) * 110
  return {
    id: node.id,
    shape: NODE_SHAPE,
    x,
    y,
    width: node.layout?.W || NODE_W,
    height: node.layout?.H || NODE_H,
    ports: { items: portItems(node.type) },
    attrs: {
      title: { text: node.name },
      subtitle: { text: meta ? `${meta.name} · ${node.type}` : `${node.type}(未注册)` },
      bar: { fill: meta?.color || DEFAULT_NODE_COLOR },
    },
    data: { type: node.type, name: node.name, extparams: node.extparams ?? {} } satisfies NodeData,
  }
}

export function edgeLabel(type: string) {
  const color = relationColor(type)
  return {
    attrs: {
      label: { text: type || 'default', fill: color, fontSize: 11 },
      body: { fill: '#fff', stroke: color, strokeWidth: 1, rx: 3, ry: 3, refWidth: 10, refHeight: 4, refX: -5, refY: -2 },
    },
    position: 0.5,
  }
}

export function edgeMetadata(conn: DesignConnection): EdgeMetadata {
  const type = conn.type || 'default'
  const { vertices, ...rest } = conn.extparams ?? {}
  return {
    shape: 'edge',
    source: { cell: conn.from_id, port: 'out' },
    target: { cell: conn.to_id, port: 'in' },
    vertices: Array.isArray(vertices) ? vertices : undefined,
    zIndex: 0,
    attrs: { line: { stroke: relationColor(type), strokeWidth: 1.5, targetMarker: { name: 'block', width: 8, height: 6 } } },
    labels: [edgeLabel(type)],
    data: { type, remark: conn.remark ?? '', extparams: rest } satisfies EdgeData,
  }
}

/** 更新连线类型的展示 */
export function applyEdgeType(edge: Edge, type: string) {
  edge.setLabels([edgeLabel(type)])
  edge.attr('line/stroke', relationColor(type))
}

/** 将模型加载到画布 */
export function loadModel(graph: Graph, nodes: DesignNode[], conns: DesignConnection[], types: Map<string, NodeTypeInfo>) {
  const cells: Cell[] = []
  nodes.forEach((node, index) => cells.push(graph.createNode(nodeMetadata(node, types.get(node.type), index))))
  const ids = new Set(nodes.map((n) => n.id))
  conns
    .filter((conn) => ids.has(conn.from_id) && ids.has(conn.to_id))
    .forEach((conn) => cells.push(graph.createEdge(edgeMetadata(conn))))
  graph.resetCells(cells)
  // 导入的模型没有布局信息时自动布局
  if (nodes.length > 1 && nodes.every((n) => !n.layout?.X && !n.layout?.Y)) {
    autoLayout(graph)
  }
}

/** 从画布导出模型 */
export function exportModel(graph: Graph): { nodes: DesignNode[]; connections: DesignConnection[] } {
  const nodes = graph.getNodes().map((node) => {
    const data = (node.getData() ?? {}) as NodeData
    const pos = node.getPosition()
    const size = node.getSize()
    const layout: DesignNode['layout'] = { X: Math.round(pos.x), Y: Math.round(pos.y) }
    if (size.width !== NODE_W) layout.W = Math.round(size.width)
    if (size.height !== NODE_H) layout.H = Math.round(size.height)
    return { id: node.id, type: data.type, name: data.name, layout, extparams: data.extparams ?? {} }
  })
  const connections = graph
    .getEdges()
    .map((edge) => {
      const data = (edge.getData() ?? {}) as EdgeData
      const ext: Record<string, any> = { ...(data.extparams ?? {}) }
      const vertices = edge.getVertices()
      if (vertices.length) ext.vertices = vertices.map((v) => ({ x: Math.round(v.x), y: Math.round(v.y) }))
      const conn: DesignConnection = {
        from_id: edge.getSourceCellId(),
        to_id: edge.getTargetCellId(),
        type: data.type || 'default',
        remark: data.remark || '',
      }
      if (Object.keys(ext).length) conn.extparams = ext
      return conn
    })
    .filter((conn) => conn.from_id && conn.to_id)
  return { nodes, connections }
}

/** 简单分层布局：按最长路径分层，层内按出现顺序排列 */
export function autoLayout(graph: Graph) {
  const nodes = graph.getNodes()
  if (!nodes.length) return
  const ids = nodes.map((n) => n.id)
  const outgoing = new Map<string, string[]>()
  const indegree = new Map<string, number>(ids.map((id) => [id, 0]))
  for (const edge of graph.getEdges()) {
    const from = edge.getSourceCellId()
    const to = edge.getTargetCellId()
    if (!indegree.has(from) || !indegree.has(to) || from === to) continue
    outgoing.set(from, [...(outgoing.get(from) ?? []), to])
    indegree.set(to, (indegree.get(to) ?? 0) + 1)
  }
  // Kahn 拓扑排序计算层级；环中节点在剩余节点里依次处理
  const level = new Map<string, number>()
  const remaining = new Map(indegree)
  const queue = ids.filter((id) => remaining.get(id) === 0)
  const visited = new Set<string>()
  const visit = (start: string[]) => {
    const q = [...start]
    while (q.length) {
      const id = q.shift()!
      if (visited.has(id)) continue
      visited.add(id)
      if (!level.has(id)) level.set(id, 0)
      for (const next of outgoing.get(id) ?? []) {
        if (visited.has(next)) continue
        level.set(next, Math.max(level.get(next) ?? 0, (level.get(id) ?? 0) + 1))
        remaining.set(next, (remaining.get(next) ?? 1) - 1)
        if ((remaining.get(next) ?? 0) <= 0) q.push(next)
      }
    }
  }
  visit(queue)
  for (const id of ids) {
    if (!visited.has(id)) visit([id])
  }
  const rows = new Map<number, string[]>()
  for (const id of ids) {
    const lv = level.get(id) ?? 0
    rows.set(lv, [...(rows.get(lv) ?? []), id])
  }
  const maxWidth = Math.max(...[...rows.values()].map((row) => row.length))
  graph.batchUpdate(() => {
    for (const [lv, row] of rows) {
      const offset = ((maxWidth - row.length) * (NODE_W + 50)) / 2
      row.forEach((id, i) => {
        graph.getCellById(id)?.setProp('position', { x: 60 + offset + i * (NODE_W + 50), y: 40 + lv * (NODE_H + 70) })
      })
    }
    graph.getEdges().forEach((edge) => edge.setVertices([]))
  })
}

export type NodeRunStatus = 'success' | 'error' | 'mocked' | 'running' | 'pending'

const statusColors: Record<NodeRunStatus, string> = {
  success: '#52a35b',
  error: '#e05252',
  mocked: '#d99a2b',
  running: '#3d7fe0',
  pending: '#a8abb2',
}

function setNodeOpacity(node: Node, opacity: number) {
  for (const selector of ['body', 'bar', 'title', 'subtitle']) {
    node.attr(`${selector}/opacity`, opacity)
  }
}

/** 按执行状态高亮节点与连线，未执行的节点置灰 */
export function applyRunStatus(graph: Graph, nodeStatus: Record<string, NodeRunStatus>, edgeKeys: Set<string>) {
  graph.batchUpdate(() => {
    for (const node of graph.getNodes()) {
      const status = nodeStatus[node.id]
      if (status) {
        node.attr('body/stroke', statusColors[status])
        node.attr('body/strokeWidth', 2)
        node.attr('badge/fill', statusColors[status])
        setNodeOpacity(node, 1)
      } else {
        node.attr('body/stroke', '#d4d7de')
        node.attr('body/strokeWidth', 1)
        node.attr('badge/fill', 'transparent')
        setNodeOpacity(node, 0.45)
      }
    }
    for (const edge of graph.getEdges()) {
      const key = edgeKey(edge.getSourceCellId(), edge.getTargetCellId(), ((edge.getData() ?? {}) as EdgeData).type)
      if (edgeKeys.has(key)) {
        edge.attr('line/strokeWidth', 2.5)
        edge.attr('line/strokeDasharray', 6)
        edge.attr('line/style', { animation: 'gf-ant-line 30s infinite linear' })
        edge.attr('line/opacity', 1)
      } else {
        edge.attr('line/strokeWidth', 1.5)
        edge.attr('line/strokeDasharray', 0)
        edge.attr('line/style', { animation: 'none' })
        edge.attr('line/opacity', 0.35)
      }
    }
  })
}

/** 清除执行状态与校验标记 */
export function clearMarks(graph: Graph) {
  graph.batchUpdate(() => {
    for (const node of graph.getNodes()) {
      node.attr('body/stroke', '#d4d7de')
      node.attr('body/strokeWidth', 1)
      node.attr('badge/fill', 'transparent')
      setNodeOpacity(node, 1)
    }
    for (const edge of graph.getEdges()) {
      edge.attr('line/strokeWidth', 1.5)
      edge.attr('line/strokeDasharray', 0)
      edge.attr('line/style', { animation: 'none' })
      edge.attr('line/opacity', 1)
    }
  })
}

/** 标记校验问题：错误节点红色边框，警告节点黄色角标 */
export function markIssues(graph: Graph, errorIds: Set<string>, warningIds: Set<string>) {
  graph.batchUpdate(() => {
    for (const node of graph.getNodes()) {
      if (errorIds.has(node.id)) {
        node.attr('body/stroke', '#e05252')
        node.attr('body/strokeWidth', 1.5)
        node.attr('badge/fill', '#e05252')
      } else if (warningIds.has(node.id)) {
        node.attr('body/stroke', '#d4d7de')
        node.attr('body/strokeWidth', 1)
        node.attr('badge/fill', '#e6a23c')
      } else {
        node.attr('body/stroke', '#d4d7de')
        node.attr('body/strokeWidth', 1)
        node.attr('badge/fill', 'transparent')
      }
    }
  })
}

export function edgeKey(from: string, to: string, type?: string) {
  return `${from}\u0000${to}\u0000${type || 'default'}`
}

/** 显示/隐藏节点端口 */
export function togglePorts(container: HTMLElement, node: Node | null, visible: boolean) {
  const selector = node ? `[data-cell-id="${node.id}"] .x6-port-body` : '.x6-port-body'
  container.querySelectorAll<SVGElement>(selector).forEach((el) => {
    el.style.visibility = visible ? 'visible' : 'hidden'
  })
}
