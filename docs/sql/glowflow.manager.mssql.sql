/*
manager-glow 数据库脚本(SQL Server 2016+)，可重复执行。

执行前提：已执行 glow-flow 引擎的建表脚本
  - glow-flow/contrib/repository/xdb/sql.mssql.sql   (流程定义：flow_chain_* / flow_node_definition / flow_basic_infra)
  - glow-flow/contrib/statestorage/xdb/sql.mssql.sql (执行状态：flow_instance_state / flow_node_state)

本脚本：
  1. 为引擎表追加管理后台需要的列(引擎查询使用显式列名，追加列不影响引擎加载)
  2. 创建节点注册表 flow_node_registry、操作审计表 flow_operate_log
*/
USE glowflow;
GO

IF OBJECT_ID(N'dbo.flow_chain_definition', N'U') IS NULL
	OR OBJECT_ID(N'dbo.flow_chain_version', N'U') IS NULL
	OR OBJECT_ID(N'dbo.flow_node_definition', N'U') IS NULL
	OR OBJECT_ID(N'dbo.flow_chain_node', N'U') IS NULL
	OR OBJECT_ID(N'dbo.flow_chain_connection', N'U') IS NULL
BEGIN
	RAISERROR(N'缺少 glow-flow 引擎表，请先执行 glow-flow/contrib/repository/xdb/sql.mssql.sql', 16, 1);
	SET NOEXEC ON;
END;
GO

/* ---------- flow_chain_definition：流程类型 ---------- */
IF COL_LENGTH(N'dbo.flow_chain_definition', N'flow_type') IS NULL
BEGIN
	ALTER TABLE dbo.flow_chain_definition
		ADD flow_type varchar(32) NOT NULL
			CONSTRAINT DF_flow_chain_definition_flow_type DEFAULT ('RULE_CHAIN')
			CONSTRAINT CK_flow_chain_definition_flow_type CHECK (flow_type IN ('RULE_CHAIN', 'WORKFLOW'));
END;
GO

/* ---------- flow_chain_version：版本号、发布状态 ---------- */
IF COL_LENGTH(N'dbo.flow_chain_version', N'version_no') IS NULL
BEGIN
	ALTER TABLE dbo.flow_chain_version
		ADD version_no int NOT NULL CONSTRAINT DF_flow_chain_version_version_no DEFAULT (0);

	-- 已有版本按 version_id 顺序编号
	EXEC(N'
		WITH numbered AS (
			SELECT version_no, ROW_NUMBER() OVER (PARTITION BY chain_no ORDER BY version_id) AS rn
			FROM dbo.flow_chain_version
		)
		UPDATE numbered SET version_no = rn;');
END;
GO

IF COL_LENGTH(N'dbo.flow_chain_version', N'publish_state') IS NULL
BEGIN
	ALTER TABLE dbo.flow_chain_version
		ADD publish_state tinyint NOT NULL
			CONSTRAINT DF_flow_chain_version_publish_state DEFAULT (0)
			CONSTRAINT CK_flow_chain_version_publish_state CHECK (publish_state IN (0, 1, 2));

	-- 已有版本：启用视为已发布，禁用视为已下线
	EXEC(N'UPDATE dbo.flow_chain_version SET publish_state = CASE WHEN [status] = 1 THEN 1 ELSE 2 END;');
END;
GO

IF COL_LENGTH(N'dbo.flow_chain_version', N'publish_time') IS NULL
	ALTER TABLE dbo.flow_chain_version ADD publish_time datetime NULL;
GO
IF COL_LENGTH(N'dbo.flow_chain_version', N'publish_by') IS NULL
	ALTER TABLE dbo.flow_chain_version ADD publish_by varchar(32) NULL;
GO
IF COL_LENGTH(N'dbo.flow_chain_version', N'base_version_id') IS NULL
	ALTER TABLE dbo.flow_chain_version ADD base_version_id bigint NULL;
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE [name] = N'IX_flow_chain_version_chain_state' AND object_id = OBJECT_ID(N'dbo.flow_chain_version'))
	CREATE INDEX IX_flow_chain_version_chain_state ON dbo.flow_chain_version (chain_no, publish_state, version_id);
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE [name] = N'UX_flow_chain_version_chain_version_no' AND object_id = OBJECT_ID(N'dbo.flow_chain_version'))
	CREATE UNIQUE INDEX UX_flow_chain_version_chain_version_no ON dbo.flow_chain_version (chain_no, version_no);
