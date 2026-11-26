package models

// DBVersionInfo flow_chain_version 记录
type DBVersionInfo struct {
	VersionID     int64  `json:"version_id"`
	ChainNo       string `json:"chain_no"`
	VersionNo     int    `json:"version_no"`
	Name          string `json:"name"`
	Status        int    `json:"status"`
	PublishState  int    `json:"publish_state"`
	ExtParams     string `json:"extparams"`
	Layout        string `json:"layout"`
	Desc          string `json:"desc"`
	BaseVersionID int64  `json:"base_version_id"`
	PublishTime   string `json:"publish_time"`
	PublishBy     string `json:"publish_by"`
	CreateTime    string `json:"create_time"`
	CreateBy      string `json:"create_by"`
	Operator      string `json:"operator"`
	UpdateTime    string `json:"update_time"`
	NodeCount     int    `json:"node_count"`
	ConnCount     int    `json:"conn_count"`
	IsEffective   bool   `json:"is_effective"` // 是否为引擎当前加载的版本
}

// DBChainNode flow_chain_node 关联 flow_node_definition
type DBChainNode struct {
	NodeID         int64  `json:"node_id"`
	VersionID      int64  `json:"version_id"`
	ChainNo        string `json:"chain_no"`
	NodeDefNo      string `json:"node_def_no"`
	NodeName       string `json:"node_name"`
	DefName        string `json:"def_name"`
	Type           string `json:"type"`
	ExtParams      string `json:"extparams"`
	Layout         string `json:"layout"`
	DefExtParams   string `json:"def_extparams"`
	DefLayout      string `json:"def_layout"`
	DefChainNo     string `json:"def_chain_no"`
	DefDescription string `json:"def_desc"`
}

// DBChainConnection flow_chain_connection 记录
type DBChainConnection struct {
	ConnID    int64  `json:"conn_id"`
	VersionID int64  `json:"version_id"`
	ChainNo   string `json:"chain_no"`
	FromID    string `json:"from_id"`
	ToID      string `json:"to_id"`
	Type      string `json:"type"`
	Remark    string `json:"remark"`
	ExtParams string `json:"extparams"`
}

// DBNodeDefinition flow_node_definition 记录
type DBNodeDefinition struct {
	NodeDefNo string `json:"node_def_no"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	ChainNo   string `json:"chain_no"`
}

// DBNodeUsage 节点定义在其他版本中的引用次数
type DBNodeUsage struct {
	NodeDefNo string `json:"node_def_no"`
	Total     int    `json:"total"`
}
