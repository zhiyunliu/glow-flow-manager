package models

import "encoding/json"

// VersionIDParams 按版本操作
type VersionIDParams struct {
	VersionID int64 `form:"version_id" json:"version_id" binding:"required"`
}

// ListVersionParams 查询流程的版本列表
type ListVersionParams struct {
	ChainNo string `form:"chain_no" binding:"required"`
}

// CreateDraftParams 创建草稿版本
type CreateDraftParams struct {
	ChainNo       string `json:"chain_no" binding:"required"`
	BaseVersionID int64  `json:"base_version_id"` // 复制来源版本，默认为生效版本，没有时为最新版本；-1 表示空白草稿
}

// SaveDraftParams 保存草稿(设计器)
type SaveDraftParams struct {
	VersionID   int64              `json:"version_id" binding:"required"`
	FlowType    string             `json:"flow_type"` // 为空时保持不变；流程从未发布过时允许切换
	Desc        string             `json:"desc"`
	ExtParams   map[string]any     `json:"extparams"`
	Layout      map[string]any     `json:"layout"`
	Nodes       []DesignNode       `json:"nodes"`
	Connections []DesignConnection `json:"connections"`
}

// PublishParams 发布版本
type PublishParams struct {
	VersionID int64  `json:"version_id" binding:"required"`
	Desc      string `json:"desc"` // 发布说明，为空时保持原描述
}

// ValidateParams 模型校验
type ValidateParams struct {
	ChainNo     string             `json:"chain_no"`
	VersionID   int64              `json:"version_id"`
	FlowType    string             `json:"flow_type" binding:"required"`
	Nodes       []DesignNode       `json:"nodes"`
	Connections []DesignConnection `json:"connections"`
}

// DebugParams 调试预览
type DebugParams struct {
	ChainNo     string             `json:"chain_no"`
	FlowType    string             `json:"flow_type" binding:"required"`
	ExtParams   map[string]any     `json:"extparams"`
	Nodes       []DesignNode       `json:"nodes"`
	Connections []DesignConnection `json:"connections"`
	Input       json.RawMessage    `json:"input"`        // 输入消息数据(JSON)
	MsgID       string             `json:"msg_id"`       // 消息ID，默认自动生成
	Timeout     int                `json:"timeout"`      // 超时时间(秒)
	MockActions *bool              `json:"mock_actions"` // 动作/集成类节点是否模拟执行，默认 true
}
