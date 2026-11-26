package version

import (
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/flow/version"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
	"github.com/zhiyunliu/manager-glow/pkg/resp"
)

// ListVersion 流程版本-列表
func ListVersion(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.ListVersionParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("ListVersion bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	list, err := version.NewVersioner(logger).List(ctx.Context(), params.ChainNo)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(list)
}

// DetailVersion 流程版本-设计模型
func DetailVersion(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.VersionIDParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("DetailVersion bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	design, err := version.NewVersioner(logger).Detail(ctx.Context(), params.VersionID)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(design)
}

// CreateDraft 流程版本-创建草稿
func CreateDraft(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.CreateDraftParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("CreateDraft bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	item, err := version.NewVersioner(logger).CreateDraft(ctx.Context(), operator.FromContext(ctx), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(item)
}

// SaveDraft 流程版本-保存草稿
func SaveDraft(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.SaveDraftParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("SaveDraft bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	result, err := version.NewVersioner(logger).SaveDraft(ctx.Context(), operator.FromContext(ctx), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(result)
}

// ValidateModel 流程版本-模型校验(发布级，含引擎编译)
func ValidateModel(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.ValidateParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("ValidateModel bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	result, err := version.NewVersioner(logger).Validate(ctx.Context(), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(result)
}

// PublishVersion 流程版本-发布
func PublishVersion(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.PublishParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("PublishVersion bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	result, err := version.NewVersioner(logger).Publish(ctx.Context(), operator.FromContext(ctx), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(result)
}

// OfflineVersion 流程版本-下线
func OfflineVersion(ctx context.Context) interface{} {
	return changeVersion(ctx, "OfflineVersion", func(svc version.Versioner, ctx context.Context, versionID int64) error {
		return svc.Offline(ctx.Context(), operator.FromContext(ctx), versionID)
	})
}

// OnlineVersion 流程版本-重新上线
func OnlineVersion(ctx context.Context) interface{} {
	return changeVersion(ctx, "OnlineVersion", func(svc version.Versioner, ctx context.Context, versionID int64) error {
		return svc.Online(ctx.Context(), operator.FromContext(ctx), versionID)
	})
}

// DeleteVersion 流程版本-删除草稿
func DeleteVersion(ctx context.Context) interface{} {
	return changeVersion(ctx, "DeleteVersion", func(svc version.Versioner, ctx context.Context, versionID int64) error {
		return svc.Delete(ctx.Context(), operator.FromContext(ctx), versionID)
	})
}

func changeVersion(ctx context.Context, name string, call func(svc version.Versioner, ctx context.Context, versionID int64) error) interface{} {
	logger := ctx.Log()
	params := models.VersionIDParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("%s bind params failed, err:%+v", name, err)
		return resp.BindError(err)
	}
	if err := call(version.NewVersioner(logger), ctx, params.VersionID); err != nil {
		return resp.Error(err)
	}
	return resp.Success()
}

// DebugModel 流程版本-调试预览
func DebugModel(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.DebugParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("DebugModel bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	result, err := version.NewVersioner(logger).Debug(ctx.Context(), params)
	if err != nil {
		if _, ok := bizerr.IsBiz(err); !ok {
			logger.Errorf("DebugModel failed, err:%+v", err)
		}
		return resp.Error(err)
	}
	return resp.Success(result)
}
