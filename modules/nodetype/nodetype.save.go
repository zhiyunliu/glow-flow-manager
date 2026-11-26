package nodetype

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/audit"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
)

var (
	nodeTypePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_\-.]{0,63}$`)
	relationPattern = regexp.MustCompile(`^[A-Za-z0-9_\-.]{1,64}$`)
	fieldKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_\-.]{0,63}$`)
	colorPattern    = regexp.MustCompile(`^(#[0-9A-Fa-f]{3,8})?$`)

	validFieldTypes = map[string]bool{
		models.FieldString: true, models.FieldText: true, models.FieldNumber: true, models.FieldBool: true,
		models.FieldSelect: true, models.FieldCode: true, models.FieldKV: true, models.FieldList: true,
		models.FieldCases: true, models.FieldNodeRef: true, models.FieldNodeRefs: true, models.FieldJSON: true,
	}
)

func (svc *NodeTyperImpl) Create(ctx context.Context, op operator.Operator, params models.SaveNodeTypeParams) (err error) {
	entry := audit.Entry{Module: enums.AuditModuleNodeType, OperateType: enums.AuditOpCreate, TargetID: params.NodeType, Summary: "注册节点类型 " + params.NodeType, After: params}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	info, err := buildInfo(params)
	if err != nil {
		return err
	}
	for _, builtin := range builtinNodeTypes {
		if builtin.NodeType == info.NodeType {
			return bizerr.Conflict(subcode.NodeTypeExists, "%s 是 glow-flow 内置节点类型，请使用“同步内置节点”", info.NodeType)
		}
	}
	existed, err := dbGetNodeType(ctx, info.NodeType)
	if err != nil {
		return err
	}
	if existed.NodeType != "" {
		return bizerr.Conflict(subcode.NodeTypeExists, "节点类型 %s 已存在", info.NodeType)
	}
	info.Source = enums.NodeSourceService
	info.Status = enums.StatusEnabled
	if info.Color == "" {
		info.Color = categoryColors[info.Category]
	}
	dbParams, err := toDBParams(info, op.ID)
	if err != nil {
		return err
	}
	if err = dbInsertNodeType(ctx, dbParams); err != nil {
		svc.logger.Errorf("insert node type failed, node_type:%s, err:%+v", info.NodeType, err)
		return err
	}
	return nil
}

func (svc *NodeTyperImpl) Update(ctx context.Context, op operator.Operator, params models.SaveNodeTypeParams) (err error) {
	entry := audit.Entry{Module: enums.AuditModuleNodeType, OperateType: enums.AuditOpUpdate, TargetID: params.NodeType, Summary: "修改节点类型 " + params.NodeType, After: params}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	row, err := svc.retrieve(ctx, params.NodeType)
	if err != nil {
		return err
	}
	current := toInfo(row)
	entry.Before = current

	info, err := buildInfo(params)
	if err != nil {
		return err
	}
	info.Source = current.Source
	info.Status = current.Status
	if current.Source == enums.NodeSourceBuiltin {
		// 内置节点的分类、输出类型、属性描述由引擎决定，只允许修改展示信息与适用流程类型
		info.Category = current.Category
		info.RelationTypes = current.RelationTypes
		info.ConfigSchema = current.ConfigSchema
		info.DefaultConfig = current.DefaultConfig
		info.ServiceName, info.ServicePath, info.Protocol = "", "", ""
	}
	dbParams, err := toDBParams(info, op.ID)
	if err != nil {
		return err
	}
	if err = dbUpdateNodeType(ctx, dbParams); err != nil {
		svc.logger.Errorf("update node type failed, node_type:%s, err:%+v", info.NodeType, err)
		return err
	}
	return nil
}

func (svc *NodeTyperImpl) ChangeStatus(ctx context.Context, op operator.Operator, params models.ChangeNodeTypeStatusParams) (err error) {
	opType := enums.AuditOpDisable
	summary := "禁用节点类型 "
	if params.Status == enums.StatusEnabled {
		opType, summary = enums.AuditOpEnable, "启用节点类型 "
	}
	entry := audit.Entry{Module: enums.AuditModuleNodeType, OperateType: opType, TargetID: params.NodeType, Summary: summary + params.NodeType, After: params}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	row, err := svc.retrieve(ctx, params.NodeType)
	if err != nil {
		return err
	}
	entry.Before = map[string]any{"node_type": row.NodeType, "status": row.Status}
	if err = dbUpdateNodeTypeStatus(ctx, params.NodeType, params.Status, op.ID); err != nil {
		svc.logger.Errorf("update node type status failed, node_type:%s, err:%+v", params.NodeType, err)
		return err
	}
	return nil
}

