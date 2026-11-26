# manager-glow

glow-flow 流程编排平台的管理后台：在 SQL Server 中维护流程链(chain)定义，提供可视化流程设计器、执行历史、节点注册与操作审计。
后端基于 [glue](https://github.com/zhiyunliu/glue)，前端为 Vue3 + Vite + TypeScript + Element Plus + AntV X6，构建后嵌入 Go 二进制。

| 功能 | 说明 |
| --- | --- |
| 流程定义 | 规则链(RULE_CHAIN)/工作流(WORKFLOW)的创建、编辑、启停、复制、导入导出；版本管理(草稿 → 发布 → 下线/上线) |
| 流程设计器 | 统一画布支持两种模式切换；工具箱按模式加载节点；拖拽建模、按端口连线、属性面板(由节点属性描述生成)、撤销重做、自动布局、小地图；实时校验 + 服务端校验(含引擎编译)；预览调试；导入导出 |
| 执行历史 | 读取引擎状态存储中的实例与节点状态，按执行版本的流程图高亮节点状态 |
| 节点注册 | 内置节点(来自 glow-flow)与独立服务节点(注册到 nacos 等注册中心)的维护，查询服务实例 |
| 操作审计 | 所有变更操作(含失败)的审计日志，记录操作前后数据 |

## 目录结构

参考 plat-server 的分层：`services` 处理 HTTP，`modules` 承载业务与数据访问，`models` 定义参数与数据结构。

```
main.go                     服务入口
config/                     业务配置(配置文件 manager 节点)
constants/                  枚举、响应码、错误子码
models/                     请求参数、数据库模型、设计器模型
modules/
  flow/chain/               流程定义：分页、创建、修改、启停、删除、复制、导入导出
  flow/version/             版本：草稿、保存、校验、发布、下线、上线、调试
  flow/designer/            纯逻辑：模型校验、引擎编译校验、调试沙箱、节点编号重映射
  flow/store/               版本图(节点/连接/节点定义)读写
  flow/dao/, flow/sqls/     流程表数据访问
  nodetype/                 节点注册(含内置节点目录 builtin.go)
  option/                   设计器动态下拉数据源(如邮件服务器)
  history/                  执行历史
  audit/                    操作审计
services/                   HTTP 处理器与路由(services.go)
pkg/                        通用组件：统一响应、业务错误、操作人、前端嵌入
web/                        前端工程，构建产物 web/dist 由 web.go 嵌入
docs/sql/                   数据库脚本
test/                       与 glow-flow 引擎的集成测试
```

## 数据库

流程定义直接存储在 glow-flow 引擎使用的表中，发布后引擎(`contrib/repository/xdb`)即可加载，管理后台不维护第二份定义。

1. 执行 glow-flow 的建表脚本：`contrib/repository/xdb/sql.mssql.sql`、`contrib/statestorage/xdb/sql.mssql.sql`
2. 执行本项目的 `docs/sql/glowflow.manager.mssql.sql`(可重复执行)：
   - 为引擎表追加列：`flow_chain_definition.flow_type`、`flow_chain_version.version_no/publish_state/publish_time/publish_by/base_version_id`、`flow_node_definition.chain_no`、`flow_chain_node.node_name`(引擎查询使用显式列名，不受影响)
   - 新建 `flow_node_registry`(节点注册)、`flow_operate_log`(操作审计)

### 版本状态与引擎加载

| 状态 | publish_state | status | 说明 |
| --- | --- | --- | --- |
| 草稿 | 0 | 0 | 可编辑；每个流程最多一个草稿 |
| 已发布 | 1 | 1 | 不可修改 |
| 已下线 | 2 | 0 | 不可修改，可重新上线 |

引擎加载 `status=1` 的流程中 version_id 最大的 `status=1` 版本(界面标记为“生效中”)。下线生效版本后自动回退到上一个已发布版本；禁用流程后引擎不再加载。

### 节点编号

设计器节点编号即 `flow_node_definition.node_def_no`(全局唯一，`n` + 11 位随机字符)。同一流程的各版本共享节点编号；已在其他版本中使用的节点不允许修改类型，节点名称按版本保存在 `flow_chain_node.node_name`。复制流程、导入时与其他流程冲突的编号会重新生成，并同步替换连接及配置中引用的编号(如 for/while 的 `do`)。

## 运行

```bash
# 数据库密码通过环境变量注入，config.json 中使用 glue 的 {ENV} 占位符
export GLOWFLOW_DB_PASSWORD=xxx

cd web && npm ci && npm run build && cd ..   # 构建前端(产物嵌入二进制)
go build -o bin/manager-glow .
./bin/manager-glow run -C config.json
```

访问 `http://127.0.0.1:8190/api/manager-glow/web/`。启动时会自动补齐内置节点类型。

`go.mod` 使用 `replace` 指向同级目录的 `../glow-flow` 与 `../glue`，与 glow-flow 保持并行开发的方式一致。

### 配置(config.json 的 manager 节点)

| 字段 | 默认值 | 说明 |
| --- | --- | --- |
| db_name | glowflow | glue `dbs` 中流程库的连接名 |
| operator_headers | X-Auth-User-Id, X-Perm-Uid, X-Operator | 网关注入的操作人请求头，取第一个非空值；均为空时为 default_operator |
| debug_timeout / debug_max_timeout | 30 / 120 | 调试执行超时(秒) |
| debug_allow_real_actions | false | 调试时是否允许真实执行动作节点；为 false 时动作/集成类节点只模拟执行 |

登录态与权限由独立网关负责，后台不做本地权限判断，只记录操作人用于审计。
注册中心用于查询独立服务节点的实例：配置 glue 的 `registry`(如 `nacos://default`)及对应 `nacos` 节点；未配置时实例查询返回 503。

### 邮件服务器(config.json 的 emails 节点)

邮件节点通过下拉选择 `smtpname`，流程中只保存名称，不保存服务器地址和账号密码。服务器参数在执行时由引擎宿主按名称读取。管理后台和引擎宿主默认都读取各自 config.json 中的 `emails`，格式相同：

```json
"emails": [
  {"smtpname": "test1", "host": "smtp.example.com", "port": 465, "ssl": true,
   "username": "noreply@example.com", "password": "授权码", "from": "流程通知 <noreply@example.com>"}
]
```

- `port` 不填时，`ssl` 为 true 用 465，否则用 25。`ssl` 为 false 时，服务器支持则自动 STARTTLS。
- `from` 不填时使用 `username`。节点上配置的发件人优先。
- 下拉接口只返回名称和服务器地址，不返回账号密码。
- 如果流程里的 `smtpname` 不在下拉列表中，校验和发布时只提示警告，不阻止发布。

**更换数据源**：管理后台的下拉选项由 `modules/option` 的 `Source` 接口提供。实现该接口，并用同一名称注册覆盖默认实现即可，比如在 `main.go` 中：

```go
option.Register(option.SourceEmailServers, option.SourceFunc(func(ctx context.Context) ([]models.Option, error) {
    // 调用外部接口 / 查询数据库，返回 [{label, value, desc}]
}))
```

引擎宿主执行邮件节点时，服务器参数由 `action.EmailServerProvider` 提供。可以用 `action.SetEmailServerProvider(...)` 换成同一个数据源，保证两边的数据一致。

### 前端开发

```bash
cd web
npm ci
npm run dev        # http://localhost:5173 ，/api/manager-glow 代理到 127.0.0.1:8190(可用 GFM_BACKEND 修改)
npm test           # 校验器等单元测试
```

## 接口

统一前缀 `/api/manager-glow`，查询用 GET、变更用 POST。成功返回 `{"code":0,"message":"成功","data":...}`；失败返回对应 HTTP 状态码及 `{"code","sub_code","message","data"}`(模型校验失败为 422，`data` 为校验结果)。

| 模块 | 接口 |
| --- | --- |
| 流程定义 `/chain` | `page` `detail` `create` `update` `change-status` `delete` `copy` `export`(`download=1` 下载文件) `import` |
| 版本 `/version` | `list` `detail` `create-draft` `save-draft` `validate` `publish` `offline` `online` `delete` `debug` |
| 节点注册 `/node-type` | `list` `page` `detail` `create` `update` `change-status` `delete` `sync-builtin` `instances` |
| 动态下拉 `/option` | `list?source=`(如 `email_servers`) |
| 执行历史 `/history` | `page` `detail` |
| 操作审计 `/audit` | `page` `detail` |

## 与引擎的约定

- **校验**：保存草稿只拦截结构性错误(编号重复、连接断开等)；发布时执行完整校验：规则链 DAG、必填配置、节点引用、节点类型与流程类型匹配，并使用 glow-flow 真实的节点工厂编译模型。内置类型在引擎中尚未实现(如工作流的 start/usertask/subchain/rulechain)时禁止发布。
- **调试**：使用引擎真实的调度器与过滤/转换/逻辑节点执行当前画布；动作、集成类节点及独立服务节点默认只模拟(原样传递消息)，`command` 节点始终模拟。
- **内置节点升级**：启动时只补齐缺失的内置节点。内置节点的属性描述有变化时(例如邮件节点改为下拉选择 `smtpname`)，需要在“节点注册”页点击“同步内置节点”，或调用 `/node-type/sync-builtin`。
- **独立服务节点 / custom 节点**：管理后台只维护注册信息，执行需要引擎宿主注册对应的节点类型(或 `RegisterCall`)。
- **执行历史**：依赖引擎使用 xdb 状态存储(`contrib/statestorage/xdb`)写入 `flow_instance_state` / `flow_node_state`。
- **热加载**：发布后需要引擎重新加载定义(当前引擎在启动时加载)，发布通知未实现。
- **时间**：时间列使用数据库 `GETDATE()`，请确保 SQL Server 时区与业务时区一致(容器可设置 `TZ` 环境变量)。

## 测试

```bash
go test ./...                                                     # 单元测试(模型校验、编译、调试沙箱等)
GLOWFLOW_IT=1 GLOWFLOW_DB_PASSWORD=xxx GLOWFLOW_IT_CHAIN=<chain_no> \
  go test ./test/ -v                                              # 用 glow-flow 引擎加载并执行已发布的流程
cd web && npm test                                                # 前端单元测试
```
