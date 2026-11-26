package nodetype

import (
	"context"
	"strings"

	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
)

func (svc *NodeTyperImpl) List(ctx context.Context, params models.ListNodeTypeParams) ([]models.NodeTypeInfo, error) {
	rows, err := dbListNodeTypes(ctx)
	if err != nil {
		svc.logger.Errorf("list node types failed, err:%+v", err)
		return nil, err
	}
	list := make([]models.NodeTypeInfo, 0, len(rows))
	for _, row := range rows {
		info := toInfo(row)
		if params.FlowMode != "" && !containsString(info.FlowModes, params.FlowMode) {
			continue
		}
		if params.Status != "" && params.Status != itoa(info.Status) {
			continue
		}
		list = append(list, info)
	}
	return list, nil
}

func (svc *NodeTyperImpl) Page(ctx context.Context, params models.PageNodeTypeParams) ([]models.NodeTypeInfo, int64, error) {
	offset := params.Normalize()
	query := map[string]any{
		"offset":    offset,
		"page_size": params.PageSize,
		"keyword":   strings.TrimSpace(params.Keyword),
	}
	if params.Category != "" {
		query["category"] = params.Category
	}
	if params.Source != "" {
		query["source"] = params.Source
	}
	if params.FlowMode != "" {
		query["flow_modes"] = params.FlowMode
	}
	if params.Status == "0" || params.Status == "1" {
		query["status"] = params.Status
	}
	rows, total, err := dbPageNodeTypes(ctx, query, params.PageSize)
	if err != nil {
		svc.logger.Errorf("page node types failed, params:%+v, err:%+v", params, err)
		return nil, 0, err
	}
	list := make([]models.NodeTypeInfo, 0, len(rows))
	for _, row := range rows {
		list = append(list, toInfo(row))
	}
	return list, total, nil
}

func (svc *NodeTyperImpl) Detail(ctx context.Context, nodeType string) (models.NodeTypeInfo, error) {
	row, err := svc.retrieve(ctx, nodeType)
	if err != nil {
		return models.NodeTypeInfo{}, err
	}
	return toInfo(row), nil
}

func (svc *NodeTyperImpl) retrieve(ctx context.Context, nodeType string) (models.DBNodeType, error) {
	row, err := dbGetNodeType(ctx, nodeType)
	if err != nil {
		svc.logger.Errorf("get node type failed, node_type:%s, err:%+v", nodeType, err)
		return row, err
	}
	if row.NodeType == "" {
		return row, bizerr.NotFound(subcode.NodeTypeNotFound, "节点类型 %s 不存在", nodeType)
	}
	return row, nil
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func itoa(val int) string {
	if val == 1 {
		return "1"
	}
	if val == 0 {
		return "0"
	}
	return ""
}
