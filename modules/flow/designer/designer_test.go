package designer_test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/flow/designer"
	"github.com/zhiyunliu/manager-glow/modules/nodetype"
)

func builtinTypes() map[string]models.NodeTypeInfo {
	types := make(map[string]models.NodeTypeInfo)
	for _, info := range nodetype.BuiltinInfos() {
		types[info.NodeType] = info
	}
	return types
}

func node(id, typ string, ext map[string]any) models.DesignNode {
	if ext == nil {
		ext = map[string]any{}
	}
	return models.DesignNode{ID: id, Type: typ, Name: id, ExtParams: ext}
}

func conn(from, to, typ string) models.DesignConnection {
	return models.DesignConnection{FromID: from, ToID: to, Type: typ}
}

func hasCode(issues []models.ValidationIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}

func TestValidateRuleChainRejectsCycle(t *testing.T) {
	nodes := []models.DesignNode{
		node("a", "exprfilter", map[string]any{"expr": "true"}),
		node("b", "exprtransform", map[string]any{"expr": "msg"}),
	}
	conns := []models.DesignConnection{conn("a", "b", "true"), conn("b", "a", "success")}
	result := designer.Validate(designer.Input{FlowType: enums.FlowTypeRuleChain, Nodes: nodes, Connections: conns, NodeTypes: builtinTypes()})
	if result.Valid || !hasCode(result.Errors, "cycle") {
		t.Fatalf("expected cycle error, got %+v", result)
	}
}

func TestValidateWorkflowAllowsLoopBackToStart(t *testing.T) {
	nodes := []models.DesignNode{
		node("s", "start", nil),
		node("task", "usertask", map[string]any{"assignees": []any{"u1"}}),
		node("e", "end", nil),
	}
	conns := []models.DesignConnection{conn("s", "task", "success"), conn("task", "e", "approve"), conn("task", "s", "reject")}
	result := designer.Validate(designer.Input{FlowType: enums.FlowTypeWorkflow, Nodes: nodes, Connections: conns, NodeTypes: builtinTypes()})
	if !result.Valid {
		t.Fatalf("workflow loop should be valid, got %+v", result.Errors)
	}
}

func TestValidateNodeModeAndReferences(t *testing.T) {
	nodes := []models.DesignNode{
		node("loop", "for", map[string]any{"do": "missing"}),
		node("task", "usertask", map[string]any{"assignees": []any{"u1"}}),
	}
	result := designer.Validate(designer.Input{FlowType: enums.FlowTypeRuleChain, Nodes: nodes, NodeTypes: builtinTypes()})
	if !hasCode(result.Errors, "node_ref_missing") {
		t.Fatalf("expected node_ref_missing, got %+v", result.Errors)
	}
	if !hasCode(result.Errors, "node_type_mode") {
		t.Fatalf("expected usertask to be rejected in rule chain, got %+v", result.Errors)
	}
}

func TestValidateRequiredFieldRespectsShowWhen(t *testing.T) {
	types := builtinTypes()
	group := node("d", "dingtalk", map[string]any{"target": "group", "msgtype": "text"})
	result := designer.Validate(designer.Input{FlowType: enums.FlowTypeRuleChain, Nodes: []models.DesignNode{group}, NodeTypes: types, ForPublish: true})
	if !hasCode(result.Errors, "config_required") || !strings.Contains(result.Errors[0].Message, "Webhook") {
		t.Fatalf("group target requires webhook, got %+v", result.Errors)
	}
	group.ExtParams["webhook"] = "https://oapi.dingtalk.com/robot/send?access_token=x"
	result = designer.Validate(designer.Input{FlowType: enums.FlowTypeRuleChain, Nodes: []models.DesignNode{group}, NodeTypes: types, ForPublish: true})
	if !result.Valid {
		t.Fatalf("user-only fields must not be required for group target, got %+v", result.Errors)
	}
}

func TestValidateDraftOnlyWarnsForMissingConfig(t *testing.T) {
	nodes := []models.DesignNode{node("h", "httpcall", nil)}
	result := designer.Validate(designer.Input{FlowType: enums.FlowTypeRuleChain, Nodes: nodes, NodeTypes: builtinTypes()})
	if !result.Valid || !hasCode(result.Warnings, "config_required") {
		t.Fatalf("missing config should be a warning for drafts, got %+v", result)
	}
}

func TestValidateUnknownRelationIsWarning(t *testing.T) {
	nodes := []models.DesignNode{
		node("sw", "switch", map[string]any{"cases": []any{map[string]any{"case": "msg.t > 1", "then": "hot"}}}),
		node("a", "end", nil),
		node("b", "end", nil),
	}
	conns := []models.DesignConnection{conn("sw", "a", "hot"), conn("sw", "b", "cold")}
	result := designer.Validate(designer.Input{FlowType: enums.FlowTypeRuleChain, Nodes: nodes, Connections: conns, NodeTypes: builtinTypes()})
	if !result.Valid {
		t.Fatalf("unexpected errors: %+v", result.Errors)
	}
	if len(result.Warnings) != 1 || result.Warnings[0].ToID != "b" || result.Warnings[0].Code != "conn_type_unknown" {
		t.Fatalf("expected warning only for the cold branch, got %+v", result.Warnings)
	}
}

