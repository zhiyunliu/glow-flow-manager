package designer

import (
	"crypto/rand"
	"strings"

	"github.com/zhiyunliu/manager-glow/models"
)

const idAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// NewNodeID 生成节点编号：n + 11 位随机小写字母数字，与前端生成规则一致
func NewNodeID() string {
	buf := make([]byte, 11)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	for i := range buf {
		buf[i] = idAlphabet[int(buf[i])%len(idAlphabet)]
	}
	return "n" + string(buf)
}

// NewIDMapping 为给定节点生成新编号
func NewIDMapping(nodes []models.DesignNode, needRemap func(id string) bool) map[string]string {
	mapping := make(map[string]string)
	for _, node := range nodes {
		if needRemap == nil || needRemap(node.ID) {
			mapping[node.ID] = NewNodeID()
		}
	}
	return mapping
}

// RemapNodeIDs 按 mapping 替换节点编号，同时替换连接两端及节点配置中引用的节点编号
// (如 for/while 的 do、groupaction 的 nodeids)。配置中与旧编号完全相同的字符串值会被替换。
func RemapNodeIDs(nodes []models.DesignNode, conns []models.DesignConnection, mapping map[string]string) ([]models.DesignNode, []models.DesignConnection) {
	if len(mapping) == 0 {
		return nodes, conns
	}
	newNodes := make([]models.DesignNode, 0, len(nodes))
	for _, node := range nodes {
		if newID, ok := mapping[node.ID]; ok {
			node.ID = newID
		}
		if node.ExtParams != nil {
			node.ExtParams, _ = remapValue(node.ExtParams, mapping).(map[string]any)
		}
		newNodes = append(newNodes, node)
	}
	newConns := make([]models.DesignConnection, 0, len(conns))
	for _, conn := range conns {
		if newID, ok := mapping[conn.FromID]; ok {
			conn.FromID = newID
		}
		if newID, ok := mapping[conn.ToID]; ok {
			conn.ToID = newID
		}
		newConns = append(newConns, conn)
	}
	return newNodes, newConns
}

func remapValue(val any, mapping map[string]string) any {
	switch v := val.(type) {
	case string:
		if newID, ok := mapping[v]; ok {
			return newID
		}
		if strings.Contains(v, ",") {
			return remapCommaList(v, mapping)
		}
		return v
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = remapValue(item, mapping)
		}
		return out
	case []string:
		out := make([]string, len(v))
		for i, item := range v {
			if newID, ok := mapping[item]; ok {
				out[i] = newID
			} else {
				out[i] = item
			}
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			out[key] = remapValue(item, mapping)
		}
		return out
	default:
		return val
	}
}

// remapCommaList 逗号分隔且每一项都是节点编号时逐项替换
func remapCommaList(text string, mapping map[string]string) string {
	parts := strings.Split(text, ",")
	for i, part := range parts {
		newID, ok := mapping[strings.TrimSpace(part)]
		if !ok {
			return text
		}
		parts[i] = newID
	}
	return strings.Join(parts, ",")
}
