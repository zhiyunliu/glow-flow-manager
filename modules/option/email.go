package option

import (
	"context"
	"net"
	"strconv"

	"github.com/zhiyunliu/glow-flow/contrib/nodedescriptor/action"
	"github.com/zhiyunliu/manager-glow/models"
)

// SourceEmailServers 邮件服务器下拉数据源，对应 email 节点的 smtpname 配置
const SourceEmailServers = "email_servers"

// ConfigEmailServerSource 从配置文件 emails 节点读取 SMTP 服务器列表(与引擎宿主读取的格式一致，见 action.EmailServer)：
//
//	"emails": [{"smtpname":"test1","host":"smtp.example.com","port":465,"ssl":true,"username":"...","password":"...","from":"..."}]
//
// 只返回名称及服务器地址，不返回账号密码
type ConfigEmailServerSource struct {
	Key string // 配置节点名称，默认 emails
}

func (s ConfigEmailServerSource) Options(context.Context) ([]models.Option, error) {
	servers, err := action.ConfigEmailServerProvider{Key: s.Key}.ListEmailServers()
	if err != nil {
		return nil, err
	}
	list := make([]models.Option, 0, len(servers))
	for _, server := range servers {
		desc := server.Host
		if server.Host != "" && server.Port > 0 {
			desc = net.JoinHostPort(server.Host, strconv.Itoa(server.Port))
		}
		list = append(list, models.Option{Label: server.SmtpName, Value: server.SmtpName, Desc: desc})
	}
	return list, nil
}

func init() {
	Register(SourceEmailServers, ConfigEmailServerSource{})
}
