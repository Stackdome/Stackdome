package stackresource

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/robfig/cron/v3"
)

var (
	portNamePattern     = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	resourceNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
	gitCommitSHAPattern = regexp.MustCompile(`^[0-9a-f]{7,40}$`)
	cronParser          = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
)

const (
	maxPortNameLength     = 15
	maxResourceNameLength = 63 // Kubernetes DNS label limit
)

func fieldErr(field, code, message string, args ...any) errors.FieldError {
	return errors.FieldError{Field: field, Code: code, Message: fmt.Sprintf(message, args...)}
}

// validateInputRules checks everything derivable from the payload alone.
// Rules mirror the cluster-agent CRD admission constraints so bad specs fail
// here instead of at CR apply or as a silent reconcile stall.
func validateInputRules(resource *models.StackResource) []errors.FieldError {
	var errs []errors.FieldError
	errs = append(errs, validateName(resource)...)
	errs = append(errs, validateSourceShape(resource)...)
	errs = append(errs, validateWorkloadType(resource)...)
	errs = append(errs, validatePorts(resource)...)
	errs = append(errs, validateEnvVars(resource)...)
	errs = append(errs, validateVolumeMountShape(resource)...)
	errs = append(errs, validateDependsOnShape(resource)...)
	return errs
}

func validateName(resource *models.StackResource) []errors.FieldError {
	if resource.Name == "" {
		return []errors.FieldError{fieldErr("name", errors.VErrResourceNameRequired, "resource name is required")}
	}
	if !resourceNamePattern.MatchString(resource.Name) || len(resource.Name) > maxResourceNameLength {
		return []errors.FieldError{fieldErr("name", errors.VErrResourceNameInvalid,
			"resource name '%s' must match %s and be at most %d characters",
			resource.Name, resourceNamePattern.String(), maxResourceNameLength)}
	}
	return nil
}

func validateSourceShape(resource *models.StackResource) []errors.FieldError {
	var errs []errors.FieldError
	hasBuild := resource.BuildConfig != nil
	hasImage := resource.ImageConfig != nil
	switch {
	case !hasBuild && !hasImage:
		return []errors.FieldError{fieldErr("source", errors.VErrSourceRequired,
			"exactly one of a git/volume build source or an image source is required")}
	case hasBuild && hasImage:
		return []errors.FieldError{fieldErr("source", errors.VErrSourceConflict,
			"a resource cannot have both a build source and an image source")}
	}

	if hasImage && resource.ImageConfig.Image == "" {
		errs = append(errs, fieldErr("source.image.ref", errors.VErrImageRefRequired, "image ref is required"))
	}
	if hasImage && resource.ImageConfig.Image != "" {
		if _, err := name.ParseReference(resource.ImageConfig.Image); err != nil {
			errs = append(errs, fieldErr("source.image.ref", errors.VErrImageRefInvalid,
				"image ref '%s' is not a valid reference: %v", resource.ImageConfig.Image, err))
		}
	}
	if hasBuild {
		errs = append(errs, validateBuildShape(resource.BuildConfig)...)
	}
	return errs
}

