package errors

// FieldError is one machine-readable validation failure, addressed to a field
// of the request body (JSON path, e.g. "ports[0].protocol").
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ValidationErrorDetails is the structured payload carried in Error.details
// for aggregated validation failures.
type ValidationErrorDetails struct {
	Errors []FieldError `json:"errors"`
}

// ValidationFailed returns a 400 carrying every collected field error.
func ValidationFailed(fieldErrors []FieldError) *ServiceError {
	return Validation("validation failed: %d error(s)", len(fieldErrors)).
		WithDetails(ValidationErrorDetails{Errors: fieldErrors})
}

// Validation error codes surfaced in FieldError.Code. Shared by create-time
// validation and release-time checks; the frontend switches on these.
const (
	// resource shape
	VErrResourceNameRequired  = "resource_name_required"
	VErrResourceNameInvalid   = "resource_name_invalid"
	VErrResourceNameDuplicate = "resource_name_duplicate"
	VErrSourceRequired        = "source_required"
	VErrSourceConflict        = "source_conflict"
	VErrWorkloadTypeInvalid   = "workload_type_invalid"
	VErrScheduleRequired      = "schedule_required"
	VErrScheduleNotAllowed    = "schedule_not_allowed"
	VErrScheduleInvalid       = "schedule_invalid"
	VErrReplicasInvalid       = "replicas_invalid"

	// ports
	VErrPortsNotAllowed     = "ports_not_allowed"
	VErrPublicPortNotHTTP   = "public_port_not_http"
	VErrPortProtocolInvalid = "port_protocol_invalid"
	VErrPortNameInvalid     = "port_name_invalid"
	VErrPortNumberInvalid   = "port_number_invalid"
	VErrPortNameDuplicate   = "port_name_duplicate"
	VErrPortNumberDuplicate = "port_number_duplicate"
	VErrSubdomainDuplicate  = "subdomain_duplicate"
	VErrDomainNotConfigured = "domain_not_configured"

	// env
	VErrEnvNameRequired  = "env_name_required"
	VErrEnvNameDuplicate = "env_name_duplicate"
	VErrEnvValueMissing  = "env_value_missing"
	VErrEnvValueConflict = "env_value_conflict"

	// volume mounts
	VErrVolumeMountInvalid = "volume_mount_invalid"
	VErrVolumeNotFound     = "volume_not_found"
	VErrVolumeHashMissing  = "volume_hash_missing"

	// referenced entities
	VErrSecretNotFound             = "secret_not_found"
	VErrGitIntegrationNotFound     = "git_integration_not_found"
	VErrRegistryCredentialNotFound = "registry_credential_not_found"

	// depends_on
	VErrDependencySelf      = "self_dependency"
	VErrDependencyDuplicate = "duplicate_dependency"
	VErrDependencyUnknown   = "unknown_dependency"
	VErrDependencyCycle     = "dependency_cycle"

	// build / git / image source shape
	VErrGitRepoURLRequired   = "git_repo_url_required"
	VErrGitBranchTagConflict = "git_branch_tag_conflict"
	VErrGitCommitInvalid     = "git_commit_invalid"
	VErrGitCommitRequiresRef = "git_commit_requires_ref"
	VErrImageRefRequired     = "image_ref_required"
	VErrImageRefInvalid      = "image_ref_invalid"
	VErrPushTargetRequired   = "push_target_required"
	VErrPushTargetConflict   = "push_target_conflict"
	VErrPushRefInvalid       = "push_ref_invalid"

	// release-time (sync resolvePins + async worker)
	VErrGitRepoUnreachable = "git_repo_unreachable"
	VErrGitAuthFailed      = "git_auth_failed"
	VErrGitBranchNotFound  = "git_branch_not_found"
	VErrGitTagNotFound     = "git_tag_not_found"
	VErrGitRateLimited     = "git_rate_limited"
	VErrImageNotFound      = "image_not_found"
	VErrRegistryAuthFailed = "registry_auth_failed"
	VErrPushAccessDenied   = "push_access_denied"

	// stack-level (settings, connections, interpolations)
	VErrStackSettingsInvalid = "stack_settings_invalid"
	VErrConnectionInvalid    = "connection_invalid"
	VErrInterpolationInvalid = "interpolation_invalid"
)
