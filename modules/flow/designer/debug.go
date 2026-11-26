package designer

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	glowflow "github.com/zhiyunliu/glow-flow"
	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/models"
)

// DebugInput 调试输入
type DebugInput struct {
	Meta        ChainMeta
	Nodes       []models.DesignNode
	Connections []models.DesignConnection
	NodeTypes   map[string]models.NodeTypeInfo
	Data        any
	MsgID       string
	Timeout     time.Duration
	// MockActions 为 true 时动作、集成、工作流类节点及独立服务节点不真实执行，原样传递消息
	MockActions bool
}

// alwaysMockTypes 调试时始终模拟执行的节点：command 会在当前进程所在主机执行系统命令，管理后台中不允许真实执行
var alwaysMockTypes = map[string]bool{"command": true}

// IsSideEffectNode 节点执行是否会产生外部副作用(发送请求、消息、邮件、执行命令等)
func IsSideEffectNode(meta models.NodeTypeInfo) bool {
	if meta.Source != enums.NodeSourceBuiltin {
		return true
	}
	switch meta.Category {
	case enums.NodeCategoryAction, enums.NodeCategoryIntegration, enums.NodeCategoryWorkflow:
		return true
	}
	return false
}

// Debug 在沙箱中执行流程：使用引擎真实的调度器与节点实现，记录每个节点的输入输出
func Debug(ctx context.Context, in DebugInput) models.DebugResult {
	tr := &tracer{start: time.Now()}
	result := models.DebugResult{
		InstanceID:    "debug-" + NewNodeID(),
		Events:        []models.DebugEvent{},
		ExecutedNodes: []string{},
		Edges:         []models.DebugEdge{},
	}

	registry := glowflow.NewNodeRegistry()
	for _, node := range in.Nodes {
		if _, ok := registry.GetDescriptor(node.Type); ok {
			continue
		}
		meta, ok := in.NodeTypes[node.Type]
		if !ok {
			result.Error = fmt.Sprintf("节点类型 %s 未注册", node.Type)
			return result
		}
		desc, real := descriptorFor(node.Type, meta)
		mock := !real || alwaysMockTypes[node.Type] || (in.MockActions && IsSideEffectNode(meta))
		_ = registry.Register(&glowflow.NodeDescriptor{
			Type:    node.Type,
			Factory: &tracingFactory{base: desc.Factory, tr: tr, mock: mock, relation: defaultRelation(meta)},
		})
	}

	compiled, err := glowflow.NewChainCompiler().Compile(ToChainDefinition(in.Meta, in.Nodes, in.Connections), registry)
	if err != nil {
		result.Error = fmt.Sprintf("引擎编译失败：%v", err)
		return result
	}

	timeout := in.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	msgID := in.MsgID
	if msgID == "" {
		msgID = result.InstanceID
	}
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		done <- glowflow.NewDispatcher().Dispatch(glowflow.NewContext(runCtx, result.InstanceID), compiled, glowflow.NewMessage(msgID, in.Data))
	}()

	select {
	case err = <-done:
		if err != nil {
			result.Error = err.Error()
		}
	case <-runCtx.Done():
		result.TimedOut = true
		result.Error = fmt.Sprintf("调试执行超过 %s 未结束", timeout)
	}
	result.DurationMs = time.Since(tr.start).Milliseconds()
	result.Events = tr.snapshot()
	result.Success = result.Error == ""
	for _, ev := range result.Events {
		if ev.Error != "" && result.Success {
			// 节点错误会结束所在分支，dispatcher 返回首个错误
			result.Success = false
		}
	}
	result.ExecutedNodes, result.Edges = traversedPath(result.Events, in.Connections)
	return result
}

// traversedPath 根据执行记录计算经过的节点与连接
func traversedPath(events []models.DebugEvent, conns []models.DesignConnection) ([]string, []models.DebugEdge) {
	executed := make(map[string]struct{})
	nodes := make([]string, 0)
	for _, ev := range events {
		if _, ok := executed[ev.NodeID]; !ok {
			executed[ev.NodeID] = struct{}{}
			nodes = append(nodes, ev.NodeID)
		}
	}
	byFrom := make(map[string][]models.DesignConnection)
	for _, conn := range conns {
		byFrom[conn.FromID] = append(byFrom[conn.FromID], conn)
	}
	edges := make([]models.DebugEdge, 0)
	seen := make(map[string]struct{})
	for _, ev := range events {
		if ev.Error != "" || ev.End {
			continue
		}
		relations := ev.Relations
		if len(relations) == 0 {
			relations = []string{ev.RelationType}
		}
		for _, relation := range relations {
			matched := matchConnections(byFrom[ev.NodeID], relation)
			for _, conn := range matched {
				if _, ok := executed[conn.ToID]; !ok {
					continue
				}
				key := conn.FromID + "\x00" + conn.ToID + "\x00" + NormalizeRelation(conn.Type)
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				edges = append(edges, models.DebugEdge{FromID: conn.FromID, ToID: conn.ToID, Type: NormalizeRelation(conn.Type)})
			}
		}
	}
	return nodes, edges
}

