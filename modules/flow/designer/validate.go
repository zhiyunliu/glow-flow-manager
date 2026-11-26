// Package designer 流程模型的纯逻辑处理：结构校验、引擎编译校验、调试执行、节点编号重映射。
// 不访问数据库，节点类型元数据由调用方传入。
package designer

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/pkg/strutil"
)

const (
	LevelError   = "error"
	LevelWarning = "warning"

	maxNodeNameBytes = 128 // flow_node_definition.name 为 varchar(128)
)

var (
	nodeIDPattern   = regexp.MustCompile(`^[A-Za-z0-9_\-]{1,32}$`)
	relationPattern = regexp.MustCompile(`^[A-Za-z0-9_\-.]{1,64}$`)
)

const (
	endNodeType   = "end"   // 结束节点：执行后不再走后续连接
	startNodeType = "start" // 工作流开始节点
)

// Input 校验输入
type Input struct {
	FlowType    string
	Nodes       []models.DesignNode
	Connections []models.DesignConnection
	NodeTypes   map[string]models.NodeTypeInfo
	// ForPublish 发布级校验：空流程、必填配置缺失、使用已禁用的节点类型视为错误(保存草稿时为警告)
	ForPublish bool
}

type validator struct {
	in     Input
	result *models.ValidationResult
	nodes  map[string]models.DesignNode
}

// Validate 校验流程模型结构
func Validate(in Input) *models.ValidationResult {
	v := &validator{
		in:     in,
		result: &models.ValidationResult{Errors: []models.ValidationIssue{}, Warnings: []models.ValidationIssue{}},
		nodes:  make(map[string]models.DesignNode, len(in.Nodes)),
	}
	v.run()
	v.result.Valid = len(v.result.Errors) == 0
	return v.result
}

func (v *validator) add(level, code, nodeID, fromID, toID, format string, args ...any) {
	issue := models.ValidationIssue{Level: level, Code: code, Message: fmt.Sprintf(format, args...), NodeID: nodeID, FromID: fromID, ToID: toID}
	if level == LevelError {
		v.result.Errors = append(v.result.Errors, issue)
	} else {
		v.result.Warnings = append(v.result.Warnings, issue)
	}
}

// strictLevel 发布时为错误，保存草稿时为警告
func (v *validator) strictLevel() string {
	if v.in.ForPublish {
		return LevelError
	}
	return LevelWarning
}

func (v *validator) run() {
	if !enums.IsValidFlowType(v.in.FlowType) {
		v.add(LevelError, "flow_type_invalid", "", "", "", "流程类型 %q 无效", v.in.FlowType)
		return
	}
	if len(v.in.Nodes) == 0 {
		v.add(v.strictLevel(), "empty", "", "", "", "流程中没有节点")
		return
	}
	v.checkNodes()
	incoming, outgoing := v.checkConnections()
	referenced := v.checkReferences()
	if v.in.FlowType == enums.FlowTypeRuleChain {
		v.checkAcyclic()
	}
	v.checkStartNodes(incoming, outgoing, referenced)
	if v.in.FlowType == enums.FlowTypeWorkflow {
		v.checkWorkflowEnd()
	}
}

