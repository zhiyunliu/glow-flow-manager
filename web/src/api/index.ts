import type {
  ChainDesign,
  ChainInfo,
  DebugResult,
  DesignConnection,
  DesignNode,
  FlowType,
  ImportResult,
  InstanceDetail,
  InstanceInfo,
  NodeTypeInfo,
  NodeTypeInstances,
  OperateLog,
  OptionItem,
  PageData,
  SyncBuiltinResult,
  ValidationResult,
  VersionInfo,
} from '@/types'
import { get, post, type RequestOptions } from './http'

// ---------- 流程定义 ----------
export const chainApi = {
  page: (params: Record<string, any>) => get<PageData<ChainInfo>>('/chain/page', params),
  detail: (chainNo: string) => get<ChainInfo>('/chain/detail', { chain_no: chainNo }),
  create: (data: { chain_no: string; name: string; flow_type: FlowType; desc: string }) => post<ChainInfo>('/chain/create', data),
  update: (data: { chain_no: string; name: string; flow_type: FlowType; desc: string }) => post<void>('/chain/update', data),
  changeStatus: (chainNo: string, status: number) => post<void>('/chain/change-status', { chain_no: chainNo, status }),
  remove: (chainNo: string) => post<void>('/chain/delete', { chain_no: chainNo }),
  copy: (data: { chain_no: string; version_id?: number; new_chain_no: string; new_name: string }) => post<ImportResult>('/chain/copy', data),
  export: (chainNo: string, versionId?: number) =>
    get<{ file_name: string; content: any }>('/chain/export', { chain_no: chainNo, version_id: versionId }),
  import: (data: { content: string; mode: string; chain_no?: string; name?: string }, opts?: RequestOptions) =>
    post<ImportResult>('/chain/import', data, opts),
}

// ---------- 流程版本与设计器 ----------
export interface ModelPayload {
  chain_no?: string
  version_id?: number
  flow_type: FlowType
  extparams?: Record<string, any>
  nodes: DesignNode[]
  connections: DesignConnection[]
}

export const versionApi = {
  list: (chainNo: string) => get<VersionInfo[]>('/version/list', { chain_no: chainNo }),
  detail: (versionId: number) => get<ChainDesign>('/version/detail', { version_id: versionId }),
  createDraft: (chainNo: string, baseVersionId?: number) =>
    post<VersionInfo>('/version/create-draft', { chain_no: chainNo, base_version_id: baseVersionId ?? 0 }),
  saveDraft: (data: ModelPayload & { version_id: number; desc: string; layout?: Record<string, any> }, opts?: RequestOptions) =>
    post<ValidationResult>('/version/save-draft', data, opts),
  validate: (data: ModelPayload) => post<ValidationResult>('/version/validate', data),
  publish: (versionId: number, desc: string, opts?: RequestOptions) =>
    post<ValidationResult>('/version/publish', { version_id: versionId, desc }, opts),
  offline: (versionId: number) => post<void>('/version/offline', { version_id: versionId }),
  online: (versionId: number) => post<void>('/version/online', { version_id: versionId }),
  remove: (versionId: number) => post<void>('/version/delete', { version_id: versionId }),
  debug: (data: ModelPayload & { input: any; msg_id?: string; timeout?: number; mock_actions?: boolean }) =>
    post<DebugResult>('/version/debug', data),
}

// ---------- 节点注册 ----------
export const nodeTypeApi = {
  list: (flowMode?: FlowType, status?: number) =>
    get<NodeTypeInfo[]>('/node-type/list', { flow_mode: flowMode, status: status === undefined ? undefined : String(status) }),
  page: (params: Record<string, any>) => get<PageData<NodeTypeInfo>>('/node-type/page', params),
  detail: (nodeType: string) => get<NodeTypeInfo>('/node-type/detail', { node_type: nodeType }),
  create: (data: Record<string, any>) => post<void>('/node-type/create', data),
  update: (data: Record<string, any>) => post<void>('/node-type/update', data),
  changeStatus: (nodeType: string, status: number) => post<void>('/node-type/change-status', { node_type: nodeType, status }),
  remove: (nodeType: string) => post<void>('/node-type/delete', { node_type: nodeType }),
  syncBuiltin: () => post<SyncBuiltinResult>('/node-type/sync-builtin'),
  instances: (nodeType: string) => get<NodeTypeInstances>('/node-type/instances', { node_type: nodeType }),
}

// ---------- 动态下拉选项 ----------
export const optionApi = {
  list: (source: string) => get<OptionItem[]>('/option/list', { source }, { silent: true }),
}

// ---------- 执行历史 ----------
export const historyApi = {
  page: (params: Record<string, any>) => get<PageData<InstanceInfo>>('/history/page', params),
  detail: (instanceId: string) => get<InstanceDetail>('/history/detail', { instance_id: instanceId }),
}

// ---------- 操作审计 ----------
export const auditApi = {
  page: (params: Record<string, any>) => get<PageData<OperateLog>>('/audit/page', params),
  detail: (logId: number) => get<OperateLog>('/audit/detail', { log_id: logId }),
}
