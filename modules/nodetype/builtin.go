package nodetype

import (
	"github.com/zhiyunliu/manager-glow/constants/enums"
	"github.com/zhiyunliu/manager-glow/models"
	"github.com/zhiyunliu/manager-glow/modules/option"
)

// builtinNodeType 引擎内置节点的注册信息。属性描述来自 glow-flow 各节点实现中的 extparams 说明。
type builtinNodeType struct {
	NodeType      string
	Name          string
	Category      string
	FlowModes     []string
	RelationTypes []string
	Fields        []models.SchemaField
	DefaultConfig map[string]any
	Icon          string
	SortNo        int
	Desc          string
}

var (
	modeRuleChain = []string{enums.FlowTypeRuleChain}
	modeWorkflow  = []string{enums.FlowTypeWorkflow}
	modeBoth      = []string{enums.FlowTypeRuleChain, enums.FlowTypeWorkflow}

	relBool        = []string{enums.RelationTrue, enums.RelationFalse, enums.RelationFailure}
	relSuccessFail = []string{enums.RelationSuccess, enums.RelationFailure}
	relSuccess     = []string{enums.RelationSuccess}
)

// 分类默认颜色
var categoryColors = map[string]string{
	enums.NodeCategoryFilter:      "#E6A23C",
	enums.NodeCategoryTransform:   "#409EFF",
	enums.NodeCategoryAction:      "#67C23A",
	enums.NodeCategoryLogical:     "#9254DE",
	enums.NodeCategoryWorkflow:    "#13C2C2",
	enums.NodeCategoryIntegration: "#F56C6C",
}

const exprVarsHelp = "可用变量：msg(消息数据)、msgid、ts(时间戳秒)、instanceid；兼容 ${msg.x} 写法"

func timeoutField(placeholder string) models.SchemaField {
	return models.SchemaField{Key: "timeout", Label: "超时时间", Type: models.FieldString, Placeholder: placeholder, Help: "Go duration 格式(如 500ms、5s)或整数秒"}
}

func scriptField(lang, placeholder, help string) models.SchemaField {
	return models.SchemaField{Key: "script", Label: "脚本", Type: models.FieldCode, Lang: lang, Required: true, Placeholder: placeholder, Help: help}
}

var casesField = models.SchemaField{
	Key: "cases", Label: "条件分支", Type: models.FieldCases, Required: true,
	Help: "case 为 expr-lang 布尔表达式，then 为满足条件时走的连接类型；可用变量 msg、msgid、ts、instanceid、loop",
}

func selectOptions(pairs ...any) []models.SchemaOption {
	options := make([]models.SchemaOption, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		options = append(options, models.SchemaOption{Label: pairs[i].(string), Value: pairs[i+1]})
	}
	return options
}