GO

/* ---------- flow_node_definition：所属流程 ---------- */
IF COL_LENGTH(N'dbo.flow_node_definition', N'chain_no') IS NULL
	ALTER TABLE dbo.flow_node_definition ADD chain_no varchar(32) NULL;
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE [name] = N'IX_flow_node_definition_chain_no' AND object_id = OBJECT_ID(N'dbo.flow_node_definition'))
	CREATE INDEX IX_flow_node_definition_chain_no ON dbo.flow_node_definition (chain_no);
GO

/* ---------- flow_chain_node：版本内节点名称 ---------- */
IF COL_LENGTH(N'dbo.flow_chain_node', N'node_name') IS NULL
	ALTER TABLE dbo.flow_chain_node ADD node_name varchar(128) NULL;
GO

/* ---------- flow_node_registry：节点注册 ---------- */
IF OBJECT_ID(N'dbo.flow_node_registry', N'U') IS NULL
BEGIN
	CREATE TABLE dbo.flow_node_registry (
		node_type varchar(64) NOT NULL,
		[name] nvarchar(128) NOT NULL,
		category varchar(32) NOT NULL,
		flow_modes varchar(64) NOT NULL CONSTRAINT DF_flow_node_registry_flow_modes DEFAULT ('RULE_CHAIN'),
		[source] varchar(16) NOT NULL CONSTRAINT DF_flow_node_registry_source DEFAULT ('service'),
		service_name varchar(128) NULL,
		service_path varchar(256) NULL,
		protocol varchar(16) NULL,
		relation_types varchar(512) NULL,
		config_schema nvarchar(max) NULL,
		default_config nvarchar(max) NULL,
		icon varchar(64) NULL,
		color varchar(16) NULL,
		sort_no int NOT NULL CONSTRAINT DF_flow_node_registry_sort_no DEFAULT (0),
		[status] tinyint NOT NULL CONSTRAINT DF_flow_node_registry_status DEFAULT (1),
		[desc] nvarchar(1024) NULL,
		create_time datetime NOT NULL CONSTRAINT DF_flow_node_registry_create_time DEFAULT (GETDATE()),
		create_by varchar(32) NULL,
		[operator] varchar(32) NULL,
		update_time datetime NULL,
		CONSTRAINT PK_flow_node_registry PRIMARY KEY CLUSTERED (node_type),
		CONSTRAINT CK_flow_node_registry_status CHECK ([status] IN (0, 1)),
		CONSTRAINT CK_flow_node_registry_source CHECK ([source] IN ('builtin', 'service')),
		CONSTRAINT CK_flow_node_registry_config_schema CHECK (config_schema IS NULL OR ISJSON(config_schema) = 1),
		CONSTRAINT CK_flow_node_registry_default_config CHECK (default_config IS NULL OR ISJSON(default_config) = 1)
	);
END;
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE [name] = N'IX_flow_node_registry_category' AND object_id = OBJECT_ID(N'dbo.flow_node_registry'))
	CREATE INDEX IX_flow_node_registry_category ON dbo.flow_node_registry (category, sort_no);
GO

/* ---------- flow_operate_log：操作审计 ---------- */
IF OBJECT_ID(N'dbo.flow_operate_log', N'U') IS NULL
BEGIN
	CREATE TABLE dbo.flow_operate_log (
		log_id bigint IDENTITY(1, 1) NOT NULL,
		trace_id varchar(64) NULL,
		module varchar(32) NOT NULL,
		operate_type varchar(32) NOT NULL,
		target_id varchar(64) NOT NULL,
		chain_no varchar(32) NULL,
		version_id bigint NULL,
		summary nvarchar(512) NULL,
		before_data nvarchar(max) NULL,
		after_data nvarchar(max) NULL,
		result tinyint NOT NULL,
		error_msg nvarchar(1024) NULL,
		[operator] varchar(64) NULL,
		client_ip varchar(64) NULL,
		create_time datetime NOT NULL CONSTRAINT DF_flow_operate_log_create_time DEFAULT (GETDATE()),
		CONSTRAINT PK_flow_operate_log PRIMARY KEY CLUSTERED (log_id),
		CONSTRAINT CK_flow_operate_log_result CHECK (result IN (0, 1))
	);
