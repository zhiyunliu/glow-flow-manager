package version

import (
	"context"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/flow/dao"
	"github.com/zhiyunliu/manager-glow/modules/flow/store"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
)

func (svc *VersionerImpl) List(ctx context.Context, chainNo string) ([]models.DBVersionInfo, error) {
	chain, err := svc.retrieveChain(ctx, chainNo)
	if err != nil {
		return nil, err
	}
	list, err := dao.ListVersions(ctx, dao.DB(), chainNo)
	if err != nil {
		svc.logger.Errorf("list versions failed, chain_no:%s, err:%+v", chainNo, err)
		return nil, err
	}
	for i := range list {
		list[i].IsEffective = list[i].VersionID == chain.EffectiveVersionID
	}
	return list, nil
}

func (svc *VersionerImpl) Detail(ctx context.Context, versionID int64) (models.ChainDesign, error) {
	design := models.ChainDesign{}
	version, err := svc.retrieveVersion(ctx, versionID)
	if err != nil {
		return design, err
	}
	chain, err := svc.retrieveChain(ctx, version.ChainNo)
	if err != nil {
		return design, err
	}
	nodes, conns, err := store.LoadDesign(ctx, dao.DB(), versionID)
	if err != nil {
		return design, err
	}
	version.IsEffective = version.VersionID == chain.EffectiveVersionID
	design.Chain = chain
	design.Version = version
	design.ExtParams, design.Layout = store.VersionMeta(version)
	design.Nodes = nodes
	design.Connections = conns
	design.Editable = version.PublishState == enums.VersionStateDraft
	return design, nil
}

func (svc *VersionerImpl) retrieveChain(ctx context.Context, chainNo string) (models.DBChainInfo, error) {
	chain, err := dao.GetChain(ctx, dao.DB(), chainNo)
	if err != nil {
		svc.logger.Errorf("get chain failed, chain_no:%s, err:%+v", chainNo, err)
		return chain, err
	}
	if chain.ChainNo == "" {
		return chain, bizerr.NotFound(subcode.ChainNotFound, "流程 %s 不存在", chainNo)
	}
	return chain, nil
}

func (svc *VersionerImpl) retrieveVersion(ctx context.Context, versionID int64) (models.DBVersionInfo, error) {
	version, err := dao.GetVersion(ctx, dao.DB(), versionID)
	if err != nil {
		svc.logger.Errorf("get version failed, version_id:%d, err:%+v", versionID, err)
		return version, err
	}
	if version.VersionID == 0 {
		return version, bizerr.NotFound(subcode.VersionNotFound, "版本 %d 不存在", versionID)
	}
	return version, nil
}

// lockVersion 在事务中按 流程 → 版本 的顺序加锁(与创建草稿、导入一致，避免并发操作死锁)
func lockVersion(ctx context.Context, tx xdb.Executer, versionID int64) (models.DBVersionInfo, models.DBChainInfo, error) {
	probe, err := dao.GetVersion(ctx, tx, versionID)
	if err != nil {
		return probe, models.DBChainInfo{}, err
	}
	if probe.VersionID == 0 {
		return probe, models.DBChainInfo{}, bizerr.NotFound(subcode.VersionNotFound, "版本 %d 不存在", versionID)
	}
	chain, err := dao.LockChain(ctx, tx, probe.ChainNo)
	if err != nil {
		return probe, chain, err
	}
	version, err := dao.LockVersion(ctx, tx, versionID)
	if err != nil {
		return version, chain, err
	}
	if version.VersionID == 0 {
		return version, chain, bizerr.NotFound(subcode.VersionNotFound, "版本 %d 不存在", versionID)
	}
	return version, chain, nil
}
