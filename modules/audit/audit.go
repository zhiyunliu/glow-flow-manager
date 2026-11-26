package audit

import (
	"context"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
)

// Entry 一条操作审计
type Entry struct {
	Module      string
	OperateType string
	TargetID    string
	ChainNo     string
	VersionID   int64
	Summary     string
	Before      any // 操作前数据，记录为 JSON
	After       any // 操作后数据(或请求参数)，记录为 JSON
	Err         error
}

// Auditor 操作审计
type Auditor interface {
	// Record 记录审计日志；写入失败只记录错误日志，不影响业务操作
	Record(ctx context.Context, op operator.Operator, entry Entry)
	// Page 分页查询审计日志
	Page(ctx context.Context, params models.PageAuditParams) ([]models.DBOperateLog, int64, error)
	// Detail 审计日志详情
	Detail(ctx context.Context, logID int64) (models.DBOperateLog, error)
}

type AuditorImpl struct {
	logger log.Logger
}

func NewAuditor(logger log.Logger) Auditor {
	return &AuditorImpl{logger: logger}
}
