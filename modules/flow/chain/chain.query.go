package chain

import (
	"context"
	"strings"

	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/flow/dao"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
)

func (svc *ChainerImpl) Page(ctx context.Context, params models.PageChainParams) ([]models.DBChainInfo, int64, error) {
	offset := params.Normalize()
	query := map[string]any{
		"offset":    offset,
		"page_size": params.PageSize,
	}
	if val := strings.TrimSpace(params.ChainNo); val != "" {
		query["chain_no"] = val
	}
	if val := strings.TrimSpace(params.Name); val != "" {
		query["name"] = val
	}
	if params.FlowType != "" {
		query["flow_type"] = params.FlowType
	}
	if params.Status == "0" || params.Status == "1" {
		query["status"] = params.Status
	}
	list, total, err := dao.PageChains(ctx, dao.DB(), query, params.PageSize)
	if err != nil {
		svc.logger.Errorf("page chains failed, params:%+v, err:%+v", params, err)
		return nil, 0, err
	}
	return list, total, nil
}

func (svc *ChainerImpl) Detail(ctx context.Context, chainNo string) (models.DBChainInfo, error) {
	return svc.retrieve(ctx, chainNo)
}

func (svc *ChainerImpl) retrieve(ctx context.Context, chainNo string) (models.DBChainInfo, error) {
	item, err := dao.GetChain(ctx, dao.DB(), chainNo)
	if err != nil {
		svc.logger.Errorf("get chain failed, chain_no:%s, err:%+v", chainNo, err)
		return item, err
	}
	if item.ChainNo == "" {
		return item, bizerr.NotFound(subcode.ChainNotFound, "流程 %s 不存在", chainNo)
	}
	return item, nil
}