// matchConnections 与引擎 NextNodes 一致：优先匹配连接类型，没有时走 default 连接
func matchConnections(conns []models.DesignConnection, relation string) []models.DesignConnection {
	relation = NormalizeRelation(relation)
	matched := make([]models.DesignConnection, 0)
	for _, conn := range conns {
		if NormalizeRelation(conn.Type) == relation {
			matched = append(matched, conn)
		}
	}
	if len(matched) == 0 && relation != enums.RelationDefault {
		for _, conn := range conns {
			if NormalizeRelation(conn.Type) == enums.RelationDefault {
				matched = append(matched, conn)
			}
		}
	}
	return matched
}

type tracer struct {
	start  time.Time
	mu     sync.Mutex
	seq    int
	events []models.DebugEvent
}

func (t *tracer) nextSeq() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.seq++
	return t.seq
}

func (t *tracer) add(ev models.DebugEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, ev)
}

func (t *tracer) snapshot() []models.DebugEvent {
	t.mu.Lock()
	defer t.mu.Unlock()
	events := append([]models.DebugEvent{}, t.events...)
	sort.SliceStable(events, func(i, j int) bool { return events[i].Seq < events[j].Seq })
	return events
}

// tracingFactory 包装节点工厂，为节点增加执行记录与模拟执行
type tracingFactory struct {
	base     glowflow.NodeFactory
	tr       *tracer
	mock     bool
	relation string
}

func (f *tracingFactory) Build(def glowflow.NodeDefinition) (glowflow.CompiledNode, error) {
	node, err := safeBuild(f.base, def)
	if err != nil {
		return nil, err
	}
	traced := &tracedNode{CompiledNode: node, tr: f.tr, mock: f.mock, relation: f.relation}
	// 保留 dispatcher 与编译器依赖的接口
	if join, ok := node.(glowflow.JoinNode); ok {
		return &tracedJoinNode{tracedNode: traced, join: join}, nil
	}
	if referrer, ok := node.(glowflow.NodeReferrer); ok {
		return &tracedReferrerNode{tracedNode: traced, referrer: referrer}, nil
	}
	return traced, nil
}

type tracedNode struct {
	glowflow.CompiledNode
	tr       *tracer
	mock     bool
	relation string
}

func (n *tracedNode) Execute(ctx glowflow.Context, msg *glowflow.Message) (result glowflow.ExecuteResult, err error) {
	ev := n.begin(snapshotMessage(msg))
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
		n.finish(ev, result, err)
	}()
	if n.mock {
		return glowflow.ExecuteResult{RelationType: n.relation, Msg: msg}, nil
	}
	return n.CompiledNode.Execute(ctx, msg)
}

func (n *tracedNode) begin(input any) *models.DebugEvent {
	start := time.Now()
	return &models.DebugEvent{
		Seq:         n.tr.nextSeq(),
		NodeID:      n.Id(),
		NodeName:    n.Name(),
		NodeType:    n.Type(),
		Mocked:      n.mock,
		Input:       input,
		StartOffset: start.Sub(n.tr.start).Milliseconds(),
		DurationMs:  start.UnixMilli(), // 暂存开始时间，finish 时换算为耗时
	}
}

func (n *tracedNode) finish(ev *models.DebugEvent, result glowflow.ExecuteResult, err error) {
	ev.DurationMs = time.Now().UnixMilli() - ev.DurationMs
	if result.Msg != nil {
		ev.Output = snapshotMessage(result.Msg)
	}
	ev.RelationType = result.RelationType
	ev.Relations = result.RelationTypes
	ev.End = result.End
	if err == nil {
		err = result.Error
	}
	if err != nil {
		ev.Error = err.Error()
	}
	n.tr.add(*ev)
}

type tracedJoinNode struct {
	*tracedNode
	join glowflow.JoinNode
}

func (n *tracedJoinNode) JoinTimeout() time.Duration {
	return n.join.JoinTimeout()
}

func (n *tracedJoinNode) Join(ctx glowflow.Context, inputs []glowflow.JoinInput, timedOut bool) (result glowflow.ExecuteResult, err error) {
	snapshot := make([]map[string]any, 0, len(inputs))
	for _, input := range inputs {
		snapshot = append(snapshot, map[string]any{"from": input.FromID, "data": snapshotMessage(input.Msg)})
	}
	ev := n.begin(map[string]any{"inputs": snapshot, "timed_out": timedOut})
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
		n.finish(ev, result, err)
	}()
	return n.join.Join(ctx, inputs, timedOut)
}

type tracedReferrerNode struct {
	*tracedNode
	referrer glowflow.NodeReferrer
}

func (n *tracedReferrerNode) ReferencedNodes() []string {
	return n.referrer.ReferencedNodes()
}

// snapshotMessage 复制消息数据为 JSON 形态，避免后续节点修改影响记录
func snapshotMessage(msg *glowflow.Message) any {
	if msg == nil {
		return nil
	}
	switch data := msg.Data.(type) {
	case nil:
		return nil
	case []byte:
		return string(data)
	case string, bool, float64, float32, int, int64, int32:
		return data
	}
	raw, err := json.Marshal(msg.Data)
	if err != nil {
		return fmt.Sprintf("%v", msg.Data)
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return string(raw)
	}
	return out
}
