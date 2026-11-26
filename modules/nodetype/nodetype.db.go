package nodetype

import (
	"context"

	"github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/manager-glow/config"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/nodetype/sqls"
)

func dbListNodeTypes(ctx context.Context) ([]models.DBNodeType, error) {
	list := make([]models.DBNodeType, 0)
	err := glue.DB(config.Sys.DBName).QueryAs(ctx, sqls.ListNodeTypes, map[string]any{}, &list)
	return list, err
}

func dbPageNodeTypes(ctx context.Context, query map[string]any, pageSize int) ([]models.DBNodeType, int64, error) {
	dbObj := glue.DB(config.Sys.DBName)
	total := models.TotalRow{}
	if err := dbObj.FirstAs(ctx, sqls.CountNodeTypes, query, &total); err != nil {
		return nil, 0, err
	}
	list := make([]models.DBNodeType, 0, pageSize)
	if total.Total == 0 {
		return list, 0, nil
	}
	err := dbObj.QueryAs(ctx, sqls.PageNodeTypes, query, &list)
	return list, total.Total, err
}

func dbGetNodeType(ctx context.Context, nodeType string) (models.DBNodeType, error) {
	item := models.DBNodeType{}
	err := glue.DB(config.Sys.DBName).FirstAs(ctx, sqls.GetNodeType, map[string]any{"node_type": nodeType}, &item)
	return item, err
}

func dbInsertNodeType(ctx context.Context, params map[string]any) error {
	_, err := glue.DB(config.Sys.DBName).Exec(ctx, sqls.InsertNodeType, params)
	return err
}

func dbUpdateNodeType(ctx context.Context, params map[string]any) error {
	_, err := glue.DB(config.Sys.DBName).Exec(ctx, sqls.UpdateNodeType, params)
	return err
}

func dbUpdateBuiltinNodeType(ctx context.Context, params map[string]any) error {
	_, err := glue.DB(config.Sys.DBName).Exec(ctx, sqls.UpdateBuiltinNodeType, params)
	return err
}

func dbUpdateNodeTypeStatus(ctx context.Context, nodeType string, status int, operator string) error {
	_, err := glue.DB(config.Sys.DBName).Exec(ctx, sqls.UpdateNodeTypeStatus, map[string]any{
		"node_type": nodeType,
		"status":    status,
		"operator":  operator,
	})
	return err
}

func dbDeleteNodeType(ctx context.Context, nodeType string) error {
	_, err := glue.DB(config.Sys.DBName).Exec(ctx, sqls.DeleteNodeType, map[string]any{"node_type": nodeType})
	return err
}

func dbListNodeTypeKeys(ctx context.Context) (map[string]struct{}, error) {
	rows := make([]struct {
		NodeType string `json:"node_type"`
	}, 0)
	if err := glue.DB(config.Sys.DBName).QueryAs(ctx, sqls.ListNodeTypeKeys, map[string]any{}, &rows); err != nil {
		return nil, err
	}
	keys := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		keys[row.NodeType] = struct{}{}
	}
	return keys, nil
}
