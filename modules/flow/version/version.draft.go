package version

import (
	"context"
	"fmt"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/audit"
	"github.com/zhiyunliu/manager-glow/modules/flow/dao"
	"github.com/zhiyunliu/manager-glow/modules/flow/designer"
	"github.com/zhiyunliu/manager-glow/modules/flow/store"
	"github.com/zhiyunliu/manager-glow/modules/nodetype"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
)

// baseVersionEmpty 创建空白草稿
const baseVersionEmpty = -1

func (svc *VersionerImpl) CreateDraft(ctx context.Context, op operator.Operator, params models.CreateDraftParams) (version models.DBVersionInfo, err error) {
	entry := audit.Entry{Module: enums.AuditModuleVersion, OperateType: enums.AuditOpCreate, ChainNo: params.ChainNo, After: params}
	created := false
	defer func() {
		// 已有草稿直接返回时不记录审计
		if created || err != nil {
			entry.Err = err
			svc.auditor.Record(ctx, op, entry)
		}
	}()

	chain, err := svc.retrieveChain(ctx, params.ChainNo)
	if err != nil {
		return version, err
	}
	var draftID int64
	err = dao.Transaction(ctx, func(ctx context.Context, tx xdb.Executer) error {
		if _, err := dao.LockChain(ctx, tx, params.ChainNo); err != nil {
			return err
		}
		existing, err := dao.GetDraftVersionID(ctx, tx, params.ChainNo)
		if err != nil {
			return err
		}
		if existing > 0 {
			draftID = existing
			return nil
		}

		spec := store.DraftSpec{ChainNo: params.ChainNo, Name: chain.Name}
		if params.BaseVersionID != baseVersionEmpty {
			baseID, err := store.ResolveSourceVersion(ctx, tx, params.ChainNo, params.BaseVersionID)
			if err != nil {
				return err
			}
			if baseID > 0 {
				base, err := dao.GetVersion(ctx, tx, baseID)
				if err != nil {
					return err
				}
				// 同一流程内复制，保留节点编号
				spec.Nodes, spec.Connections, err = store.LoadDesign(ctx, tx, baseID)
				if err != nil {
					return err
				}
				spec.ExtParams, spec.Layout = store.VersionMeta(base)
				spec.BaseVersionID = baseID
				spec.Desc = fmt.Sprintf("基于 v%d", base.VersionNo)
			}
		}
		draftID, _, err = store.CreateDraft(ctx, tx, spec, op.ID)
		created = err == nil
		return err
	})
	if err != nil {
		svc.logger.Errorf("create draft failed, params:%+v, err:%+v", params, err)
		return version, err
	}
	version, err = svc.retrieveVersion(ctx, draftID)
	entry.TargetID = fmt.Sprint(draftID)
	entry.VersionID = draftID
	entry.Summary = fmt.Sprintf("创建草稿 %s v%d", params.ChainNo, version.VersionNo)
	return version, err
}