func (v *validator) checkNodes() {
	for _, node := range v.in.Nodes {
		if !nodeIDPattern.MatchString(node.ID) {
			v.add(LevelError, "node_id_invalid", node.ID, "", "", "节点编号 %q 无效：只能包含字母、数字、下划线和中划线，长度 1-32", node.ID)
			continue
		}
		if _, dup := v.nodes[node.ID]; dup {
			v.add(LevelError, "node_id_duplicate", node.ID, "", "", "节点编号 %s 重复", node.ID)
			continue
		}
		v.nodes[node.ID] = node

		label := nodeLabel(node)
		if strings.TrimSpace(node.Name) == "" {
			v.add(LevelError, "node_name_empty", node.ID, "", "", "节点 %s 名称为空", node.ID)
		} else if strutil.DBLen(node.Name) > maxNodeNameBytes {
			v.add(LevelError, "node_name_too_long", node.ID, "", "", "节点 %s 名称过长：最多 %d 个字符(中文按 2 个计算)", label, maxNodeNameBytes)
		}
		if strings.TrimSpace(node.Type) == "" {
			v.add(LevelError, "node_type_empty", node.ID, "", "", "节点 %s 未指定类型", label)
			continue
		}
		meta, ok := v.in.NodeTypes[node.Type]
		if !ok {
			v.add(LevelError, "node_type_unregistered", node.ID, "", "", "节点 %s 的类型 %s 未注册", label, node.Type)
			continue
		}
		if meta.Status != enums.StatusEnabled {
			v.add(v.strictLevel(), "node_type_disabled", node.ID, "", "", "节点 %s 的类型 %s 已禁用", label, node.Type)
		}
		if !containsString(meta.FlowModes, v.in.FlowType) {
			v.add(LevelError, "node_type_mode", node.ID, "", "", "节点 %s 的类型 %s 不能用于%s", label, node.Type, flowTypeLabel(v.in.FlowType))
		}
		v.checkRequiredConfig(node, meta)
	}
}

func (v *validator) checkRequiredConfig(node models.DesignNode, meta models.NodeTypeInfo) {
	if meta.ConfigSchema == nil {
		return
	}
	for _, field := range meta.ConfigSchema.Fields {
		if !field.Required || !fieldVisible(field, node.ExtParams) {
			continue
		}
		if isEmptyValue(node.ExtParams[field.Key]) {
			v.add(v.strictLevel(), "config_required", node.ID, "", "", "节点 %s 缺少必填配置：%s", nodeLabel(node), fieldLabel(field))
		}
	}
}

func (v *validator) checkConnections() (incoming, outgoing map[string]int) {
	incoming = make(map[string]int)
	outgoing = make(map[string]int)
	seen := make(map[string]struct{})
	for _, conn := range v.in.Connections {
		relation := NormalizeRelation(conn.Type)
		from, fromOK := v.nodes[conn.FromID]
		to, toOK := v.nodes[conn.ToID]
		if !fromOK || !toOK {
			v.add(LevelError, "conn_node_missing", "", conn.FromID, conn.ToID, "连接 %s → %s 引用了不存在的节点", conn.FromID, conn.ToID)
			continue
		}
		if !relationPattern.MatchString(relation) {
			v.add(LevelError, "conn_type_invalid", "", conn.FromID, conn.ToID, "连接 %s → %s 的类型 %q 无效", nodeLabel(from), nodeLabel(to), conn.Type)
			continue
		}
		key := conn.FromID + "\x00" + conn.ToID + "\x00" + relation
		if _, dup := seen[key]; dup {
			v.add(LevelError, "conn_duplicate", "", conn.FromID, conn.ToID, "连接 %s → %s(%s) 重复", nodeLabel(from), nodeLabel(to), relation)
			continue
		}
		seen[key] = struct{}{}
		if conn.FromID == conn.ToID && v.in.FlowType == enums.FlowTypeWorkflow {
			v.add(LevelWarning, "conn_self_loop", conn.FromID, conn.FromID, conn.ToID, "节点 %s 连接到自身", nodeLabel(from))
		}
		incoming[conn.ToID]++
		outgoing[conn.FromID]++

		if from.Type == endNodeType {
			v.add(LevelWarning, "end_outgoing", from.ID, conn.FromID, conn.ToID, "结束节点 %s 的后续连接不会执行", nodeLabel(from))
		}
		if meta, ok := v.in.NodeTypes[from.Type]; ok {
			allowed := AllowedRelations(from, meta)
			if len(allowed) > 0 && !containsString(allowed, relation) {
				v.add(LevelWarning, "conn_type_unknown", "", conn.FromID, conn.ToID,
					"连接 %s → %s 的类型 %s 不在节点输出类型 [%s] 中，执行时不会走该连接(未匹配时走 default)",
					nodeLabel(from), nodeLabel(to), relation, strings.Join(allowed, ", "))
			}
		}
	}
	return incoming, outgoing
}

