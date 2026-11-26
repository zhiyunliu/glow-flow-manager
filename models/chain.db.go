package models

// DBChainInfo flow_chain_definition 记录及版本汇总
type DBChainInfo struct {
	ChainNo    string `json:"chain_no"`
	Name       string `json:"name"`
	FlowType   string `json:"flow_type"`
	Status     int    `json:"status"`
	ExtParams  string `json:"extparams"`
	Layout     string `json:"layout"`
	Desc       string `json:"desc"`
	CreateTime string `json:"create_time"`
	CreateBy   string `json:"create_by"`
	Operator   string `json:"operator"`
	UpdateTime string `json:"update_time"`

	EffectiveVersionID int64 `json:"effective_version_id"` // 生效版本：version_id 最大的已发布版本
	EffectiveVersionNo int   `json:"effective_version_no"`
	DraftVersionID     int64 `json:"draft_version_id"`
	DraftVersionNo     int   `json:"draft_version_no"`
	LatestVersionNo    int   `json:"latest_version_no"`
	VersionCount       int   `json:"version_count"`
	PublishedCount     int   `json:"published_count"` // 发布过的版本数(已发布 + 已下线)
}
