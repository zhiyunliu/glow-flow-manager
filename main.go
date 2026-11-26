package main

import (
	"context"

	"github.com/zhiyunliu/manager-glow/config"
	"github.com/zhiyunliu/manager-glow/modules/nodetype"
	"github.com/zhiyunliu/manager-glow/services"

	"github.com/zhiyunliu/glue"
	_ "github.com/zhiyunliu/glue/contrib/metrics/prometheus"
	_ "github.com/zhiyunliu/glue/contrib/registry/nacos"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/log"
	"github.com/zhiyunliu/glue/server/api"
	_ "github.com/zhiyunliu/xdb-mssql"
)

func main() {
	global.AppName = "manager-glow"

	apiSrv := api.New("apiserver", api.WithServiceName(global.AppName), api.Log(log.WithRequest()))
	services.BindAPI(apiSrv)

	app := glue.NewApp(glue.Server(apiSrv), glue.StartingHook(func(ctx context.Context) error {
		if err := config.Load(global.Config); err != nil {
			return err
		}
		// 补齐内置节点类型，数据库暂不可用时不阻止服务启动
		nodetype.EnsureBuiltin(ctx, log.DefaultLogger)
		return nil
	}))
	_ = app.Start()
}