END;
GO

IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE [name] = N'IX_flow_operate_log_create_time' AND object_id = OBJECT_ID(N'dbo.flow_operate_log'))
	CREATE INDEX IX_flow_operate_log_create_time ON dbo.flow_operate_log (create_time DESC);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE [name] = N'IX_flow_operate_log_module_time' AND object_id = OBJECT_ID(N'dbo.flow_operate_log'))
	CREATE INDEX IX_flow_operate_log_module_time ON dbo.flow_operate_log (module, operate_type, create_time DESC);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE [name] = N'IX_flow_operate_log_chain_time' AND object_id = OBJECT_ID(N'dbo.flow_operate_log'))
	CREATE INDEX IX_flow_operate_log_chain_time ON dbo.flow_operate_log (chain_no, create_time DESC);
GO
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE [name] = N'IX_flow_operate_log_target' AND object_id = OBJECT_ID(N'dbo.flow_operate_log'))
	CREATE INDEX IX_flow_operate_log_target ON dbo.flow_operate_log (target_id);
GO

/* ---------- 执行历史查询索引(flow_instance_state 为引擎状态表) ---------- */
IF OBJECT_ID(N'dbo.flow_instance_state', N'U') IS NOT NULL
	AND NOT EXISTS (SELECT 1 FROM sys.indexes WHERE [name] = N'IX_flow_instance_state_create_time' AND object_id = OBJECT_ID(N'dbo.flow_instance_state'))
	CREATE INDEX IX_flow_instance_state_create_time ON dbo.flow_instance_state (create_time DESC);
GO

/* ---------- 字段说明 ---------- */
DECLARE @descriptions TABLE (
	table_name sysname NOT NULL,
	column_name sysname NULL,
	description nvarchar(4000) NOT NULL
);

INSERT INTO @descriptions (table_name, column_name, description)
VALUES
	(N'flow_chain_definition', N'flow_type', N'流程类型：RULE_CHAIN 规则链，WORKFLOW 工作流'),
	(N'flow_chain_version', N'version_no', N'流程内版本序号，从 1 递增'),
	(N'flow_chain_version', N'publish_state', N'发布状态：0 草稿，1 已发布，2 已下线'),
	(N'flow_chain_version', N'publish_time', N'发布时间'),
	(N'flow_chain_version', N'publish_by', N'发布人'),
	(N'flow_chain_version', N'base_version_id', N'复制来源版本 ID'),
	(N'flow_node_definition', N'chain_no', N'所属流程链编号'),
	(N'flow_chain_node', N'node_name', N'版本内节点名称'),

	(N'flow_node_registry', NULL, N'节点注册：设计器可用的节点类型'),
	(N'flow_node_registry', N'node_type', N'节点类型，对应 flow_node_definition.type'),
	(N'flow_node_registry', N'name', N'节点名称'),
	(N'flow_node_registry', N'category', N'分类：filter/transform/action/logical/workflow/integration'),
	(N'flow_node_registry', N'flow_modes', N'适用流程类型，逗号分隔：RULE_CHAIN,WORKFLOW'),
	(N'flow_node_registry', N'source', N'来源：builtin 引擎内置，service 独立服务'),
	(N'flow_node_registry', N'service_name', N'注册中心(nacos)中的服务名'),
	(N'flow_node_registry', N'service_path', N'服务调用路径'),
	(N'flow_node_registry', N'protocol', N'服务协议：grpc/http'),
	(N'flow_node_registry', N'relation_types', N'输出连接类型，JSON 数组'),
	(N'flow_node_registry', N'config_schema', N'属性面板配置描述，JSON'),
	(N'flow_node_registry', N'default_config', N'默认配置，JSON'),
	(N'flow_node_registry', N'icon', N'图标'),
	(N'flow_node_registry', N'color', N'颜色'),
	(N'flow_node_registry', N'sort_no', N'排序'),
	(N'flow_node_registry', N'status', N'状态：1 启用，0 禁用'),
	(N'flow_node_registry', N'desc', N'描述'),

	(N'flow_operate_log', NULL, N'操作审计日志'),
	(N'flow_operate_log', N'log_id', N'日志 ID'),
	(N'flow_operate_log', N'trace_id', N'请求链路 ID'),
	(N'flow_operate_log', N'module', N'模块：chain/version/node_type'),
	(N'flow_operate_log', N'operate_type', N'操作类型'),
	(N'flow_operate_log', N'target_id', N'操作对象标识'),
	(N'flow_operate_log', N'chain_no', N'流程链编号'),
	(N'flow_operate_log', N'version_id', N'版本 ID'),
	(N'flow_operate_log', N'summary', N'操作摘要'),
	(N'flow_operate_log', N'before_data', N'操作前数据，JSON'),
	(N'flow_operate_log', N'after_data', N'操作后数据，JSON'),
	(N'flow_operate_log', N'result', N'结果：1 成功，0 失败'),
	(N'flow_operate_log', N'error_msg', N'失败原因'),
	(N'flow_operate_log', N'operator', N'操作人'),
	(N'flow_operate_log', N'client_ip', N'客户端 IP'),
	(N'flow_operate_log', N'create_time', N'操作时间');

