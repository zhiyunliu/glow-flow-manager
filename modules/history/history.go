package history

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	glowflow "github.com/zhiyunliu/glow-flow"
	"github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/manager-glow/config"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/flow/dao"
	"github.com/zhiyunliu/manager-glow/modules/flow/store"
	"github.com/zhiyunliu/manager-glow/modules/history/sqls"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
	"github.com/zhiyunliu/manager-glow/pkg/timefmt"
)

// Historian 执行历史：读取 glow-flow 引擎写入的实例状态与节点状态
type Historian interface {
	// Page 分页查询执行实例
	Page(ctx context.Context, params models.PageInstanceParams) ([]models.InstanceInfo, int64, error)
	// Detail 执行实例详情，包含节点状态与执行版本的流程图
	Detail(ctx context.Context, instanceID string) (models.InstanceDetail, error)
}

type HistorianImpl struct {
	logger log.Logger
}

func NewHistorian(logger log.Logger) Historian {
	return &HistorianImpl{logger: logger}
}

func (svc *HistorianImpl) Page(ctx context.Context, params models.PageInstanceParams) ([]models.InstanceInfo, int64, error) {
	offset := params.Normalize()
	query := map[string]any{"offset": offset, "page_size": params.PageSize}
	setIfNotEmpty(query, "chain_id", params.ChainNo)
	setIfNotEmpty(query, "version", params.Version)
	setIfNotEmpty(query, "status", params.Status)
	setIfNotEmpty(query, "instance_id", params.InstanceID)
	for key, text := range map[string]string{"start_time": params.StartTime, "end_time": params.EndTime} {
		if strings.TrimSpace(text) == "" {
			continue
		}
		val, err := timefmt.ToSQL(text)
		if err != nil {
			return nil, 0, bizerr.BadRequest(subcode.ParamInvalid, "时间格式错误：%s", text)
		}
		query[key] = val
	}

	dbObj := glue.DB(config.Sys.DBName)
	total := models.TotalRow{}
	if err := dbObj.FirstAs(ctx, sqls.CountInstances, query, &total); err != nil {
		svc.logger.Errorf("count instances failed, params:%+v, err:%+v", params, err)
		return nil, 0, err
	}
	list := make([]models.InstanceInfo, 0, params.PageSize)
	if total.Total == 0 {
		return list, 0, nil
	}
	rows := make([]models.DBInstanceState, 0, params.PageSize)
	if err := dbObj.QueryAs(ctx, sqls.PageInstances, query, &rows); err != nil {
		svc.logger.Errorf("page instances failed, params:%+v, err:%+v", params, err)
		return nil, 0, err
	}
	for _, row := range rows {
		info := toInstanceInfo(row)
		// 列表不返回变量，避免数据过大
		info.Variables = nil
		list = append(list, info)
	}
	return list, total.Total, nil
}

func (svc *HistorianImpl) Detail(ctx context.Context, instanceID string) (models.InstanceDetail, error) {
	detail := models.InstanceDetail{Nodes: []models.NodeStateInfo{}, DesignNodes: []models.DesignNode{}, DesignConns: []models.DesignConnection{}}
	dbObj := glue.DB(config.Sys.DBName)
	row := models.DBInstanceState{}
	if err := dbObj.FirstAs(ctx, sqls.GetInstance, map[string]any{"instance_id": instanceID}, &row); err != nil {
		svc.logger.Errorf("get instance failed, instance_id:%s, err:%+v", instanceID, err)
		return detail, err
	}
	if row.InstanceID == "" {
		return detail, bizerr.NotFound(subcode.InstanceNotFound, "执行实例 %s 不存在", instanceID)
	}
	detail.Instance = toInstanceInfo(row)

	// 执行版本的流程图，用于展示节点名称及执行路径
	nodeMeta := map[string]models.DesignNode{}
	if versionID, err := strconv.ParseInt(row.Version, 10, 64); err == nil && versionID > 0 {
		nodes, conns, err := store.LoadDesign(ctx, dbObj, versionID)
		if err != nil {
			svc.logger.Warnf("load instance design failed, instance_id:%s, version:%s, err:%v", instanceID, row.Version, err)
		} else {
			detail.DesignNodes, detail.DesignConns = nodes, conns
			for _, node := range nodes {
				nodeMeta[node.ID] = node
			}
		}
	}
	if chain, err := dao.GetChain(ctx, dbObj, row.ChainID); err == nil {
		detail.FlowType = chain.FlowType
	}

	nodeRows := make([]models.DBNodeState, 0)
	if err := dbObj.QueryAs(ctx, sqls.ListNodeStates, map[string]any{"instance_id": instanceID}, &nodeRows); err != nil {
		svc.logger.Errorf("list node states failed, instance_id:%s, err:%+v", instanceID, err)
		return detail, err
	}
	for _, nodeRow := range nodeRows {
		info := toNodeStateInfo(nodeRow)
		if node, ok := nodeMeta[info.NodeID]; ok {
			info.NodeName, info.NodeType = node.Name, node.Type
		}
		detail.Nodes = append(detail.Nodes, info)
	}
	return detail, nil
}

func toInstanceInfo(row models.DBInstanceState) models.InstanceInfo {
	info := models.InstanceInfo{
		InstanceID: row.InstanceID,
		ChainNo:    row.ChainID,
		ChainName:  row.ChainName,
		Version:    row.Version,
		VersionNo:  row.VersionNo,
		Status:     row.Status,
		CreatedAt:  row.CreateTime,
		UpdatedAt:  row.UpdateTime,
	}
	state := glowflow.InstanceState{}
	if err := json.Unmarshal([]byte(row.StateData), &state); err != nil {
		return info
	}
	info.CurrentNodeIDs = state.CurrentNodeIDs
	info.Variables = state.Variables
	info.Error = state.Error
	if !state.CreatedAt.IsZero() {
		info.CreatedAt = timefmt.Format(state.CreatedAt)
	}
	if !state.UpdatedAt.IsZero() {
		info.UpdatedAt = timefmt.Format(state.UpdatedAt)
	}
	if state.FinishedAt != nil && !state.FinishedAt.IsZero() {
		info.FinishedAt = timefmt.Format(*state.FinishedAt)
		if !state.CreatedAt.IsZero() {
			info.DurationMs = state.FinishedAt.Sub(state.CreatedAt).Milliseconds()
		}
	}
	return info
}

func toNodeStateInfo(row models.DBNodeState) models.NodeStateInfo {
	info := models.NodeStateInfo{NodeID: row.NodeID, Status: row.Status, Attempts: row.Attempts, StartedAt: row.CreateTime}
	state := glowflow.NodeState{}
	if err := json.Unmarshal([]byte(row.StateData), &state); err != nil {
		return info
	}
	info.Input = state.Input
	info.Output = state.Output
	info.Metadata = state.Metadata
	info.Error = state.Error
	info.StartedAt = formatTimePtr(state.StartedAt)
	info.FinishedAt = formatTimePtr(state.FinishedAt)
	if state.StartedAt != nil && state.FinishedAt != nil {
		info.DurationMs = state.FinishedAt.Sub(*state.StartedAt).Milliseconds()
	}
	return info
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return timefmt.Format(*t)
}

func setIfNotEmpty(query map[string]any, key, val string) {
	if val = strings.TrimSpace(val); val != "" {
		query[key] = val
	}
}
