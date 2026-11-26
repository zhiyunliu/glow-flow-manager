export interface PageData<T> {
  list: T[]
  total: number
}

export type FlowType = 'RULE_CHAIN' | 'WORKFLOW'

export interface ChainInfo {
  chain_no: string
  name: string
  flow_type: FlowType
  status: number
  extparams: string
  layout: string
  desc: string
  create_time: string
  create_by: string
  operator: string
  update_time: string
  effective_version_id: number
  effective_version_no: number
  draft_version_id: number
  draft_version_no: number
  latest_version_no: number
  version_count: number
  published_count: number
}

export interface VersionInfo {
  version_id: number
  chain_no: string
  version_no: number
  name: string
  status: number
  publish_state: number
  extparams: string
  layout: string
  desc: string
  base_version_id: number
  publish_time: string
  publish_by: string
  create_time: string
  create_by: string
  operator: string
  update_time: string
  node_count: number
  conn_count: number
  is_effective: boolean
}

/** 与 glow-flow Layout 的 JSON 字段一致 */
export interface Layout {
  icon?: string
  X?: number
  Y?: number
  W?: number
  H?: number
}

export interface DesignNode {
  id: string
  type: string
  name: string
  layout: Layout
  extparams: Record<string, any>
}

export interface DesignConnection {
  from_id: string
  to_id: string
  type: string
  remark?: string
  extparams?: Record<string, any>
}

export interface ChainDesign {
  chain: ChainInfo
  version: VersionInfo
  extparams: Record<string, any>
  layout: Record<string, any>
  nodes: DesignNode[]
  connections: DesignConnection[]
  editable: boolean
}

export interface ValidationIssue {
  level: 'error' | 'warning'
  code: string
  message: string
  node_id?: string
  from_id?: string
  to_id?: string
}

export interface ValidationResult {
  valid: boolean
  errors: ValidationIssue[]
  warnings: ValidationIssue[]
}

export interface DebugEvent {
  seq: number
  node_id: string
  node_name: string
  node_type: string
  mocked: boolean
  input: any
  output: any
  relation_type: string
  relations?: string[]
  end?: boolean
  error?: string
  start_offset_ms: number
  duration_ms: number
}

export interface DebugEdge {
  from_id: string
  to_id: string
  type: string
}

export interface DebugResult {
  instance_id: string
  success: boolean
  error?: string
  timed_out: boolean
  duration_ms: number
  events: DebugEvent[]
  executed_nodes: string[]
  edges: DebugEdge[]
  validation?: ValidationResult
}

export interface SchemaOption {
  label: string
  value: any
}

export type FieldType =
  | 'string'
  | 'text'
  | 'number'
  | 'bool'
  | 'select'
  | 'code'
  | 'kv'
  | 'list'
  | 'cases'
  | 'nodeRef'
  | 'nodeRefs'
  | 'json'

export interface SchemaField {
  key: string
  label: string
  type: FieldType
  required?: boolean
  default?: any
  placeholder?: string
  help?: string
  lang?: string
  options?: SchemaOption[]
  /** 动态下拉数据源名称(select)，选项通过 /option/list?source= 获取 */
  options_source?: string
  group?: string
  show_when?: Record<string, any>
}

/** 动态下拉选项 */
export interface OptionItem {
  label: string
  value: string
  desc?: string
}

export interface NodeTypeInfo {
  node_type: string
  name: string
  category: string
  flow_modes: FlowType[]
  source: 'builtin' | 'service'
  service_name: string
  service_path: string
  protocol: string
  relation_types: string[]
  config_schema: { fields: SchemaField[] } | null
  default_config: Record<string, any>
  icon: string
  color: string
  sort_no: number
  status: number
  desc: string
  create_time: string
  create_by: string
  operator: string
  update_time: string
  usage_count: number
  engine_supported: boolean
}

export interface RegistryInstance {
  id: string
  name: string
  version: string
  healthy: boolean
  weight: number
  endpoints: string[]
  metadata: Record<string, string>
}

export interface NodeTypeInstances {
  node_type: string
  service_name: string
  registry: string
  instances: RegistryInstance[]
}

export interface InstanceInfo {
  instance_id: string
  chain_no: string
  chain_name: string
  version: string
  version_no: number
  status: string
  current_node_ids?: string[]
  variables?: Record<string, any>
  error?: string
  created_at: string
  updated_at: string
  finished_at?: string
  duration_ms: number
}

export interface NodeStateInfo {
  node_id: string
  node_name: string
  node_type: string
  status: string
  attempts: number
  input?: any
  output?: any
  metadata?: Record<string, any>
  error?: string
  started_at?: string
  finished_at?: string
  duration_ms: number
}

export interface InstanceDetail {
  instance: InstanceInfo
  nodes: NodeStateInfo[]
  flow_type: FlowType
  design_nodes: DesignNode[]
  design_connections: DesignConnection[]
}

export interface OperateLog {
  log_id: number
  trace_id: string
  module: string
  operate_type: string
  target_id: string
  chain_no: string
  version_id: number
  summary: string
  before_data?: string
  after_data?: string
  result: number
  error_msg: string
  operator: string
  client_ip: string
  create_time: string
}

export interface ImportResult {
  chain_no: string
  version_id: number
  version_no: number
  remapped?: Record<string, string>
}

export interface SyncBuiltinResult {
  inserted: string[]
  updated: string[]
}
