package models

const (
	defaultPageSize = 20
	maxPageSize     = 200
)

// PageParams 分页参数
type PageParams struct {
	PageNum  int `form:"page_num" json:"page_num"`
	PageSize int `form:"page_size" json:"page_size"`
}

// Normalize 修正分页参数并返回偏移量
func (p *PageParams) Normalize() (offset int) {
	if p.PageNum <= 0 {
		p.PageNum = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = defaultPageSize
	}
	if p.PageSize > maxPageSize {
		p.PageSize = maxPageSize
	}
	return (p.PageNum - 1) * p.PageSize
}

// TotalRow 统计行数
type TotalRow struct {
	Total int64 `json:"total"`
}
