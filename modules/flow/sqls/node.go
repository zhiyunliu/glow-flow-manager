package sqls

// ListVersionNodes 版本节点：名称优先使用版本内名称，配置为空时与引擎一致回退到节点定义
const ListVersionNodes = `
select n.node_id, n.version_id, n.chain_no, n.node_def_no,
	isnull(n.node_name, '') node_name, isnull(d.[name], '') def_name, isnull(d.[type], '') [type],
	isnull(n.extparams, '') extparams, isnull(n.layout, '') layout,
	isnull(d.extparams, '') def_extparams, isnull(d.layout_definition, '') def_layout,
	isnull(d.chain_no, '') def_chain_no, isnull(d.[desc], '') def_desc
from dbo.flow_chain_node n with (nolock)
left join dbo.flow_node_definition d with (nolock) on d.node_def_no = n.node_def_no
where n.version_id = @{version_id}
order by n.node_id`

const ListVersionConnections = `
select cn.conn_id, cn.version_id, cn.chain_no, cn.from_id, cn.to_id, cn.[type],
	isnull(cn.remark, '') remark, isnull(cn.extparams, '') extparams
from dbo.flow_chain_connection cn with (nolock)
where cn.version_id = @{version_id}
order by cn.conn_id`

// ListNodeDefinitions 查询节点定义(调用方保证 node_def_no 列表非空且已校验格式)
const ListNodeDefinitions = `
select d.node_def_no, d.[name], d.[type], isnull(d.chain_no, '') chain_no
from dbo.flow_node_definition d with (updlock, rowlock)
where 1 = 1 &{in d.node_def_no}`

// ListNodeUsageOutsideVersion 节点定义在其他版本中的引用次数
const ListNodeUsageOutsideVersion = `
select n.node_def_no, count(1) as total
from dbo.flow_chain_node n
where n.version_id <> @{version_id} &{in n.node_def_no}
group by n.node_def_no`

const InsertNodeDefinition = `
insert into dbo.flow_node_definition (
	node_def_no, [name], [type], extparams, layout_definition, create_time, create_by, [operator], update_time, [desc], chain_no
) values (
	@{node_def_no}, @{name}, @{type}, nullif(@{extparams}, ''), null, getdate(), @{operator}, @{operator}, getdate(), null, @{chain_no}
)`

const UpdateNodeDefinition = `
update dbo.flow_node_definition
set [name] = @{name},
	[type] = @{type},
	[operator] = @{operator},
	update_time = getdate()
where node_def_no = @{node_def_no}`

const DeleteVersionNodes = `delete from dbo.flow_chain_node where version_id = @{version_id}`
const DeleteVersionConnections = `delete from dbo.flow_chain_connection where version_id = @{version_id}`

const InsertChainNode = `
insert into dbo.flow_chain_node (version_id, chain_no, node_def_no, extparams, layout, node_name)
values (@{version_id}, @{chain_no}, @{node_def_no}, @{extparams}, nullif(@{layout}, ''), @{node_name})`

const InsertChainConnection = `
insert into dbo.flow_chain_connection (version_id, chain_no, from_id, to_id, [type], remark, extparams)
values (@{version_id}, @{chain_no}, @{from_id}, @{to_id}, @{type}, nullif(@{remark}, ''), nullif(@{extparams}, ''))`

// DeleteOrphanNodeDefinitions 删除流程中不再被任何版本引用的节点定义
const DeleteOrphanNodeDefinitions = `
delete d
from dbo.flow_node_definition d
where d.chain_no = @{chain_no}
	and not exists (select 1 from dbo.flow_chain_node n where n.node_def_no = d.node_def_no)`
