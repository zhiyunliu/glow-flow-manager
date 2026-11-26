package models

// PageChainParams 流程定义分页查询
type PageChainParams struct {
	PageParams
	ChainNo  string `form:"chain_no"`
	Name     string `form:"name"`
	FlowType string `form:"flow_type"`
	Status   string `form:"status"` // 空为全部，0 禁用，1 启用
}

// ChainNoParams 按流程编号操作
type ChainNoParams struct {
	ChainNo string `form:"chain_no" json:"chain_no" binding:"required"`
}

// CreateChainParams 创建流程定义，同时创建 v1 草稿
type CreateChainParams struct {
	ChainNo  string `json:"chain_no" binding:"required"`
	Name     string `json:"name" binding:"required"`
	FlowType string `json:"flow_type" binding:"required"`
	Desc     string `json:"desc"`
}

// UpdateChainParams 修改流程定义基本信息
type UpdateChainParams struct {
	ChainNo  string `json:"chain_no" binding:"required"`
	Name     string `json:"name" binding:"required"`
	FlowType string `json:"flow_type" binding:"required"`
	Desc     string `json:"desc"`
}

// ChangeChainStatusParams 启用/禁用流程
type ChangeChainStatusParams struct {
	ChainNo string `json:"chain_no" binding:"required"`
	Status  int    `json:"status" binding:"oneof=0 1"`
}

// CopyChainParams 复制流程
type CopyChainParams struct {
	ChainNo    string `json:"chain_no" binding:"required"`
	VersionID  int64  `json:"version_id"` // 复制的版本，默认为生效版本，没有生效版本时为最新版本
	NewChainNo string `json:"new_chain_no" binding:"required"`
	NewName    string `json:"new_name" binding:"required"`
}

// ExportChainParams 导出流程
type ExportChainParams struct {
	ChainNo   string `form:"chain_no" binding:"required"`
	VersionID int64  `form:"version_id"` // 默认为生效版本，没有生效版本时为最新版本
}

// ImportChainParams 导入流程
type ImportChainParams struct {
	Content string `json:"content" binding:"required"` // 导出的 JSON 内容，也支持 glow-flow ChainDefinition JSON
	Mode    string `json:"mode" binding:"required"`    // new_chain 创建新流程；new_version 作为已有流程的草稿
	ChainNo string `json:"chain_no"`                   // 目标流程编号，为空时使用导入内容中的编号
	Name    string `json:"name"`                       // 新流程名称，为空时使用导入内容中的名称
}

// ChainExport 导出文件格式
type ChainExport struct {
	Format     string         `json:"format"`
	ExportTime string         `json:"export_time"`
	FlowType   string         `json:"flow_type"`
	Desc       string         `json:"desc,omitempty"`
	VersionNo  int            `json:"version_no"`
	Definition map[string]any `json:"definition"` // glow-flow ChainDefinition
	Edges      []EdgeExtra    `json:"edges,omitempty"`
}

// EdgeExtra 连接在设计器中的附加信息(引擎不使用)
type EdgeExtra struct {
	FromID    string         `json:"from_id"`
	ToID      string         `json:"to_id"`
	Type      string         `json:"type"`
	ExtParams map[string]any `json:"extparams,omitempty"`
}

// ImportResult 导入结果
type ImportResult struct {
	ChainNo   string            `json:"chain_no"`
	VersionID int64             `json:"version_id"`
	VersionNo int               `json:"version_no"`
	Remapped  map[string]string `json:"remapped,omitempty"` // 与其他流程冲突而重新生成的节点编号
}
