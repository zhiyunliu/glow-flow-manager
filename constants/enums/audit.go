package enums

// 审计模块
const (
	AuditModuleChain    = "chain"     // 流程定义
	AuditModuleVersion  = "version"   // 流程版本
	AuditModuleNodeType = "node_type" // 节点注册
)

// 审计操作类型
const (
	AuditOpCreate    = "create"
	AuditOpUpdate    = "update"
	AuditOpDelete    = "delete"
	AuditOpEnable    = "enable"
	AuditOpDisable   = "disable"
	AuditOpCopy      = "copy"
	AuditOpImport    = "import"
	AuditOpSaveDraft = "save_draft"
	AuditOpPublish   = "publish"
	AuditOpOffline   = "offline"
	AuditOpOnline    = "online"
	AuditOpSync      = "sync"
)

// 审计结果
const (
	AuditResultFail    = 0
	AuditResultSuccess = 1
)
