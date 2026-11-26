// Package bizerr 构造带 HTTP 状态码、业务码、子码的业务错误，由 glue 默认错误编码器输出为
// {"code":..., "sub_code":..., "message":..., "data":...}。
package bizerr

import (
	"fmt"
	"net/http"

	"github.com/zhiyunliu/glue/errors"
	"github.com/zhiyunliu/manager-glow/constants/respcode"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
)

// New 创建业务错误
func New(httpStatus, code int, subCode, message string, opts ...errors.Option) errors.Error {
	opts = append([]errors.Option{errors.WithSubCode(subCode), errors.WithStatusCode(httpStatus)}, opts...)
	return errors.New(code, message, opts...)
}

// BadRequest 参数错误(400)
func BadRequest(subCode, format string, args ...any) errors.Error {
	return New(http.StatusBadRequest, respcode.BadRequest, subCode, fmt.Sprintf(format, args...))
}

// NotFound 数据不存在(404)
func NotFound(subCode, format string, args ...any) errors.Error {
	return New(http.StatusNotFound, respcode.NotFound, subCode, fmt.Sprintf(format, args...))
}

// Conflict 状态冲突或数据重复(409)
func Conflict(subCode, format string, args ...any) errors.Error {
	return New(http.StatusConflict, respcode.Conflict, subCode, fmt.Sprintf(format, args...))
}

// Unprocessable 模型校验失败(422)，data 为校验结果
func Unprocessable(subCode, message string, data any) errors.Error {
	return New(http.StatusUnprocessableEntity, respcode.Unprocessable, subCode, message, errors.WithData(data))
}

// Unavailable 依赖的外部服务不可用(503)
func Unavailable(subCode, format string, args ...any) errors.Error {
	return New(http.StatusServiceUnavailable, respcode.Unavailable, subCode, fmt.Sprintf(format, args...))
}

// System 系统错误(500)，不向调用方暴露内部错误信息
func System(err error) errors.Error {
	return New(http.StatusInternalServerError, respcode.SystemError, subcode.SystemError, "系统繁忙，请稍后重试", errors.WithInnerErr(err))
}

// IsBiz 判断是否为已经构造好的业务错误
func IsBiz(err error) (errors.Error, bool) {
	var xerr errors.Error
	if errors.As(err, &xerr) {
		return xerr, true
	}
	return nil, false
}

// Message 返回适合展示的错误消息
func Message(err error) string {
	if err == nil {
		return ""
	}
	if xerr, ok := IsBiz(err); ok {
		return xerr.GetMessage()
	}
	return err.Error()
}
