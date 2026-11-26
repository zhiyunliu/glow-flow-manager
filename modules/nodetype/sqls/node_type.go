package sqls

const nodeTypeColumns = `
	r.node_type, r.[name], r.category, r.flow_modes, r.[source],
	isnull(r.service_name, '') service_name, isnull(r.service_path, '') service_path, isnull(r.protocol, '') protocol,
	isnull(r.relation_types, '') relation_types, isnull(r.config_schema, '') config_schema, isnull(r.default_config, '') default_config,
	isnull(r.icon, '') icon, isnull(r.color, '') color, r.sort_no, r.[status], isnull(r.[desc], '') [desc],
	convert(varchar(19), r.create_time, 120) create_time, isnull(r.create_by, '') create_by,
	isnull(r.[operator], '') [operator], isnull(convert(varchar(19), r.update_time, 120), '') update_time,
	(select count(1) from dbo.flow_node_definition d with (nolock) where d.[type] = r.node_type) usage_count`

const ListNodeTypes = `
select ` + nodeTypeColumns + `
from dbo.flow_node_registry r with (nolock)
order by r.sort_no, r.node_type`

const nodeTypeWhere = `
where (@{keyword} = '' or r.node_type like '%' + @{keyword} + '%' or r.[name] like '%' + @{keyword} + '%')
	&{r.category}
	&{r.source}
	&{like %r.flow_modes%}
	&{r.status}`

const CountNodeTypes = `
select count(1) as total
from dbo.flow_node_registry r with (nolock)` + nodeTypeWhere

const PageNodeTypes = `
select ` + nodeTypeColumns + `
from dbo.flow_node_registry r with (nolock)` + nodeTypeWhere + `
order by r.sort_no, r.node_type
offset @{offset} rows fetch next @{page_size} rows only`

const GetNodeType = `
select ` + nodeTypeColumns + `
from dbo.flow_node_registry r with (nolock)
where r.node_type = @{node_type}`

const InsertNodeType = `
insert into dbo.flow_node_registry (
	node_type, [name], category, flow_modes, [source], service_name, service_path, protocol,
	relation_types, config_schema, default_config, icon, color, sort_no, [status], [desc],
	create_time, create_by, [operator], update_time
) values (
	@{node_type}, @{name}, @{category}, @{flow_modes}, @{source}, nullif(@{service_name}, ''), nullif(@{service_path}, ''), nullif(@{protocol}, ''),
	nullif(@{relation_types}, ''), nullif(@{config_schema}, ''), nullif(@{default_config}, ''), nullif(@{icon}, ''), nullif(@{color}, ''), @{sort_no}, @{status}, nullif(@{desc}, ''),
	getdate(), @{operator}, @{operator}, getdate()
)`

const UpdateNodeType = `
update dbo.flow_node_registry
set [name] = @{name},
	category = @{category},
	flow_modes = @{flow_modes},
	service_name = nullif(@{service_name}, ''),
	service_path = nullif(@{service_path}, ''),
	protocol = nullif(@{protocol}, ''),
	relation_types = nullif(@{relation_types}, ''),
	config_schema = nullif(@{config_schema}, ''),
	default_config = nullif(@{default_config}, ''),
	icon = nullif(@{icon}, ''),
	color = nullif(@{color}, ''),
	sort_no = @{sort_no},
	[desc] = nullif(@{desc}, ''),
	[operator] = @{operator},
	update_time = getdate()
where node_type = @{node_type}`

// UpdateBuiltinNodeType 同步内置节点：只更新由引擎决定的部分，保留名称、图标、排序、状态等自定义
const UpdateBuiltinNodeType = `
update dbo.flow_node_registry
set category = @{category},
	[source] = 'builtin',
	relation_types = nullif(@{relation_types}, ''),
	config_schema = nullif(@{config_schema}, ''),
	default_config = nullif(@{default_config}, ''),
	[operator] = @{operator},
	update_time = getdate()
where node_type = @{node_type}`

const UpdateNodeTypeStatus = `
update dbo.flow_node_registry
set [status] = @{status},
	[operator] = @{operator},
	update_time = getdate()
where node_type = @{node_type}`

const DeleteNodeType = `
delete from dbo.flow_node_registry
where node_type = @{node_type}
	and [source] = 'service'`

const ListNodeTypeKeys = `
select r.node_type
from dbo.flow_node_registry r with (nolock)`
