package models

import "encoding/json"

// ListNodeTypeParams 查询节点类型(设计器工具箱)
type ListNodeTypeParams struct {
	FlowMode string `form:"flow_mode"` // RULE_CHAIN / WORKFLOW，空为全部
	Status   string `form:"status"`    // 空为全部
}

// PageNodeTypeParams 节点类型分页查询
type PageNodeTypeParams struct {
	PageParams
	Keyword  string `form:"keyword"` // 匹配类型或名称
	Category string `form:"category"`
	Source   string `form:"source"`
	FlowMode string `form:"flow_mode"`
	Status   string `form:"status"`
}

// NodeTypeParams 按节点类型操作
type NodeTypeParams struct {
	NodeType string `form:"node_type" json:"node_type" binding:"required"`
}

// SaveNodeTypeParams 创建/修改节点类型
type SaveNodeTypeParams struct {
	NodeType      string          `json:"node_type" binding:"required"`
	Name          string          `json:"name" binding:"required"`
	Category      string          `json:"category" binding:"required"`
	FlowModes     []string        `json:"flow_modes" binding:"required,min=1"`
	ServiceName   string          `json:"service_name"`
	ServicePath   string          `json:"service_path"`
	Protocol      string          `json:"protocol"`
	RelationTypes []string        `json:"relation_types"`
	ConfigSchema  json.RawMessage `json:"config_schema"`
	DefaultConfig map[string]any  `json:"default_config"`
	Icon          string          `json:"icon"`
	Color         string          `json:"color"`
	SortNo        int             `json:"sort_no"`
	Desc          string          `json:"desc"`
}

// ChangeNodeTypeStatusParams 启用/禁用节点类型
type ChangeNodeTypeStatusParams struct {
	NodeType string `json:"node_type" binding:"required"`
	Status   int    `json:"status" binding:"oneof=0 1"`
}

// NodeTypeInfo 节点类型(JSON 字段已解析)
type NodeTypeInfo struct {
	NodeType      string         `json:"node_type"`
	Name          string         `json:"name"`
	Category      string         `json:"category"`
	FlowModes     []string       `json:"flow_modes"`
	Source        string         `json:"source"`
	ServiceName   string         `json:"service_name"`
	ServicePath   string         `json:"service_path"`
	Protocol      string         `json:"protocol"`
	RelationTypes []string       `json:"relation_types"`
	ConfigSchema  *ConfigSchema  `json:"config_schema"`
	DefaultConfig map[string]any `json:"default_config"`
	Icon          string         `json:"icon"`
	Color         string         `json:"color"`
	SortNo        int            `json:"sort_no"`
	Status        int            `json:"status"`
	Desc          string         `json:"desc"`
	CreateTime    string         `json:"create_time"`
	CreateBy      string         `json:"create_by"`
	Operator      string         `json:"operator"`
	UpdateTime    string         `json:"update_time"`
	UsageCount    int            `json:"usage_count"` // 被流程节点引用的次数
	// EngineSupported 内置节点在 glow-flow 引擎中是否已实现；独立服务节点需要引擎宿主注册，固定为 false
	EngineSupported bool `json:"engine_supported"`
}

// ConfigSchema 节点属性面板描述
type ConfigSchema struct {
	Fields []SchemaField `json:"fields"`
}

// 属性字段类型
const (
	FieldString   = "string"   // 单行文本
	FieldText     = "text"     // 多行文本
	FieldNumber   = "number"   // 数字
	FieldBool     = "bool"     // 开关
	FieldSelect   = "select"   // 下拉选择
	FieldCode     = "code"     // 代码/表达式，lang 指定语言
	FieldKV       = "kv"       // 键值对对象
	FieldList     = "list"     // 字符串数组
	FieldCases    = "cases"    // 条件分支 [{"case":"", "then":""}]
	FieldNodeRef  = "nodeRef"  // 引用流程中的节点ID
	FieldNodeRefs = "nodeRefs" // 引用流程中的多个节点ID
	FieldJSON     = "json"     // 任意 JSON
)

// SchemaField 属性字段
type SchemaField struct {
	Key         string         `json:"key"`
	Label       string         `json:"label"`
	Type        string         `json:"type"`
	Required    bool           `json:"required,omitempty"`
	Default     any            `json:"default,omitempty"`
	Placeholder string         `json:"placeholder,omitempty"`
	Help        string         `json:"help,omitempty"`
	Lang        string         `json:"lang,omitempty"`
	Options     []SchemaOption `json:"options,omitempty"`
	// OptionsSource 动态下拉数据源名称(select 类型)，选项通过 /option/list?source= 获取，见 modules/option
	OptionsSource string         `json:"options_source,omitempty"`
	Group         string         `json:"group,omitempty"`
	ShowWhen      map[string]any `json:"show_when,omitempty"` // 其他字段取值匹配时才显示，如 {"target":"group"}
}

// SchemaOption 下拉选项
type SchemaOption struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

// RegistryInstance 注册中心中的服务实例
type RegistryInstance struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Healthy   bool              `json:"healthy"`
	Weight    int64             `json:"weight"`
	Endpoints []string          `json:"endpoints"`
	Metadata  map[string]string `json:"metadata"`
}

// NodeTypeInstances 节点服务实例
type NodeTypeInstances struct {
	NodeType    string             `json:"node_type"`
	ServiceName string             `json:"service_name"`
	Registry    string             `json:"registry"`
	Instances   []RegistryInstance `json:"instances"`
}

// SyncBuiltinResult 同步内置节点结果
type SyncBuiltinResult struct {
	Inserted []string `json:"inserted"`
	Updated  []string `json:"updated"`
}
