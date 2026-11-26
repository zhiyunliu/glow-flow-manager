package audit

import (
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/audit"
	"github.com/zhiyunliu/manager-glow/pkg/resp"
)

// PageAudit 操作审计-分页
func PageAudit(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.PageAuditParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("PageAudit bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	list, total, err := audit.NewAuditor(logger).Page(ctx.Context(), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Page(list, total)
}

// DetailAudit 操作审计-详情(含操作前后数据)
func DetailAudit(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.AuditIDParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("DetailAudit bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	item, err := audit.NewAuditor(logger).Detail(ctx.Context(), params.LogID)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(item)
}
