// Package webembed 将嵌入的前端构建产物注册为 glue API 路由，支持单页应用路由回退到 index.html。
package webembed

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/zhiyunliu/glue/context"
	"github.com/zhiyunliu/glue/engine"
)

// Response 实现 glue 的 ResponseEntity，直接输出状态码、响应头与内容
type Response struct {
	Code    int
	Headers map[string]string
	Data    []byte
}

func (r Response) StatusCode() int                 { return r.Code }
func (r Response) Header() map[string]string       { return r.Headers }
func (r Response) Body() (bytes []byte, err error) { return r.Data, nil }

// Server 解耦 api.Server
type Server interface {
	Handle(path string, handler interface{}, opts ...engine.RouterOption)
}

// Register 在 webRoot 下注册前端页面，dist 为前端构建产物目录(index.html 位于其根目录)
func Register(server Server, webRoot string, dist fs.FS) {
	webRoot = strings.TrimRight(webRoot, "/")
	handler := NewHandler(webRoot, dist)
	server.Handle(webRoot+"/*any", handler, engine.MethodGet)
}

// NewHandler 创建前端页面处理器
func NewHandler(webRoot string, dist fs.FS) func(ctx context.Context) any {
	return func(ctx context.Context) any {
		reqPath := ctx.Request().Path().GetURL().Path
		filePath := strings.TrimPrefix(strings.TrimPrefix(reqPath, webRoot), "/")
		filePath = strings.TrimPrefix(path.Clean("/"+filePath), "/")
		if filePath == "" {
			filePath = "index.html"
		}
		return serveFile(dist, filePath)
	}
}

func serveFile(dist fs.FS, filePath string) Response {
	content, err := fs.ReadFile(dist, filePath)
	if err != nil {
		// 带扩展名的资源不存在时返回 404，其他路径视为前端路由，返回 index.html
		if path.Ext(filePath) != "" && filePath != "index.html" {
			return Response{Code: http.StatusNotFound, Headers: map[string]string{"Content-Type": "text/plain; charset=utf-8"}, Data: []byte("File not found")}
		}
		filePath = "index.html"
		content, err = fs.ReadFile(dist, filePath)
		if err != nil {
			return Response{
				Code:    http.StatusNotFound,
				Headers: map[string]string{"Content-Type": "text/plain; charset=utf-8"},
				Data:    []byte("前端尚未构建，请在 web 目录执行 npm install && npm run build"),
			}
		}
	}
	headers := map[string]string{"Content-Type": contentType(filePath)}
	if filePath == "index.html" {
		headers["Cache-Control"] = "no-cache"
	} else if strings.HasPrefix(filePath, "assets/") {
		// vite 构建产物文件名带内容哈希，可长期缓存
		headers["Cache-Control"] = "public, max-age=31536000, immutable"
	}
	return Response{Code: http.StatusOK, Headers: headers, Data: content}
}

func contentType(filePath string) string {
	switch strings.ToLower(path.Ext(filePath)) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "application/javascript; charset=utf-8"
	case ".json", ".map":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".wasm":
		return "application/wasm"
	default:
		return "application/octet-stream"
	}
}