func (svc *VersionerImpl) SaveDraft(ctx context.Context, op operator.Operator, params models.SaveDraftParams) (result *models.ValidationResult, err error) {
	entry := audit.Entry{Module: enums.AuditModuleVersion, OperateType: enums.AuditOpSaveDraft, TargetID: fmt.Sprint(params.VersionID), VersionID: params.VersionID}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	if err = store.CheckDesc(params.Desc); err != nil {
		return nil, err
	}
	nodes, conns := store.Normalize(params.Nodes, params.Connections)
	types, err := nodetype.LoadMetaMap(ctx)
	if err != nil {
		return nil, err
	}

	err = dao.Transaction(ctx, func(ctx context.Context, tx xdb.Executer) error {
		version, chain, err := lockVersion(ctx, tx, params.VersionID)
		if err != nil {
			return err
		}
		entry.ChainNo = version.ChainNo
		entry.Summary = fmt.Sprintf("保存草稿 %s v%d", version.ChainNo, version.VersionNo)
		if version.PublishState != enums.VersionStateDraft {
			return bizerr.Conflict(subcode.VersionNotDraft, "v%d 不是草稿，不能修改", version.VersionNo)
		}
		flowType := chain.FlowType
		if params.FlowType != "" && params.FlowType != chain.FlowType {
			if !enums.IsValidFlowType(params.FlowType) {
				return bizerr.BadRequest(subcode.ParamInvalid, "流程类型 %q 无效", params.FlowType)
			}
			if published, err := hasPublished(ctx, tx, version.ChainNo); err != nil {
				return err
			} else if published {
				return bizerr.Conflict(subcode.ChainTypeLocked, "流程已有发布记录，不能切换流程类型")
			}
			if err := dao.UpdateChainFlowType(ctx, tx, version.ChainNo, params.FlowType, op.ID); err != nil {
				return err
			}
			flowType = params.FlowType
		}

		result = designer.Validate(designer.Input{FlowType: flowType, Nodes: nodes, Connections: conns, NodeTypes: types})
		if err := store.CheckIntegrity(result, conns); err != nil {
			return err
		}

		beforeNodes, beforeConns, err := store.LoadDesign(ctx, tx, version.VersionID)
		if err != nil {
			return err
		}
		entry.Before = map[string]any{"desc": version.Desc, "nodes": beforeNodes, "connections": beforeConns}
		entry.After = map[string]any{"flow_type": flowType, "desc": params.Desc, "nodes": nodes, "connections": conns}

		if err := store.WriteGraph(ctx, tx, version.ChainNo, version.VersionID, nodes, conns, op.ID); err != nil {
			return err
		}
		ext, err := designer.EncodeJSON(params.ExtParams)
		if err != nil {
			return err
		}
		layout, err := designer.EncodeJSON(params.Layout)
		if err != nil {
			return err
		}
		return dao.UpdateDraftVersion(ctx, tx, map[string]any{
			"version_id": version.VersionID,
			"name":       chain.Name,
			"extparams":  ext,
			"layout":     layout,
			"desc":       params.Desc,
			"operator":   op.ID,
		})
	})
	if err != nil {
		svc.logger.Errorf("save draft failed, version_id:%d, err:%+v", params.VersionID, err)
		return result, err
	}
	return result, nil
}

func (svc *VersionerImpl) Delete(ctx context.Context, op operator.Operator, versionID int64) (err error) {
	entry := audit.Entry{Module: enums.AuditModuleVersion, OperateType: enums.AuditOpDelete, TargetID: fmt.Sprint(versionID), VersionID: versionID}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	err = dao.Transaction(ctx, func(ctx context.Context, tx xdb.Executer) error {
		version, _, err := lockVersion(ctx, tx, versionID)
		if err != nil {
			return err
		}
		entry.ChainNo = version.ChainNo
		entry.Summary = fmt.Sprintf("删除草稿 %s v%d", version.ChainNo, version.VersionNo)
		if version.PublishState != enums.VersionStateDraft {
			return bizerr.Conflict(subcode.VersionNotDraft, "只能删除草稿版本")
		}
		nodes, conns, err := store.LoadDesign(ctx, tx, versionID)
		if err != nil {
			return err
		}
		entry.Before = map[string]any{"version": version, "nodes": nodes, "connections": conns}
		if err := dao.ClearVersionGraph(ctx, tx, versionID); err != nil {
			return err
		}
		if _, err := dao.ChangeVersionState(ctx, tx, deleteVersionSQL, map[string]any{"version_id": versionID}); err != nil {
			return err
		}
		return dao.DeleteOrphanNodeDefinitions(ctx, tx, version.ChainNo)
	})
	if err != nil {
		svc.logger.Errorf("delete draft failed, version_id:%d, err:%+v", versionID, err)
	}
	return err
}

// hasPublished 流程是否有发布记录(已发布或已下线)
func hasPublished(ctx context.Context, tx xdb.Executer, chainNo string) (bool, error) {
	versions, err := dao.ListVersions(ctx, tx, chainNo)
	if err != nil {
		return false, err
	}
	for _, v := range versions {
		if v.PublishState != enums.VersionStateDraft {
			return true, nil
		}
	}
	return false, nil
}