var builtinNodeTypes = []builtinNodeType{
	// ---------- 过滤 ----------
	{
		NodeType: "exprfilter", Name: "表达式过滤", Category: enums.NodeCategoryFilter, FlowModes: modeBoth, RelationTypes: relBool,
		Icon: "Filter", SortNo: 100,
		Desc: "使用 expr-lang 布尔表达式判定消息：true/false 分支，表达式出错走 failure",
		Fields: []models.SchemaField{
			{Key: "expr", Label: "表达式", Type: models.FieldCode, Lang: "expr", Required: true, Placeholder: `msg.temperature > 50 && msg.status == "ok"`, Help: exprVarsHelp},
		},
	},
	{
		NodeType: "jsfilter", Name: "JS 过滤", Category: enums.NodeCategoryFilter, FlowModes: modeRuleChain, RelationTypes: relBool,
		Icon: "Filter", SortNo: 110,
		Desc: "JavaScript 脚本返回布尔值判定消息，脚本出错或超时走 failure",
		Fields: []models.SchemaField{
			scriptField("javascript", "return msg.temperature > 50;", "脚本作为函数体执行：function Filter(msg, msgId) { ... }，可用 $ctx 获取流程上下文"),
			timeoutField("2s"),
		},
	},
	{
		NodeType: "luafilter", Name: "Lua 过滤", Category: enums.NodeCategoryFilter, FlowModes: modeRuleChain, RelationTypes: relBool,
		Icon: "Filter", SortNo: 120,
		Desc: "Lua 脚本返回布尔值判定消息，脚本出错或超时走 failure",
		Fields: []models.SchemaField{
			scriptField("lua", "return msg.temperature > 50", "脚本作为函数体执行：function Filter(msg, msgId) ... end，可用 ctx:instance_id()"),
			timeoutField("2s"),
		},
	},

	// ---------- 转换 ----------
	{
		NodeType: "exprtransform", Name: "表达式转换", Category: enums.NodeCategoryTransform, FlowModes: modeBoth, RelationTypes: relSuccessFail,
		Icon: "Operation", SortNo: 200,
		Desc: "使用 expr-lang 表达式生成新的消息数据，expr 与 mapping 二选一(同时配置时 expr 优先)",
		Fields: []models.SchemaField{
			{Key: "expr", Label: "表达式", Type: models.FieldCode, Lang: "expr", Placeholder: `{"total": msg.price * msg.qty}`, Help: "结果作为新的消息数据；" + exprVarsHelp},
			{Key: "mapping", Label: "字段映射", Type: models.FieldKV, Help: "字段名 -> 表达式，结果组成对象作为新的消息数据"},
		},
	},
	{
		NodeType: "jstransform", Name: "JS 转换", Category: enums.NodeCategoryTransform, FlowModes: modeRuleChain, RelationTypes: relSuccessFail,
		Icon: "Operation", SortNo: 210,
		Desc: "JavaScript 脚本返回新的消息数据",
		Fields: []models.SchemaField{
			scriptField("javascript", "msg.total = msg.price * msg.qty;\nreturn msg;", "脚本作为函数体执行：function (msg, msgId) { ... }，return 的值作为新的消息数据"),
			timeoutField("2s"),
		},
	},
	{
		NodeType: "template", Name: "模板渲染", Category: enums.NodeCategoryTransform, FlowModes: modeBoth, RelationTypes: relSuccessFail,
		Icon: "Document", SortNo: 220,
		Desc: "使用 Go text/template 将消息渲染为字符串",
		Fields: []models.SchemaField{
			{Key: "template", Label: "模板", Type: models.FieldText, Required: true, Placeholder: "订单{{ .msg.order_no }}已支付，金额{{ .msg.amount }}", Help: "可用 .msg、.msgid、.ts、.instanceid；函数 escape、json"},
		},
	},

	// ---------- 动作 ----------
	{
		NodeType: "httpcall", Name: "HTTP 调用", Category: enums.NodeCategoryAction, FlowModes: modeBoth, RelationTypes: relSuccessFail,
		Icon: "Link", SortNo: 300, DefaultConfig: map[string]any{"method": "POST"},
		Desc: "通过 glue http 客户端发起请求，2xx 走 success 并以响应内容作为消息数据",
		Fields: []models.SchemaField{
			{Key: "service", Label: "请求地址", Type: models.FieldString, Required: true, Placeholder: "http://host/api/orders/${order_no}", Help: "支持 http/https/xhttp 及 ${} 模板"},
			{Key: "method", Label: "请求方法", Type: models.FieldSelect, Default: "POST", Options: selectOptions("POST", "POST", "GET", "GET", "PUT", "PUT", "DELETE", "DELETE", "PATCH", "PATCH", "HEAD", "HEAD")},
			{Key: "clientname", Label: "客户端配置", Type: models.FieldString, Placeholder: "default", Help: "glue xhttp 配置名称"},
			{Key: "headers", Label: "请求头", Type: models.FieldKV, Help: "值支持 ${} 模板"},
			timeoutField("5s"),
		},
	},
	{
		NodeType: "rpccall", Name: "RPC 调用", Category: enums.NodeCategoryAction, FlowModes: modeBoth, RelationTypes: relSuccessFail,
		Icon: "Connection", SortNo: 310,
		Desc: "通过 glue rpc 客户端调用服务，2xx 走 success 并以响应内容作为消息数据",
		Fields: []models.SchemaField{
			{Key: "service", Label: "服务地址", Type: models.FieldString, Required: true, Placeholder: "grpc://order-service/api/order/query", Help: "支持 ${} 模板"},
			{Key: "method", Label: "请求方法", Type: models.FieldString, Placeholder: "POST"},
			{Key: "clientname", Label: "客户端配置", Type: models.FieldString, Placeholder: "default", Help: "glue rpc 配置名称"},
			{Key: "headers", Label: "请求头", Type: models.FieldKV},
			timeoutField("5s"),
		},
	},
	{
		NodeType: "mqp", Name: "消息发送", Category: enums.NodeCategoryAction, FlowModes: modeBoth, RelationTypes: relSuccessFail,
		Icon: "Promotion", SortNo: 320,
		Desc: "通过 glue 队列发送消息数据",
		Fields: []models.SchemaField{
			{Key: "queuekey", Label: "队列 Key", Type: models.FieldString, Required: true, Help: "支持 ${} 模板"},
			{Key: "queuename", Label: "队列配置", Type: models.FieldString, Placeholder: "default", Help: "glue queues 配置名称"},
			{Key: "delay", Label: "延迟发送", Type: models.FieldString, Placeholder: "30s", Help: ">0 时延迟发送(按秒向上取整)"},
			timeoutField("5s"),
		},
	},
	{
		NodeType: "xdb", Name: "数据库操作", Category: enums.NodeCategoryAction, FlowModes: modeBoth, RelationTypes: relSuccessFail,
		Icon: "Coin", SortNo: 330, DefaultConfig: map[string]any{"execmethod": "queryas"},
		Desc: "通过 glue xdb 执行 SQL，消息数据作为 SQL 参数",
		Fields: []models.SchemaField{
			{Key: "dbconn", Label: "数据库配置", Type: models.FieldString, Required: true, Help: "glue dbs 配置名称"},
			{Key: "execmethod", Label: "执行方式", Type: models.FieldSelect, Required: true, Default: "queryas", Options: selectOptions("查询列表(queryas)", "queryas", "查询首行(firstas)", "firstas", "执行(exec)", "exec")},
			{Key: "execsql", Label: "SQL", Type: models.FieldCode, Lang: "sql", Required: true, Placeholder: "select * from orders where order_no = @{order_no}"},
			{Key: "stream", Label: "流式输出", Type: models.FieldBool, Help: "仅 queryas 支持"},
		},
	},
	{
		NodeType: "command", Name: "执行命令", Category: enums.NodeCategoryAction, FlowModes: modeRuleChain, RelationTypes: relSuccessFail,
		Icon: "Monitor", SortNo: 340,
		Desc: "在引擎所在主机执行命令(不经过 shell)，退出码为 0 走 success",
		Fields: []models.SchemaField{
			{Key: "command", Label: "命令", Type: models.FieldString, Required: true, Placeholder: "/usr/bin/python3"},
			{Key: "args", Label: "参数", Type: models.FieldList, Help: "每一项支持 ${} 模板"},
			{Key: "workdir", Label: "工作目录", Type: models.FieldString},
			{Key: "env", Label: "环境变量", Type: models.FieldKV},
			{Key: "stdin", Label: "消息写入标准输入", Type: models.FieldBool},
			timeoutField("30s"),
		},
	},
	{
		NodeType: "email", Name: "发送邮件", Category: enums.NodeCategoryAction, FlowModes: modeBoth, RelationTypes: relSuccessFail,
		Icon: "Message", SortNo: 350, DefaultConfig: map[string]any{"contenttype": "html"},
		Desc: "通过 SMTP 发送邮件",
		Fields: []models.SchemaField{
			{Key: "smtpname", Label: "邮件服务器", Type: models.FieldSelect, Required: true, OptionsSource: option.SourceEmailServers, Group: "服务器",
				Help: "SMTP 服务器配置名称，对应配置 emails[].smtpname；服务器地址、账号密码由引擎宿主按名称读取，不保存在流程中"},
			{Key: "from", Label: "发件人", Type: models.FieldString, Placeholder: "名称 <addr@example.com>", Group: "邮件", Help: "为空时使用服务器配置的发件人"},
			{Key: "to", Label: "收件人", Type: models.FieldList, Group: "邮件", Help: "to/cc/bcc 至少一个不为空，支持 ${} 模板"},
			{Key: "cc", Label: "抄送", Type: models.FieldList, Group: "邮件"},
			{Key: "bcc", Label: "密送", Type: models.FieldList, Group: "邮件"},
			{Key: "subject", Label: "主题", Type: models.FieldString, Group: "邮件"},
			{Key: "content", Label: "正文", Type: models.FieldText, Group: "邮件", Help: "为空时使用消息数据"},
			{Key: "contenttype", Label: "正文格式", Type: models.FieldSelect, Default: "html", Options: selectOptions("HTML", "html", "纯文本", "text"), Group: "邮件"},
			timeoutField("30s"),
		},
	},
	{
		NodeType: "dingtalk", Name: "钉钉通知", Category: enums.NodeCategoryAction, FlowModes: modeBoth, RelationTypes: relSuccessFail,
		Icon: "Bell", SortNo: 360, DefaultConfig: map[string]any{"target": "group", "msgtype": "text"},
		Desc: "发送钉钉群机器人消息或企业内部应用工作通知",
		Fields: []models.SchemaField{
			{Key: "target", Label: "发送对象", Type: models.FieldSelect, Required: true, Default: "group", Options: selectOptions("群机器人", "group", "指定到人", "user")},
			{Key: "msgtype", Label: "消息类型", Type: models.FieldSelect, Default: "text", Options: selectOptions("文本", "text", "Markdown", "markdown", "图片", "image")},
			{Key: "title", Label: "标题", Type: models.FieldString, Help: "markdown/image 消息使用，默认为节点名称"},
			{Key: "content", Label: "内容", Type: models.FieldText, Help: "支持 ${} 模板；为空时使用消息数据"},
			{Key: "imageurl", Label: "图片地址", Type: models.FieldString, Required: true, ShowWhen: map[string]any{"msgtype": "image"}},
			{Key: "webhook", Label: "Webhook", Type: models.FieldString, Required: true, ShowWhen: map[string]any{"target": "group"}},
			{Key: "secret", Label: "加签密钥", Type: models.FieldString, ShowWhen: map[string]any{"target": "group"}},
			{Key: "atmobiles", Label: "@手机号", Type: models.FieldList, ShowWhen: map[string]any{"target": "group"}},
			{Key: "atuserids", Label: "@userId", Type: models.FieldList, ShowWhen: map[string]any{"target": "group"}},
			{Key: "atall", Label: "@所有人", Type: models.FieldBool, ShowWhen: map[string]any{"target": "group"}},
			{Key: "appkey", Label: "AppKey", Type: models.FieldString, Required: true, ShowWhen: map[string]any{"target": "user"}},
			{Key: "appsecret", Label: "AppSecret", Type: models.FieldString, Required: true, ShowWhen: map[string]any{"target": "user"}},
			{Key: "agentid", Label: "AgentId", Type: models.FieldString, Required: true, ShowWhen: map[string]any{"target": "user"}},
			{Key: "userids", Label: "接收人 userId", Type: models.FieldList, Required: true, ShowWhen: map[string]any{"target": "user"}},
			timeoutField("10s"),
		},
	},
	{
		NodeType: "delay", Name: "延迟", Category: enums.NodeCategoryAction, FlowModes: modeBoth, RelationTypes: relSuccess,
		Icon: "Timer", SortNo: 370,
		Desc: "等待指定时间后原样传递消息",
		Fields: []models.SchemaField{
			{Key: "delay", Label: "延迟时间", Type: models.FieldString, Required: true, Placeholder: "5s", Help: "Go duration 格式，如 500ms、5s、1m"},
		},
	},
	{
		NodeType: "custom", Name: "自定义函数", Category: enums.NodeCategoryAction, FlowModes: modeBoth, RelationTypes: relSuccessFail,
		Icon: "MagicStick", SortNo: 380,
		Desc: "调用引擎宿主进程通过 RegisterCall 注册的函数",
		Fields: []models.SchemaField{
			{Key: "nodetype", Label: "函数类型", Type: models.FieldString, Required: true, Help: "引擎中 RegisterCall 注册的名称"},
		},
	},

	// ---------- 逻辑控制 ----------
	{
		NodeType: "switch", Name: "条件分支", Category: enums.NodeCategoryLogical, FlowModes: modeBoth,
		Icon: "Switch", SortNo: 400,
		Desc:   "按顺序计算条件，走第一个满足条件的分支；都不满足时走 default",
		Fields: []models.SchemaField{casesField},
	},
	{
		NodeType: "inclusive", Name: "包容分支", Category: enums.NodeCategoryLogical, FlowModes: modeBoth, RelationTypes: []string{enums.RelationFailure},
		Icon: "Share", SortNo: 410,
		Desc:   "同时走所有满足条件的分支；都不满足时走 default，可在下游使用汇聚节点",
		Fields: []models.SchemaField{casesField},
	},
	{
		NodeType: "fork", Name: "并行分支", Category: enums.NodeCategoryLogical, FlowModes: modeBoth, RelationTypes: relSuccess,
		Icon: "Share", SortNo: 420,
		Desc: "将消息发送给所有 success/default 连接的后续节点并行执行，通常与汇聚节点配合",
	},
	{
		NodeType: "join", Name: "汇聚", Category: enums.NodeCategoryLogical, FlowModes: modeBoth, RelationTypes: relSuccessFail,
		Icon: "Aim", SortNo: 430,
		Desc: "等待所有可能到达的上游分支结束后合并消息，只执行一次后续节点；超时走 failure",
		Fields: []models.SchemaField{
			timeoutField("30s"),
			{Key: "mergetomap", Label: "合并为对象", Type: models.FieldBool, Help: "默认输出数组 [{nodeId, data}]"},
		},
	},
	{
		NodeType: "for", Name: "遍历循环", Category: enums.NodeCategoryLogical, FlowModes: modeRuleChain, RelationTypes: relSuccessFail,
		Icon: "Refresh", SortNo: 440, DefaultConfig: map[string]any{"mode": 0},
		Desc: "遍历数组/map，对每一项执行循环体",
		Fields: []models.SchemaField{
			{Key: "range", Label: "遍历对象", Type: models.FieldCode, Lang: "expr", Placeholder: "msg.items", Help: "expr-lang 表达式，为空时遍历消息数据本身"},
			{Key: "do", Label: "循环体起始节点", Type: models.FieldNodeRef, Required: true, Help: "该节点及其后续节点每次迭代执行一次"},
			{Key: "mode", Label: "结果处理", Type: models.FieldSelect, Default: 0, Options: selectOptions("忽略结果(ignore)", 0, "合并为数组(merge)", 1, "逐次替换(replace)", 2, "异步执行(async)", 3)},
		},
	},
	{
		NodeType: "while", Name: "条件循环", Category: enums.NodeCategoryLogical, FlowModes: modeRuleChain, RelationTypes: relSuccessFail,
		Icon: "RefreshRight", SortNo: 450, DefaultConfig: map[string]any{"mode": 2},
		Desc: "条件满足时重复执行循环体",
		Fields: []models.SchemaField{
			{Key: "condition", Label: "循环条件", Type: models.FieldCode, Lang: "expr", Required: true, Placeholder: "msg.count < 10"},
			{Key: "do", Label: "循环体起始节点", Type: models.FieldNodeRef, Required: true},
			{Key: "mode", Label: "结果处理", Type: models.FieldSelect, Default: 2, Options: selectOptions("输出原消息(ignore)", 0, "合并为数组(merge)", 1, "最后一次结果(replace)", 2)},
			{Key: "maxloops", Label: "最大迭代次数", Type: models.FieldNumber, Help: "超过时走 failure，0 表示不限制"},
		},
	},
	{
		NodeType: "break", Name: "跳出循环", Category: enums.NodeCategoryLogical, FlowModes: modeRuleChain, RelationTypes: relSuccess,
		Icon: "Close", SortNo: 460,
		Desc: "在循环体中使用：本次迭代结束后退出最近一层循环",
	},
	{
		NodeType: "groupaction", Name: "节点组", Category: enums.NodeCategoryLogical, FlowModes: modeRuleChain, RelationTypes: relSuccessFail,
		Icon: "Grid", SortNo: 470,
		Desc: "并行执行一组节点(只执行节点本身)，匹配数达到要求走 success",
		Fields: []models.SchemaField{
			{Key: "nodeids", Label: "节点列表", Type: models.FieldNodeRefs, Required: true},
			{Key: "matchrelationtype", Label: "匹配的结果类型", Type: models.FieldString, Placeholder: "success"},
			{Key: "matchnum", Label: "匹配数量", Type: models.FieldNumber, Help: "0 表示全部节点都需匹配"},
			timeoutField("30s"),
			{Key: "mergetomap", Label: "合并为对象", Type: models.FieldBool},
		},
	},
	{
		NodeType: "end", Name: "结束", Category: enums.NodeCategoryLogical, FlowModes: modeBoth,
		Icon: "CircleCheck", SortNo: 480,
		Desc: "显式结束当前分支，当前消息作为分支结果",
	},

	// ---------- 工作流 ----------
	{
		NodeType: "start", Name: "开始", Category: enums.NodeCategoryWorkflow, FlowModes: modeWorkflow, RelationTypes: relSuccess,
		Icon: "VideoPlay", SortNo: 500,
		Desc: "工作流开始事件(glow-flow 引擎暂未实现)",
	},
	{
		NodeType: "usertask", Name: "人工任务", Category: enums.NodeCategoryWorkflow, FlowModes: modeWorkflow, RelationTypes: []string{"approve", "reject", enums.RelationFailure},
		Icon: "User", SortNo: 510,
		Desc: "等待人工处理的任务节点，支持审批通过/驳回(glow-flow 引擎暂未实现)",
		Fields: []models.SchemaField{
			{Key: "assignees", Label: "处理人", Type: models.FieldList, Required: true},
			{Key: "mode", Label: "会签方式", Type: models.FieldSelect, Default: "any", Options: selectOptions("或签", "any", "会签", "all")},
			{Key: "form", Label: "表单", Type: models.FieldJSON},
			timeoutField("72h"),
		},
	},
	{
		NodeType: "subchain", Name: "子流程", Category: enums.NodeCategoryWorkflow, FlowModes: modeWorkflow, RelationTypes: relSuccessFail,
		Icon: "Files", SortNo: 520,
		Desc: "调用另一个流程(glow-flow 引擎暂未实现)",
		Fields: []models.SchemaField{
			{Key: "chain_no", Label: "子流程编号", Type: models.FieldString, Required: true},
		},
	},
	{
		NodeType: "rulechain", Name: "规则链调用", Category: enums.NodeCategoryWorkflow, FlowModes: modeWorkflow, RelationTypes: relSuccessFail,
		Icon: "Guide", SortNo: 530,
		Desc: "工作流中同步调用规则链并将结果写回流程变量(glow-flow 引擎暂未实现)",
		Fields: []models.SchemaField{
			{Key: "chain_no", Label: "规则链编号", Type: models.FieldString, Required: true},
			{Key: "result_var", Label: "结果变量", Type: models.FieldString},
		},
	},
}

// builtinInfos 内置节点注册信息
func builtinInfos() []models.NodeTypeInfo {
	infos := make([]models.NodeTypeInfo, 0, len(builtinNodeTypes))
	for _, item := range builtinNodeTypes {
		fields := item.Fields
		if fields == nil {
			fields = []models.SchemaField{}
		}
		relations := item.RelationTypes
		if relations == nil {
			relations = []string{}
		}
		defaults := item.DefaultConfig
		if defaults == nil {
			defaults = map[string]any{}
		}
		infos = append(infos, models.NodeTypeInfo{
			NodeType:      item.NodeType,
			Name:          item.Name,
			Category:      item.Category,
			FlowModes:     item.FlowModes,
			Source:        enums.NodeSourceBuiltin,
			RelationTypes: relations,
			ConfigSchema:  &models.ConfigSchema{Fields: fields},
			DefaultConfig: defaults,
			Icon:          item.Icon,
			Color:         categoryColors[item.Category],
			SortNo:        item.SortNo,
			Status:        enums.StatusEnabled,
			Desc:          item.Desc,
		})
	}
	return infos
}

// BuiltinInfos 内置节点注册信息(副本)
func BuiltinInfos() []models.NodeTypeInfo {
	return builtinInfos()
}
