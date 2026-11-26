package services

import (
	"github.com/zhiyunliu/glue/engine"
	"github.com/zhiyunliu/glue/server/api"
	"github.com/zhiyunliu/manager-glow/pkg/webembed"
	"github.com/zhiyunliu/manager-glow/services/audit"
	"github.com/zhiyunliu/manager-glow/services/chain"
	"github.com/zhiyunliu/manager-glow/services/history"
	"github.com/zhiyunliu/manager-glow/services/nodetype"
	"github.com/zhiyunliu/manager-glow/services/option"
	"github.com/zhiyunliu/manager-glow/services/version"
	distweb "github.com/zhiyunliu/manager-glow/web"
)

// APIRoot 接口根路径；前端页面挂载在 APIRoot/web
const APIRoot = "/api/manager-glow"

// BindAPI 注册接口。登录态与权限由独立网关负责，查询使用 GET，变更使用 POST。
func BindAPI(server *api.Server) {
	chainGroup := server.Group(APIRoot + "/chain") // 流程定义
	chainGroup.Handle("/page", chain.PageChain, engine.MethodGet)
	chainGroup.Handle("/detail", chain.DetailChain, engine.MethodGet)
	chainGroup.Handle("/create", chain.CreateChain, engine.MethodPost)
	chainGroup.Handle("/update", chain.UpdateChain, engine.MethodPost)
	chainGroup.Handle("/change-status", chain.ChangeChainStatus, engine.MethodPost)
	chainGroup.Handle("/delete", chain.DeleteChain, engine.MethodPost)
	chainGroup.Handle("/copy", chain.CopyChain, engine.MethodPost)
	chainGroup.Handle("/export", chain.ExportChain, engine.MethodGet)
	chainGroup.Handle("/import", chain.ImportChain, engine.MethodPost)

	versionGroup := server.Group(APIRoot + "/version") // 流程版本与设计器
	versionGroup.Handle("/list", version.ListVersion, engine.MethodGet)
	versionGroup.Handle("/detail", version.DetailVersion, engine.MethodGet)
	versionGroup.Handle("/create-draft", version.CreateDraft, engine.MethodPost)
	versionGroup.Handle("/save-draft", version.SaveDraft, engine.MethodPost)
	versionGroup.Handle("/validate", version.ValidateModel, engine.MethodPost)
	versionGroup.Handle("/publish", version.PublishVersion, engine.MethodPost)
	versionGroup.Handle("/offline", version.OfflineVersion, engine.MethodPost)
	versionGroup.Handle("/online", version.OnlineVersion, engine.MethodPost)
	versionGroup.Handle("/delete", version.DeleteVersion, engine.MethodPost)
	versionGroup.Handle("/debug", version.DebugModel, engine.MethodPost)

	nodeTypeGroup := server.Group(APIRoot + "/node-type") // 节点注册
	nodeTypeGroup.Handle("/list", nodetype.ListNodeType, engine.MethodGet)
	nodeTypeGroup.Handle("/page", nodetype.PageNodeType, engine.MethodGet)
	nodeTypeGroup.Handle("/detail", nodetype.DetailNodeType, engine.MethodGet)
	nodeTypeGroup.Handle("/create", nodetype.CreateNodeType, engine.MethodPost)
	nodeTypeGroup.Handle("/update", nodetype.UpdateNodeType, engine.MethodPost)
	nodeTypeGroup.Handle("/change-status", nodetype.ChangeNodeTypeStatus, engine.MethodPost)
	nodeTypeGroup.Handle("/delete", nodetype.DeleteNodeType, engine.MethodPost)
	nodeTypeGroup.Handle("/sync-builtin", nodetype.SyncBuiltinNodeType, engine.MethodPost)
	nodeTypeGroup.Handle("/instances", nodetype.NodeTypeInstances, engine.MethodGet)

	optionGroup := server.Group(APIRoot + "/option") // 设计器动态下拉选项
	optionGroup.Handle("/list", option.ListOption, engine.MethodGet)

	historyGroup := server.Group(APIRoot + "/history") // 执行历史
	historyGroup.Handle("/page", history.PageInstance, engine.MethodGet)
	historyGroup.Handle("/detail", history.DetailInstance, engine.MethodGet)

	auditGroup := server.Group(APIRoot + "/audit") // 操作审计
	auditGroup.Handle("/page", audit.PageAudit, engine.MethodGet)
	auditGroup.Handle("/detail", audit.DetailAudit, engine.MethodGet)

	webembed.Register(server, APIRoot+"/web", distweb.Dist())
}
