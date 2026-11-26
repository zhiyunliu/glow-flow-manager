package version

import (
	"context"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/audit"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
)

// Versioner 流程版本与设计器
type Versioner interface {
	// List 查询流程的所有版本
	List(ctx context.Context, chainNo string) ([]models.DBVersionInfo, error)
	// Detail 加载版本的完整设计模型
	Detail(ctx context.Context, versionID int64) (models.ChainDesign, error)
	// CreateDraft 创建草稿版本；已有草稿时直接返回该草稿
	CreateDraft(ctx context.Context, op operator.Operator, params models.CreateDraftParams) (models.DBVersionInfo, error)
	// SaveDraft 保存草稿，返回校验结果(仅结构错误会阻止保存)
	SaveDraft(ctx context.Context, op operator.Operator, params models.SaveDraftParams) (*models.ValidationResult, error)
	// Validate 发布级模型校验(含引擎编译)
	Validate(ctx context.Context, params models.ValidateParams) (*models.ValidationResult, error)
	// Publish 发布草稿，发布后版本不可修改
	Publish(ctx context.Context, op operator.Operator, params models.PublishParams) (*models.ValidationResult, error)
	// Offline 下线已发布版本
	Offline(ctx context.Context, op operator.Operator, versionID int64) error
	// Online 重新上线已下线的版本
	Online(ctx context.Context, op operator.Operator, versionID int64) error
	// Delete 删除草稿
	Delete(ctx context.Context, op operator.Operator, versionID int64) error
	// Debug 调试预览
	Debug(ctx context.Context, params models.DebugParams) (models.DebugResult, error)
}

type VersionerImpl struct {
	logger  log.Logger
	auditor audit.Auditor
}

func NewVersioner(logger log.Logger) Versioner {
	return &VersionerImpl{logger: logger, auditor: audit.NewAuditor(logger)}
}