// checkReferences 校验 nodeRef/nodeRefs 字段引用的节点，返回被引用的节点集合
func (v *validator) checkReferences() map[string]struct{} {
	referenced := make(map[string]struct{})
	for _, node := range v.in.Nodes {
		meta, ok := v.in.NodeTypes[node.Type]
		if !ok || meta.ConfigSchema == nil {
			continue
		}
		for _, field := range meta.ConfigSchema.Fields {
			if field.Type != models.FieldNodeRef && field.Type != models.FieldNodeRefs {
				continue
			}
			for _, refID := range NodeRefValues(node.ExtParams[field.Key]) {
				switch {
				case refID == node.ID:
					v.add(LevelError, "node_ref_self", node.ID, "", "", "节点 %s 的配置 %s 引用了自身", nodeLabel(node), fieldLabel(field))
				case !v.hasNode(refID):
					v.add(LevelError, "node_ref_missing", node.ID, "", "", "节点 %s 的配置 %s 引用了不存在的节点 %s", nodeLabel(node), fieldLabel(field), refID)
				default:
					referenced[refID] = struct{}{}
				}
			}
		}
	}
	return referenced
}

func (v *validator) hasNode(id string) bool {
	_, ok := v.nodes[id]
	return ok
}

// checkAcyclic 规则链必须是有向无环图：Kahn 拓扑排序后仍有剩余节点则存在环
func (v *validator) checkAcyclic() {
	indegree := make(map[string]int, len(v.nodes))
	adjacency := make(map[string][]string)
	for id := range v.nodes {
		indegree[id] = 0
	}
	edgeSeen := make(map[string]struct{})
	for _, conn := range v.in.Connections {
		if !v.hasNode(conn.FromID) || !v.hasNode(conn.ToID) {
			continue
		}
		key := conn.FromID + "\x00" + conn.ToID
		if _, ok := edgeSeen[key]; ok {
			continue
		}
		edgeSeen[key] = struct{}{}
		adjacency[conn.FromID] = append(adjacency[conn.FromID], conn.ToID)
		indegree[conn.ToID]++
	}
	queue := make([]string, 0)
	for id, degree := range indegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	visited := 0
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		visited++
		for _, next := range adjacency[id] {
			indegree[next]--
			if indegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	if visited == len(indegree) {
		return
	}
	cycle := make([]string, 0)
	for id, degree := range indegree {
		if degree > 0 {
			cycle = append(cycle, id)
		}
	}
	sort.Strings(cycle)
	labels := make([]string, 0, len(cycle))
	for _, id := range cycle {
		labels = append(labels, nodeLabel(v.nodes[id]))
	}
	v.add(LevelError, "cycle", cycle[0], "", "", "规则链必须是有向无环图，以下节点存在循环连接：%s", strings.Join(labels, "、"))
	for _, id := range cycle[1:] {
		v.add(LevelError, "cycle", id, "", "", "节点 %s 处于循环连接中", nodeLabel(v.nodes[id]))
	}
}

// checkStartNodes 引擎以没有入边且未被引用的节点作为起始节点；
// 工作流中显式的开始节点(start)允许被回退连线指向，同样视为起始节点
func (v *validator) checkStartNodes(incoming, outgoing map[string]int, referenced map[string]struct{}) {
	starts := 0
	counted := make(map[string]struct{}, len(v.nodes))
	for _, node := range v.in.Nodes {
		// 只统计通过编号校验的节点，重复编号只统计一次
		if _, done := counted[node.ID]; done || !v.hasNode(node.ID) {
			continue
		}
		counted[node.ID] = struct{}{}
		if v.in.FlowType == enums.FlowTypeWorkflow && node.Type == startNodeType {
			starts++
			continue
		}
		if incoming[node.ID] > 0 {
			continue
		}
		if _, ok := referenced[node.ID]; ok {
			continue
		}
		starts++
		if outgoing[node.ID] == 0 && len(v.nodes) > 1 {
			v.add(LevelWarning, "isolated", node.ID, "", "", "节点 %s 没有任何连接，将作为独立的起始节点执行", nodeLabel(node))
		}
	}
	if starts == 0 && len(v.nodes) > 0 {
		v.add(LevelError, "no_start", "", "", "", "没有起始节点：每个节点都有入边或被其他节点引用")
	}
}

func (v *validator) checkWorkflowEnd() {
	for _, node := range v.in.Nodes {
		if node.Type == endNodeType {
			return
		}
	}
	v.add(LevelWarning, "no_end", "", "", "", "工作流中没有结束节点")
}

// NormalizeRelation 空连接类型视为 default
func NormalizeRelation(relation string) string {
	relation = strings.TrimSpace(relation)
	if relation == "" {
		return enums.RelationDefault
	}
	return relation
}

// AllowedRelations 节点可输出的连接类型：注册的输出类型、条件分支的 then 值以及 default
func AllowedRelations(node models.DesignNode, meta models.NodeTypeInfo) []string {
	if len(meta.RelationTypes) == 0 && !hasCasesField(meta) {
		return nil
	}
	allowed := append([]string{}, meta.RelationTypes...)
	if hasCasesField(meta) {
		allowed = append(allowed, CaseRelations(node.ExtParams["cases"])...)
	}
	allowed = append(allowed, enums.RelationDefault)
	return uniqueStrings(allowed)
}

func hasCasesField(meta models.NodeTypeInfo) bool {
	if meta.ConfigSchema == nil {
		return false
	}
	for _, field := range meta.ConfigSchema.Fields {
		if field.Type == models.FieldCases {
			return true
		}
	}
	return false
}

// CaseRelations 读取条件分支配置中的 then 值
func CaseRelations(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		if text, isText := raw.(string); isText && strings.TrimSpace(text) != "" {
			var parsed []any
			if err := jsonUnmarshalString(text, &parsed); err == nil {
				items = parsed
			}
		}
	}
	relations := make([]string, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if then, ok := entry["then"].(string); ok && strings.TrimSpace(then) != "" {
			relations = append(relations, strings.TrimSpace(then))
		}
	}
	return relations
}

