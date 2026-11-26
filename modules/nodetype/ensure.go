package nodetype

import (
	"context"
	"time"

	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/audit"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
)

// SyncBuiltin 同步内置节点：插入缺失的内置节点；已存在的内置节点更新分类、输出类型与属性描述
func (svc *NodeTyperImpl) SyncBuiltin(ctx context.Context, op operator.Operator) (result models.SyncBuiltinResult, err error) {
	entry := audit.Entry{Module: enums.AuditModuleNodeType, OperateType: enums.AuditOpSync, TargetID: "builtin", Summary: "同步内置节点类型"}
	defer func() {
		entry.Err = err
		entry.After = result
		svc.auditor.Record(ctx, op, entry)
	}()
	return syncBuiltin(ctx, op.ID, true)
}

// EnsureBuiltin 服务启动时补齐缺失的内置节点，不修改已存在的记录；失败只记录日志
func EnsureBuiltin(ctx context.Context, logger log.Logger) {
	// glue.DB 在连接失败时 panic，这里不应阻止服务启动
	defer func() {
		if r := recover(); r != nil {
			logger.Warnf("ensure builtin node types failed: %v", r)
		}
	}()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	result, err := syncBuiltin(ctx, operator.System().ID, false)
	if err != nil {
		logger.Warnf("ensure builtin node types failed: %v", err)
		return
	}
	if len(result.Inserted) > 0 {
		logger.Infof("builtin node types inserted: %v", result.Inserted)
	}
}

func syncBuiltin(ctx context.Context, operatorID string, overwrite bool) (models.SyncBuiltinResult, error) {
	result := models.SyncBuiltinResult{Inserted: []string{}, Updated: []string{}}
	existed, err := dbListNodeTypeKeys(ctx)
	if err != nil {
		return result, err
	}
	for _, info := range builtinInfos() {
		params, err := toDBParams(info, operatorID)
		if err != nil {
			return result, err
		}
		if _, ok := existed[info.NodeType]; !ok {
			if err := dbInsertNodeType(ctx, params); err != nil {
				return result, err
			}
			result.Inserted = append(result.Inserted, info.NodeType)
			continue
		}
		if !overwrite {
			continue
		}
		if err := dbUpdateBuiltinNodeType(ctx, params); err != nil {
			return result, err
		}
		result.Updated = append(result.Updated, info.NodeType)
	}
	return result, nil
}
