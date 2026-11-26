package models

// PageInstanceParams 执行历史分页查询
type PageInstanceParams struct {
	PageParams
	ChainNo    string `form:"chain_no"`
	Version    string `form:"version"` // 引擎版本，即 version_id
	Status     string `form:"status"`
	InstanceID string `form:"instance_id"`
	StartTime  string `form:"start_time"` // yyyy-MM-dd HH:mm:ss
	EndTime    string `form:"end_time"`
}

// InstanceIDParams 按实例查询
type InstanceIDParams struct {
	InstanceID string `form:"instance_id" binding:"required"`
}

// DBInstanceState flow_instance_state 记录
type DBInstanceState struct {
	InstanceID string `json:"instance_id"`
	ChainID    string `json:"chain_id"`
	Version    string `json:"version"`
	Status     string `json:"status"`
	StateData  string `json:"state_data"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
	ChainName  string `json:"chain_name"`
	VersionNo  int    `json:"version_no"`
}

// DBNodeState flow_node_state 记录
type DBNodeState struct {
	InstanceID string `json:"instance_id"`
	NodeID     string `json:"node_id"`
	ChainID    string `json:"chain_id"`
	Version    string `json:"version"`
	Status     string `json:"status"`
	Attempts   int    `json:"attempts"`
	StateData  string `json:"state_data"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

// InstanceInfo 执行实例
type InstanceInfo struct {
	InstanceID     string         `json:"instance_id"`
	ChainNo        string         `json:"chain_no"`
	ChainName      string         `json:"chain_name"`
	Version        string         `json:"version"`
	VersionNo      int            `json:"version_no"`
	Status         string         `json:"status"`
	CurrentNodeIDs []string       `json:"current_node_ids,omitempty"`
	Variables      map[string]any `json:"variables,omitempty"`
	Error          string         `json:"error,omitempty"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
	FinishedAt     string         `json:"finished_at,omitempty"`
	DurationMs     int64          `json:"duration_ms"`
}

// NodeStateInfo 节点执行状态
type NodeStateInfo struct {
	NodeID     string         `json:"node_id"`
	NodeName   string         `json:"node_name"`
	NodeType   string         `json:"node_type"`
	Status     string         `json:"status"`
	Attempts   int            `json:"attempts"`
	Input      any            `json:"input,omitempty"`
	Output     any            `json:"output,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	Error      string         `json:"error,omitempty"`
	StartedAt  string         `json:"started_at,omitempty"`
	FinishedAt string         `json:"finished_at,omitempty"`
	DurationMs int64          `json:"duration_ms"`
}

// InstanceDetail 执行实例详情
type InstanceDetail struct {
	Instance    InstanceInfo       `json:"instance"`
	Nodes       []NodeStateInfo    `json:"nodes"`
	FlowType    string             `json:"flow_type"`
	DesignNodes []DesignNode       `json:"design_nodes"`
	DesignConns []DesignConnection `json:"design_connections"`
}
