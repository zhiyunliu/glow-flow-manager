package nodetype

import (
	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/nodetype"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
	"github.com/zhiyunliu/manager-glow/pkg/resp"
)

// ListNodeType 节点注册-列表(设计器工具箱)
func ListNodeType(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.ListNodeTypeParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("ListNodeType bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	list, err := nodetype.NewNodeTyper(logger).List(ctx.Context(), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(list)
}

// PageNodeType 节点注册-分页
func PageNodeType(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.PageNodeTypeParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("PageNodeType bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	list, total, err := nodetype.NewNodeTyper(logger).Page(ctx.Context(), params)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Page(list, total)
}

// DetailNodeType 节点注册-详情
func DetailNodeType(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.NodeTypeParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("DetailNodeType bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	item, err := nodetype.NewNodeTyper(logger).Detail(ctx.Context(), params.NodeType)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(item)
}

// CreateNodeType 节点注册-注册独立服务节点
func CreateNodeType(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.SaveNodeTypeParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("CreateNodeType bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	if err := nodetype.NewNodeTyper(logger).Create(ctx.Context(), operator.FromContext(ctx), params); err != nil {
		return resp.Error(err)
	}
	return resp.Success()
}

// UpdateNodeType 节点注册-修改
func UpdateNodeType(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.SaveNodeTypeParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("UpdateNodeType bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	if err := nodetype.NewNodeTyper(logger).Update(ctx.Context(), operator.FromContext(ctx), params); err != nil {
		return resp.Error(err)
	}
	return resp.Success()
}

// ChangeNodeTypeStatus 节点注册-启用/禁用
func ChangeNodeTypeStatus(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.ChangeNodeTypeStatusParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("ChangeNodeTypeStatus bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	if err := nodetype.NewNodeTyper(logger).ChangeStatus(ctx.Context(), operator.FromContext(ctx), params); err != nil {
		return resp.Error(err)
	}
	return resp.Success()
}

// DeleteNodeType 节点注册-删除
func DeleteNodeType(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.NodeTypeParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("DeleteNodeType bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	if err := nodetype.NewNodeTyper(logger).Delete(ctx.Context(), operator.FromContext(ctx), params.NodeType); err != nil {
		return resp.Error(err)
	}
	return resp.Success()
}

// SyncBuiltinNodeType 节点注册-同步内置节点
func SyncBuiltinNodeType(ctx context.Context) interface{} {
	logger := ctx.Log()
	result, err := nodetype.NewNodeTyper(logger).SyncBuiltin(ctx.Context(), operator.FromContext(ctx))
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(result)
}

// NodeTypeInstances 节点注册-注册中心实例
func NodeTypeInstances(ctx context.Context) interface{} {
	logger := ctx.Log()
	params := models.NodeTypeParams{}
	if err := ctx.Bind(&params); err != nil {
		logger.Errorf("NodeTypeInstances bind params failed, err:%+v", err)
		return resp.BindError(err)
	}
	result, err := nodetype.NewNodeTyper(logger).Instances(ctx.Context(), params.NodeType)
	if err != nil {
		return resp.Error(err)
	}
	return resp.Success(result)
}
