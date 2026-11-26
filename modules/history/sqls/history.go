package sqls

// flow_instance_state / flow_node_state 为 glow-flow 引擎 xdb 状态存储表，完整状态保存在 state_data(JSON)

const instanceWhere = `
where 1 = 1
	&{s.chain_id}
	&{s.version}
	&{s.status}
	&{like %s.instance_id%}
	&{s.create_time >= start_time}
	&{s.create_time <= end_time}`

const instanceColumns = `
	s.instance_id, s.chain_id, s.[version], s.[status], s.state_data,
	convert(varchar(19), s.create_time, 120) create_time, convert(varchar(19), s.update_time, 120) update_time,
	isnull(c.[name], '') chain_name, isnull(v.version_no, 0) version_no`

const instanceJoins = `
left join dbo.flow_chain_definition c with (nolock) on c.chain_no = s.chain_id
left join dbo.flow_chain_version v with (nolock) on v.version_id = try_cast(s.[version] as bigint)`

const CountInstances = `
select count(1) as total
from dbo.flow_instance_state s with (nolock)` + instanceWhere

const PageInstances = `
select ` + instanceColumns + `
from dbo.flow_instance_state s with (nolock)` + instanceJoins + instanceWhere + `
order by s.create_time desc, s.instance_id
offset @{offset} rows fetch next @{page_size} rows only`

const GetInstance = `
select ` + instanceColumns + `
from dbo.flow_instance_state s with (nolock)` + instanceJoins + `
where s.instance_id = @{instance_id}`

const ListNodeStates = `
select n.instance_id, n.node_id, n.chain_id, n.[version], n.[status], n.attempts, n.state_data,
	convert(varchar(19), n.create_time, 120) create_time, convert(varchar(19), n.update_time, 120) update_time
from dbo.flow_node_state n with (nolock)
where n.instance_id = @{instance_id}
order by n.create_time, n.node_id`
