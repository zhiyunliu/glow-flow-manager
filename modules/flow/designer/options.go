package designer

import (
	"fmt"

	"github.com/zhiyunliu/manager-glow/models"
)

// OptionLookup 返回动态下拉数据源的可选值集合
type OptionLookup func(source string) (map[string]struct{}, error)

// OptionSourceIssues 检查引用动态下拉数据源(SchemaField.OptionsSource)的配置值是否在可选列表中。
// 管理后台的数据源与引擎宿主的配置可能不一致，因此只提示警告，不阻止发布。
func OptionSourceIssues(nodes []models.DesignNode, types map[string]models.NodeTypeInfo, lookup OptionLookup) []models.ValidationIssue {
	issues := make([]models.ValidationIssue, 0)
	cache := make(map[string]map[string]struct{})
	failed := make(map[string]bool)
	for _, node := range nodes {
		meta, ok := types[node.Type]
		if !ok || meta.ConfigSchema == nil {
			continue
		}
		for _, field := range meta.ConfigSchema.Fields {
			source := field.OptionsSource
			if source == "" || !fieldVisible(field, node.ExtParams) {
				continue
			}
			val := node.ExtParams[field.Key]
			if isEmptyValue(val) || failed[source] {
				continue
			}
			values, ok := cache[source]
			if !ok {
				var err error
				if values, err = lookup(source); err != nil {
					failed[source] = true
					issues = append(issues, models.ValidationIssue{Level: LevelWarning, Code: "option_source_unavailable",
						Message: fmt.Sprintf("无法获取%s的可选项，未校验其取值：%v", fieldLabel(field), err)})
					continue
				}
				cache[source] = values
			}
			if _, ok := values[fmt.Sprint(val)]; !ok {
				issues = append(issues, models.ValidationIssue{Level: LevelWarning, Code: "option_not_found", NodeID: node.ID,
					Message: fmt.Sprintf("节点 %s 的%s「%v」不在可选列表中，请确认引擎宿主已配置", nodeLabel(node), fieldLabel(field), val)})
			}
		}
	}
	return issues
}
