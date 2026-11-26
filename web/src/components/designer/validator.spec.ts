import { describe, expect, it } from 'vitest'
import type { DesignConnection, DesignNode, NodeTypeInfo } from '@/types'
import { parseImportContent, regenerateIds } from './model'
import { allowedRelations, defaultRelation, initialConfig, validateModel } from './validator'

function meta(partial: Partial<NodeTypeInfo> & { node_type: string }): NodeTypeInfo {
  return {
    name: partial.node_type,
    category: 'filter',
    flow_modes: ['RULE_CHAIN'],
    source: 'builtin',
    service_name: '',
    service_path: '',
    protocol: '',
    relation_types: [],
    config_schema: { fields: [] },
    default_config: {},
    icon: '',
    color: '',
    sort_no: 0,
    status: 1,
    desc: '',
    create_time: '',
    create_by: '',
    operator: '',
    update_time: '',
    usage_count: 0,
    engine_supported: true,
    ...partial,
  }
}

const types = new Map<string, NodeTypeInfo>([
  ['exprfilter', meta({ node_type: 'exprfilter', relation_types: ['true', 'false', 'failure'], config_schema: { fields: [{ key: 'expr', label: '表达式', type: 'code', required: true }] } })],
  ['switch', meta({ node_type: 'switch', flow_modes: ['RULE_CHAIN', 'WORKFLOW'], config_schema: { fields: [{ key: 'cases', label: '条件分支', type: 'cases', required: true }] } })],
  ['for', meta({ node_type: 'for', relation_types: ['success', 'failure'], config_schema: { fields: [{ key: 'do', label: '循环体', type: 'nodeRef', required: true }, { key: 'mode', label: '模式', type: 'select', default: 0 }] } })],
  ['start', meta({ node_type: 'start', flow_modes: ['WORKFLOW'], relation_types: ['success'] })],
  ['end', meta({ node_type: 'end', flow_modes: ['RULE_CHAIN', 'WORKFLOW'] })],
  ['dingtalk', meta({
    node_type: 'dingtalk',
    default_config: { target: 'group' },
    config_schema: {
      fields: [
        { key: 'target', label: '对象', type: 'select' },
        { key: 'webhook', label: 'Webhook', type: 'string', required: true, show_when: { target: 'group' } },
        { key: 'appkey', label: 'AppKey', type: 'string', required: true, show_when: { target: 'user' } },
      ],
    },
  })],
])

const node = (id: string, type: string, extparams: Record<string, any> = {}): DesignNode => ({ id, type, name: id, layout: {}, extparams })
const conn = (from_id: string, to_id: string, type = 'default'): DesignConnection => ({ from_id, to_id, type })

