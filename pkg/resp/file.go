package resp

import (
	"fmt"
	"net/http"
	"net/url"
)

// FileResponse 文件下载响应，实现 glue ResponseEntity
type FileResponse struct {
	Name        string
	ContentType string
	Data        []byte
}

func (f FileResponse) StatusCode() int { return http.StatusOK }

func (f FileResponse) Header() map[string]string {
	return map[string]string{
		"Content-Type":        f.ContentType,
		"Content-Disposition": fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", f.Name, url.PathEscape(f.Name)),
	}
}

func (f FileResponse) Body() ([]byte, error) { return f.Data, nil }

// File 返回文件下载
func File(name, contentType string, data []byte) any {
	return FileResponse{Name: name, ContentType: contentType, Data: data}
}
