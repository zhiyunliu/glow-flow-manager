package chain

import (
	"context"
	"strings"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/audit"
	"github.com/zhiyunliu/manager-glow/modules/flow/dao"
	"github.com/zhiyunliu/manager-glow/modules/flow/store"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
)

func checkChainParams(chainNo, name, flowType, desc string) error {
	if err := store.CheckChainNo(chainNo); err != nil {
		return err
	}
	if err := store.CheckName(name); err != nil {
		return err
	}
	if !enums.IsValidFlowType(flowType) {
		return bizerr.BadRequest(subcode.ParamInvalid, "流程类型 %q 无效", flowType)
	}
	return store.CheckDesc(desc)
}

func (svc *ChainerImpl) Create(ctx context.Context, op operator.Operator, params models.CreateChainParams) (chain models.DBChainInfo, err error) {
	params.ChainNo = strings.TrimSpace(params.ChainNo)
	params.Name = strings.TrimSpace(params.Name)
	entry := audit.Entry{Module: enums.AuditModuleChain, OperateType: enums.AuditOpCreate, TargetID: params.ChainNo, ChainNo: params.ChainNo, Summary: "创建流程 " + params.Name, After: params}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	if err = checkChainParams(params.ChainNo, params.Name, params.FlowType, params.Desc); err != nil {
		return chain, err
	}
	err = dao.Transaction(ctx, func(ctx context.Context, tx xdb.Executer) error {
		existed, err := dao.LockChain(ctx, tx, params.ChainNo)
		if err != nil {
			return err
		}
		if existed.ChainNo != "" {
			return bizerr.Conflict(subcode.ChainExists, "流程编号 %s 已存在", params.ChainNo)
		}
		if err := dao.InsertChain(ctx, tx, map[string]any{
			"chain_no":  params.ChainNo,
			"name":      params.Name,
			"extparams": "",
			"layout":    "",
			"operator":  op.ID,
			"desc":      params.Desc,
			"flow_type": params.FlowType,
		}); err != nil {
			return err
		}
		versionID, _, err := store.CreateDraft(ctx, tx, store.DraftSpec{ChainNo: params.ChainNo, Name: params.Name}, op.ID)
		entry.VersionID = versionID
		return err
	})
	if err != nil {
		if store.IsDuplicateKey(err) {
			err = bizerr.Conflict(subcode.ChainExists, "流程编号 %s 已存在", params.ChainNo)
		}
		svc.logger.Errorf("create chain failed, params:%+v, err:%+v", params, err)
		return chain, err
	}
	return svc.retrieve(ctx, params.ChainNo)
}

func (svc *ChainerImpl) Update(ctx context.Context, op operator.Operator, params models.UpdateChainParams) (err error) {
	params.Name = strings.TrimSpace(params.Name)
	entry := audit.Entry{Module: enums.AuditModuleChain, OperateType: enums.AuditOpUpdate, TargetID: params.ChainNo, ChainNo: params.ChainNo, Summary: "修改流程 " + params.Name, After: params}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	if err = checkChainParams(params.ChainNo, params.Name, params.FlowType, params.Desc); err != nil {
		return err
	}
	current, err := svc.retrieve(ctx, params.ChainNo)
	if err != nil {
		return err
	}
	entry.Before = map[string]any{"chain_no": current.ChainNo, "name": current.Name, "flow_type": current.FlowType, "desc": current.Desc}
	if current.FlowType != params.FlowType && current.PublishedCount > 0 {
		return bizerr.Conflict(subcode.ChainTypeLocked, "流程已有发布记录，不能修改流程类型")
	}
	return dao.Transaction(ctx, func(ctx context.Context, tx xdb.Executer) error {
		if err := dao.UpdateChain(ctx, tx, map[string]any{
			"chain_no":  params.ChainNo,
			"name":      params.Name,
			"flow_type": params.FlowType,
			"desc":      params.Desc,
			"operator":  op.ID,
		}); err != nil {
			return err
		}
		return dao.UpdateDraftVersionName(ctx, tx, params.ChainNo, params.Name)
	})
}

func (svc *ChainerImpl) ChangeStatus(ctx context.Context, op operator.Operator, params models.ChangeChainStatusParams) (err error) {
	opType, summary := enums.AuditOpDisable, "禁用流程 "
	if params.Status == enums.StatusEnabled {
		opType, summary = enums.AuditOpEnable, "启用流程 "
	}
	entry := audit.Entry{Module: enums.AuditModuleChain, OperateType: opType, TargetID: params.ChainNo, ChainNo: params.ChainNo, Summary: summary + params.ChainNo, After: params}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	current, err := svc.retrieve(ctx, params.ChainNo)
	if err != nil {
		return err
	}
	entry.Before = map[string]any{"chain_no": current.ChainNo, "status": current.Status}
	if err = dao.UpdateChainStatus(ctx, dao.DB(), params.ChainNo, params.Status, op.ID); err != nil {
		svc.logger.Errorf("update chain status failed, params:%+v, err:%+v", params, err)
	}
	return err
}

func (svc *ChainerImpl) Delete(ctx context.Context, op operator.Operator, chainNo string) (err error) {
	entry := audit.Entry{Module: enums.AuditModuleChain, OperateType: enums.AuditOpDelete, TargetID: chainNo, ChainNo: chainNo, Summary: "删除流程 " + chainNo}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	current, err := svc.retrieve(ctx, chainNo)
	if err != nil {
		return err
	}
	entry.Before = current
	if current.PublishedCount > 0 {
		return bizerr.Conflict(subcode.ChainHasPublished, "流程已有发布记录，不能删除，可禁用流程")
	}
	err = dao.Transaction(ctx, func(ctx context.Context, tx xdb.Executer) error {
		return dao.DeleteChain(ctx, tx, chainNo)
	})
	if err != nil {
		svc.logger.Errorf("delete chain failed, chain_no:%s, err:%+v", chainNo, err)
	}
	return err
}
