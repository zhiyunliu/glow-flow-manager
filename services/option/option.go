package option

import (
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/option"
	"github.com/zhiyunliu/manager-glow/pkg/resp"
)

// ListOption 动态下拉选项-列表
func ListOption(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.ListOptionParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("ListOption bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	list, err := option.List(ctx.Context(), params.Source)
	if err != nil {
		logger.Warnf("ListOption failed, source:%s, err:%v", params.Source, err)
		return resp.Error(err)
	}
	return resp.Success(list)
}