func (svc *NodeTyperImpl) Delete(ctx context.Context, op operator.Operator, nodeType string) (err error) {
	entry := audit.Entry{Module: enums.AuditModuleNodeType, OperateType: enums.AuditOpDelete, TargetID: nodeType, Summary: "删除节点类型 " + nodeType}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	row, err := svc.retrieve(ctx, nodeType)
	if err != nil {
		return err
	}
	entry.Before = toInfo(row)
	if row.Source == enums.NodeSourceBuiltin {
		return bizerr.Conflict(subcode.NodeTypeBuiltinReadonly, "内置节点类型不能删除，可禁用")
	}
	if row.UsageCount > 0 {
		return bizerr.Conflict(subcode.NodeTypeInUse, "节点类型 %s 已被 %d 个流程节点使用，不能删除", nodeType, row.UsageCount)
	}
	if err = dbDeleteNodeType(ctx, nodeType); err != nil {
		svc.logger.Errorf("delete node type failed, node_type:%s, err:%+v", nodeType, err)
		return err
	}
	return nil
}

// buildInfo 校验并转换请求参数
func buildInfo(params models.SaveNodeTypeParams) (models.NodeTypeInfo, error) {
	info := models.NodeTypeInfo{
		NodeType:      strings.TrimSpace(params.NodeType),
		Name:          strings.TrimSpace(params.Name),
		Category:      params.Category,
		FlowModes:     uniqueStrings(params.FlowModes),
		ServiceName:   strings.TrimSpace(params.ServiceName),
		ServicePath:   strings.TrimSpace(params.ServicePath),
		Protocol:      strings.TrimSpace(params.Protocol),
		RelationTypes: []string{},
		DefaultConfig: params.DefaultConfig,
		Icon:          strings.TrimSpace(params.Icon),
		Color:         strings.TrimSpace(params.Color),
		SortNo:        params.SortNo,
		Desc:          strings.TrimSpace(params.Desc),
	}
	if !nodeTypePattern.MatchString(info.NodeType) {
		return info, bizerr.BadRequest(subcode.ParamInvalid, "节点类型 %q 无效：以字母开头，只能包含字母、数字、下划线、中划线和点，最长 64 位", params.NodeType)
	}
	if info.Name == "" {
		return info, bizerr.BadRequest(subcode.ParamInvalid, "节点名称不能为空")
	}
	if !enums.IsValidNodeCategory(info.Category) {
		return info, bizerr.BadRequest(subcode.ParamInvalid, "节点分类 %q 无效", params.Category)
	}
	if len(info.FlowModes) == 0 {
		return info, bizerr.BadRequest(subcode.ParamInvalid, "至少选择一种适用的流程类型")
	}
	for _, mode := range info.FlowModes {
		if !enums.IsValidFlowType(mode) {
			return info, bizerr.BadRequest(subcode.ParamInvalid, "流程类型 %q 无效", mode)
		}
	}
	if !colorPattern.MatchString(info.Color) {
		return info, bizerr.BadRequest(subcode.ParamInvalid, "颜色 %q 无效，应为 #RRGGBB 格式", params.Color)
	}
	for _, relation := range uniqueStrings(params.RelationTypes) {
		if !relationPattern.MatchString(relation) {
			return info, bizerr.BadRequest(subcode.ParamInvalid, "输出类型 %q 无效", relation)
		}
		info.RelationTypes = append(info.RelationTypes, relation)
	}
	schema, err := parseSchema(params.ConfigSchema)
	if err != nil {
		return info, bizerr.BadRequest(subcode.ParamInvalid, "属性描述无效：%s", err.Error())
	}
	info.ConfigSchema = schema
	if info.DefaultConfig == nil {
		info.DefaultConfig = map[string]any{}
	}
	return info, nil
}

// parseSchema 解析属性描述，支持 {"fields":[...]} 或直接传字段数组
func parseSchema(raw json.RawMessage) (*models.ConfigSchema, error) {
	schema := &models.ConfigSchema{Fields: []models.SchemaField{}}
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "null" {
		return schema, nil
	}
	if strings.HasPrefix(text, "[") {
		if err := json.Unmarshal(raw, &schema.Fields); err != nil {
			return nil, err
		}
	} else if err := json.Unmarshal(raw, schema); err != nil {
		return nil, err
	}
	if schema.Fields == nil {
		schema.Fields = []models.SchemaField{}
	}
	seen := make(map[string]struct{}, len(schema.Fields))
	for i, field := range schema.Fields {
		if !fieldKeyPattern.MatchString(field.Key) {
			return nil, fmt.Errorf("第 %d 个字段的 key %q 无效", i+1, field.Key)
		}
		if _, dup := seen[field.Key]; dup {
			return nil, fmt.Errorf("字段 %s 重复", field.Key)
		}
		seen[field.Key] = struct{}{}
		if field.Type == "" {
			schema.Fields[i].Type = models.FieldString
		} else if !validFieldTypes[field.Type] {
			return nil, fmt.Errorf("字段 %s 的类型 %q 无效", field.Key, field.Type)
		}
	}
	return schema, nil
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
