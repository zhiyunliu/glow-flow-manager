package audit

import (
	"context"
	"strings"

	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
	"github.com/zhiyunliu/manager-glow/pkg/timefmt"
)

func (a *AuditorImpl) Page(ctx context.Context, params models.PageAuditParams) ([]models.DBOperateLog, int64, error) {
	offset := params.Normalize()
	query := map[string]any{
		"offset":    offset,
		"page_size": params.PageSize,
	}
	setIfNotEmpty(query, "module", params.Module)
	setIfNotEmpty(query, "operate_type", params.OperateType)
	setIfNotEmpty(query, "target_id", params.TargetID)
	setIfNotEmpty(query, "chain_no", params.ChainNo)
	setIfNotEmpty(query, "operator", params.Operator)
	if params.Result == "0" || params.Result == "1" {
		query["result"] = params.Result
	}
	if params.StartTime != "" {
		start, err := timefmt.ToSQL(params.StartTime)
		if err != nil {
			return nil, 0, bizerr.BadRequest(subcode.ParamInvalid, "开始时间格式错误：%s", params.StartTime)
		}
		query["start_time"] = start
	}
	if params.EndTime != "" {
		end, err := timefmt.ToSQL(params.EndTime)
		if err != nil {
			return nil, 0, bizerr.BadRequest(subcode.ParamInvalid, "结束时间格式错误：%s", params.EndTime)
		}
		query["end_time"] = end
	}
	list, total, err := dbPageOperateLog(ctx, query, params.PageSize)
	if err != nil {
		a.logger.Errorf("page operate log failed, params:%+v, err:%+v", params, err)
		return nil, 0, err
	}
	return list, total, nil
}

func (a *AuditorImpl) Detail(ctx context.Context, logID int64) (models.DBOperateLog, error) {
	item, err := dbGetOperateLog(ctx, logID)
	if err != nil {
		a.logger.Errorf("get operate log failed, log_id:%d, err:%+v", logID, err)
		return item, err
	}
	if item.LogID == 0 {
		return item, bizerr.NotFound(subcode.AuditNotFound, "审计日志不存在")
	}
	return item, nil
}

func setIfNotEmpty(query map[string]any, key, val string) {
	if val = strings.TrimSpace(val); val != "" {
		query[key] = val
	}
}
