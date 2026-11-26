package sqls

const versionColumns = `
	v.version_id, v.chain_no, v.version_no, v.[name], v.[status], v.publish_state,
	isnull(v.extparams, '') extparams, isnull(v.layout, '') layout, isnull(v.[desc], '') [desc],
	isnull(v.base_version_id, 0) base_version_id,
	isnull(convert(varchar(19), v.publish_time, 120), '') publish_time, isnull(v.publish_by, '') publish_by,
	convert(varchar(19), v.create_time, 120) create_time, isnull(v.create_by, '') create_by,
	isnull(v.[operator], '') [operator], isnull(convert(varchar(19), v.update_time, 120), '') update_time`

const versionCounts = `,
	(select count(1) from dbo.flow_chain_node n with (nolock) where n.version_id = v.version_id) node_count,
	(select count(1) from dbo.flow_chain_connection cn with (nolock) where cn.version_id = v.version_id) conn_count`

const ListVersions = `
select ` + versionColumns + versionCounts + `
from dbo.flow_chain_version v with (nolock)
where v.chain_no = @{chain_no}
order by v.version_id desc`

const GetVersion = `
select ` + versionColumns + versionCounts + `
from dbo.flow_chain_version v with (nolock)
where v.version_id = @{version_id}`

// LockVersion 在事务中锁定版本
const LockVersion = `
select ` + versionColumns + `
from dbo.flow_chain_version v with (updlock, rowlock)
where v.version_id = @{version_id}`

// GetEffectiveVersionID 引擎加载的版本：version_id 最大的 status=1 版本
const GetEffectiveVersionID = `
select isnull(max(v.version_id), 0) as total
from dbo.flow_chain_version v
where v.chain_no = @{chain_no} and v.[status] = 1`

const GetLatestVersionID = `
select isnull(max(v.version_id), 0) as total
from dbo.flow_chain_version v
where v.chain_no = @{chain_no}`

const GetDraftVersionID = `
select isnull(max(v.version_id), 0) as total
from dbo.flow_chain_version v
where v.chain_no = @{chain_no} and v.publish_state = 0`

const NextVersionNo = `
select isnull(max(v.version_no), 0) + 1 as total
from dbo.flow_chain_version v with (updlock, holdlock)
where v.chain_no = @{chain_no}`

const InsertVersion = `
insert into dbo.flow_chain_version (
	chain_no, [name], [status], extparams, layout, create_time, create_by, [operator], update_time, [desc],
	version_no, publish_state, base_version_id
)
output inserted.version_id as total
values (
	@{chain_no}, @{name}, 0, nullif(@{extparams}, ''), nullif(@{layout}, ''), getdate(), @{operator}, @{operator}, getdate(), nullif(@{desc}, ''),
	@{version_no}, 0, nullif(@{base_version_id}, 0)
)`

const UpdateDraftVersion = `
update dbo.flow_chain_version
set [name] = @{name},
	extparams = nullif(@{extparams}, ''),
	layout = nullif(@{layout}, ''),
	[desc] = nullif(@{desc}, ''),
	[operator] = @{operator},
	update_time = getdate()
where version_id = @{version_id} and publish_state = 0`

const PublishVersion = `
update dbo.flow_chain_version
set [status] = 1,
	publish_state = 1,
	publish_time = getdate(),
	publish_by = @{operator},
	[desc] = isnull(nullif(@{desc}, ''), [desc]),
	[operator] = @{operator},
	update_time = getdate()
where version_id = @{version_id} and publish_state = 0`

const OfflineVersion = `
update dbo.flow_chain_version
set [status] = 0,
	publish_state = 2,
	[operator] = @{operator},
	update_time = getdate()
where version_id = @{version_id} and publish_state = 1`

const OnlineVersion = `
update dbo.flow_chain_version
set [status] = 1,
	publish_state = 1,
	[operator] = @{operator},
	update_time = getdate()
where version_id = @{version_id} and publish_state = 2`

const DeleteVersion = `
delete from dbo.flow_chain_version
where version_id = @{version_id} and publish_state = 0`

// UpdateDraftVersionName 流程改名时同步草稿版本名称(已发布版本保持不变)
const UpdateDraftVersionName = `
update dbo.flow_chain_version
set [name] = @{name},
	update_time = getdate()
where chain_no = @{chain_no} and publish_state = 0`
