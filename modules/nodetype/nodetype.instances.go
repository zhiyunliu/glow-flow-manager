package nodetype

import (
	"context"
	"sort"
	"time"

	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/glue/registry"
	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
)

// registrarResolver 获取注册中心，测试时可替换
var registrarResolver = func() (registry.Registrar, error) {
	return registry.GetRegistrar(global.Config)
}

func (svc *NodeTyperImpl) Instances(ctx context.Context, nodeType string) (models.NodeTypeInstances, error) {
	row, err := svc.retrieve(ctx, nodeType)
	if err != nil {
		return models.NodeTypeInstances{}, err
	}
	result := models.NodeTypeInstances{NodeType: row.NodeType, ServiceName: row.ServiceName, Instances: []models.RegistryInstance{}}
	if row.Source != enums.NodeSourceService || row.ServiceName == "" {
		return result, bizerr.BadRequest(subcode.ParamInvalid, "节点类型 %s 未配置服务名称，无法查询注册中心实例", nodeType)
	}
	registrar, err := registrarResolver()
	if err != nil {
		return result, bizerr.Unavailable(subcode.RegistryUnavailable, "注册中心未配置或不可用：%v", err)
	}
	result.Registry = registrar.Name()

	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	instances, err := registrar.GetService(queryCtx, row.ServiceName)
	if err != nil {
		svc.logger.Errorf("query registry instances failed, service:%s, err:%+v", row.ServiceName, err)
		return result, bizerr.Unavailable(subcode.RegistryUnavailable, "查询注册中心失败：%v", err)
	}
	for _, instance := range instances {
		if instance == nil {
			continue
		}
		item := models.RegistryInstance{
			ID:        instance.ID,
			Name:      instance.Name,
			Version:   instance.Version,
			Healthy:   instance.Healthy,
			Weight:    instance.Weight,
			Endpoints: make([]string, 0, len(instance.Endpoints)),
			Metadata:  instance.Metadata,
		}
		for _, endpoint := range instance.Endpoints {
			item.Endpoints = append(item.Endpoints, endpoint.EndpointURL)
		}
		result.Instances = append(result.Instances, item)
	}
	sort.Slice(result.Instances, func(i, j int) bool { return result.Instances[i].ID < result.Instances[j].ID })
	return result, nil
}
