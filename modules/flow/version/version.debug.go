package version

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/zhiyunliu/manager-glow/config"
	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/flow/designer"
	"github.com/zhiyunliu/manager-glow/modules/flow/store"
	"github.com/zhiyunliu/manager-glow/modules/nodetype"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
)

// Debug 调试预览：使用设计器中当前(未保存)的模型，在沙箱中执行一次。
// 默认模拟执行动作/集成类节点，避免调试时发送请求、消息或执行命令。
func (svc *VersionerImpl) Debug(ctx context.Context, params models.DebugParams) (models.DebugResult, error) {
	result := models.DebugResult{Events: []models.DebugEvent{}, ExecutedNodes: []string{}, Edges: []models.DebugEdge{}}
	if len(params.Nodes) == 0 {
		return result, bizerr.BadRequest(subcode.ParamInvalid, "流程中没有节点")
	}
	var data any
	if text := strings.TrimSpace(string(params.Input)); text != "" && text != "null" {
		if err := json.Unmarshal(params.Input, &data); err != nil {
			return result, bizerr.BadRequest(subcode.ParamInvalid, "输入消息不是有效的 JSON：%v", err)
		}
	}
	types, err := nodetype.LoadMetaMap(ctx)
	if err != nil {
		return result, err
	}
	nodes, conns := store.Normalize(params.Nodes, params.Connections)
	validation := designer.Validate(designer.Input{FlowType: params.FlowType, Nodes: nodes, Connections: conns, NodeTypes: types})
	if !validation.Valid {
		result.Validation = validation
		result.Error = "模型校验未通过：" + validation.Errors[0].Message
		return result, nil
	}

	timeout := params.Timeout
	if timeout <= 0 {
		timeout = config.Sys.DebugTimeout
	}
	if timeout > config.Sys.DebugMaxTimeout {
		timeout = config.Sys.DebugMaxTimeout
	}
	mockActions := true
	if params.MockActions != nil {
		mockActions = *params.MockActions
	}
	if !mockActions && !config.Sys.DebugAllowRealActions {
		return result, bizerr.BadRequest(subcode.ParamInvalid, "服务端未允许调试时真实执行动作节点(配置 manager.debug_allow_real_actions)")
	}
	chainNo := params.ChainNo
	if chainNo == "" {
		chainNo = "debug"
	}
	result = designer.Debug(ctx, designer.DebugInput{
		Meta:        designer.ChainMeta{ChainNo: chainNo, Version: "debug", Status: enums.StatusEnabled, ExtParams: params.ExtParams},
		Nodes:       nodes,
		Connections: conns,
		NodeTypes:   types,
		Data:        data,
		MsgID:       params.MsgID,
		Timeout:     time.Duration(timeout) * time.Second,
		MockActions: mockActions,
	})
	if len(validation.Warnings) > 0 {
		result.Validation = validation
	}
	svc.logger.Infof("debug chain, chain_no:%s, success:%v, events:%d, duration:%dms", chainNo, result.Success, len(result.Events), result.DurationMs)
	return result, nil
}
