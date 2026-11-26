<template>
  <div class="flow-canvas">
    <div ref="containerRef" class="flow-canvas__graph" />
    <div v-show="minimap" ref="minimapRef" class="flow-canvas__minimap" />
    <div class="flow-canvas__zoom">
      <el-button-group size="small">
        <el-button :icon="'ZoomOut'" @click="zoomBy(-0.1)" />
        <el-button class="flow-canvas__scale" @click="zoomTo(1)">{{ Math.round(scale * 100) }}%</el-button>
        <el-button :icon="'ZoomIn'" @click="zoomBy(0.1)" />
        <el-button :icon="'FullScreen'" title="适应画布" @click="fit" />
      </el-button-group>
    </div>
    <div v-if="!readonly" class="flow-canvas__tips">拖动空白处平移画布 · Shift + 拖动框选 · Ctrl + 滚轮缩放 · Delete 删除</div>
  </div>
</template>

<script setup lang="ts">
import { Dnd, History, Keyboard, MiniMap, Selection, Snapline, type Cell, type Edge, type Graph, type Node } from '@antv/x6'
import { onBeforeUnmount, onMounted, ref, shallowRef } from 'vue'
import type { DebugResult, DesignConnection, DesignNode, NodeTypeInfo, ValidationResult } from '@/types'
import { newNodeId } from '@/utils'
import {
  applyEdgeType,
  applyRunStatus,
  autoLayout,
  clearMarks,
  createGraph,
  edgeKey,
  type EdgeData,
  exportModel,
  loadModel,
  markIssues,
  nodeMetadata,
  type NodeData,
  type NodeRunStatus,
  togglePorts,
} from './graph'
import { defaultRelation, initialConfig } from './validator'

const props = withDefaults(defineProps<{ readonly?: boolean; types: Map<string, NodeTypeInfo>; minimap?: boolean }>(), {
  readonly: false,
  minimap: true,
})

export type CanvasSelection = { kind: 'none' } | { kind: 'node'; id: string } | { kind: 'edge'; id: string }

const emit = defineEmits<{
  (e: 'select', selection: CanvasSelection): void
  (e: 'change'): void
}>()

const containerRef = ref<HTMLDivElement>()
const minimapRef = ref<HTMLDivElement>()
const graph = shallowRef<Graph>()
const scale = ref(1)
let history: History | undefined
let selection: Selection | undefined
let dnd: Dnd | undefined
let suppressChange = false

function notifyChange() {
  if (!suppressChange) emit('change')
}

let activeEdgeId = ''

/** 选中连线时显示拐点编辑与删除工具 */
function activateEdge(edge: Edge | null) {
  const g = graph.value
  if (!g) return
  if (activeEdgeId && activeEdgeId !== edge?.id) {
    const prev = g.getCellById(activeEdgeId)
    if (prev?.isEdge()) prev.removeTools()
  }
  activeEdgeId = edge?.id ?? ''
  if (edge) {
    edge.removeTools()
    edge.addTools([{ name: 'vertices' }, { name: 'button-remove', args: { distance: '25%' } }])
  }
}

onMounted(() => {
  const g = createGraph(containerRef.value!, { readonly: props.readonly })
  graph.value = g
  if (minimapRef.value && props.minimap) {
    g.use(new MiniMap({ container: minimapRef.value, width: 180, height: 120, padding: 10 }))
  }
  if (!props.readonly) {
    selection = new Selection({ enabled: true, multiple: true, rubberband: true, modifiers: 'shift', showNodeSelectionBox: true, pointerEvents: 'none' })
    // 只记录结构与数据变化，高亮等样式变化不进入撤销栈
    history = new History({ enabled: true, beforeAddCommand: (_event, args: any) => !(args && args.key === 'attrs') })
    g.use(selection)
    g.use(new Snapline({ enabled: true }))
    g.use(new Keyboard({ enabled: true, global: false }))
    g.use(history)
    // 拖拽与放置时都保留节点编号(默认 clone 会生成 UUID，超出节点编号长度限制)
    dnd = new Dnd({
      target: g,
      scaled: false,
      getDragNode: (node) => node.clone({ keepId: true }),
      getDropNode: (node) => node.clone({ keepId: true }),
    })
    bindEditEvents(g)
  }
  g.on('scale', ({ sx }) => (scale.value = sx))
  g.on('node:click', ({ node }) => emit('select', { kind: 'node', id: node.id }))
  g.on('edge:click', ({ edge }) => emit('select', { kind: 'edge', id: edge.id }))
  g.on('blank:click', () => emit('select', { kind: 'none' }))
})

