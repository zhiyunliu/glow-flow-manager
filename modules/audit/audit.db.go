package audit

import (
	"context"

	"github.com/zhiyunliu/glue"
	"github.com/zhiyunliu/manager-glow/config"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/audit/sqls"
)

func dbInsertOperateLog(ctx context.Context, params map[string]any) error {
	_, err := glue.DB(config.Sys.DBName).Exec(ctx, sqls.InsertOperateLog, params)
	return err
}

func dbPageOperateLog(ctx context.Context, query map[string]any, pageSize int) ([]models.DBOperateLog, int64, error) {
	dbObj := glue.DB(config.Sys.DBName)
	total := models.TotalRow{}
	if err := dbObj.FirstAs(ctx, sqls.CountOperateLog, query, &total); err != nil {
		return nil, 0, err
	}
	list := make([]models.DBOperateLog, 0, pageSize)
	if total.Total == 0 {
		return list, 0, nil
	}
	err := dbObj.QueryAs(ctx, sqls.PageOperateLog, query, &list)
	return list, total.Total, err
}

func dbGetOperateLog(ctx context.Context, logID int64) (models.DBOperateLog, error) {
	item := models.DBOperateLog{}
	err := glue.DB(config.Sys.DBName).FirstAs(ctx, sqls.GetOperateLog, map[string]any{"log_id": logID}, &item)
	return item, err
}
