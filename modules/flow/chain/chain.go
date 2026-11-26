package chain

import (
	"context"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/audit"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
)

// Chainer 流程定义管理
type Chainer interface {
	// Page 分页查询流程定义
	Page(ctx context.Context, params models.PageChainParams) ([]models.DBChainInfo, int64, error)
	// Detail 流程定义详情
	Detail(ctx context.Context, chainNo string) (models.DBChainInfo, error)
	// Create 创建流程定义，同时创建 v1 草稿
	Create(ctx context.Context, op operator.Operator, params models.CreateChainParams) (models.DBChainInfo, error)
	// Update 修改流程基本信息
	Update(ctx context.Context, op operator.Operator, params models.UpdateChainParams) error
	// ChangeStatus 启用/禁用流程(禁用后引擎不加载)
	ChangeStatus(ctx context.Context, op operator.Operator, params models.ChangeChainStatusParams) error
	// Delete 删除从未发布过的流程
	Delete(ctx context.Context, op operator.Operator, chainNo string) error
	// Copy 复制流程的某个版本为新流程
	Copy(ctx context.Context, op operator.Operator, params models.CopyChainParams) (models.ImportResult, error)
	// Export 导出流程版本
	Export(ctx context.Context, params models.ExportChainParams) (models.ChainExport, string, error)
	// Import 导入流程
	Import(ctx context.Context, op operator.Operator, params models.ImportChainParams) (models.ImportResult, error)
}

type ChainerImpl struct {
	logger  log.Logger
	auditor audit.Auditor
}

func NewChainer(logger log.Logger) Chainer {
	return &ChainerImpl{logger: logger, auditor: audit.NewAuditor(logger)}
}
