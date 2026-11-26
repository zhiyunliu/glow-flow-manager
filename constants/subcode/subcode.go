package subcode

// 稳定的机器可读错误子码，前端只依赖子码和中文消息
const (
	ParamInvalid = "param.invalid"
	SystemError  = "system.error"

	ChainNotFound     = "chain.not_found"
	ChainExists       = "chain.exists"
	ChainHasPublished = "chain.has_published"
	ChainTypeLocked   = "chain.flow_type_locked"

	VersionNotFound     = "version.not_found"
	VersionNotDraft     = "version.not_draft"
	VersionStateInvalid = "version.state_invalid"
	VersionNoPublished  = "version.no_published"

	ModelInvalid       = "model.invalid"
	ModelCompileFailed = "model.compile_failed"
	NodeIDConflict     = "model.node_id_conflict"
	NodeTypeChanged    = "model.node_type_changed"

	ImportInvalid = "import.invalid"

	NodeTypeNotFound        = "node_type.not_found"
	NodeTypeExists          = "node_type.exists"
	NodeTypeBuiltinReadonly = "node_type.builtin_readonly"
	NodeTypeInUse           = "node_type.in_use"

	RegistryUnavailable = "registry.unavailable"

	OptionSourceNotFound    = "option_source.not_found"
	OptionSourceUnavailable = "option_source.unavailable"

	InstanceNotFound = "history.instance_not_found"
	AuditNotFound    = "audit.not_found"

	DebugFailed = "debug.failed"
)
