package designer

import (
	"fmt"

	glowflow "github.com/zhiyunliu/glow-flow"
	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/models"

	// 注册 glow-flow 内置节点，编译校验与调试使用引擎真实的节点工厂
	_ "github.com/zhiyunliu/glow-flow/contrib/nodedescriptor/action"
	_ "github.com/zhiyunliu/glow-flow/contrib/nodedescriptor/filter"
	_ "github.com/zhiyunliu/glow-flow/contrib/nodedescriptor/logical"
	_ "github.com/zhiyunliu/glow-flow/contrib/nodedescriptor/transform"
)

// stubOnlyTypes 依赖引擎宿主进程注册信息的内置节点，管理后台中无法真实构建，使用桩节点
var stubOnlyTypes = map[string]bool{
	"custom": true, // 需要在引擎进程中通过 RegisterCall 注册函数
}

// HasEngineDescriptor 判断节点类型是否在 glow-flow 引擎中有内置实现
func HasEngineDescriptor(nodeType string) bool {
	_, ok := glowflow.DefaultRegistry.GetDescriptor(nodeType)
	return ok
}

// descriptorFor 返回节点类型的描述器：内置节点使用引擎真实工厂，其他节点(独立服务节点等)使用桩节点
func descriptorFor(nodeType string, meta models.NodeTypeInfo) (desc *glowflow.NodeDescriptor, real bool) {
	if meta.Source == enums.NodeSourceBuiltin && !stubOnlyTypes[nodeType] {
		if desc, ok := glowflow.DefaultRegistry.GetDescriptor(nodeType); ok {
			return desc, true
		}
	}
	return &glowflow.NodeDescriptor{Type: nodeType, Factory: stubFactory{relation: defaultRelation(meta)}}, false
}

// defaultRelation 桩节点/模拟执行时走的连接类型
func defaultRelation(meta models.NodeTypeInfo) string {
	if containsString(meta.RelationTypes, enums.RelationSuccess) {
		return enums.RelationSuccess
	}
	if len(meta.RelationTypes) > 0 {
		return meta.RelationTypes[0]
	}
	return enums.RelationDefault
}

// Compile 使用 glow-flow 引擎编译模型：逐个构建节点以收集所有节点的配置错误，再整体编译校验节点引用
func Compile(meta ChainMeta, nodes []models.DesignNode, conns []models.DesignConnection, types map[string]models.NodeTypeInfo) []models.ValidationIssue {
	issues := make([]models.ValidationIssue, 0)
	registry := glowflow.NewNodeRegistry()
	for _, node := range nodes {
		nodeMeta, ok := types[node.Type]
		if !ok {
			// 未注册的节点类型已在结构校验中报告
			return issues
		}
		desc, exists := registry.GetDescriptor(node.Type)
		if !exists {
			desc, _ = descriptorFor(node.Type, nodeMeta)
			if err := registry.Register(desc); err != nil {
				issues = append(issues, compileIssue(node.ID, "节点类型 %s 注册失败：%v", node.Type, err))
				continue
			}
		}
		if _, err := safeBuild(desc.Factory, ToEngineNode(node)); err != nil {
			issues = append(issues, compileIssue(node.ID, "节点 %s 配置无效：%v", nodeLabel(node), err))
		}
	}
	if len(issues) > 0 {
		return issues
	}
	if _, err := glowflow.NewChainCompiler().Compile(ToChainDefinition(meta, nodes, conns), registry); err != nil {
		issues = append(issues, compileIssue("", "引擎编译失败：%v", err))
	}
	return issues
}

func compileIssue(nodeID, format string, args ...any) models.ValidationIssue {
	return models.ValidationIssue{Level: LevelError, Code: "compile", Message: fmt.Sprintf(format, args...), NodeID: nodeID}
}

// safeBuild 构建节点，节点工厂 panic 时转换为错误
func safeBuild(factory glowflow.NodeFactory, def glowflow.NodeDefinition) (node glowflow.CompiledNode, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return factory.Build(def)
}

// stubFactory 桩节点工厂：节点原样传递消息并走默认的连接类型
type stubFactory struct {
	relation string
}

func (f stubFactory) Build(def glowflow.NodeDefinition) (glowflow.CompiledNode, error) {
	return &stubNode{Node: def, relation: f.relation}, nil
}

type stubNode struct {
	glowflow.Node
	relation string
}

func (n *stubNode) Execute(ctx glowflow.Context, msg *glowflow.Message) (glowflow.ExecuteResult, error) {
	return glowflow.ExecuteResult{RelationType: n.relation, Msg: msg}, nil
}

// MergeValidation 将编译问题合并到校验结果
func MergeValidation(result *models.ValidationResult, issues []models.ValidationIssue) *models.ValidationResult {
	for _, issue := range issues {
		if issue.Level == LevelError {
			result.Errors = append(result.Errors, issue)
		} else {
			result.Warnings = append(result.Warnings, issue)
		}
	}
	result.Valid = len(result.Errors) == 0
	return result
}

// EngineSupportIssues 发布前检查节点类型在引擎中是否可用：
//   - 内置类型在 glow-flow 中没有实现时为错误(引擎加载会编译失败)
//   - 独立服务节点需要引擎宿主注册对应的节点类型，提示警告
func EngineSupportIssues(nodes []models.DesignNode, types map[string]models.NodeTypeInfo) []models.ValidationIssue {
	issues := make([]models.ValidationIssue, 0)
	reported := make(map[string]bool)
	for _, node := range nodes {
		meta, ok := types[node.Type]
		if !ok || reported[node.Type] {
			continue
		}
		reported[node.Type] = true
		if meta.Source == enums.NodeSourceBuiltin {
			if !stubOnlyTypes[node.Type] && !HasEngineDescriptor(node.Type) {
				issues = append(issues, models.ValidationIssue{Level: LevelError, Code: "engine_unsupported", NodeID: node.ID,
					Message: fmt.Sprintf("节点类型 %s(%s) 在 glow-flow 引擎中尚未实现，发布后引擎无法加载该流程", meta.Name, node.Type)})
			}
			if node.Type == "custom" {
				issues = append(issues, models.ValidationIssue{Level: LevelWarning, Code: "engine_host_required", NodeID: node.ID,
					Message: "自定义函数节点需要引擎宿主通过 RegisterCall 注册对应函数"})
			}
			continue
		}
		issues = append(issues, models.ValidationIssue{Level: LevelWarning, Code: "engine_host_required", NodeID: node.ID,
			Message: fmt.Sprintf("独立服务节点 %s(%s) 需要引擎宿主注册该节点类型才能执行", meta.Name, node.Type)})
	}
	return issues
}

// ValidateForPublish 发布级校验：结构校验、引擎可用性检查、引擎编译校验
func ValidateForPublish(in Input, meta ChainMeta) *models.ValidationResult {
	in.ForPublish = true
	result := ValidateAndCompile(in, meta)
	return MergeValidation(result, EngineSupportIssues(in.Nodes, in.NodeTypes))
}

// ValidateAndCompile 结构校验通过后再进行引擎编译校验
func ValidateAndCompile(in Input, meta ChainMeta) *models.ValidationResult {
	result := Validate(in)
	if !result.Valid || len(in.Nodes) == 0 {
		return result
	}
	return MergeValidation(result, Compile(meta, in.Nodes, in.Connections, in.NodeTypes))
}
