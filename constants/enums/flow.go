package enums

// 流程类型
const (
	FlowTypeRuleChain = "RULE_CHAIN" // 规则链：无状态、事件驱动，必须为 DAG
	FlowTypeWorkflow  = "WORKFLOW"   // 工作流：长周期，允许循环与回退
)

// IsValidFlowType 判断流程类型是否合法
func IsValidFlowType(flowType string) bool {
	return flowType == FlowTypeRuleChain || flowType == FlowTypeWorkflow
}

// 启用状态，对应 flow_chain_definition.status / flow_chain_version.status / flow_node_registry.status
const (
	StatusDisabled = 0
	StatusEnabled  = 1
)

// 版本发布状态，对应 flow_chain_version.publish_state。
// 引擎只加载 status=1 的版本，因此草稿与已下线版本的 status 均为 0，已发布版本的 status 为 1。
const (
	VersionStateDraft     = 0 // 草稿：可编辑，引擎不加载
	VersionStatePublished = 1 // 已发布：不可修改，引擎加载同一流程中 version_id 最大的已发布版本
	VersionStateOffline   = 2 // 已下线：不可修改，引擎不加载，可重新上线
)

// 导入方式
const (
	ImportModeNewChain   = "new_chain"   // 创建新流程
	ImportModeNewVersion = "new_version" // 作为已有流程的草稿版本
)