func TestCompileReportsEveryNodeError(t *testing.T) {
	nodes := []models.DesignNode{
		node("f", "exprfilter", map[string]any{"expr": "msg.t >"}),
		node("d", "delay", map[string]any{"delay": "abc"}),
	}
	issues := designer.Compile(designer.ChainMeta{ChainNo: "c", Version: "1"}, nodes, nil, builtinTypes())
	if len(issues) != 2 || issues[0].NodeID != "f" || issues[1].NodeID != "d" {
		t.Fatalf("expected compile errors for both nodes, got %+v", issues)
	}
}

func TestEngineSupportIssues(t *testing.T) {
	types := builtinTypes()
	types["risk.check"] = models.NodeTypeInfo{NodeType: "risk.check", Name: "风控", Source: enums.NodeSourceService, FlowModes: []string{enums.FlowTypeRuleChain}}
	nodes := []models.DesignNode{node("s", "start", nil), node("r", "risk.check", nil), node("f", "exprfilter", map[string]any{"expr": "true"})}
	issues := designer.EngineSupportIssues(nodes, types)
	if len(issues) != 2 || issues[0].Code != "engine_unsupported" || issues[1].Code != "engine_host_required" {
		t.Fatalf("unexpected engine support issues: %+v", issues)
	}
}

func TestDebugRunsLoopAndJoinWithRealEngine(t *testing.T) {
	nodes := []models.DesignNode{
		node("loop", "for", map[string]any{"range": "msg.items", "do": "double", "mode": 1}),
		node("double", "exprtransform", map[string]any{"expr": "msg * 2"}),
		node("fork", "fork", nil),
		node("left", "exprtransform", map[string]any{"expr": `{"side": "left"}`}),
		node("right", "exprtransform", map[string]any{"expr": `{"side": "right"}`}),
		node("join", "join", map[string]any{"mergetomap": false}),
		node("push", "httpcall", map[string]any{"service": "http://127.0.0.1:1/never-called"}),
	}
	conns := []models.DesignConnection{
		conn("loop", "fork", "success"),
		conn("fork", "left", "success"),
		conn("fork", "right", "success"),
		conn("left", "join", "success"),
		conn("right", "join", "success"),
		conn("join", "push", "success"),
	}
	result := designer.Debug(context.Background(), designer.DebugInput{
		Meta:        designer.ChainMeta{ChainNo: "debug", Version: "1"},
		Nodes:       nodes,
		Connections: conns,
		NodeTypes:   builtinTypes(),
		Data:        map[string]any{"items": []any{1, 2, 3}},
		Timeout:     10 * time.Second,
		MockActions: true,
	})
	if !result.Success {
		t.Fatalf("debug failed: %+v", result)
	}
	counts := map[string]int{}
	var loopOutput, pushEvent any
	for _, ev := range result.Events {
		counts[ev.NodeID]++
		if ev.NodeID == "loop" {
			loopOutput = ev.Output
		}
		if ev.NodeID == "push" {
			pushEvent = ev
			if !ev.Mocked {
				t.Fatalf("httpcall must be mocked in debug: %+v", ev)
			}
		}
	}
	if counts["double"] != 3 || counts["join"] != 1 || counts["push"] != 1 {
		t.Fatalf("unexpected execution counts %v", counts)
	}
	if !reflect.DeepEqual(loopOutput, []any{float64(2), float64(4), float64(6)}) {
		t.Fatalf("for merge output = %#v", loopOutput)
	}
	if pushEvent == nil || len(result.Edges) != 6 {
		t.Fatalf("expected 6 traversed edges, got %+v", result.Edges)
	}
}

func TestRemapNodeIDsRewritesReferences(t *testing.T) {
	nodes := []models.DesignNode{
		node("a", "for", map[string]any{"do": "b"}),
		node("b", "exprtransform", map[string]any{"expr": "msg"}),
		node("g", "groupaction", map[string]any{"nodeids": "a, b", "list": []any{"b", "x"}}),
	}
	conns := []models.DesignConnection{conn("a", "g", "success")}
	mapping := map[string]string{"a": "a2", "b": "b2"}
	outNodes, outConns := designer.RemapNodeIDs(nodes, conns, mapping)
	if outNodes[0].ID != "a2" || outNodes[0].ExtParams["do"] != "b2" {
		t.Fatalf("for node not remapped: %+v", outNodes[0])
	}
	if outNodes[2].ExtParams["nodeids"] != "a2,b2" || !reflect.DeepEqual(outNodes[2].ExtParams["list"], []any{"b2", "x"}) {
		t.Fatalf("group node not remapped: %+v", outNodes[2].ExtParams)
	}
	if outConns[0].FromID != "a2" || outConns[0].ToID != "g" {
		t.Fatalf("connection not remapped: %+v", outConns[0])
	}
	if nodes[0].ID != "a" || nodes[0].ExtParams["do"] != "b" {
		t.Fatalf("input must not be modified: %+v", nodes[0])
	}
}

func TestNewNodeIDFormat(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		id := designer.NewNodeID()
		if len(id) != 12 || id[0] != 'n' || seen[id] {
			t.Fatalf("invalid or duplicate id %q", id)
		}
		seen[id] = true
	}
}
