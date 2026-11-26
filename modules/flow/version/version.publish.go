package version

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/audit"
	"github.com/zhiyunliu/manager-glow/modules/flow/dao"
	"github.com/zhiyunliu/manager-glow/modules/flow/designer"
	"github.com/zhiyunliu/manager-glow/modules/flow/sqls"
	"github.com/zhiyunliu/manager-glow/modules/flow/store"
	"github.com/zhiyunliu/manager-glow/modules/nodetype"
	"github.com/zhiyunliu/manager-glow/modules/option"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
)

const deleteVersionSQL = sqls.DeleteVersion

func (svc *VersionerImpl) Validate(ctx context.Context, params models.ValidateParams) (*models.ValidationResult, error) {
	types, err := nodetype.LoadMetaMap(ctx)
	if err != nil {
		return nil, err
	}
	nodes, conns := store.Normalize(params.Nodes, params.Connections)
	meta := designer.ChainMeta{ChainNo: params.ChainNo, Version: strconv.FormatInt(params.VersionID, 10), Status: enums.StatusEnabled}
	if meta.ChainNo == "" {
		meta.ChainNo = "validate"
	}
	result := designer.ValidateForPublish(designer.Input{FlowType: params.FlowType, Nodes: nodes, Connections: conns, NodeTypes: types}, meta)
	return designer.MergeValidation(result, optionIssues(ctx, nodes, types)), nil
}

// optionIssues 检查引用动态下拉数据源的配置值是否仍在可选列表中(仅警告)
func optionIssues(ctx context.Context, nodes []models.DesignNode, types map[string]models.NodeTypeInfo) []models.ValidationIssue {
	return designer.OptionSourceIssues(nodes, types, func(source string) (map[string]struct{}, error) {
		return option.Values(ctx, source)
	})
}

func (svc *VersionerImpl) Publish(ctx context.Context, op operator.Operator, params models.PublishParams) (result *models.ValidationResult, err error) {
	entry := audit.Entry{Module: enums.AuditModuleVersion, OperateType: enums.AuditOpPublish, TargetID: fmt.Sprint(params.VersionID), VersionID: params.VersionID, After: params}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	if err = store.CheckDesc(params.Desc); err != nil {
		return nil, err
	}
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
		entry.Summary = fmt.Sprintf("发布 %s v%d", version.ChainNo, version.VersionNo)
		if version.PublishState != enums.VersionStateDraft {
			return bizerr.Conflict(subcode.VersionNotDraft, "v%d 不是草稿，不能发布", version.VersionNo)
		}
		// 发布已保存的草稿内容
		nodes, conns, err := store.LoadDesign(ctx, tx, version.VersionID)
		if err != nil {
			return err
		}
		ext, _ := store.VersionMeta(version)
		result = designer.ValidateForPublish(
			designer.Input{FlowType: chain.FlowType, Nodes: nodes, Connections: conns, NodeTypes: types},
			designer.ChainMeta{ChainNo: version.ChainNo, Version: strconv.FormatInt(version.VersionID, 10), Name: version.Name, Status: chain.Status, ExtParams: ext},
		)
		if !result.Valid {
			return bizerr.Unprocessable(subcode.ModelInvalid, fmt.Sprintf("发布校验未通过(%d 个错误)：%s", len(result.Errors), result.Errors[0].Message), result)
		}
		result = designer.MergeValidation(result, optionIssues(ctx, nodes, types))
		affected, err := dao.ChangeVersionState(ctx, tx, sqls.PublishVersion, map[string]any{"version_id": version.VersionID, "desc": params.Desc, "operator": op.ID})
		if err != nil {
			return err
		}
		if affected == 0 {
			return bizerr.Conflict(subcode.VersionStateInvalid, "版本状态已变化，请刷新后重试")
		}
		entry.After = map[string]any{"version_no": version.VersionNo, "desc": params.Desc, "nodes": len(nodes), "connections": len(conns), "warnings": result.Warnings}
		return nil
	})
	if err != nil {
		svc.logger.Errorf("publish version failed, version_id:%d, err:%+v", params.VersionID, err)
	}
	return result, err
}

func (svc *VersionerImpl) Offline(ctx context.Context, op operator.Operator, versionID int64) error {
	return svc.changeState(ctx, op, versionID, enums.AuditOpOffline, enums.VersionStatePublished, sqls.OfflineVersion, "下线", "已发布")
}

func (svc *VersionerImpl) Online(ctx context.Context, op operator.Operator, versionID int64) error {
	return svc.changeState(ctx, op, versionID, enums.AuditOpOnline, enums.VersionStateOffline, sqls.OnlineVersion, "上线", "已下线")
}

func (svc *VersionerImpl) changeState(ctx context.Context, op operator.Operator, versionID int64, opType string, fromState int, query, action, fromLabel string) (err error) {
	entry := audit.Entry{Module: enums.AuditModuleVersion, OperateType: opType, TargetID: fmt.Sprint(versionID), VersionID: versionID}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	err = dao.Transaction(ctx, func(ctx context.Context, tx xdb.Executer) error {
		version, _, err := lockVersion(ctx, tx, versionID)
		if err != nil {
			return err
		}
		entry.ChainNo = version.ChainNo
		entry.Summary = fmt.Sprintf("%s %s v%d", action, version.ChainNo, version.VersionNo)
		entry.Before = map[string]any{"publish_state": version.PublishState, "status": version.Status}
		if version.PublishState != fromState {
			return bizerr.Conflict(subcode.VersionStateInvalid, "只有%s的版本才能%s", fromLabel, action)
		}
		affected, err := dao.ChangeVersionState(ctx, tx, query, map[string]any{"version_id": versionID, "operator": op.ID})
		if err != nil {
			return err
		}
		if affected == 0 {
			return bizerr.Conflict(subcode.VersionStateInvalid, "版本状态已变化，请刷新后重试")
		}
		return nil
	})
	if err != nil {
		svc.logger.Errorf("%s version failed, version_id:%d, err:%+v", opType, versionID, err)
	}
	return err
}