function bindEditEvents(g: Graph) {
  const container = containerRef.value!
  g.on('node:mouseenter', ({ node }) => togglePorts(container, node, true))
  g.on('node:mouseleave', ({ node }) => togglePorts(container, node, false))
  // 悬停只显示删除按钮；拐点编辑工具只加在选中的连线上，避免其点击区域覆盖输出端口导致无法再次连线
  g.on('edge:mouseenter', ({ edge }) => {
    if (edge.id !== activeEdgeId) edge.addTools([{ name: 'button-remove', args: { distance: '25%' } }])
  })
  g.on('edge:mouseleave', ({ edge }) => {
    if (edge.id !== activeEdgeId) edge.removeTools()
  })
  g.on('edge:click', ({ edge }) => activateEdge(edge))
  g.on('node:click', () => activateEdge(null))
  g.on('blank:click', () => activateEdge(null))
  g.on('edge:removed', ({ edge }) => {
    if (edge.id === activeEdgeId) activeEdgeId = ''
  })

  // 新建连线：按源节点类型设置默认连接类型
  g.on('edge:connected', ({ edge, isNew }) => {
    if (!isNew) return notifyChange()
    const source = edge.getSourceNode()
    const data = (source?.getData() ?? {}) as NodeData
    const used = g
      .getOutgoingEdges(source!)
      ?.filter((e) => e.id !== edge.id)
      .map((e) => ((e.getData() ?? {}) as EdgeData).type) ?? []
    const type = defaultRelation({ id: source!.id, name: data.name, type: data.type, layout: {}, extparams: data.extparams }, props.types.get(data.type), used)
    edge.setData({ type, remark: '', extparams: {} } satisfies EdgeData, { overwrite: true })
    applyEdgeType(edge, type)
    emit('select', { kind: 'edge', id: edge.id })
    notifyChange()
  })
  // 撤销/重做时同步数据对应的展示
  g.on('edge:change:data', ({ edge }) => applyEdgeType(edge, ((edge.getData() ?? {}) as EdgeData).type || 'default'))
  g.on('node:change:data', ({ node }) => node.attr('title/text', ((node.getData() ?? {}) as NodeData).name))

  for (const event of ['node:added', 'node:removed', 'node:moved', 'edge:removed', 'cell:change:data', 'edge:change:vertices'] as const) {
    g.on(event, notifyChange)
  }

  const keyboard = g.getPlugin<Keyboard>('keyboard')!
  keyboard.bindKey(['delete', 'backspace'], () => {
    removeSelected()
    return false
  })
  keyboard.bindKey(['ctrl+z', 'meta+z'], () => {
    undo()
    return false
  })
  keyboard.bindKey(['ctrl+shift+z', 'meta+shift+z', 'ctrl+y'], () => {
    redo()
    return false
  })
  keyboard.bindKey(['ctrl+a', 'meta+a'], () => {
    selection?.select(g.getCells())
    return false
  })
}

/** 加载模型(不进入撤销栈) */
function load(nodes: DesignNode[], conns: DesignConnection[]) {
  const g = graph.value!
  suppressChange = true
  history?.disable()
  loadModel(g, nodes, conns, props.types)
  history?.enable()
  history?.clean()
  suppressChange = false
  requestAnimationFrame(() => fit())
}

function getModel() {
  return exportModel(graph.value!)
}

/** 从工具箱拖入节点 */
function startDrag(meta: NodeTypeInfo, evt: MouseEvent) {
  if (!dnd || props.readonly) return
  const node = graph.value!.createNode(
    nodeMetadata({ id: newNodeId(), type: meta.node_type, name: meta.name, layout: {}, extparams: initialConfig(meta) }, meta),
  )
  dnd.start(node, evt)
}

function getNode(id: string): Node | undefined {
  const cell = graph.value?.getCellById(id)
  return cell?.isNode() ? cell : undefined
}

function getEdge(id: string): Edge | undefined {
  const cell = graph.value?.getCellById(id)
  return cell?.isEdge() ? cell : undefined
}

