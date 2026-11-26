package sqls

const InsertOperateLog = `
insert into dbo.flow_operate_log (
	trace_id, module, operate_type, target_id, chain_no, version_id, summary,
	before_data, after_data, result, error_msg, [operator], client_ip, create_time
) values (
	nullif(@{trace_id}, ''), @{module}, @{operate_type}, @{target_id}, nullif(@{chain_no}, ''), nullif(@{version_id}, 0), nullif(@{summary}, ''),
	nullif(@{before_data}, ''), nullif(@{after_data}, ''), @{result}, nullif(@{error_msg}, ''), @{operator}, nullif(@{client_ip}, ''), getdate()
)`

const auditWhere = `
where 1 = 1
	&{l.module}
	&{l.operate_type}
	&{l.target_id}
	&{l.chain_no}
	&{like %l.operator%}
	&{l.result}
	&{l.create_time >= start_time}
	&{l.create_time <= end_time}`

const CountOperateLog = `
select count(1) as total
from dbo.flow_operate_log l with (nolock)` + auditWhere

const PageOperateLog = `
select l.log_id, isnull(l.trace_id, '') trace_id, l.module, l.operate_type, l.target_id, isnull(l.chain_no, '') chain_no,
	isnull(l.version_id, 0) version_id, isnull(l.summary, '') summary, l.result, isnull(l.error_msg, '') error_msg,
	isnull(l.[operator], '') [operator], isnull(l.client_ip, '') client_ip, convert(varchar(19), l.create_time, 120) create_time
from dbo.flow_operate_log l with (nolock)` + auditWhere + `
order by l.log_id desc
offset @{offset} rows fetch next @{page_size} rows only`

const GetOperateLog = `
select l.log_id, isnull(l.trace_id, '') trace_id, l.module, l.operate_type, l.target_id, isnull(l.chain_no, '') chain_no,
	isnull(l.version_id, 0) version_id, isnull(l.summary, '') summary, isnull(l.before_data, '') before_data, isnull(l.after_data, '') after_data,
	l.result, isnull(l.error_msg, '') error_msg, isnull(l.[operator], '') [operator], isnull(l.client_ip, '') client_ip,
	convert(varchar(19), l.create_time, 120) create_time
from dbo.flow_operate_log l with (nolock)
where l.log_id = @{log_id}`
