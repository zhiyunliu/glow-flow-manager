package audit

import (
	"context"
	"encoding/json"
	"unicode/utf8"

	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
)

const (
	maxSummaryLen  = 512
	maxErrorMsgLen = 1024
	maxTargetIDLen = 64
)

func (a *AuditorImpl) Record(ctx context.Context, op operator.Operator, entry Entry) {
	result := enums.AuditResultSuccess
	errMsg := ""
	if entry.Err != nil {
		result = enums.AuditResultFail
		errMsg = truncate(bizerr.Message(entry.Err), maxErrorMsgLen)
	}
	params := map[string]any{
		"trace_id":     op.TraceID,
		"module":       entry.Module,
		"operate_type": entry.OperateType,
		"target_id":    truncate(entry.TargetID, maxTargetIDLen),
		"chain_no":     entry.ChainNo,
		"version_id":   entry.VersionID,
		"summary":      truncate(entry.Summary, maxSummaryLen),
		"before_data":  toJSON(entry.Before),
		"after_data":   toJSON(entry.After),
		"result":       result,
		"error_msg":    errMsg,
		"operator":     op.ID,
		"client_ip":    op.ClientIP,
	}
	// 请求取消(如客户端断开)后仍需写入审计
	if err := dbInsertOperateLog(context.WithoutCancel(ctx), params); err != nil {
		a.logger.Errorf("record operate log failed, module:%s, operate:%s, target:%s, err:%+v", entry.Module, entry.OperateType, entry.TargetID, err)
	}
}

func toJSON(val any) string {
	if val == nil {
		return ""
	}
	if text, ok := val.(string); ok {
		return text
	}
	data, err := json.Marshal(val)
	if err != nil {
		return ""
	}
	if string(data) == "null" {
		return ""
	}
	return string(data)
}

func truncate(text string, max int) string {
	if utf8.RuneCountInString(text) <= max {
		return text
	}
	runes := []rune(text)
	return string(runes[:max])
}