describe('validateModel', () => {
  it('rejects cycles in rule chains but allows them in workflows', () => {
    const nodes = [node('a', 'exprfilter', { expr: 'true' }), node('b', 'end')]
    const conns = [conn('a', 'b', 'true'), conn('b', 'a')]
    expect(validateModel('RULE_CHAIN', nodes, conns, types).errors.some((e) => e.code === 'cycle')).toBe(true)

    const wf = [node('s', 'start'), node('sw', 'switch', { cases: [{ case: 'x', then: 'back' }] }), node('e', 'end')]
    const wfConns = [conn('s', 'sw', 'success'), conn('sw', 's', 'back'), conn('sw', 'e')]
    const result = validateModel('WORKFLOW', wf, wfConns, types)
    expect(result.errors).toEqual([])
  })

  it('reports missing required config as warning for drafts and error when strict', () => {
    const nodes = [node('a', 'exprfilter')]
    expect(validateModel('RULE_CHAIN', nodes, [], types).warnings.map((w) => w.code)).toContain('config_required')
    expect(validateModel('RULE_CHAIN', nodes, [], types, true).errors.map((e) => e.code)).toContain('config_required')
  })

  it('only requires fields visible under show_when', () => {
    const nodes = [node('d', 'dingtalk', { target: 'group', webhook: 'https://x' })]
    expect(validateModel('RULE_CHAIN', nodes, [], types, true).valid).toBe(true)
    nodes[0].extparams = { target: 'user' }
    const errors = validateModel('RULE_CHAIN', nodes, [], types, true).errors
    expect(errors.map((e) => e.message).join()).toContain('AppKey')
    expect(errors.map((e) => e.message).join()).not.toContain('Webhook')
  })

  it('checks node references and mode restrictions', () => {
    const nodes = [node('loop', 'for', { do: 'missing' }), node('s', 'start')]
    const codes = validateModel('RULE_CHAIN', nodes, [], types).errors.map((e) => e.code)
    expect(codes).toContain('node_ref_missing')
    expect(codes).toContain('node_type_mode')
  })

  it('treats referenced nodes as non-start nodes', () => {
    const nodes = [node('loop', 'for', { do: 'body' }), node('body', 'end')]
    const result = validateModel('RULE_CHAIN', nodes, [], types)
    expect(result.warnings.filter((w) => w.code === 'isolated').map((w) => w.node_id)).toEqual(['loop'])
  })

  it('flags duplicate connections and unknown relations', () => {
    const nodes = [node('a', 'exprfilter', { expr: 'true' }), node('b', 'end')]
    const result = validateModel('RULE_CHAIN', nodes, [conn('a', 'b', 'true'), conn('a', 'b', 'true'), conn('a', 'b', 'maybe')], types)
    expect(result.errors.map((e) => e.code)).toContain('conn_duplicate')
    expect(result.warnings.map((w) => w.code)).toContain('conn_type_unknown')
  })
})

describe('relations', () => {
  it('includes switch case targets and default', () => {
    expect(allowedRelations({ extparams: { cases: [{ case: 'a', then: 'hot' }, { case: 'b', then: 'cold' }] } }, types.get('switch'))).toEqual(['hot', 'cold', 'default'])
  })

  it('picks the first unused relation for new connections', () => {
    const filter = node('a', 'exprfilter', { expr: 'true' })
    expect(defaultRelation(filter, types.get('exprfilter'), [])).toBe('true')
    expect(defaultRelation(filter, types.get('exprfilter'), ['true'])).toBe('false')
  })

  it('builds initial config from defaults', () => {
    expect(initialConfig(types.get('for')!)).toEqual({ mode: 0 })
    expect(initialConfig(types.get('dingtalk')!)).toEqual({ target: 'group' })
  })
})

describe('import helpers', () => {
  it('parses export files with edge extras', () => {
    const text = JSON.stringify({
      flow_type: 'RULE_CHAIN',
      definition: { nodes: [{ id: 'a', type: 'end', name: 'A' }, { id: 'b', type: 'end', name: 'B' }], connections: [{ from_id: 'a', to_id: 'b', type: 'success' }] },
      edges: [{ from_id: 'a', to_id: 'b', type: 'success', extparams: { vertices: [{ x: 1, y: 2 }] } }],
    })
    const result = parseImportContent(text)
    expect(result.flowType).toBe('RULE_CHAIN')
    expect(result.connections[0].extparams).toEqual({ vertices: [{ x: 1, y: 2 }] })
  })

  it('regenerates ids and rewrites references', () => {
    const nodes = [node('a', 'for', { do: 'b' }), node('b', 'end'), node('g', 'end', { nodeids: 'a,b' })]
    const { nodes: out, connections } = regenerateIds(nodes, [conn('a', 'g')])
    const [a, b, g] = out
    expect(a.id).toMatch(/^n[a-z0-9]{11}$/)
    expect(a.extparams.do).toBe(b.id)
    expect(g.extparams.nodeids).toBe(`${a.id},${b.id}`)
    expect(connections[0]).toMatchObject({ from_id: a.id, to_id: g.id })
  })
})
