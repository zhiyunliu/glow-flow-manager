// Package test 集成测试：验证管理后台写入的流程定义能被 glow-flow 引擎直接加载与执行。
//
// 需要可访问的 SQL Server(见 config.json)，并已通过管理后台发布流程，默认跳过：
//
//	GLOWFLOW_IT=1 GLOWFLOW_DB_PASSWORD=xxx GLOWFLOW_IT_CHAIN=temp_alarm go test ./test/ -v
package test

import (
	"context"
	"os"
	"testing"

	glowflow "github.com/zhiyunliu/glow-flow"
	xdbrepo "github.com/zhiyunliu/glow-flow/contrib/repository/xdb"
	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/config/file"
	_ "github.com/zhiyunliu/glue/contrib/xhttp/http"
	"github.com/zhiyunliu/glue/global"
	_ "github.com/zhiyunliu/xdb-mssql"

	// 引擎宿主需要注册的内置节点
	_ "github.com/zhiyunliu/glow-flow/contrib/nodedescriptor/action"
	_ "github.com/zhiyunliu/glow-flow/contrib/nodedescriptor/filter"
	_ "github.com/zhiyunliu/glow-flow/contrib/nodedescriptor/logical"
	_ "github.com/zhiyunliu/glow-flow/contrib/nodedescriptor/transform"
)

func TestEngineLoadsPublishedChain(t *testing.T) {
	if os.Getenv("GLOWFLOW_IT") != "1" {
		t.Skip("set GLOWFLOW_IT=1 to run the SQL Server integration test")
	}
	chainNo := os.Getenv("GLOWFLOW_IT_CHAIN")
	if chainNo == "" {
		chainNo = "temp_alarm"
	}
	// 引擎宿主的 http 客户端配置(httpcall 节点使用)
	cfg := config.New(config.WithSource(file.NewSource("../config.json"), config.NewStrSource(`{"xhttp":{"default":{}}}`)))
	if err := cfg.Load(); err != nil {
		t.Fatalf("load config: %v", err)
	}
	global.Config = cfg

	ctx := context.Background()
	repo := xdbrepo.NewRepository("glowflow")
	def, err := repo.LoadChainDefinition(ctx, chainNo)
	if err != nil {
		t.Fatalf("engine repository failed to load %s: %v", chainNo, err)
	}
	t.Logf("loaded %s version %s: %d nodes, %d connections", def.ID, def.Version, len(def.Nodes), len(def.Connections))
	for _, node := range def.Nodes {
		t.Logf("  node %s type=%s name=%s extparams=%v", node.ID, node.NodeType, node.NodeName, node.ExtParams)
	}
	if len(def.Nodes) == 0 {
		t.Fatalf("published chain has no nodes")
	}

	all, err := repo.LoadChainDefinitions(ctx)
	if err != nil {
		t.Fatalf("engine repository failed to load all chains: %v", err)
	}
	t.Logf("engine would load %d active chains", len(all))

	engine := glowflow.NewEngine()
	if err := engine.Load(*def); err != nil {
		t.Fatalf("engine failed to compile published chain: %v", err)
	}
	for _, data := range []map[string]any{
		{"temperature": 20, "device": "D-02"},
		{"temperature": 61, "device": "D-01"},
	} {
		instanceID, err := engine.Dispatch(ctx, chainNo, data)
		if err != nil {
			t.Fatalf("dispatch %v failed: %v", data, err)
		}
		t.Logf("dispatched %v -> instance %s", data, instanceID)
	}
}
