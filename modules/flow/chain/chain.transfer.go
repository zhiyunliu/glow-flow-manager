package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	glowflow "github.com/zhiyunliu/glow-flow"
	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/audit"
	"github.com/zhiyunliu/manager-glow/modules/flow/dao"
	"github.com/zhiyunliu/manager-glow/modules/flow/designer"
	"github.com/zhiyunliu/manager-glow/modules/flow/store"
	"github.com/zhiyunliu/manager-glow/modules/nodetype"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
	"github.com/zhiyunliu/manager-glow/pkg/operator"
)

// ExportFormat 导出文件格式标识
const ExportFormat = "manager-glow/chain@1"

func (svc *ChainerImpl) Copy(ctx context.Context, op operator.Operator, params models.CopyChainParams) (result models.ImportResult, err error) {
	params.NewChainNo = strings.TrimSpace(params.NewChainNo)
	params.NewName = strings.TrimSpace(params.NewName)
	entry := audit.Entry{Module: enums.AuditModuleChain, OperateType: enums.AuditOpCopy, TargetID: params.NewChainNo, ChainNo: params.NewChainNo,
		Summary: fmt.Sprintf("复制流程 %s 为 %s", params.ChainNo, params.NewChainNo), After: params}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	source, err := svc.retrieve(ctx, params.ChainNo)
	if err != nil {
		return result, err
	}
	if err = checkChainParams(params.NewChainNo, params.NewName, source.FlowType, source.Desc); err != nil {
		return result, err
	}
	sourceVersionID, err := store.ResolveSourceVersion(ctx, dao.DB(), source.ChainNo, params.VersionID)
	if err != nil {
		return result, err
	}
	spec := store.DraftSpec{ChainNo: params.NewChainNo, Name: params.NewName}
	if sourceVersionID > 0 {
		version, err := dao.GetVersion(ctx, dao.DB(), sourceVersionID)
		if err != nil {
			return result, err
		}
		nodes, conns, err := store.LoadDesign(ctx, dao.DB(), sourceVersionID)
		if err != nil {
			return result, err
		}
		// 节点编号全局唯一，复制时全部重新生成
		spec.Nodes, spec.Connections = designer.RemapNodeIDs(nodes, conns, designer.NewIDMapping(nodes, nil))
		spec.ExtParams, spec.Layout = store.VersionMeta(version)
		spec.Desc = fmt.Sprintf("复制自 %s v%d", source.ChainNo, version.VersionNo)
	}

	err = dao.Transaction(ctx, func(ctx context.Context, tx xdb.Executer) error {
		existed, err := dao.LockChain(ctx, tx, params.NewChainNo)
		if err != nil {
			return err
		}
		if existed.ChainNo != "" {
			return bizerr.Conflict(subcode.ChainExists, "流程编号 %s 已存在", params.NewChainNo)
		}
		if err := dao.InsertChain(ctx, tx, map[string]any{
			"chain_no": params.NewChainNo, "name": params.NewName, "extparams": "", "layout": "",
			"operator": op.ID, "desc": source.Desc, "flow_type": source.FlowType,
		}); err != nil {
			return err
		}
		result.ChainNo = params.NewChainNo
		result.VersionID, result.VersionNo, err = store.CreateDraft(ctx, tx, spec, op.ID)
		return err
	})
	if err != nil && store.IsDuplicateKey(err) {
		err = bizerr.Conflict(subcode.ChainExists, "流程编号 %s 已存在", params.NewChainNo)
	}
	entry.VersionID = result.VersionID
	return result, err
}

