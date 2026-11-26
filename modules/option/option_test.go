package option

import (
	"context"
	"errors"
	"testing"

	"github.com/zhiyunliu/glue/config"
	"github.com/zhiyunliu/glue/global"
	"github.com/zhiyunliu/manager-glow/constants/subcode"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/pkg/bizerr"
)

func TestListNormalizesOptions(t *testing.T) {
	Register("test_source", SourceFunc(func(context.Context) ([]models.Option, error) {
		return []models.Option{{Value: " a "}, {Label: "B", Value: "b"}, {Value: "a"}, {Value: ""}}, nil
	}))
	t.Cleanup(func() { Register("test_source", nil) })

	list, err := List(context.Background(), "test_source")
	if err != nil || len(list) != 2 || list[0] != (models.Option{Label: "a", Value: "a"}) || list[1].Label != "B" {
		t.Fatalf("list = %+v, %v", list, err)
	}
	if _, err := List(context.Background(), "missing"); err == nil {
		t.Fatal("missing source should fail")
	} else if e, ok := bizerr.IsBiz(err); !ok || e.GetSubCode() != subcode.OptionSourceNotFound {
		t.Fatalf("err = %v", err)
	}

	Register("test_source", SourceFunc(func(context.Context) ([]models.Option, error) { return nil, errors.New("down") }))
	if _, err := List(context.Background(), "test_source"); err == nil {
		t.Fatal("source error should be returned")
	}
}

func TestConfigEmailServerSource(t *testing.T) {
	old := global.Config
	t.Cleanup(func() { global.Config = old })
	cfg := config.New(config.WithSource(config.NewStrSource(`{"emails":[
		{"smtpname":"test1","host":"smtp1.example.com","port":465,"ssl":true,"username":"u","password":"secret"},
		{"smtpname":"test2","host":"smtp2.example.com"}
	]}`)))
	if err := cfg.Load(); err != nil {
		t.Fatalf("load config: %v", err)
	}
	global.Config = cfg

	list, err := List(context.Background(), SourceEmailServers)
	want := []models.Option{
		{Label: "test1", Value: "test1", Desc: "smtp1.example.com:465"},
		{Label: "test2", Value: "test2", Desc: "smtp2.example.com"},
	}
	if err != nil || len(list) != 2 || list[0] != want[0] || list[1] != want[1] {
		t.Fatalf("list = %+v, %v", list, err)
	}
}
