package models

// PageAuditParams 审计日志分页查询
type PageAuditParams struct {
	PageParams
	Module      string `form:"module"`
	OperateType string `form:"operate_type"`
	TargetID    string `form:"target_id"`
	ChainNo     string `form:"chain_no"`
	Operator    string `form:"operator"`
	Result      string `form:"result"` // 空为全部，0 失败，1 成功
	StartTime   string `form:"start_time"`
	EndTime     string `form:"end_time"`
}

// AuditIDParams 按日志 ID 查询
type AuditIDParams struct {
	LogID int64 `form:"log_id" binding:"required"`
}

// DBOperateLog flow_operate_log 记录
type DBOperateLog struct {
	LogID       int64  `json:"log_id"`
	TraceID     string `json:"trace_id"`
	Module      string `json:"module"`
	OperateType string `json:"operate_type"`
	TargetID    string `json:"target_id"`
	ChainNo     string `json:"chain_no"`
	VersionID   int64  `json:"version_id"`
	Summary     string `json:"summary"`
	BeforeData  string `json:"before_data,omitempty"`
	AfterData   string `json:"after_data,omitempty"`
	Result      int    `json:"result"`
	ErrorMsg    string `json:"error_msg"`
	Operator    string `json:"operator"`
	ClientIP    string `json:"client_ip"`
	CreateTime  string `json:"create_time"`
}