func validateBuildShape(build *models.BuildConfigSpec) []errors.FieldError {
	var errs []errors.FieldError

	srcCtx := build.SourceContext
	if (srcCtx.Git == nil && srcCtx.Volume == nil) || (srcCtx.Git != nil && srcCtx.Volume != nil) {
		errs = append(errs, fieldErr("source", errors.VErrSourceConflict,
			"exactly one of source.git or source.volume must be set"))
		return errs
	}

	if srcCtx.Git != nil {
		if srcCtx.Git.RepoURL == "" {
			errs = append(errs, fieldErr("source.git.repo_url", errors.VErrGitRepoURLRequired, "repo_url is required"))
		}
		if rev := build.SourceRevision.Git; rev != nil {
			if rev.Branch != "" && rev.Tag != "" {
				errs = append(errs, fieldErr("source.git", errors.VErrGitBranchTagConflict,
					"branch and tag cannot both be set"))
			}
			if rev.Commit != "" && !gitCommitSHAPattern.MatchString(rev.Commit) {
				errs = append(errs, fieldErr("source.git.commit", errors.VErrGitCommitInvalid,
					"commit '%s' is not a valid git SHA", rev.Commit))
			}
			if rev.Commit != "" && rev.Branch == "" && rev.Tag == "" {
				errs = append(errs, fieldErr("source.git", errors.VErrGitCommitRequiresRef,
					"a commit pin requires a branch or tag (the cluster needs a fetchable ref)"))
			}
		}
	}
	if srcCtx.Volume != nil && srcCtx.Volume.SourceVolumeName == "" && srcCtx.Volume.SourceVolumeID == "" {
		errs = append(errs, fieldErr("source.volume", errors.VErrVolumeMountInvalid,
			"a volume build source needs volume_name or volume_id"))
	}

	// Field paths below assume a git source; volume-source pushes share the shape.
	pushField := "source.git.push"
	if srcCtx.Volume != nil {
		pushField = "source.volume.push"
	}
	repo := build.BuildImageRepository
	switch {
	case repo.UseInClusterRegistry && repo.ExternalImageRef != "":
		errs = append(errs, fieldErr(pushField, errors.VErrPushTargetConflict,
			"external repository cannot be set when using the in-cluster registry"))
	case !repo.UseInClusterRegistry && repo.ExternalImageRef == "":
		errs = append(errs, fieldErr(pushField, errors.VErrPushTargetRequired,
			"a push repository is required when not using the in-cluster registry"))
	case repo.ExternalImageRef != "":
		if _, err := name.NewRepository(repo.ExternalImageRef); err != nil {
			errs = append(errs, fieldErr(pushField+".repository", errors.VErrPushRefInvalid,
				"push repository '%s' is not a valid registry reference: %v", repo.ExternalImageRef, err))
		}
	}
	return errs
}

func validateWorkloadType(resource *models.StackResource) []errors.FieldError {
	var errs []errors.FieldError
	wt := resource.WorkloadType
	switch wt {
	case "", models.WorkloadTypeService, models.WorkloadTypeStatefulService,
		models.WorkloadTypeWorker, models.WorkloadTypeJob, models.WorkloadTypeCronJob:
	default:
		return []errors.FieldError{fieldErr("workload_type", errors.VErrWorkloadTypeInvalid,
			"unknown workload type '%s'", wt)}
	}

	if wt == models.WorkloadTypeCronJob && resource.Schedule == "" {
		errs = append(errs, fieldErr("schedule", errors.VErrScheduleRequired, "CronJob requires a schedule"))
	}
	if wt != models.WorkloadTypeCronJob && resource.Schedule != "" {
		errs = append(errs, fieldErr("schedule", errors.VErrScheduleNotAllowed,
			"schedule is only allowed for CronJob workloads"))
	}
	if resource.Schedule != "" {
		if _, err := cronParser.Parse(resource.Schedule); err != nil {
			errs = append(errs, fieldErr("schedule", errors.VErrScheduleInvalid,
				"schedule '%s' is not a valid cron expression: %v", resource.Schedule, err))
		}
	}

	switch wt {
	case models.WorkloadTypeWorker, models.WorkloadTypeJob, models.WorkloadTypeCronJob:
		if len(resource.Ports) > 0 {
			errs = append(errs, fieldErr("ports", errors.VErrPortsNotAllowed,
				"%s workloads cannot declare ports", wt))
		}
	}

	if resource.Replicas != nil {
		if *resource.Replicas < 0 {
			errs = append(errs, fieldErr("replicas", errors.VErrReplicasInvalid, "replicas cannot be negative"))
		}
	}
	return errs
}

func validatePorts(resource *models.StackResource) []errors.FieldError {
	var errs []errors.FieldError
	seenNames := map[string]bool{}
	seenNumbers := map[int]bool{}
	for i, p := range resource.Ports {
		f := func(sub string) string { return fmt.Sprintf("ports[%d].%s", i, sub) }

		if p.Name == "" || !portNamePattern.MatchString(p.Name) || len(p.Name) > maxPortNameLength {
			errs = append(errs, fieldErr(f("name"), errors.VErrPortNameInvalid,
				"port name '%s' must match %s and be at most %d characters", p.Name, portNamePattern.String(), maxPortNameLength))
		}
		if p.Number < 1 || p.Number > 65535 {
			errs = append(errs, fieldErr(f("number"), errors.VErrPortNumberInvalid,
				"port number %d must be between 1 and 65535", p.Number))
		}
		if seenNames[p.Name] {
			errs = append(errs, fieldErr(f("name"), errors.VErrPortNameDuplicate,
				"duplicate port name '%s'", p.Name))
		}
		seenNames[p.Name] = true
		if seenNumbers[p.Number] {
			errs = append(errs, fieldErr(f("number"), errors.VErrPortNumberDuplicate,
				"duplicate port number %d", p.Number))
		}
		seenNumbers[p.Number] = true

		proto := strings.ToLower(p.Protocol)
		switch proto {
		case "", models.PortProtocolHTTP, models.PortProtocolGRPC, models.PortProtocolTCP:
		default:
			errs = append(errs, fieldErr(f("protocol"), errors.VErrPortProtocolInvalid,
				"protocol '%s' must be one of http, grpc, tcp", p.Protocol))
			continue
		}
		if p.ExposedToPublic && proto != models.PortProtocolHTTP && proto != "" {
			errs = append(errs, fieldErr(f("protocol"), errors.VErrPublicPortNotHTTP,
				"port '%s' (%s) cannot be exposed to public; only http ports can", p.Name, proto))
		}
	}
	return errs
}