func (svc *ChainerImpl) Export(ctx context.Context, params models.ExportChainParams) (models.ChainExport, string, error) {
	export := models.ChainExport{}
	chain, err := svc.retrieve(ctx, params.ChainNo)
	if err != nil {
		return export, "", err
	}
	versionID, err := store.ResolveSourceVersion(ctx, dao.DB(), chain.ChainNo, params.VersionID)
	if err != nil {
		return export, "", err
	}
	if versionID == 0 {
		return export, "", bizerr.NotFound(subcode.VersionNotFound, "流程 %s 没有可导出的版本", chain.ChainNo)
	}
	version, err := dao.GetVersion(ctx, dao.DB(), versionID)
	if err != nil {
		return export, "", err
	}
	nodes, conns, err := store.LoadDesign(ctx, dao.DB(), versionID)
	if err != nil {
		return export, "", err
	}
	ext, _ := store.VersionMeta(version)
	name := version.Name
	if name == "" {
		name = chain.Name
	}
	def := designer.ToChainDefinition(designer.ChainMeta{
		ChainNo: chain.ChainNo, Version: strconv.FormatInt(versionID, 10), Name: name, Status: chain.Status, ExtParams: ext,
	}, nodes, conns)
	defMap, err := toMap(def)
	if err != nil {
		return export, "", err
	}
	export = models.ChainExport{
		Format:     ExportFormat,
		ExportTime: time.Now().Format("2006-01-02 15:04:05"),
		FlowType:   chain.FlowType,
		Desc:       chain.Desc,
		VersionNo:  version.VersionNo,
		Definition: defMap,
	}
	for _, conn := range conns {
		if len(conn.ExtParams) > 0 {
			export.Edges = append(export.Edges, models.EdgeExtra{FromID: conn.FromID, ToID: conn.ToID, Type: conn.Type, ExtParams: conn.ExtParams})
		}
	}
	svc.logger.Infof("export chain, chain_no:%s, version_id:%d", chain.ChainNo, versionID)
	return export, fmt.Sprintf("%s_v%d.json", chain.ChainNo, version.VersionNo), nil
}

func (svc *ChainerImpl) Import(ctx context.Context, op operator.Operator, params models.ImportChainParams) (result models.ImportResult, err error) {
	entry := audit.Entry{Module: enums.AuditModuleChain, OperateType: enums.AuditOpImport, TargetID: strings.TrimSpace(params.ChainNo), ChainNo: strings.TrimSpace(params.ChainNo)}
	defer func() { entry.Err = err; svc.auditor.Record(ctx, op, entry) }()

	imported, err := parseImport(params.Content)
	if err != nil {
		return result, err
	}
	entry.After = imported.raw
	nodes, conns := designer.FromChainDefinition(imported.def)
	conns = applyEdgeExtras(conns, imported.edges)
	nodes, conns = store.Normalize(nodes, conns)

	chainNo := strings.TrimSpace(params.ChainNo)
	if chainNo == "" {
		chainNo = strings.TrimSpace(imported.def.ID)
	}
	entry.TargetID, entry.ChainNo = chainNo, chainNo
	flowType, name, desc := imported.flowType, strings.TrimSpace(params.Name), imported.desc
	switch params.Mode {
	case enums.ImportModeNewChain:
		if name == "" {
			name = strings.TrimSpace(imported.def.Metadata.Name)
		}
		if name == "" {
			name = chainNo
		}
		if flowType == "" {
			flowType = enums.FlowTypeRuleChain
		}
		if err = checkChainParams(chainNo, name, flowType, desc); err != nil {
			return result, err
		}
		entry.Summary = fmt.Sprintf("导入为新流程 %s", chainNo)
	case enums.ImportModeNewVersion:
		chain, err := svc.retrieve(ctx, chainNo)
		if err != nil {
			return result, err
		}
		flowType, name = chain.FlowType, chain.Name
		entry.Summary = fmt.Sprintf("导入为流程 %s 的草稿", chainNo)
	default:
		return result, bizerr.BadRequest(subcode.ParamInvalid, "导入方式 %q 无效", params.Mode)
	}

	types, err := nodetype.LoadMetaMap(ctx)
	if err != nil {
		return result, err
	}
	validation := designer.Validate(designer.Input{FlowType: flowType, Nodes: nodes, Connections: conns, NodeTypes: types})
	if err = store.CheckIntegrity(validation, conns); err != nil {
		return result, err
	}

	err = dao.Transaction(ctx, func(ctx context.Context, tx xdb.Executer) error {
		existed, err := dao.LockChain(ctx, tx, chainNo)
		if err != nil {
			return err
		}
		if params.Mode == enums.ImportModeNewChain {
			if existed.ChainNo != "" {
				return bizerr.Conflict(subcode.ChainExists, "流程编号 %s 已存在", chainNo)
			}
			if err := dao.InsertChain(ctx, tx, map[string]any{
				"chain_no": chainNo, "name": name, "extparams": "", "layout": "",
				"operator": op.ID, "desc": desc, "flow_type": flowType,
			}); err != nil {
				return err
			}
		} else if existed.ChainNo == "" {
			return bizerr.NotFound(subcode.ChainNotFound, "流程 %s 不存在", chainNo)
		}

		// 节点编号已被其他流程使用时重新生成
		ids := make([]string, 0, len(nodes))
		for _, node := range nodes {
			ids = append(ids, node.ID)
		}
		defs, err := dao.ListNodeDefinitions(ctx, tx, ids)
		if err != nil {
			return err
		}
		mapping := designer.NewIDMapping(nodes, func(id string) bool {
			def, ok := defs[id]
			return ok && def.ChainNo != chainNo
		})
		nodes, conns = designer.RemapNodeIDs(nodes, conns, mapping)
		if len(mapping) > 0 {
			result.Remapped = mapping
		}
		result.ChainNo = chainNo

		spec := store.DraftSpec{ChainNo: chainNo, Name: name, Desc: "导入", ExtParams: imported.def.Metadata.ExtParams, Nodes: nodes, Connections: conns}
		if params.Mode == enums.ImportModeNewVersion {
			draftID, err := dao.GetDraftVersionID(ctx, tx, chainNo)
			if err != nil {
				return err
			}
			if draftID > 0 {
				// 已有草稿时覆盖草稿内容
				draft, err := dao.LockVersion(ctx, tx, draftID)
				if err != nil {
					return err
				}
				ext, err := designer.EncodeJSON(spec.ExtParams)
				if err != nil {
					return err
				}
				if err := dao.UpdateDraftVersion(ctx, tx, map[string]any{
					"version_id": draftID, "name": name, "extparams": ext, "layout": draft.Layout, "desc": spec.Desc, "operator": op.ID,
				}); err != nil {
					return err
				}
				result.VersionID, result.VersionNo = draftID, draft.VersionNo
				return store.WriteGraph(ctx, tx, chainNo, draftID, nodes, conns, op.ID)
			}
		}
		result.VersionID, result.VersionNo, err = store.CreateDraft(ctx, tx, spec, op.ID)
		return err
	})
	if err != nil && store.IsDuplicateKey(err) {
		err = bizerr.Conflict(subcode.ChainExists, "流程编号 %s 已存在", chainNo)
	}
	entry.VersionID = result.VersionID
	return result, err
}

