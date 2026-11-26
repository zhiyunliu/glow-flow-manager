package sqls

const chainColumns = `
	c.chain_no, c.[name], c.flow_type, c.[status],
	isnull(c.extparams, '') extparams, isnull(c.layout, '') layout, isnull(c.[desc], '') [desc],
	convert(varchar(19), c.create_time, 120) create_time, isnull(c.create_by, '') create_by,
	isnull(c.[operator], '') [operator], isnull(convert(varchar(19), c.update_time, 120), '') update_time,
	isnull(eff.version_id, 0) effective_version_id, isnull(eff.version_no, 0) effective_version_no,
	isnull(dr.version_id, 0) draft_version_id, isnull(dr.version_no, 0) draft_version_no,
	isnull(st.version_count, 0) version_count, isnull(st.latest_version_no, 0) latest_version_no,
	isnull(st.published_count, 0) published_count`

// chainApply 生效版本与引擎一致：version_id 最大的 status=1 版本
const chainApply = `
outer apply (
	select top 1 v.version_id, v.version_no
	from dbo.flow_chain_version v with (nolock)
	where v.chain_no = c.chain_no and v.[status] = 1
	order by v.version_id desc
) eff
outer apply (
	select top 1 v.version_id, v.version_no
	from dbo.flow_chain_version v with (nolock)
	where v.chain_no = c.chain_no and v.publish_state = 0
	order by v.version_id desc
) dr
outer apply (
	select count(1) version_count, max(v.version_no) latest_version_no,
		sum(case when v.publish_state in (1, 2) then 1 else 0 end) published_count
	from dbo.flow_chain_version v with (nolock)
	where v.chain_no = c.chain_no
) st`

const chainWhere = `
where 1 = 1
	&{like %c.chain_no%}
	&{like %c.name%}
	&{c.flow_type}
	&{c.status}`

const CountChains = `
select count(1) as total
from dbo.flow_chain_definition c with (nolock)` + chainWhere

const PageChains = `
select ` + chainColumns + `
from dbo.flow_chain_definition c with (nolock)` + chainApply + chainWhere + `
order by c.create_time desc, c.chain_no
offset @{offset} rows fetch next @{page_size} rows only`

const GetChain = `
select ` + chainColumns + `
from dbo.flow_chain_definition c with (nolock)` + chainApply + `
where c.chain_no = @{chain_no}`

// LockChain 在事务中锁定流程，串行化同一流程的版本操作
const LockChain = `
select c.chain_no, c.[name], c.flow_type, c.[status]
from dbo.flow_chain_definition c with (updlock, rowlock)
where c.chain_no = @{chain_no}`

const InsertChain = `
insert into dbo.flow_chain_definition (
	chain_no, [name], [status], extparams, layout, create_time, create_by, [operator], update_time, [desc], flow_type
) values (
	@{chain_no}, @{name}, 1, nullif(@{extparams}, ''), nullif(@{layout}, ''), getdate(), @{operator}, @{operator}, getdate(), nullif(@{desc}, ''), @{flow_type}
)`

const UpdateChain = `
update dbo.flow_chain_definition
set [name] = @{name},
	flow_type = @{flow_type},
	[desc] = nullif(@{desc}, ''),
	[operator] = @{operator},
	update_time = getdate()
where chain_no = @{chain_no}`

const UpdateChainFlowType = `
update dbo.flow_chain_definition
set flow_type = @{flow_type},
	[operator] = @{operator},
	update_time = getdate()
where chain_no = @{chain_no}`

const UpdateChainStatus = `
update dbo.flow_chain_definition
set [status] = @{status},
	[operator] = @{operator},
	update_time = getdate()
where chain_no = @{chain_no}`

const DeleteChainConnections = `delete from dbo.flow_chain_connection where chain_no = @{chain_no}`
const DeleteChainNodes = `delete from dbo.flow_chain_node where chain_no = @{chain_no}`
const DeleteChainVersions = `delete from dbo.flow_chain_version where chain_no = @{chain_no}`
const DeleteChainNodeDefinitions = `delete from dbo.flow_node_definition where chain_no = @{chain_no}`
const DeleteChain = `delete from dbo.flow_chain_definition where chain_no = @{chain_no}`
