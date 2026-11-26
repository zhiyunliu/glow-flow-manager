package models

// DBNodeType flow_node_registry 记录
type DBNodeType struct {
	NodeType      string `json:"node_type"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	FlowModes     string `json:"flow_modes"`
	Source        string `json:"source"`
	ServiceName   string `json:"service_name"`
	ServicePath   string `json:"service_path"`
	Protocol      string `json:"protocol"`
	RelationTypes string `json:"relation_types"`
	ConfigSchema  string `json:"config_schema"`
	DefaultConfig string `json:"default_config"`
	Icon          string `json:"icon"`
	Color         string `json:"color"`
	SortNo        int    `json:"sort_no"`
	Status        int    `json:"status"`
	Desc          string `json:"desc"`
	CreateTime    string `json:"create_time"`
	CreateBy      string `json:"create_by"`
	Operator      string `json:"operator"`
	UpdateTime    string `json:"update_time"`
	UsageCount    int    `json:"usage_count"`
}