type importContent struct {
	def      glowflow.ChainDefinition
	flowType string
	desc     string
	edges    []models.EdgeExtra
	raw      json.RawMessage
}

// parseImport 解析导入内容：支持本系统导出文件与 glow-flow ChainDefinition JSON
func parseImport(content string) (importContent, error) {
	result := importContent{raw: json.RawMessage(content)}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal([]byte(content), &probe); err != nil {
		return result, bizerr.BadRequest(subcode.ImportInvalid, "导入内容不是有效的 JSON 对象：%v", err)
	}
	if _, ok := probe["definition"]; ok {
		var export struct {
			FlowType   string                   `json:"flow_type"`
			Desc       string                   `json:"desc"`
			Definition glowflow.ChainDefinition `json:"definition"`
			Edges      []models.EdgeExtra       `json:"edges"`
		}
		if err := json.Unmarshal([]byte(content), &export); err != nil {
			return result, bizerr.BadRequest(subcode.ImportInvalid, "导入文件格式错误：%v", err)
		}
		if export.FlowType != "" && !enums.IsValidFlowType(export.FlowType) {
			return result, bizerr.BadRequest(subcode.ImportInvalid, "导入文件中的流程类型 %q 无效", export.FlowType)
		}
		result.def, result.flowType, result.desc, result.edges = export.Definition, export.FlowType, export.Desc, export.Edges
		return result, nil
	}
	if _, ok := probe["nodes"]; ok {
		if err := json.Unmarshal([]byte(content), &result.def); err != nil {
			return result, bizerr.BadRequest(subcode.ImportInvalid, "ChainDefinition 格式错误：%v", err)
		}
		return result, nil
	}
	return result, bizerr.BadRequest(subcode.ImportInvalid, "无法识别的导入内容：需要本系统导出的文件或 glow-flow ChainDefinition JSON")
}

// applyEdgeExtras 恢复导出文件中连接的设计器附加信息
func applyEdgeExtras(conns []models.DesignConnection, edges []models.EdgeExtra) []models.DesignConnection {
	if len(edges) == 0 {
		return conns
	}
	extras := make(map[string]map[string]any, len(edges))
	for _, edge := range edges {
		extras[edge.FromID+"\x00"+edge.ToID+"\x00"+designer.NormalizeRelation(edge.Type)] = edge.ExtParams
	}
	for i, conn := range conns {
		if ext, ok := extras[conn.FromID+"\x00"+conn.ToID+"\x00"+designer.NormalizeRelation(conn.Type)]; ok {
			conns[i].ExtParams = ext
		}
	}
	return conns
}

func toMap(val any) (map[string]any, error) {
	data, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	result := map[string]any{}
	err = json.Unmarshal(data, &result)
	return result, err
}
