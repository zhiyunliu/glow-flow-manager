// Package dao 流程定义相关表的数据访问，函数接收 xdb.Executer，可在事务中使用
package dao

import (
	"context"

	"github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/manager-glow/config"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/flow/sqls"
)

// DB 流程库连接
func DB() xdb.IDB {
	return glue.DB(config.Sys.DBName)
}

// Transaction 在事务中执行
func Transaction(ctx context.Context, callback func(ctx context.Context, tx xdb.Executer) error) error {
	return DB().Transaction(ctx, callback)
}

// ---------- 流程 ----------

func PageChains(ctx context.Context, dbObj xdb.Executer, query map[string]any, pageSize int) ([]models.DBChainInfo, int64, error) {
	total := models.TotalRow{}
	if err := dbObj.FirstAs(ctx, sqls.CountChains, query, &total); err != nil {
		return nil, 0, err
	}
	list := make([]models.DBChainInfo, 0, pageSize)
	if total.Total == 0 {
		return list, 0, nil
	}
	err := dbObj.QueryAs(ctx, sqls.PageChains, query, &list)
	return list, total.Total, err
}

func GetChain(ctx context.Context, dbObj xdb.Executer, chainNo string) (models.DBChainInfo, error) {
	item := models.DBChainInfo{}
	err := dbObj.FirstAs(ctx, sqls.GetChain, map[string]any{"chain_no": chainNo}, &item)
	return item, err
}

// LockChain 锁定流程行，返回的 ChainNo 为空表示不存在
func LockChain(ctx context.Context, tx xdb.Executer, chainNo string) (models.DBChainInfo, error) {
	item := models.DBChainInfo{}
	err := tx.FirstAs(ctx, sqls.LockChain, map[string]any{"chain_no": chainNo}, &item)
	return item, err
}

func InsertChain(ctx context.Context, tx xdb.Executer, params map[string]any) error {
	_, err := tx.Exec(ctx, sqls.InsertChain, params)
	return err
}

func UpdateChain(ctx context.Context, tx xdb.Executer, params map[string]any) error {
	_, err := tx.Exec(ctx, sqls.UpdateChain, params)
	return err
}

func UpdateChainFlowType(ctx context.Context, tx xdb.Executer, chainNo, flowType, operator string) error {
	_, err := tx.Exec(ctx, sqls.UpdateChainFlowType, map[string]any{"chain_no": chainNo, "flow_type": flowType, "operator": operator})
	return err
}

func UpdateChainStatus(ctx context.Context, dbObj xdb.Executer, chainNo string, status int, operator string) error {
	_, err := dbObj.Exec(ctx, sqls.UpdateChainStatus, map[string]any{"chain_no": chainNo, "status": status, "operator": operator})
	return err
}

// DeleteChain 删除流程及其所有版本、节点、连接、节点定义
func DeleteChain(ctx context.Context, tx xdb.Executer, chainNo string) error {
	params := map[string]any{"chain_no": chainNo}
	for _, query := range []string{
		sqls.DeleteChainConnections,
		sqls.DeleteChainNodes,
		sqls.DeleteChainVersions,
		sqls.DeleteChainNodeDefinitions,
		sqls.DeleteChain,
	} {
		if _, err := tx.Exec(ctx, query, params); err != nil {
			return err
		}
	}
	return nil
}

// ---------- 版本 ----------

func ListVersions(ctx context.Context, dbObj xdb.Executer, chainNo string) ([]models.DBVersionInfo, error) {
	list := make([]models.DBVersionInfo, 0)
	err := dbObj.QueryAs(ctx, sqls.ListVersions, map[string]any{"chain_no": chainNo}, &list)
	return list, err
}

func GetVersion(ctx context.Context, dbObj xdb.Executer, versionID int64) (models.DBVersionInfo, error) {
	item := models.DBVersionInfo{}
	err := dbObj.FirstAs(ctx, sqls.GetVersion, map[string]any{"version_id": versionID}, &item)
	return item, err
}

// LockVersion 锁定版本行，返回的 VersionID 为 0 表示不存在
func LockVersion(ctx context.Context, tx xdb.Executer, versionID int64) (models.DBVersionInfo, error) {
	item := models.DBVersionInfo{}
	err := tx.FirstAs(ctx, sqls.LockVersion, map[string]any{"version_id": versionID}, &item)
	return item, err
}

func scalarInt64(ctx context.Context, dbObj xdb.Executer, query string, params map[string]any) (int64, error) {
	row := models.TotalRow{}
	err := dbObj.FirstAs(ctx, query, params, &row)
	return row.Total, err
}

func GetEffectiveVersionID(ctx context.Context, dbObj xdb.Executer, chainNo string) (int64, error) {
	return scalarInt64(ctx, dbObj, sqls.GetEffectiveVersionID, map[string]any{"chain_no": chainNo})
}

func GetLatestVersionID(ctx context.Context, dbObj xdb.Executer, chainNo string) (int64, error) {
	return scalarInt64(ctx, dbObj, sqls.GetLatestVersionID, map[string]any{"chain_no": chainNo})
}

