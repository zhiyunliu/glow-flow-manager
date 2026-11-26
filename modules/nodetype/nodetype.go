package nodetype

import (
	"context"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/audit"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
)

// NodeTyper 节点注册管理
type NodeTyper interface {
	// List 查询节点类型(设计器工具箱)
	List(ctx context.Context, params models.ListNodeTypeParams) ([]models.NodeTypeInfo, error)
	// Page 分页查询节点类型
	Page(ctx context.Context, params models.PageNodeTypeParams) ([]models.NodeTypeInfo, int64, error)
	// Detail 节点类型详情
	Detail(ctx context.Context, nodeType string) (models.NodeTypeInfo, error)
	// Create 注册独立服务节点
	Create(ctx context.Context, op operator.Operator, params models.SaveNodeTypeParams) error
	// Update 修改节点类型
	Update(ctx context.Context, op operator.Operator, params models.SaveNodeTypeParams) error
	// ChangeStatus 启用/禁用节点类型
	ChangeStatus(ctx context.Context, op operator.Operator, params models.ChangeNodeTypeStatusParams) error
	// Delete 删除未被使用的独立服务节点
	Delete(ctx context.Context, op operator.Operator, nodeType string) error
	// SyncBuiltin 同步 glow-flow 内置节点
	SyncBuiltin(ctx context.Context, op operator.Operator) (models.SyncBuiltinResult, error)
	// Instances 查询独立服务节点在注册中心中的实例
	Instances(ctx context.Context, nodeType string) (models.NodeTypeInstances, error)
}

type NodeTyperImpl struct {
	logger  log.Logger
	auditor audit.Auditor
}

func NewNodeTyper(logger log.Logger) NodeTyper {
	return &NodeTyperImpl{logger: logger, auditor: audit.NewAuditor(logger)}
}

// LoadMetaMap 加载全部节点类型，供模型校验使用
func LoadMetaMap(ctx context.Context) (map[string]models.NodeTypeInfo, error) {
	rows, err := dbListNodeTypes(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]models.NodeTypeInfo, len(rows))
	for _, row := range rows {
		result[row.NodeType] = toInfo(row)
	}
	return result, nil
}
