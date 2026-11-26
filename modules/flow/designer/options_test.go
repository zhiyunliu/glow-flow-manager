package designer

import (
	"errors"
	"testing"

	"github.com/zhiyunliu/manager-glow/models"
)

func TestOptionSourceIssues(t *testing.T) {
	types := map[string]models.NodeTypeInfo{
		"email": {NodeType: "email", ConfigSchema: &models.ConfigSchema{Fields: []models.SchemaField{
			{Key: "smtpname", Label: "邮件服务器", Type: models.FieldSelect, OptionsSource: "email_servers"},
			{Key: "contenttype", Type: models.FieldSelect},
		}}},
	}
	nodes := []models.DesignNode{
		{ID: "n1", Type: "email", Name: "a", ExtParams: map[string]any{"smtpname": "test1"}},
		{ID: "n2", Type: "email", Name: "b", ExtParams: map[string]any{"smtpname": "gone"}},
		{ID: "n3", Type: "email", Name: "c", ExtParams: map[string]any{}},
	}
	calls := 0
	issues := OptionSourceIssues(nodes, types, func(source string) (map[string]struct{}, error) {
		calls++
		if source != "email_servers" {
			t.Fatalf("source = %s", source)
		}
		return map[string]struct{}{"test1": {}, "test2": {}}, nil
	})
	if calls != 1 {
		t.Fatalf("lookup calls = %d, want cached once", calls)
	}
	if len(issues) != 1 || issues[0].NodeID != "n2" || issues[0].Level != LevelWarning || issues[0].Code != "option_not_found" {
		t.Fatalf("issues = %+v", issues)
	}

	issues = OptionSourceIssues(nodes, types, func(string) (map[string]struct{}, error) {
		return nil, errors.New("down")
	})
	if len(issues) != 1 || issues[0].Code != "option_source_unavailable" || issues[0].Level != LevelWarning {
		t.Fatalf("issues = %+v", issues)
	}
}