// NodeRefValues 读取 nodeRef/nodeRefs 字段的节点编号，支持字符串、逗号分隔字符串、数组
func NodeRefValues(raw any) []string {
	values := make([]string, 0)
	switch val := raw.(type) {
	case string:
		for _, part := range strings.Split(val, ",") {
			if part = strings.TrimSpace(part); part != "" {
				values = append(values, part)
			}
		}
	case []any:
		for _, item := range val {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				values = append(values, strings.TrimSpace(text))
			}
		}
	case []string:
		for _, text := range val {
			if strings.TrimSpace(text) != "" {
				values = append(values, strings.TrimSpace(text))
			}
		}
	}
	return values
}

// fieldVisible 判断字段在当前配置下是否显示(show_when 条件)
func fieldVisible(field models.SchemaField, ext map[string]any) bool {
	for key, want := range field.ShowWhen {
		got, ok := ext[key]
		if !ok || isEmptyValue(got) {
			// 未配置时按字段默认值处理由前端完成，这里视为条件不满足
			return false
		}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			return false
		}
	}
	return true
}

func isEmptyValue(val any) bool {
	if val == nil {
		return true
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case bool:
		return false
	}
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() == 0
	}
	return false
}

func nodeLabel(node models.DesignNode) string {
	if strings.TrimSpace(node.Name) == "" {
		return node.ID
	}
	return fmt.Sprintf("%s(%s)", node.Name, node.ID)
}

func fieldLabel(field models.SchemaField) string {
	if field.Label != "" {
		return field.Label
	}
	return field.Key
}

func flowTypeLabel(flowType string) string {
	if flowType == enums.FlowTypeWorkflow {
		return "工作流"
	}
	return "规则链"
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
