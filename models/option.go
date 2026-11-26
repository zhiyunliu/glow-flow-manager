package models

// Option 动态下拉选项
type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Desc  string `json:"desc,omitempty"` // 辅助说明，如邮件服务器地址
}

// ListOptionParams 获取下拉选项
type ListOptionParams struct {
	Source string `form:"source" json:"source" binding:"required"` // 数据源名称，如 email_servers
}
