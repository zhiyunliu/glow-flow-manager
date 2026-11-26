package history

import (
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/history"
	"github.com/zhiyunliu/manager-glow/pkg/resp"
)

// PageInstance 执行历史-分页
func PageInstance(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.PageInstanceParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("PageInstance bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	list, total, err := history.NewHistorian(logger).Page(ctx.Context(), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Page(list, total)
}

// DetailInstance 执行历史-详情
func DetailInstance(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.InstanceIDParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("DetailInstance bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	detail, err := history.NewHistorian(logger).Detail(ctx.Context(), params.InstanceID)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(detail)
}