func GetDraftVersionID(ctx context.Context, dbObj xdb.Executer, chainNo string) (int64, error) {
	return scalarInt64(ctx, dbObj, sqls.GetDraftVersionID, map[string]any{"chain_no": chainNo})
}

// NextVersionNo 下一个版本序号，需在锁定流程的事务中调用
func NextVersionNo(ctx context.Context, tx xdb.Executer, chainNo string) (int, error) {
	no, err := scalarInt64(ctx, tx, sqls.NextVersionNo, map[string]any{"chain_no": chainNo})
	return int(no), err
}

// InsertVersion 新增草稿版本，返回 version_id
func InsertVersion(ctx context.Context, tx xdb.Executer, params map[string]any) (int64, error) {
	return scalarInt64(ctx, tx, sqls.InsertVersion, params)
}

func UpdateDraftVersion(ctx context.Context, tx xdb.Executer, params map[string]any) error {
	_, err := tx.Exec(ctx, sqls.UpdateDraftVersion, params)
	return err
}

// ChangeVersionState 执行发布/下线/上线/删除等状态变更，返回受影响行数
func ChangeVersionState(ctx context.Context, tx xdb.Executer, query string, params map[string]any) (int64, error) {
	result, err := tx.Exec(ctx, query, params)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ---------- 节点与连接 ----------

func ListVersionNodes(ctx context.Context, dbObj xdb.Executer, versionID int64) ([]models.DBChainNode, error) {
	list := make([]models.DBChainNode, 0)
	err := dbObj.QueryAs(ctx, sqls.ListVersionNodes, map[string]any{"version_id": versionID}, &list)
	return list, err
}

func ListVersionConnections(ctx context.Context, dbObj xdb.Executer, versionID int64) ([]models.DBChainConnection, error) {
	list := make([]models.DBChainConnection, 0)
	err := dbObj.QueryAs(ctx, sqls.ListVersionConnections, map[string]any{"version_id": versionID}, &list)
	return list, err
}

// ListNodeDefinitions 查询并锁定节点定义；ids 为空时返回空
func ListNodeDefinitions(ctx context.Context, tx xdb.Executer, ids []string) (map[string]models.DBNodeDefinition, error) {
	result := make(map[string]models.DBNodeDefinition)
	if len(ids) == 0 {
		return result, nil
	}
	list := make([]models.DBNodeDefinition, 0, len(ids))
	if err := tx.QueryAs(ctx, sqls.ListNodeDefinitions, map[string]any{"node_def_no": ids}, &list); err != nil {
		return nil, err
	}
	for _, item := range list {
		result[item.NodeDefNo] = item
	}
	return result, nil
}

// ListNodeUsageOutsideVersion 节点定义在其他版本中的引用次数；ids 为空时返回空
func ListNodeUsageOutsideVersion(ctx context.Context, tx xdb.Executer, versionID int64, ids []string) (map[string]int, error) {
	result := make(map[string]int)
	if len(ids) == 0 {
		return result, nil
	}
	list := make([]models.DBNodeUsage, 0)
	if err := tx.QueryAs(ctx, sqls.ListNodeUsageOutsideVersion, map[string]any{"version_id": versionID, "node_def_no": ids}, &list); err != nil {
		return nil, err
	}
	for _, item := range list {
		result[item.NodeDefNo] = item.Total
	}
	return result, nil
}

func InsertNodeDefinition(ctx context.Context, tx xdb.Executer, params map[string]any) error {
	_, err := tx.Exec(ctx, sqls.InsertNodeDefinition, params)
	return err
}

func UpdateNodeDefinition(ctx context.Context, tx xdb.Executer, params map[string]any) error {
	_, err := tx.Exec(ctx, sqls.UpdateNodeDefinition, params)
	return err
}

// ClearVersionGraph 删除版本的节点与连接
func ClearVersionGraph(ctx context.Context, tx xdb.Executer, versionID int64) error {
	params := map[string]any{"version_id": versionID}
	if _, err := tx.Exec(ctx, sqls.DeleteVersionConnections, params); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, sqls.DeleteVersionNodes, params)
	return err
}

func InsertChainNode(ctx context.Context, tx xdb.Executer, params map[string]any) error {
	_, err := tx.Exec(ctx, sqls.InsertChainNode, params)
	return err
}

func InsertChainConnection(ctx context.Context, tx xdb.Executer, params map[string]any) error {
	_, err := tx.Exec(ctx, sqls.InsertChainConnection, params)
	return err
}

func DeleteOrphanNodeDefinitions(ctx context.Context, tx xdb.Executer, chainNo string) error {
	_, err := tx.Exec(ctx, sqls.DeleteOrphanNodeDefinitions, map[string]any{"chain_no": chainNo})
	return err
}

func UpdateDraftVersionName(ctx context.Context, tx xdb.Executer, chainNo, name string) error {
	_, err := tx.Exec(ctx, sqls.UpdateDraftVersionName, map[string]any{"chain_no": chainNo, "name": name})
	return err
}
