package nodetype

import (
	"encoding/json"
	"strings"

	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/flow/designer"
)

// toInfo 将数据库记录转换为节点类型信息，解析 JSON 字段
func toInfo(row models.DBNodeType) models.NodeTypeInfo {
	info := models.NodeTypeInfo{
		NodeType:      row.NodeType,
		Name:          row.Name,
		Category:      row.Category,
		FlowModes:     splitModes(row.FlowModes),
		Source:        row.Source,
		ServiceName:   row.ServiceName,
		ServicePath:   row.ServicePath,
		Protocol:      row.Protocol,
		RelationTypes: []string{},
		ConfigSchema:  &models.ConfigSchema{Fields: []models.SchemaField{}},
		DefaultConfig: map[string]any{},
		Icon:          row.Icon,
		Color:         row.Color,
		SortNo:        row.SortNo,
		Status:        row.Status,
		Desc:          row.Desc,
		CreateTime:    row.CreateTime,
		CreateBy:      row.CreateBy,
		Operator:      row.Operator,
		UpdateTime:    row.UpdateTime,
		UsageCount:    row.UsageCount,
	}
	info.EngineSupported = row.Source == enums.NodeSourceBuiltin && (designer.HasEngineDescriptor(row.NodeType) || row.NodeType == "custom")
	if strings.TrimSpace(row.RelationTypes) != "" {
		_ = json.Unmarshal([]byte(row.RelationTypes), &info.RelationTypes)
	}
	if strings.TrimSpace(row.ConfigSchema) != "" {
		schema := &models.ConfigSchema{}
		if err := json.Unmarshal([]byte(row.ConfigSchema), schema); err == nil {
			if schema.Fields == nil {
				schema.Fields = []models.SchemaField{}
			}
			info.ConfigSchema = schema
		}
	}
	if strings.TrimSpace(row.DefaultConfig) != "" {
		_ = json.Unmarshal([]byte(row.DefaultConfig), &info.DefaultConfig)
		if info.DefaultConfig == nil {
			info.DefaultConfig = map[string]any{}
		}
	}
	return info
}

// toDBParams 将节点类型信息转换为 SQL 参数
func toDBParams(info models.NodeTypeInfo, operator string) (map[string]any, error) {
	relations, err := marshalOrEmpty(info.RelationTypes)
	if err != nil {
		return nil, err
	}
	schema, err := marshalOrEmpty(info.ConfigSchema)
	if err != nil {
		return nil, err
	}
	defaults, err := marshalOrEmpty(info.DefaultConfig)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"node_type":      info.NodeType,
		"name":           info.Name,
		"category":       info.Category,
		"flow_modes":     strings.Join(info.FlowModes, ","),
		"source":         info.Source,
		"service_name":   info.ServiceName,
		"service_path":   info.ServicePath,
		"protocol":       info.Protocol,
		"relation_types": relations,
		"config_schema":  schema,
		"default_config": defaults,
		"icon":           info.Icon,
		"color":          info.Color,
		"sort_no":        info.SortNo,
		"status":         info.Status,
		"desc":           info.Desc,
		"operator":       operator,
	}, nil
}

func marshalOrEmpty(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	data, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	if string(data) == "null" {
		return "", nil
	}
	return string(data), nil
}

func splitModes(text string) []string {
	modes := make([]string, 0, 2)
	for _, part := range strings.Split(text, ",") {
		if part = strings.TrimSpace(part); part != "" {
			modes = append(modes, part)
		}
	}
	return modes
}
