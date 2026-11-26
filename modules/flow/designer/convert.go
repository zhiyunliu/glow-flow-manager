package designer

import (
	"encoding/json"
	"strings"

	glowflow "github.com/zhiyunliu/glow-flow"
	"github.com/zhiyunliu/golibs/xtypes"
	"github.com/zhiyunliu/manager-glow/models"
)

// ChainMeta 引擎 ChainDefinition 中的流程元数据
type ChainMeta struct {
	ChainNo   string
	Version   string
	Name      string
	Status    int
	ExtParams map[string]any
}

// ToChainDefinition 将设计器模型转换为 glow-flow 引擎的流程定义
func ToChainDefinition(meta ChainMeta, nodes []models.DesignNode, conns []models.DesignConnection) glowflow.ChainDefinition {
	def := glowflow.ChainDefinition{
		ID:      meta.ChainNo,
		Version: meta.Version,
		Metadata: glowflow.ChainMetadata{
			ID:        meta.ChainNo,
			Name:      meta.Name,
			Status:    meta.Status,
			ExtParams: meta.ExtParams,
		},
		Endpoints:   []glowflow.EndpointDefinition{},
		Nodes:       make([]glowflow.NodeDefinition, 0, len(nodes)),
		Connections: make([]glowflow.ConnectionDefinition, 0, len(conns)),
	}
	for _, node := range nodes {
		def.Nodes = append(def.Nodes, ToEngineNode(node))
	}
	for _, conn := range conns {
		def.Connections = append(def.Connections, glowflow.ConnectionDefinition{
			FromID: conn.FromID,
			ToID:   conn.ToID,
			Type:   NormalizeRelation(conn.Type),
			Remark: conn.Remark,
		})
	}
	return def
}

// ToEngineNode 设计器节点转换为引擎节点定义
func ToEngineNode(node models.DesignNode) glowflow.NodeDefinition {
	ext := xtypes.XMap{}
	for key, val := range node.ExtParams {
		ext[key] = val
	}
	return glowflow.NodeDefinition{
		ID:       node.ID,
		NodeType: node.Type,
		NodeName: node.Name,
		NodeLayout: glowflow.Layout{
			Icon: node.Layout.Icon,
			H:    node.Layout.H,
			X:    node.Layout.X,
			Y:    node.Layout.Y,
			W:    node.Layout.W,
		},
		ExtParams: ext,
	}
}

// FromChainDefinition 将引擎流程定义转换为设计器模型(用于导入)
func FromChainDefinition(def glowflow.ChainDefinition) ([]models.DesignNode, []models.DesignConnection) {
	nodes := make([]models.DesignNode, 0, len(def.Nodes))
	for _, node := range def.Nodes {
		ext := map[string]any{}
		for key, val := range node.ExtParams {
			ext[key] = val
		}
		nodes = append(nodes, models.DesignNode{
			ID:   node.ID,
			Type: node.NodeType,
			Name: node.NodeName,
			Layout: models.Layout{
				Icon: node.NodeLayout.Icon,
				H:    node.NodeLayout.H,
				X:    node.NodeLayout.X,
				Y:    node.NodeLayout.Y,
				W:    node.NodeLayout.W,
			},
			ExtParams: ext,
		})
	}
	conns := make([]models.DesignConnection, 0, len(def.Connections))
	for _, conn := range def.Connections {
		conns = append(conns, models.DesignConnection{
			FromID: conn.FromID,
			ToID:   conn.ToID,
			Type:   NormalizeRelation(conn.Type),
			Remark: conn.Remark,
		})
	}
	return nodes, conns
}

// EncodeJSON 将对象编码为 JSON 字符串；nil 或空对象返回空字符串
func EncodeJSON(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	switch v := val.(type) {
	case map[string]any:
		if len(v) == 0 {
			return "", nil
		}
	case string:
		return v, nil
	}
	data, err := json.Marshal(val)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// EncodeNodeExtParams 节点配置编码：始终写入 JSON 对象(至少为 {})，避免引擎回退到节点定义的默认配置
func EncodeNodeExtParams(ext map[string]any) (string, error) {
	if len(ext) == 0 {
		return "{}", nil
	}
	data, err := json.Marshal(ext)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DecodeObject 解析 JSON 对象，空字符串或 null 返回空 map
func DecodeObject(text string) (map[string]any, error) {
	result := map[string]any{}
	if strings.TrimSpace(text) == "" || strings.EqualFold(strings.TrimSpace(text), "null") {
		return result, nil
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, err
	}
	if result == nil {
		result = map[string]any{}
	}
	return result, nil
}

// DecodeLayout 解析节点布局
func DecodeLayout(text string) (models.Layout, error) {
	var layout models.Layout
	if strings.TrimSpace(text) == "" || strings.EqualFold(strings.TrimSpace(text), "null") {
		return layout, nil
	}
	err := json.Unmarshal([]byte(text), &layout)
	return layout, err
}

func jsonUnmarshalString(text string, out any) error {
	return json.Unmarshal([]byte(text), out)
}