func validateEnvVars(resource *models.StackResource) []errors.FieldError {
	if resource.ExecutionConfig == nil {
		return nil
	}
	var errs []errors.FieldError
	// Built lazily on first self_output use: env.SelfOutput must reference
	// one of the resource's own declared outputs, otherwise the env renders
	// broken/empty at deploy time with no signal to the user.
	var allowedSelfOutputs map[string]bool
	seen := map[string]bool{}
	for i, env := range resource.ExecutionConfig.Env {
		f := fmt.Sprintf("execution_config.env[%d]", i)
		if env.Name == "" {
			errs = append(errs, fieldErr(f+".name", errors.VErrEnvNameRequired, "env var name is required"))
			continue
		}
		if seen[env.Name] {
			errs = append(errs, fieldErr(f+".name", errors.VErrEnvNameDuplicate,
				"duplicate env var '%s'", env.Name))
		}
		seen[env.Name] = true

		set := 0
		if env.Value != "" {
			set++
		}
		if env.SelfOutput != "" {
			set++
		}
		if env.SecretKeyRef != nil {
			set++
		}
		switch {
		case set == 0:
			errs = append(errs, fieldErr(f, errors.VErrEnvValueMissing,
				"env '%s' needs exactly one of value, self_output, or secret_key_ref", env.Name))
		case set > 1:
			errs = append(errs, fieldErr(f, errors.VErrEnvValueConflict,
				"env '%s' has more than one of value, self_output, secret_key_ref", env.Name))
		}
		if env.SecretKeyRef != nil && (env.SecretKeyRef.SecretName == "" || env.SecretKeyRef.Key == "") {
			errs = append(errs, fieldErr(f+".secret_key_ref", errors.VErrEnvValueMissing,
				"env '%s' secret_key_ref needs secret_name and key", env.Name))
		}
		if env.SelfOutput != "" {
			if allowedSelfOutputs == nil {
				allowedSelfOutputs = map[string]bool{}
				for _, output := range resource.EnsureDeclaredOutputs() {
					allowedSelfOutputs[output.Name] = true
				}
			}
			if !allowedSelfOutputs[env.SelfOutput] {
				errs = append(errs, fieldErr(f+".self_output", errors.VErrEnvSelfOutputUnknown,
					"env '%s' references unsupported self_output '%s'; it must be one of the resource's declared outputs", env.Name, env.SelfOutput))
			}
		}
	}
	return errs
}

func validateVolumeMountShape(resource *models.StackResource) []errors.FieldError {
	var errs []errors.FieldError
	for i, m := range resource.VolumeMounts {
		f := fmt.Sprintf("volume_mounts[%d]", i)
		if m.TargetPath == "" || (m.SourceVolumeName == "" && m.SourceVolumeID == "") {
			errs = append(errs, fieldErr(f, errors.VErrVolumeMountInvalid,
				"volume mount needs a target path and a source volume"))
		}
	}
	return errs
}

func validateDependsOnShape(resource *models.StackResource) []errors.FieldError {
	var errs []errors.FieldError
	seen := map[string]bool{}
	for i, dep := range resource.DependsOn {
		f := fmt.Sprintf("depends_on[%d]", i)
		if dep == resource.Name {
			errs = append(errs, fieldErr(f, errors.VErrDependencySelf,
				"resource cannot depend on itself"))
		}
		if seen[dep] {
			errs = append(errs, fieldErr(f, errors.VErrDependencyDuplicate,
				"duplicate dependency '%s'", dep))
		}
		seen[dep] = true
	}
	return errs
}
