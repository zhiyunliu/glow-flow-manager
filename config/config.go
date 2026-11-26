package config

import (
	"github.com/zhiyunliu/glue/config"
)

// Config 管理后台业务配置，对应配置文件中的 manager 节点
type Config struct {
	DBName          string   `json:"db_name"`           // glue dbs 中流程库的连接名称
	OperatorHeaders []string `json:"operator_headers"`  // 网关注入的操作人请求头，按顺序取第一个非空值
	DefaultOperator string   `json:"default_operator"`  // 请求中没有操作人时使用的名称
	DebugTimeout    int      `json:"debug_timeout"`     // 调试预览的默认超时时间(秒)
	DebugMaxTimeout int      `json:"debug_max_timeout"` // 调试预览允许的最大超时时间(秒)
	// DebugAllowRealActions 调试时是否允许真实执行动作/集成类节点(发送请求、消息、邮件等)，默认不允许。
	// command 节点在任何情况下都不会在管理后台真实执行。
	DebugAllowRealActions bool `json:"debug_allow_real_actions"`
}

var Sys = &Config{
	DBName:          "glowflow",
	OperatorHeaders: []string{"X-Auth-User-Id", "X-Perm-Uid", "X-Operator"},
	DefaultOperator: "anonymous",
	DebugTimeout:    30,
	DebugMaxTimeout: 120,
}

// Load 从 glue 配置的 manager 节点加载业务配置，未配置的字段保留默认值
func Load(cfg config.Config) error {
	node := cfg.Get("manager")
	if node == nil {
		return nil
	}
	if err := node.ScanTo(Sys); err != nil {
		return err
	}
	if Sys.DBName == "" {
		Sys.DBName = "glowflow"
	}
	if Sys.DebugTimeout <= 0 {
		Sys.DebugTimeout = 30
	}
	if Sys.DebugMaxTimeout < Sys.DebugTimeout {
		Sys.DebugMaxTimeout = Sys.DebugTimeout
	}
	return nil
}
