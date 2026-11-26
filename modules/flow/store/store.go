// Package store 流程版本图(节点、连接、节点定义)的读写，供流程定义与版本管理共用
package store

import (
	"context"
	"regexp"
	"strings"

	"github.com/zhiyunliu/glue/xdb"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/flow/dao"
	"github.com/zhiyunliu/manager-glow/modules/flow/designer"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
	"github.com/zhiyunliu/manager-glow/pkg/strutil"
)

const (
	maxRemarkBytes = 256  // flow_chain_connection.remark varchar(256)
	maxDescBytes   = 1024 // [desc] varchar(1024)
	maxNameBytes   = 256  // flow_chain_definition.name / flow_chain_version.name varchar(256)
)

var chainNoPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_\-]{0,31}$`)

// CheckChainNo 校验流程编号
func CheckChainNo(chainNo string) error {
	if !chainNoPattern.MatchString(chainNo) {
		return bizerr.BadRequest(subcode.ParamInvalid, "流程编号 %q 无效：以字母开头，只能包含字母、数字、下划线和中划线，最长 32 位", chainNo)
	}
	return nil
}

// CheckName 校验流程名称
func CheckName(name string) error {
	if strings.TrimSpace(name) == "" {
		return bizerr.BadRequest(subcode.ParamInvalid, "流程名称不能为空")
	}
	if strutil.DBLen(name) > maxNameBytes {
		return bizerr.BadRequest(subcode.ParamInvalid, "流程名称过长：最多 %d 个字符(中文按 2 个计算)", maxNameBytes)
	}
	return nil
}

// CheckDesc 校验描述
func CheckDesc(desc string) error {
	if strutil.DBLen(desc) > maxDescBytes {
		return bizerr.BadRequest(subcode.ParamInvalid, "描述过长：最多 %d 个字符(中文按 2 个计算)", maxDescBytes)
	}
	return nil
}

// Normalize 规范化设计器模型：去除名称首尾空白、补齐空配置、空连接类型视为 default
func Normalize(nodes []models.DesignNode, conns []models.DesignConnection) ([]models.DesignNode, []models.DesignConnection) {
	outNodes := make([]models.DesignNode, 0, len(nodes))
	for _, node := range nodes {
		node.ID = strings.TrimSpace(node.ID)
		node.Name = strings.TrimSpace(node.Name)
		node.Type = strings.TrimSpace(node.Type)
		if node.ExtParams == nil {
			node.ExtParams = map[string]any{}
		}
		outNodes = append(outNodes, node)
	}
	outConns := make([]models.DesignConnection, 0, len(conns))
	for _, conn := range conns {
		conn.FromID = strings.TrimSpace(conn.FromID)
		conn.ToID = strings.TrimSpace(conn.ToID)
		conn.Type = designer.NormalizeRelation(conn.Type)
		conn.Remark = strings.TrimSpace(conn.Remark)
		outConns = append(outConns, conn)
	}
	return outNodes, outConns
}

// integrityCodes 影响数据完整性的校验错误，存在时不允许保存草稿
var integrityCodes = map[string]bool{
	"node_id_invalid":    true,
	"node_id_duplicate":  true,
	"node_name_empty":    true,
	"node_name_too_long": true,
	"node_type_empty":    true,
	"conn_node_missing":  true,
	"conn_type_invalid":  true,
	"conn_duplicate":     true,
}

// CheckIntegrity 草稿保存前检查：存在完整性错误时返回 422 及校验结果
func CheckIntegrity(result *models.ValidationResult, conns []models.DesignConnection) error {
	for _, issue := range result.Errors {
		if integrityCodes[issue.Code] {
			return bizerr.Unprocessable(subcode.ModelInvalid, "模型存在结构错误，无法保存："+issue.Message, result)
		}
	}
	for _, conn := range conns {
		if strutil.DBLen(conn.Remark) > maxRemarkBytes {
			return bizerr.BadRequest(subcode.ParamInvalid, "连接 %s → %s 的备注过长", conn.FromID, conn.ToID)
		}
	}
	return nil
}

// LoadDesign 读取版本的节点与连接
func LoadDesign(ctx context.Context, dbObj xdb.Executer, versionID int64) ([]models.DesignNode, []models.DesignConnection, error) {
	rows, err := dao.ListVersionNodes(ctx, dbObj, versionID)
	if err != nil {
		return nil, nil, err
	}
	nodes := make([]models.DesignNode, 0, len(rows))
	for _, row := range rows {
		name := row.NodeName
		if name == "" {
			name = row.DefName
		}
		// 与引擎一致：版本节点配置为空时使用节点定义的配置
		extText := row.ExtParams
		if strings.TrimSpace(extText) == "" {
			extText = row.DefExtParams
		}
		ext, err := designer.DecodeObject(extText)
		if err != nil {
			return nil, nil, bizerr.Conflict(subcode.ModelInvalid, "节点 %s 的配置不是有效的 JSON：%v", row.NodeDefNo, err)
		}
		layoutText := row.Layout
		if strings.TrimSpace(layoutText) == "" {
			layoutText = row.DefLayout
		}
		layout, err := designer.DecodeLayout(layoutText)
		if err != nil {
			return nil, nil, bizerr.Conflict(subcode.ModelInvalid, "节点 %s 的布局不是有效的 JSON：%v", row.NodeDefNo, err)
		}
		nodes = append(nodes, models.DesignNode{ID: row.NodeDefNo, Type: row.Type, Name: name, Layout: layout, ExtParams: ext})
	}

	connRows, err := dao.ListVersionConnections(ctx, dbObj, versionID)
	if err != nil {
		return nil, nil, err
	}
	conns := make([]models.DesignConnection, 0, len(connRows))
	for _, row := range connRows {
		conn := models.DesignConnection{FromID: row.FromID, ToID: row.ToID, Type: designer.NormalizeRelation(row.Type), Remark: row.Remark}
		if ext, err := designer.DecodeObject(row.ExtParams); err == nil && len(ext) > 0 {
			conn.ExtParams = ext
		}
		conns = append(conns, conn)
	}
	return nodes, conns, nil
}

// WriteGraph 覆盖写入版本的节点与连接，并维护节点定义：
//   - 节点编号(node_def_no)全局唯一，属于其他流程时拒绝
//   - 已在其他版本中使用的节点不允许修改类型，避免改变已发布版本的执行逻辑
//   - 节点名称写入版本内名称(node_name)，已发布版本的展示名称不受影响
func WriteGraph(ctx context.Context, tx xdb.Executer, chainNo string, versionID int64, nodes []models.DesignNode, conns []models.DesignConnection, operator string) error {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}
	defs, err := dao.ListNodeDefinitions(ctx, tx, ids)
	if err != nil {
		return err
	}
	usage, err := dao.ListNodeUsageOutsideVersion(ctx, tx, versionID, ids)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		def, exists := defs[node.ID]
		if !exists {
			defaultExt, err := designer.EncodeNodeExtParams(node.ExtParams)
			if err != nil {
				return err
			}
			if err := dao.InsertNodeDefinition(ctx, tx, map[string]any{
				"node_def_no": node.ID,
				"name":        node.Name,
				"type":        node.Type,
				"extparams":   defaultExt,
				"operator":    operator,
				"chain_no":    chainNo,
			}); err != nil {
				return err
			}
			continue
		}
		if def.ChainNo != "" && def.ChainNo != chainNo {
			return bizerr.Conflict(subcode.NodeIDConflict, "节点编号 %s 已被流程 %s 使用，请重新添加该节点", node.ID, def.ChainNo)
		}
		usedElsewhere := usage[node.ID] > 0
		if def.Type != node.Type && usedElsewhere {
			return bizerr.Conflict(subcode.NodeTypeChanged, "节点 %s 已在其他版本中使用，不能将类型从 %s 修改为 %s，请删除后重新添加节点", node.ID, def.Type, node.Type)
		}
		// 未归属流程的历史节点定义被其他版本使用时保持不变
		if def.ChainNo == "" && usedElsewhere {
			continue
		}
		if def.Name != node.Name || def.Type != node.Type {
			if err := dao.UpdateNodeDefinition(ctx, tx, map[string]any{
				"node_def_no": node.ID,
				"name":        node.Name,
				"type":        node.Type,
				"operator":    operator,
			}); err != nil {
				return err
			}
		}
	}

	if err := dao.ClearVersionGraph(ctx, tx, versionID); err != nil {
		return err
	}
	for _, node := range nodes {
		ext, err := designer.EncodeNodeExtParams(node.ExtParams)
		if err != nil {
			return err
		}
		layout, err := designer.EncodeJSON(node.Layout)
		if err != nil {
			return err
		}
		if err := dao.InsertChainNode(ctx, tx, map[string]any{
			"version_id":  versionID,
			"chain_no":    chainNo,
			"node_def_no": node.ID,
			"extparams":   ext,
			"layout":      layout,
			"node_name":   node.Name,
		}); err != nil {
			return err
		}
	}
	for _, conn := range conns {
		ext, err := designer.EncodeJSON(conn.ExtParams)
		if err != nil {
			return err
		}
		if err := dao.InsertChainConnection(ctx, tx, map[string]any{
			"version_id": versionID,
			"chain_no":   chainNo,
			"from_id":    conn.FromID,
			"to_id":      conn.ToID,
			"type":       designer.NormalizeRelation(conn.Type),
			"remark":     conn.Remark,
			"extparams":  ext,
		}); err != nil {
			return err
		}
	}
	return dao.DeleteOrphanNodeDefinitions(ctx, tx, chainNo)
}

// DraftSpec 新建草稿版本的内容
type DraftSpec struct {
	ChainNo       string
	Name          string
	Desc          string
	ExtParams     map[string]any
	Layout        map[string]any
	BaseVersionID int64
	Nodes         []models.DesignNode
	Connections   []models.DesignConnection
}

// CreateDraft 新建草稿版本并写入节点与连接，需在锁定流程的事务中调用
func CreateDraft(ctx context.Context, tx xdb.Executer, spec DraftSpec, operator string) (versionID int64, versionNo int, err error) {
	versionNo, err = dao.NextVersionNo(ctx, tx, spec.ChainNo)
	if err != nil {
		return 0, 0, err
	}
	ext, err := designer.EncodeJSON(spec.ExtParams)
	if err != nil {
		return 0, 0, err
	}
	layout, err := designer.EncodeJSON(spec.Layout)
	if err != nil {
		return 0, 0, err
	}
	versionID, err = dao.InsertVersion(ctx, tx, map[string]any{
		"chain_no":        spec.ChainNo,
		"name":            spec.Name,
		"extparams":       ext,
		"layout":          layout,
		"operator":        operator,
		"desc":            spec.Desc,
		"version_no":      versionNo,
		"base_version_id": spec.BaseVersionID,
	})
	if err != nil {
		return 0, 0, err
	}
	if err = WriteGraph(ctx, tx, spec.ChainNo, versionID, spec.Nodes, spec.Connections, operator); err != nil {
		return 0, 0, err
	}
	return versionID, versionNo, nil
}

// ResolveSourceVersion 确定复制/导出使用的版本：指定版本 > 生效版本 > 最新版本；没有版本时返回 0
func ResolveSourceVersion(ctx context.Context, dbObj xdb.Executer, chainNo string, versionID int64) (int64, error) {
	if versionID > 0 {
		version, err := dao.GetVersion(ctx, dbObj, versionID)
		if err != nil {
			return 0, err
		}
		if version.VersionID == 0 || version.ChainNo != chainNo {
			return 0, bizerr.NotFound(subcode.VersionNotFound, "流程 %s 中不存在版本 %d", chainNo, versionID)
		}
		return versionID, nil
	}
	effective, err := dao.GetEffectiveVersionID(ctx, dbObj, chainNo)
	if err != nil || effective > 0 {
		return effective, err
	}
	return dao.GetLatestVersionID(ctx, dbObj, chainNo)
}

// VersionMeta 解析版本的扩展参数与布局
func VersionMeta(version models.DBVersionInfo) (ext map[string]any, layout map[string]any) {
	ext, err := designer.DecodeObject(version.ExtParams)
	if err != nil {
		ext = map[string]any{}
	}
	layout, err = designer.DecodeObject(version.Layout)
	if err != nil {
		layout = map[string]any{}
	}
	return ext, layout
}

// IsDuplicateKey 判断是否为主键/唯一索引冲突
func IsDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "primary key constraint") || strings.Contains(msg, "unique index")
}
