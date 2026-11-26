package enums

// 节点分类
const (
	NodeCategoryFilter      = "filter"      // 过滤
	NodeCategoryTransform   = "transform"   // 转换
	NodeCategoryAction      = "action"      // 动作
	NodeCategoryLogical     = "logical"     // 逻辑控制
	NodeCategoryWorkflow    = "workflow"    // 工作流(人工任务、子流程等)
	NodeCategoryIntegration = "integration" // 集成(外部服务节点)
)

// IsValidNodeCategory 判断节点分类是否合法
func IsValidNodeCategory(category string) bool {
	switch category {
	case NodeCategoryFilter, NodeCategoryTransform, NodeCategoryAction,
		NodeCategoryLogical, NodeCategoryWorkflow, NodeCategoryIntegration:
		return true
	}
	return false
}

// 节点来源
const (
	NodeSourceBuiltin = "builtin" // glow-flow 引擎内置节点
	NodeSourceService = "service" // 独立服务提供的节点，注册到注册中心(如 nacos)
)

// 内置连接关系类型，与 glow-flow constants/enums/relationtype 保持一致
const (
	RelationSuccess = "success"
	RelationFailure = "failure"
	RelationDefault = "default"
	RelationTrue    = "true"
	RelationFalse   = "false"
)
