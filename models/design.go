package models

// Layout 节点布局，字段与 glow-flow Layout 的 JSON 保持一致，引擎可直接解析
type Layout struct {
	Icon string `json:"icon,omitempty"`
	H    int    `json:"H,omitempty"`
	X    int    `json:"X,omitempty"`
	Y    int    `json:"Y,omitempty"`
	W    int    `json:"W,omitempty"`
}

// DesignNode 设计器节点，对应引擎 NodeDefinition；id 即 flow_node_definition.node_def_no
type DesignNode struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Name      string         `json:"name"`
	Layout    Layout         `json:"layout"`
	ExtParams map[string]any `json:"extparams"`
}

// DesignConnection 设计器连接，对应引擎 ConnectionDefinition
type DesignConnection struct {
	FromID    string         `json:"from_id"`
	ToID      string         `json:"to_id"`
	Type      string         `json:"type"`
	Remark    string         `json:"remark,omitempty"`
	ExtParams map[string]any `json:"extparams,omitempty"` // 设计器附加信息(如连线拐点)，引擎不使用
}

// ChainDesign 设计器加载的完整模型
type ChainDesign struct {
	Chain       DBChainInfo        `json:"chain"`
	Version     DBVersionInfo      `json:"version"`
	ExtParams   map[string]any     `json:"extparams"`
	Layout      map[string]any     `json:"layout"`
	Nodes       []DesignNode       `json:"nodes"`
	Connections []DesignConnection `json:"connections"`
	Editable    bool               `json:"editable"` // 仅草稿可编辑
}

// ValidationIssue 模型校验问题
type ValidationIssue struct {
	Level   string `json:"level"` // error / warning
	Code    string `json:"code"`
	Message string `json:"message"`
	NodeID  string `json:"node_id,omitempty"`
	FromID  string `json:"from_id,omitempty"`
	ToID    string `json:"to_id,omitempty"`
}

// ValidationResult 模型校验结果
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationIssue `json:"errors"`
	Warnings []ValidationIssue `json:"warnings"`
}

// DebugEvent 调试执行中单个节点的一次执行
type DebugEvent struct {
	Seq          int      `json:"seq"`
	NodeID       string   `json:"node_id"`
	NodeName     string   `json:"node_name"`
	NodeType     string   `json:"node_type"`
	Mocked       bool     `json:"mocked"`
	Input        any      `json:"input"`
	Output       any      `json:"output"`
	RelationType string   `json:"relation_type"`
	Relations    []string `json:"relations,omitempty"`
	End          bool     `json:"end,omitempty"`
	Error        string   `json:"error,omitempty"`
	StartOffset  int64    `json:"start_offset_ms"` // 相对调试开始的毫秒数
	DurationMs   int64    `json:"duration_ms"`
}

// DebugEdge 调试中实际经过的连接
type DebugEdge struct {
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
	Type   string `json:"type"`
}

// DebugResult 调试结果
type DebugResult struct {
	InstanceID    string            `json:"instance_id"`
	Success       bool              `json:"success"`
	Error         string            `json:"error,omitempty"`
	TimedOut      bool              `json:"timed_out"`
	DurationMs    int64             `json:"duration_ms"`
	Events        []DebugEvent      `json:"events"`
	ExecutedNodes []string          `json:"executed_nodes"`
	Edges         []DebugEdge       `json:"edges"`
	Validation    *ValidationResult `json:"validation,omitempty"` // 校验未通过时返回
}