function updateNodeData(id: string, patch: Partial<NodeData>) {
  const node = getNode(id)
  if (!node) return
  const data = { ...((node.getData() ?? {}) as NodeData), ...patch }
  node.setData(data, { overwrite: true })
  if (patch.name !== undefined) node.attr('title/text', patch.name)
}

function updateEdgeData(id: string, patch: Partial<EdgeData>) {
  const edge = getEdge(id)
  if (!edge) return
  const data = { ...((edge.getData() ?? {}) as EdgeData), ...patch }
  edge.setData(data, { overwrite: true })
  if (patch.type !== undefined) applyEdgeType(edge, patch.type || 'default')
}

function removeSelected() {
  const cells: Cell[] = selection?.getSelectedCells() ?? []
  if (!cells.length) return
  graph.value!.removeCells(cells)
  emit('select', { kind: 'none' })
}

function removeCell(id: string) {
  graph.value?.removeCell(id)
  emit('select', { kind: 'none' })
}

function undo() {
  if (history?.canUndo()) history.undo()
}

function redo() {
  if (history?.canRedo()) history.redo()
}

function zoomBy(delta: number) {
  graph.value?.zoom(delta, { minScale: 0.2, maxScale: 2.5 })
}

function zoomTo(value: number) {
  graph.value?.zoomTo(value)
}

function fit() {
  const g = graph.value
  if (!g) return
  if (g.getNodes().length) g.zoomToFit({ padding: 40, maxScale: 1 })
  else g.zoomTo(1)
}

function layout() {
  autoLayout(graph.value!)
  requestAnimationFrame(() => fit())
}

/** 选中并定位节点 */
function focusNode(id: string) {
  const node = getNode(id)
  if (!node) return
  selection?.reset(node)
  graph.value!.centerCell(node)
  emit('select', { kind: 'node', id })
}

function focusEdge(from: string, to: string) {
  const edge = graph.value?.getEdges().find((e) => e.getSourceCellId() === from && e.getTargetCellId() === to)
  if (!edge) return
  selection?.reset(edge)
  emit('select', { kind: 'edge', id: edge.id })
}

function withoutHistory(fn: () => void) {
  history?.disable()
  try {
    fn()
  } finally {
    history?.enable()
  }
}

function showIssues(result?: ValidationResult) {
  const g = graph.value
  if (!g) return
  const errors = new Set((result?.errors ?? []).map((i) => i.node_id).filter(Boolean) as string[])
  const warnings = new Set((result?.warnings ?? []).map((i) => i.node_id).filter(Boolean) as string[])
  withoutHistory(() => markIssues(g, errors, warnings))
}

function showDebug(result: DebugResult) {
  const status: Record<string, NodeRunStatus> = {}
  for (const ev of result.events) {
    if (ev.error) status[ev.node_id] = 'error'
    else if (status[ev.node_id] !== 'error') status[ev.node_id] = ev.mocked ? 'mocked' : 'success'
  }
  const edges = new Set(result.edges.map((e) => edgeKey(e.from_id, e.to_id, e.type)))
  withoutHistory(() => applyRunStatus(graph.value!, status, edges))
}

/** 执行历史：按节点状态高亮 */
function showStatus(status: Record<string, NodeRunStatus>, edges: Set<string>) {
  withoutHistory(() => applyRunStatus(graph.value!, status, edges))
}

function clearHighlight() {
  withoutHistory(() => clearMarks(graph.value!))
}

onBeforeUnmount(() => {
  graph.value?.dispose()
})

defineExpose({
  load,
  getModel,
  startDrag,
  getNode,
  getEdge,
  updateNodeData,
  updateEdgeData,
  removeSelected,
  removeCell,
  undo,
  redo,
  fit,
  layout,
  focusNode,
  focusEdge,
  showIssues,
  showDebug,
  showStatus,
  clearHighlight,
  graph,
})
</script>

<style scoped>
.flow-canvas {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.flow-canvas__graph {
  width: 100%;
  height: 100%;
}

.flow-canvas__minimap {
  position: absolute;
  right: 12px;
  bottom: 12px;
  border: 1px solid #e4e7ed;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.flow-canvas__zoom {
  position: absolute;
  left: 12px;
  bottom: 12px;
}

.flow-canvas__scale {
  width: 64px;
}

.flow-canvas__tips {
  position: absolute;
  left: 12px;
  top: 10px;
  color: #a8abb2;
  font-size: 12px;
  pointer-events: none;
}
</style>
