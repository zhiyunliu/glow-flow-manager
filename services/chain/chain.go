package chain

import (
	"encoding/json"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/flow/chain"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
	"github.com/zhiyunliu/manager-glow/pkg/resp"
)

// PageChain 流程定义-分页
func PageChain(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.PageChainParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("PageChain bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	list, total, err := chain.NewChainer(logger).Page(ctx.Context(), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Page(list, total)
}

// DetailChain 流程定义-详情
func DetailChain(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.ChainNoParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("DetailChain bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	item, err := chain.NewChainer(logger).Detail(ctx.Context(), params.ChainNo)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(item)
}

// CreateChain 流程定义-创建(同时创建 v1 草稿)
func CreateChain(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.CreateChainParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("CreateChain bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	item, err := chain.NewChainer(logger).Create(ctx.Context(), operator.FromContext(ctx), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(item)
}

// UpdateChain 流程定义-修改
func UpdateChain(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.UpdateChainParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("UpdateChain bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	if err := chain.NewChainer(logger).Update(ctx.Context(), operator.FromContext(ctx), params); err != nil {
		return resp.Error(err)
	}
	return resp.Success()
}

// ChangeChainStatus 流程定义-启用/禁用
func ChangeChainStatus(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.ChangeChainStatusParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("ChangeChainStatus bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	if err := chain.NewChainer(logger).ChangeStatus(ctx.Context(), operator.FromContext(ctx), params); err != nil {
		return resp.Error(err)
	}
	return resp.Success()
}

// DeleteChain 流程定义-删除(仅未发布过的流程)
func DeleteChain(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.ChainNoParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("DeleteChain bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	if err := chain.NewChainer(logger).Delete(ctx.Context(), operator.FromContext(ctx), params.ChainNo); err != nil {
		return resp.Error(err)
	}
	return resp.Success()
}

// CopyChain 流程定义-复制为新流程
func CopyChain(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.CopyChainParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("CopyChain bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	result, err := chain.NewChainer(logger).Copy(ctx.Context(), operator.FromContext(ctx), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(result)
}

// ExportChain 流程定义-导出；download=1 时以文件下载
func ExportChain(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.ExportChainParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("ExportChain bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	export, fileName, err := chain.NewChainer(logger).Export(ctx.Context(), params)
	if err != nil {
		return resp.Error(err)
	}
	if ctx.Request().Query().Get("download") == "1" {
		data, err := json.MarshalIndent(export, "", "  ")
		if err != nil {
			return resp.Error(err)
		}
		return resp.File(fileName, "application/json; charset=utf-8", data)
	}
	return resp.Success(map[string]any{"file_name": fileName, "content": export})
}

// ImportChain 流程定义-导入
func ImportChain(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.ImportChainParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("ImportChain bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	result, err := chain.NewChainer(logger).Import(ctx.Context(), operator.FromContext(ctx), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(result)
}
