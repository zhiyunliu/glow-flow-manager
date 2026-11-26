// Package operator 读取独立网关注入的操作人信息。
// 登录态与权限由网关负责，管理后台只记录操作人用于审计，不做本地权限判断。
package operator

import (
	"strings"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/manager-glow/config"
)

// Operator 当前请求的操作人
type Operator struct {
	ID       string `json:"id"`        // 操作人标识
	ClientIP string `json:"client_ip"` // 客户端 IP
	TraceID  string `json:"trace_id"`  // 请求链路 ID
}

// FromContext 从请求头中读取操作人，未携带时使用配置的默认操作人
func FromContext(ctx context.Context) Operator {
	req := ctx.Request()
	op := Operator{
		ClientIP: req.GetClientIP(),
		TraceID:  req.RequestID(),
	}
	for _, header := range config.Sys.OperatorHeaders {
		if val := strings.TrimSpace(req.GetHeader(header)); val != "" {
			op.ID = val
			break
		}
	}
	if op.ID == "" {
		op.ID = config.Sys.DefaultOperator
	}
	// 操作人字段在库中为 varchar(32)
	if len(op.ID) > 32 {
		op.ID = op.ID[:32]
	}
	return op
}

// System 系统内部操作(如启动时补齐内置节点)
func System() Operator {
	return Operator{ID: "system"}
}
