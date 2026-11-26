package respcode

// 统一响应码：成功为 0；错误码前三位与 HTTP 状态码一致
const (
	Success       = 0
	BadRequest    = 400001 // 参数错误
	Unauthorized  = 401001 // 未登录
	NotFound      = 404001 // 数据不存在
	Conflict      = 409001 // 状态冲突、数据重复
	Unprocessable = 422001 // 模型校验失败
	SystemError   = 500001 // 系统错误
	Unavailable   = 503001 // 依赖的外部服务不可用
)