DECLARE @table_name sysname, @column_name sysname, @description nvarchar(4000);
DECLARE description_cursor CURSOR LOCAL FAST_FORWARD FOR
	SELECT table_name, column_name, description FROM @descriptions;
OPEN description_cursor;
FETCH NEXT FROM description_cursor INTO @table_name, @column_name, @description;
WHILE @@FETCH_STATUS = 0
BEGIN
	IF @column_name IS NULL
	BEGIN
		IF EXISTS (SELECT 1 FROM sys.extended_properties WHERE class = 1 AND major_id = OBJECT_ID(N'dbo.' + @table_name) AND minor_id = 0 AND [name] = N'MS_Description')
			EXEC sys.sp_updateextendedproperty @name = N'MS_Description', @value = @description,
				@level0type = N'SCHEMA', @level0name = N'dbo', @level1type = N'TABLE', @level1name = @table_name;
		ELSE
			EXEC sys.sp_addextendedproperty @name = N'MS_Description', @value = @description,
				@level0type = N'SCHEMA', @level0name = N'dbo', @level1type = N'TABLE', @level1name = @table_name;
	END
	ELSE IF COL_LENGTH(N'dbo.' + @table_name, @column_name) IS NOT NULL
	BEGIN
		IF EXISTS (
			SELECT 1 FROM sys.extended_properties
			WHERE class = 1 AND major_id = OBJECT_ID(N'dbo.' + @table_name)
				AND minor_id = COLUMNPROPERTY(OBJECT_ID(N'dbo.' + @table_name), @column_name, 'ColumnId')
				AND [name] = N'MS_Description')
			EXEC sys.sp_updateextendedproperty @name = N'MS_Description', @value = @description,
				@level0type = N'SCHEMA', @level0name = N'dbo', @level1type = N'TABLE', @level1name = @table_name,
				@level2type = N'COLUMN', @level2name = @column_name;
		ELSE
			EXEC sys.sp_addextendedproperty @name = N'MS_Description', @value = @description,
				@level0type = N'SCHEMA', @level0name = N'dbo', @level1type = N'TABLE', @level1name = @table_name,
				@level2type = N'COLUMN', @level2name = @column_name;
	END;
	FETCH NEXT FROM description_cursor INTO @table_name, @column_name, @description;
END;
CLOSE description_cursor;
DEALLOCATE description_cursor;
GO

SET NOEXEC OFF;
GO
