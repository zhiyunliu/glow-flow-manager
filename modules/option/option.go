// Package option 设计器动态下拉选项的数据源。
//
// 节点属性字段通过 SchemaField.OptionsSource 引用数据源名称，前端调用 /option/list?source= 获取选项。
// 更换数据来源(如改为调用外部接口、查询数据库)时，实现 Source 接口并以相同名称 Register 覆盖默认实现即可，
// 例如在 main 中：
//
//	option.Register(option.SourceEmailServers, option.SourceFunc(func(ctx context.Context) ([]models.Option, error) {
//		// 调用外部接口获取 SMTP 服务器列表
//	}))
package option

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
)

// Source 下拉选项数据源
type Source interface {
	Options(ctx context.Context) ([]models.Option, error)
}

// SourceFunc 函数形式的 Source
type SourceFunc func(ctx context.Context) ([]models.Option, error)

func (f SourceFunc) Options(ctx context.Context) ([]models.Option, error) {
	return f(ctx)
}

var (
	mu      sync.RWMutex
	sources = map[string]Source{}
)

// Register 注册下拉数据源，同名时覆盖(用于更换数据来源)；src 为 nil 时移除
func Register(name string, src Source) {
	mu.Lock()
	defer mu.Unlock()
	if src == nil {
		delete(sources, name)
		return
	}
	sources[name] = src
}

// Get 获取已注册的下拉数据源
func Get(name string) (Source, bool) {
	mu.RLock()
	defer mu.RUnlock()
	src, ok := sources[name]
	return src, ok
}

// Names 已注册的数据源名称
func Names() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(sources))
	for name := range sources {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// List 获取指定数据源的下拉选项：去除空值及重复值
func List(ctx context.Context, name string) ([]models.Option, error) {
	name = strings.TrimSpace(name)
	src, ok := Get(name)
	if !ok {
		return nil, bizerr.NotFound(subcode.OptionSourceNotFound, "下拉数据源 %s 不存在", name)
	}
	items, err := src.Options(ctx)
	if err != nil {
		return nil, bizerr.Unavailable(subcode.OptionSourceUnavailable, "获取下拉数据源 %s 失败：%v", name, err)
	}
	list := make([]models.Option, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.Value = strings.TrimSpace(item.Value); item.Value == "" {
			continue
		}
		if _, ok := seen[item.Value]; ok {
			continue
		}
		seen[item.Value] = struct{}{}
		if strings.TrimSpace(item.Label) == "" {
			item.Label = item.Value
		}
		list = append(list, item)
	}
	return list, nil
}

// Values 指定数据源的可选值集合，用于校验
func Values(ctx context.Context, name string) (map[string]struct{}, error) {
	list, err := List(ctx, name)
	if err != nil {
		return nil, err
	}
	values := make(map[string]struct{}, len(list))
	for _, item := range list {
		values[item.Value] = struct{}{}
	}
	return values, nil
}
