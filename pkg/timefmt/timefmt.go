// Package timefmt 时间参数转换
package timefmt

import (
	"strings"
	"time"
)

var inputLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02 15:04",
	"2006-01-02",
}

// Parse 解析前端传入的时间
func Parse(text string) (time.Time, error) {
	text = strings.TrimSpace(text)
	var lastErr error
	for _, layout := range inputLayouts {
		t, err := time.ParseInLocation(layout, text, time.Local)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

// ToSQL 转换为 SQL Server 不受语言设置影响的 ISO 8601 格式
func ToSQL(text string) (string, error) {
	t, err := Parse(text)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02T15:04:05"), nil
}

// Format 格式化为展示时间，零值返回空字符串
func Format(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02 15:04:05")
}
