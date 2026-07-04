package stack

import (
	"context"
	stderrors "errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Stackdome/stackdome/pkg/clients"
	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stackdeploy"
	"github.com/Stackdome/stackdome/pkg/validator"
	"github.com/robfig/cron/v3"
)

//go:generate mockgen -source=stack_validator.go -destination=../../mocks/mock_stack_validator_dependencies.go -package=mocks
type organisationDomainService interface {
	ListByOrganisationID(ctx context.Context, orgID string) ([]*models.OrganisationDomain, *errors.ServiceError)
}

type secretService interface {
	ValidateImageRegistrySecretForStackResource(ctx context.Context, secretID string) *errors.ServiceError
	ValidateGitSecretForStackResource(ctx context.Context, secretID string) *errors.ServiceError
	ValidateSecretHasKeys(ctx context.Context, secretID string, requiredKeys []string) (bool, []string, *errors.ServiceError)
	ValidateSecretExists(ctx context.Context, secretID string) (bool, *errors.ServiceError)
	InternalGetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError)
}

type postgresAddonService interface {
	GetPostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError)
}

// registryClientProvider builds registry clients from resolved credentials so
// the image/push probes can be faked in tests.
type registryClientProvider interface {
	ClientFor(resolved *credentials.ResolvedRegistryCredential) (clients.RegistryClient, error)
}

type defaultRegistryClientProvider struct{}

func (defaultRegistryClientProvider) ClientFor(resolved *credentials.ResolvedRegistryCredential) (clients.RegistryClient, error) {
	if resolved != nil && resolved.Username != "" && resolved.Password != "" {
		return clients.NewRegistryClientWithAuth(resolved.Username, resolved.Password)
	}
	return clients.NewRegistryClientAnonymous()
}

// gitClientProvider builds git clients from resolved credentials so the clone
// probe can be faked in tests.
type gitClientProvider interface {
	ClientFor(repoURL string, creds gitclient.GitCredentials) (gitclient.GitClient, error)
}

type defaultGitClientProvider struct{}

func (defaultGitClientProvider) ClientFor(repoURL string, creds gitclient.GitCredentials) (gitclient.GitClient, error) {
	return gitclient.NewGitClientForRepo(repoURL, creds)
}

// Add only validations that take reasonable time to complete.
// Avoid validations that require network calls or long-running operations.
// Long running validations should be handled in the stack worker.
type stackValidator struct {
	interpolationValidator validator.InterpolationValidation
	domainService          organisationDomainService
	secretService          secretService
	postgresAddonService   postgresAddonService
	credentialResolver     credentials.Resolver
	registryClients        registryClientProvider
	gitClients             gitClientProvider
	logger                 logger.Logger
}

type StackValidatorSpec struct {
	DomainService        organisationDomainService
	SecretService        secretService
	PostgresAddonService postgresAddonService
	CredentialResolver   credentials.Resolver
	// RegistryClients is optional; it defaults to real registry clients.
	RegistryClients registryClientProvider
	// GitClients is optional; it defaults to real git clients.
	GitClients gitClientProvider
}

func NewStackValidator(
	spec StackValidatorSpec,
) validator.StackValidator {
	registryClients := spec.RegistryClients
	if registryClients == nil {
		registryClients = defaultRegistryClientProvider{}
	}
	gitClients := spec.GitClients
	if gitClients == nil {
		gitClients = defaultGitClientProvider{}
	}
	return &stackValidator{
		interpolationValidator: NewInterpolationValidation(),
		domainService:          spec.DomainService,
		secretService:          spec.SecretService,
		postgresAddonService:   spec.PostgresAddonService,
		credentialResolver:     spec.CredentialResolver,
		registryClients:        registryClients,
		gitClients:             gitClients,
		logger:                 logger.NewLoggerWithPrefix(context.Background(), "stack-validator"),
	}
}

func (v *stackValidator) ValidateForCreate(ctx context.Context, spec *models.Stack) *errors.ServiceError {
	if err := v.validateUniqueResourceNames(spec); err != nil {
		return err
	}
	// On create every resource is new, so probe all sources.
	if err := v.validateImageSource(ctx, nil, spec); err != nil {
		return err
	}
	if err := v.validateStackEnvVars(spec); err != nil {
		return err
	}
	if err := v.validateStackPorts(spec); err != nil {
		return err
	}
	if err := validateWorkloadTypes(spec); err != nil {
		return err
	}
	if err := v.validateVolumeMounts(spec); err != nil {
		return err
	}
	if err := v.validateDomainExistence(ctx, spec); err != nil {
		return err
	}
	if err := v.validateBuildSourceVolumes(spec); err != nil {
		return err
	}
	if err := v.validateConnections(ctx, nil, spec); err != nil {
		return err
	}
	if err := validateStackSettings(spec); err != nil {
		return err
	}

	return nil
}

func (v *stackValidator) ValidateForUpdate(ctx context.Context, existing *models.Stack, spec *models.Stack) *errors.ServiceError {
	// Validate immutable fields
	if spec.Name != existing.Name {
		return errors.BadRequest("stack name cannot be updated")
	}
	if spec.UserID != existing.UserID {
		return errors.BadRequest("stack user cannot be updated")
	}
	if spec.OrganisationID != existing.OrganisationID {
		return errors.BadRequest("stack organisation cannot be updated")
	}

	if err := v.validateUniqueResourceNames(spec); err != nil {
		return err
	}
	// On update only re-probe resources whose source-relevant config changed;
	// shape/XOR/static validation still runs for every resource.
	if err := v.validateImageSource(ctx, existing, spec); err != nil {
		return err
	}
	if err := v.validateStackEnvVars(spec); err != nil {
		return err
	}
	if err := v.validateStackPorts(spec); err != nil {
		return err
	}
	if err := validateWorkloadTypes(spec); err != nil {
		return err
	}
	if err := v.validateVolumeMounts(spec); err != nil {
		return err
	}
	if err := v.validateDomainExistence(ctx, spec); err != nil {
		return err
	}
	if err := v.validateBuildSourceVolumes(spec); err != nil {
		return err
	}
	if err := v.validateConnections(ctx, existing, spec); err != nil {
		return err
	}
	if err := validateStackSettings(spec); err != nil {
		return err
	}
	return nil
}

func validateStackSettings(spec *models.Stack) *errors.ServiceError {
	if spec.Settings == nil {
		return nil
	}
	s := spec.Settings
	if s.ReleaseRetentionLimit > models.MaxReleaseRetentionLimit {
		return errors.BadRequest("release_retention_limit must be at most %d", models.MaxReleaseRetentionLimit)
	}
	if s.MinSuccessfulReleases > models.MaxMinSuccessfulReleases {
		return errors.BadRequest("min_successful_releases must be at most %d", models.MaxMinSuccessfulReleases)
	}
	if s.DeployTimeoutMinutes > models.MaxDeployTimeoutMinutes {
		return errors.BadRequest("deploy_timeout_minutes must be at most %d", models.MaxDeployTimeoutMinutes)
	}
	if s.MinSuccessfulReleases > 0 && s.ReleaseRetentionLimit > 0 && s.MinSuccessfulReleases > s.ReleaseRetentionLimit {
		return errors.BadRequest("min_successful_releases (%d) must not exceed release_retention_limit (%d)", s.MinSuccessfulReleases, s.ReleaseRetentionLimit)
	}
	return nil
}

var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)

func validateWorkloadTypes(spec *models.Stack) *errors.ServiceError {
	for _, r := range spec.StackResources {
		wt := r.WorkloadType
		if wt == "" {
			wt = models.WorkloadTypeService
		}

		switch wt {
		case models.WorkloadTypeService, models.WorkloadTypeStatefulService, models.WorkloadTypeWorker, models.WorkloadTypeJob, models.WorkloadTypeCronJob:
		default:
			return errors.BadRequest("stack resource '%s' has unsupported workload_type '%s'", r.Name, wt)
		}

		if wt == models.WorkloadTypeCronJob {
			if r.Schedule == "" {
				return errors.BadRequest("stack resource '%s' requires schedule for workload_type '%s'", r.Name, wt)
			}
			if _, err := cronParser.Parse(r.Schedule); err != nil {
				return errors.BadRequest("stack resource '%s' has invalid cron schedule '%s': %s", r.Name, r.Schedule, err.Error())
			}
		} else if r.Schedule != "" {
			return errors.BadRequest("stack resource '%s' cannot set schedule for workload_type '%s'", r.Name, wt)
		}

		switch wt {
		case models.WorkloadTypeWorker, models.WorkloadTypeJob, models.WorkloadTypeCronJob:
			if len(r.Ports) > 0 {
				return errors.BadRequest("stack resource '%s' with workload_type '%s' cannot declare ports", r.Name, wt)
			}
		}

		if r.Replicas != nil && *r.Replicas < 0 {
			return errors.BadRequest("stack resource '%s' replicas cannot be negative", r.Name)
		}

		switch wt {
		case models.WorkloadTypeJob, models.WorkloadTypeCronJob:
			if r.Replicas != nil {
				return errors.BadRequest("stack resource '%s' with workload_type '%s' cannot set replicas", r.Name, wt)
			}
		}
	}
	return nil
}

func (v *stackValidator) validateUniqueResourceNames(spec *models.Stack) *errors.ServiceError {
	seen := make(map[string]struct{}, len(spec.StackResources))
	for _, r := range spec.StackResources {
		if _, exists := seen[r.Name]; exists {
			return errors.BadRequest("duplicate stack resource name '%s'", r.Name)
		}
		seen[r.Name] = struct{}{}
	}
	return nil
}

// validateImageSource validates every resource's source config. Shape/XOR and
// static credential checks run for all resources; the network probes
// (image-pullable, push-access, git-clone) only run for resources whose
// source-relevant config changed. On create, existing is nil and everything is
// probed.
func (v *stackValidator) validateImageSource(ctx context.Context, existing *models.Stack, spec *models.Stack) *errors.ServiceError {
	existingResources := resourcesByName(existing)

	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]

		// Validate that resource has exactly one config type
		if err := v.validateResourceConfigType(currentResource); err != nil {
			return err
		}

		probe := resourceNeedsSourceProbe(existingResources[currentResource.Name], currentResource)

		// Validate image config if present
		if currentResource.ImageConfig != nil {
			if err := v.validateImageConfig(ctx, spec.OrganisationID, currentResource, probe); err != nil {
				return err
			}
		}

		// Validate build config if present
		if currentResource.BuildConfig != nil {
			if err := v.validateBuildConfig(ctx, spec.OrganisationID, currentResource, probe); err != nil {
				return err
			}
		}
	}
	return nil
}

// resourcesByName indexes a stack's resources by name; returns nil for a nil
// stack (the create path).
func resourcesByName(stack *models.Stack) map[string]*models.StackResource {
	if stack == nil {
		return nil
	}
	byName := make(map[string]*models.StackResource, len(stack.StackResources))
	for i := range stack.StackResources {
		byName[stack.StackResources[i].Name] = stack.StackResources[i]
	}
	return byName
}

// resourceNeedsSourceProbe reports whether the desired resource's registry/git
// sources must be re-probed. A resource with no stored counterpart (new, or the
// create path) is always probed. Re-submitted inline credentials force a probe
// because the backing managed secret's data may have changed even when the ref
// string did not. Otherwise the probe runs only when a source-relevant field
// (image/repo ref, revision, push target, or credential selector) changed.
func resourceNeedsSourceProbe(existing, desired *models.StackResource) bool {
	if existing == nil {
		return true
	}
	if desiredHasInlineCredentials(desired) {
		return true
	}
	return imageSourceOf(existing) != imageSourceOf(desired) ||
		buildSourceOf(existing) != buildSourceOf(desired)
}

// desiredHasInlineCredentials reports whether the request carried inline
// credentials on any source slot. These are transient (never persisted), so an
// existing stored resource never has them; their presence on the desired spec
// means credentials were (re-)submitted and the source must be re-probed.
func desiredHasInlineCredentials(r *models.StackResource) bool {
	if r.ImageConfig != nil && r.ImageConfig.InlineCredentials != nil {
		return true
	}
	if r.BuildConfig != nil {
		if r.BuildConfig.PushInlineCredentials != nil {
			return true
		}
		if git := r.BuildConfig.SourceContext.Git; git != nil && git.InlineCredentials != nil {
			return true
		}
	}
	return false
}

// imageSourceView is a flat, comparable projection of the pull-probe-relevant
// fields of a resource's image config.
type imageSourceView struct {
	present              bool
	image                string
	pullSecretID         string
	registryCredentialID string
}

func imageSourceOf(r *models.StackResource) imageSourceView {
	if r == nil || r.ImageConfig == nil {
		return imageSourceView{}
	}
	view := imageSourceView{
		present:              true,
		image:                r.ImageConfig.Image,
		registryCredentialID: r.ImageConfig.RegistryCredentialID,
	}
	if r.ImageConfig.PullSecretRef != nil {
		view.pullSecretID = r.ImageConfig.PullSecretRef.SecretID
	}
	return view
}

// buildSourceView is a flat, comparable projection of the push/git-probe-relevant
// fields of a resource's build config.
type buildSourceView struct {
	present            bool
	gitRepoURL         string
	gitSecretID        string
	gitIntegrationID   string
	volumeSourceName   string
	revBranch          string
	revTag             string
	revCommit          string
	revVolumeHash      string
	dockerfilePath     string
	contextPath        string
	insecureRegistry   bool
	useInClusterReg    bool
	clusterRegName     string
	externalImageRef   string
	pushSecretID       string
	pushRegistryCredID string
}

func buildSourceOf(r *models.StackResource) buildSourceView {
	if r == nil || r.BuildConfig == nil {
		return buildSourceView{}
	}
	b := r.BuildConfig
	view := buildSourceView{
		present:            true,
		dockerfilePath:     b.DockerfilePath,
		contextPath:        b.ContextPathWithinSource,
		insecureRegistry:   b.BuildImageRepository.InsecureRegistry,
		useInClusterReg:    b.BuildImageRepository.UseInClusterRegistry,
		clusterRegName:     b.BuildImageRepository.ClusterRegistryName,
		externalImageRef:   b.BuildImageRepository.ExternalImageRef,
		pushRegistryCredID: b.PushRegistryCredentialID,
	}
	if git := b.SourceContext.Git; git != nil {
		view.gitRepoURL = git.RepoURL
		view.gitIntegrationID = git.IntegrationID
		if git.GitSecretRef != nil {
			view.gitSecretID = git.GitSecretRef.SecretID
		}
	}
	if vol := b.SourceContext.Volume; vol != nil {
		view.volumeSourceName = vol.SourceVolumeName
	}
	if rev := b.SourceRevision.Git; rev != nil {
		view.revBranch = rev.Branch
		view.revTag = rev.Tag
		view.revCommit = rev.Commit
	}
	if rev := b.SourceRevision.Volume; rev != nil {
		view.revVolumeHash = rev.CurrentVolumeHash
	}
	if b.RegistrySecretRef != nil {
		view.pushSecretID = b.RegistrySecretRef.SecretID
	}
	return view
}

func (v *stackValidator) validateResourceConfigType(resource *models.StackResource) *errors.ServiceError {
	hasBuild := resource.BuildConfig != nil
	hasImage := resource.ImageConfig != nil

	if hasBuild && hasImage {
		return errors.BadRequest("stack resource '%s' cannot have both build and image config", resource.Name)
	}
	if !hasBuild && !hasImage {
		return errors.BadRequest("stack resource '%s' must have either build or image config", resource.Name)
	}
	return nil
}

func (v *stackValidator) validateImageConfig(ctx context.Context, orgID string, resource *models.StackResource, probe bool) *errors.ServiceError {
	if err := resource.ImageConfig.Validate(); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid image config: %s", resource.Name, err.Error())
	}

	if resource.ImageConfig.PullSecretRef != nil {
		if err := v.validateRegistrySecretRef(ctx, resource, resource.ImageConfig.PullSecretRef, credentials.RegistryPurposePull); err != nil {
			return err
		}
	}

	if !probe {
		return nil
	}
	return v.validateImagePullable(ctx, orgID, resource)
}

func (v *stackValidator) validateImagePullable(ctx context.Context, orgID string, resource *models.StackResource) *errors.ServiceError {
	imageRef := resource.ImageConfig.Image
	if isClusterLocalRegistryRef(imageRef) {
		// In-cluster registries aren't reachable from the hub.
		return nil
	}

	resolved, serr := v.credentialResolver.RegistryCredentials(ctx, orgID, imageRef, credentials.RegistryPurposePull, credentials.RegistryAuthSelector{
		SecretRef:            resource.ImageConfig.PullSecretRef,
		RegistryCredentialID: resource.ImageConfig.RegistryCredentialID,
	})
	if serr != nil {
		return errors.BadRequest("stack resource '%s' has invalid pull credentials: %s", resource.Name, serr.Error())
	}

	client, err := v.registryClients.ClientFor(resolved)
	if err != nil {
		return errors.GeneralError("failed to create registry client for stack resource '%s': %s", resource.Name, err.Error())
	}

	exists, err := client.CheckImage(ctx, imageRef)
	if err != nil {
		target := errors.CredentialErrorTarget{
			Kind: errors.CredentialTargetKindImagePull,
			Host: registryHostForRef(imageRef),
			Ref:  imageRef,
		}
		switch {
		case stderrors.Is(err, clients.ErrRateLimited) && resolved.Source == credentials.SourceAnonymous:
			v.logger.Warnf("registry rate limited while checking image '%s' anonymously for stack resource '%s'; skipping existence check", imageRef, resource.Name)
			return nil
		case stderrors.Is(err, clients.ErrAuthFailed) && resolved.Source == credentials.SourceAnonymous:
			return errors.CredentialsRequired(target, "stack resource '%s': image '%s' requires registry credentials", resource.Name, imageRef)
		case stderrors.Is(err, clients.ErrAuthFailed):
			return errors.CredentialsInvalid(target, "stack resource '%s': configured registry credentials were rejected for image '%s'", resource.Name, imageRef)
		default:
			return errors.GeneralError("failed to check image for stack resource '%s': %s", resource.Name, err.Error())
		}
	}
	if !exists {
		return errors.BadRequest("stack resource '%s' image '%s' does not exist or is not pullable", resource.Name, imageRef)
	}
	return nil
}

// gitCommitSHAPattern matches full or abbreviated lowercase-hex commit SHAs,
// mirroring the CRD's CEL validation.
var gitCommitSHAPattern = regexp.MustCompile(`^[0-9a-f]{7,40}$`)

func (v *stackValidator) validateBuildConfig(ctx context.Context, orgID string, resource *models.StackResource, probe bool) *errors.ServiceError {
	if err := resource.BuildConfig.Validate(); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid build config: %s", resource.Name, err.Error())
	}

	if rev := resource.BuildConfig.SourceRevision.Git; rev != nil && rev.Commit != "" {
		if !gitCommitSHAPattern.MatchString(rev.Commit) {
			return errors.BadRequest("stack resource '%s' has invalid commit SHA '%s': must be 7-40 lowercase hex characters", resource.Name, rev.Commit)
		}
	}

	if resource.BuildConfig.SourceContext.Git != nil {
		if err := v.validateGitSource(ctx, orgID, resource, probe); err != nil {
			return err
		}
	}

	if resource.BuildConfig.RegistrySecretRef != nil {
		if err := v.validateRegistrySecretRef(ctx, resource, resource.BuildConfig.RegistrySecretRef, credentials.RegistryPurposePush); err != nil {
			return err
		}
	}

	if !probe {
		return nil
	}
	return v.validatePushAccess(ctx, orgID, resource)
}

func (v *stackValidator) validateRegistrySecretRef(ctx context.Context, resource *models.StackResource, secretRef *models.SecretReference, purpose credentials.RegistryPurpose) *errors.ServiceError {
	if secretRef.SecretID == "" {
		return errors.BadRequest("stack resource '%s' has empty %s secret ID", resource.Name, purpose)
	}

	if err := v.secretService.ValidateImageRegistrySecretForStackResource(ctx, secretRef.SecretID); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid %s secret: %s", resource.Name, purpose, err.Error())
	}

	return nil
}

func (v *stackValidator) validatePushAccess(ctx context.Context, orgID string, resource *models.StackResource) *errors.ServiceError {
	repo := resource.BuildConfig.BuildImageRepository
	// Push probes only make sense for external registries the hub can reach.
	if repo.UseInClusterRegistry || repo.InsecureRegistry || repo.ExternalImageRef == "" ||
		isClusterLocalRegistryRef(repo.ExternalImageRef) {
		return nil
	}

	resolved, serr := v.credentialResolver.RegistryCredentials(ctx, orgID, repo.ExternalImageRef, credentials.RegistryPurposePush, credentials.RegistryAuthSelector{
		SecretRef:            resource.BuildConfig.RegistrySecretRef,
		RegistryCredentialID: resource.BuildConfig.PushRegistryCredentialID,
	})
	if serr != nil {
		return errors.BadRequest("stack resource '%s' has invalid push credentials: %s", resource.Name, serr.Error())
	}

	client, err := v.registryClients.ClientFor(resolved)
	if err != nil {
		return errors.GeneralError("failed to create registry client for stack resource '%s': %s", resource.Name, err.Error())
	}

	if err := client.CheckPushAccess(ctx, repo.ExternalImageRef); err != nil {
		host := registryHostForRef(repo.ExternalImageRef)
		target := errors.CredentialErrorTarget{
			Kind: errors.CredentialTargetKindImagePush,
			Host: host,
			Ref:  repo.ExternalImageRef,
		}
		switch {
		case stderrors.Is(err, clients.ErrRateLimited) && resolved.Source == credentials.SourceAnonymous:
			v.logger.Warnf("registry rate limited while checking push access to '%s' anonymously for stack resource '%s'; skipping probe", repo.ExternalImageRef, resource.Name)
			return nil
		case stderrors.Is(err, clients.ErrAuthFailed) && resolved.Source == credentials.SourceAnonymous:
			return errors.CredentialsRequired(target, "stack resource '%s': pushing to '%s' requires registry credentials", resource.Name, repo.ExternalImageRef)
		case stderrors.Is(err, clients.ErrAuthFailed):
			return errors.CredentialsInvalid(target, "stack resource '%s': configured registry credentials cannot push to '%s'", resource.Name, repo.ExternalImageRef)
		default:
			// Non-auth failures (DNS, connection refused, TLS, 5xx) surface as
			// their own validation error naming the host and real cause rather
			// than being masked as a credentials problem.
			return errors.BadRequest("stack resource '%s' cannot push to registry '%s': %s", resource.Name, host, err.Error())
		}
	}

	return nil
}

// registryHostForRef extracts the normalized registry host from an
// image/repository ref, falling back to the raw ref when it can't be parsed.
func registryHostForRef(ref string) string {
	host, err := clients.NormalizeRegistryHost(ref)
	if err != nil {
		return ref
	}
	return host
}

func isClusterLocalRegistryRef(ref string) bool {
	return strings.HasSuffix(registryHostForRef(ref), clients.ClusterLocalRegistrySuffix)
}

func (v *stackValidator) validateGitSource(ctx context.Context, orgID string, resource *models.StackResource, probe bool) *errors.ServiceError {
	git := resource.BuildConfig.SourceContext.Git

	if git.GitSecretRef != nil {
		if err := v.validateGitWithSecret(ctx, resource); err != nil {
			return err
		}
	}

	if !probe {
		return nil
	}
	return v.validateGitCloneAccess(ctx, orgID, resource)
}

func (v *stackValidator) validateGitCloneAccess(ctx context.Context, orgID string, resource *models.StackResource) *errors.ServiceError {
	git := resource.BuildConfig.SourceContext.Git

	resolved, serr := v.credentialResolver.GitCredentials(ctx, orgID, git.RepoURL, credentials.GitAuthSelector{
		SecretRef:     git.GitSecretRef,
		IntegrationID: git.IntegrationID,
	})
	if serr != nil {
		return errors.BadRequest("stack resource '%s' has invalid git credentials: %s", resource.Name, serr.Error())
	}

	client, err := v.gitClients.ClientFor(git.RepoURL, resolved.Credentials)
	if err != nil {
		return errors.GeneralError("failed to create git client for stack resource '%s': %s", resource.Name, err.Error())
	}

	if _, err := client.CheckAccess(ctx, git.RepoURL); err != nil {
		target := errors.CredentialErrorTarget{
			Kind: errors.CredentialTargetKindGitClone,
			Host: gitclient.RepoHost(git.RepoURL),
			Ref:  git.RepoURL,
		}
		anonymous := resolved.Source == credentials.SourceAnonymous
		switch {
		case stderrors.Is(err, gitclient.ErrRateLimited) && anonymous:
			v.logger.Warnf("git host rate limited while checking repo '%s' anonymously for stack resource '%s'; skipping access check", git.RepoURL, resource.Name)
			return nil
		case anonymous && (stderrors.Is(err, gitclient.ErrAuthFailed) || stderrors.Is(err, gitclient.ErrNotFound)):
			// Private repos often report not-found to anonymous callers.
			return errors.CredentialsRequired(target, "stack resource '%s': repository '%s' was not found or requires credentials", resource.Name, git.RepoURL)
		case stderrors.Is(err, gitclient.ErrAuthFailed):
			return errors.CredentialsInvalid(target, "stack resource '%s': configured git credentials were rejected by '%s'", resource.Name, git.RepoURL)
		case stderrors.Is(err, gitclient.ErrNotFound):
			return errors.BadRequest("stack resource '%s': repository '%s' not found", resource.Name, git.RepoURL)
		default:
			return errors.GeneralError("failed to check git access for stack resource '%s': %s", resource.Name, err.Error())
		}
	}
	return nil
}

func (v *stackValidator) validateGitWithSecret(ctx context.Context, resource *models.StackResource) *errors.ServiceError {
	secretRef := resource.BuildConfig.SourceContext.Git.GitSecretRef

	if secretRef.SecretID == "" {
		return errors.BadRequest("stack resource '%s' has empty git secret ID", resource.Name)
	}

	if err := v.secretService.ValidateGitSecretForStackResource(ctx, secretRef.SecretID); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid git secret: %s", resource.Name, err.Error())
	}

	return nil
}

func (v *stackValidator) validateBuildSourceVolumes(spec *models.Stack) *errors.ServiceError {
	// Populate the volume mounts with the source volume names.
	definedVolumesMap := spec.VolumesMap()
	for i := range spec.StackResources {
		spec.StackResources[i].UserID = spec.UserID
		if spec.StackResources[i].BuildConfig != nil {
			buildConfig := spec.StackResources[i].BuildConfig
			if buildConfig.SourceContext.Volume != nil {
				volume, found := definedVolumesMap[buildConfig.SourceContext.Volume.SourceVolumeName]
				if !found {
					return errors.BadRequest("volume '%s' does not exist", buildConfig.SourceContext.Volume.SourceVolumeName)
				}
				buildConfig.SourceContext.Volume.SourceVolumeName = volume.Name
			}
		}
	}
	return nil
}

func (v *stackValidator) validateVolumeMounts(spec *models.Stack) *errors.ServiceError {
	if len(spec.Volumes) == 0 && spec.HasVolumeMounts() {
		return errors.BadRequest("stack '%s' has volume mounts but no volumes defined", spec.Name)
	}
	if !spec.HasVolumeMounts() {
		return nil
	}

	definedVolumes := spec.Volumes
	definedVolumesMap := make(map[string]*models.Volume)
	for i := range definedVolumes {
		definedVolumesMap[definedVolumes[i].Name] = definedVolumes[i]
	}

	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		for j := range spec.StackResources[i].VolumeMounts {
			currentVolumeMount := currentResource.VolumeMounts[j]
			if _, found := definedVolumesMap[currentVolumeMount.SourceVolumeName]; !found {
				return errors.BadRequest("volume '%s' does not exist", currentVolumeMount.SourceVolumeName)
			}
		}
	}
	return nil
}

func (v *stackValidator) validateStackEnvVars(spec *models.Stack) *errors.ServiceError {
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		if currentResource.ExecutionConfig == nil || currentResource.ExecutionConfig.Env == nil {
			continue
		}
		currentEnvVars := currentResource.ExecutionConfig.Env
		allowedSelfOutputs := make(map[string]struct{}, len(currentResource.EnsureDeclaredOutputs()))
		for _, output := range currentResource.EnsureDeclaredOutputs() {
			allowedSelfOutputs[output.Name] = struct{}{}
		}
		keys := make(map[string]struct{})
		for _, envVar := range currentEnvVars {
			if len(envVar.Name) == 0 {
				return errors.BadRequest("stack resource '%s' has empty env var name", currentResource.Name)
			}
			hasValue := len(envVar.Value) > 0
			hasSelfOutput := len(envVar.SelfOutput) > 0
			if hasValue == hasSelfOutput {
				return errors.BadRequest("stack resource '%s' env var '%s' must set exactly one of value or self_output", currentResource.Name, envVar.Name)
			}
			if hasSelfOutput {
				if _, ok := allowedSelfOutputs[envVar.SelfOutput]; !ok {
					return errors.BadRequest("stack resource '%s' env var '%s' references unsupported self_output '%s'", currentResource.Name, envVar.Name, envVar.SelfOutput)
				}
			}
			if _, exists := keys[envVar.Name]; exists {
				return errors.BadRequest("stack resource '%s' has duplicate env var name '%s'", currentResource.Name, envVar.Name)
			}
			keys[envVar.Name] = struct{}{}
		}

	}

	if err := v.interpolationValidator.ValidateStackInterpolations(spec); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid interpolation: %s", spec.Name, err.Error())
	}

	return nil
}

func (v *stackValidator) validateConnections(ctx context.Context, existing *models.Stack, spec *models.Stack) *errors.ServiceError {
	if len(spec.Connections) == 0 {
		return nil
	}

	resourceMap := spec.ResourcesMap()
	volumeMap := connectionVolumeMap(existing, spec)

	for i, connection := range spec.Connections {
		label := connectionLabel(connection, i)

		if err := validateConnectionKind(label, connection.Kind); err != nil {
			return err
		}
		if err := validateConnectionTargetResource(resourceMap, label, connection.To); err != nil {
			return err
		}

		sourceOutputs, err := v.validateConnectionSource(ctx, resourceMap, volumeMap, label, connection)
		if err != nil {
			return err
		}
		if err := validateConnectionMappings(label, connection, sourceOutputs); err != nil {
			return err
		}
	}

	return nil
}

func validateConnectionKind(label string, kind models.ConnectionKind) *errors.ServiceError {
	switch kind {
	case models.ConnectionKindEnv,
		models.ConnectionKindVolumeMount,
		models.ConnectionKindBuildArtifactSource:
		return nil
	default:
		return errors.BadRequest("connection '%s' has unsupported kind '%s'", label, kind)
	}
}

func (v *stackValidator) validateConnectionSource(
	ctx context.Context,
	resourceMap map[string]*models.StackResource,
	volumeMap map[string]*models.Volume,
	label string,
	connection models.StackConnection,
) ([]models.OutputDescriptor, *errors.ServiceError) {
	switch connection.From.Type {
	case models.TopologyNodeTypeStackResource:
		if connection.From.Name == "" {
			return nil, errors.BadRequest("connection '%s' is missing from.name for stack_resource source", label)
		}
		resource, ok := resourceMap[connection.From.Name]
		if !ok {
			return nil, errors.BadRequest("connection '%s' references unknown stack resource '%s'", label, connection.From.Name)
		}
		if connection.Kind == models.ConnectionKindBuildArtifactSource {
			return nil, validateBuildArtifactSourceConfig(volumeMap, label, connection)
		}
		if len(connection.Config) > 0 {
			return nil, errors.BadRequest("connection '%s' does not support config for from.type '%s'", label, connection.From.Type)
		}
		return resource.EnsureDeclaredOutputs(), nil
	case models.TopologyNodeTypePostgresAddon:
		addon, err := v.validatePostgresConnectionConfig(ctx, label, connection)
		if err != nil {
			return nil, err
		}
		return addon.EnsureDeclaredOutputs(), nil
	case models.TopologyNodeTypeSecret:
		if connection.From.Id == "" {
			return nil, errors.BadRequest("connection '%s' is missing from.id for secret source", label)
		}
		if len(connection.Config) > 0 {
			return nil, errors.BadRequest("connection '%s' does not support config for from.type '%s'", label, connection.From.Type)
		}
		secret, serviceErr := v.secretService.InternalGetByID(ctx, connection.From.Id)
		if serviceErr != nil {
			return nil, errors.BadRequest("connection '%s' references non-existent secret '%s'", label, connection.From.Id)
		}
		return secret.EnsureDeclaredOutputs(), nil
	case models.TopologyNodeTypeVolume:
		if err := validateVolumeConnectionConfig(volumeMap, label, connection); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		if len(connection.Config) > 0 {
			return nil, errors.BadRequest("connection '%s' does not support config for from.type '%s'", label, connection.From.Type)
		}
		return nil, nil
	}
}

func validateConnectionMappings(label string, connection models.StackConnection, sourceOutputs []models.OutputDescriptor) *errors.ServiceError {
	if connection.Kind != models.ConnectionKindEnv {
		return nil
	}

	allowedOutputs := make(map[string]struct{}, len(sourceOutputs))
	for _, output := range sourceOutputs {
		allowedOutputs[output.Name] = struct{}{}
	}

	for _, mapping := range connection.Mappings {
		if mapping.Target.Type != models.ConnectionTargetTypeEnv {
			return errors.BadRequest("connection '%s' env mappings only support target type '%s'", label, models.ConnectionTargetTypeEnv)
		}
		if mapping.Target.Name == "" {
			return errors.BadRequest("connection '%s' has env mapping with empty target name", label)
		}
		if err := validateValueRef(label, connection.From, mapping.Value, allowedOutputs); err != nil {
			return err
		}
	}

	return nil
}

func validateValueRef(label string, source models.TopologyNodeRef, valueRef models.ValueRef, allowedOutputs map[string]struct{}) *errors.ServiceError {
	hasOutput := valueRef.Output != ""
	hasTemplate := valueRef.Template != ""

	if !hasOutput && !hasTemplate {
		return errors.BadRequest("connection '%s' mapping must specify either 'output' or 'template'", label)
	}
	if hasOutput && hasTemplate {
		return errors.BadRequest("connection '%s' mapping must specify either 'output' or 'template', not both", label)
	}
	if hasTemplate && len(valueRef.Values) == 0 {
		return errors.BadRequest("connection '%s' mapping has template but no values", label)
	}

	if hasOutput {
		if _, ok := allowedOutputs[valueRef.Output]; !ok {
			return errors.BadRequest("connection '%s' references unsupported output '%s' for source '%s'", label, valueRef.Output, topologyNodeRefLabel(source))
		}
	}

	if hasTemplate {
		if err := stackdeploy.ValidateTemplateKeys(valueRef.Template, valueRef.Values); err != nil {
			return errors.BadRequest("connection '%s' has invalid template: %s", label, err.Error())
		}
	}

	for _, ref := range valueRef.Values {
		if _, ok := allowedOutputs[ref.Output]; !ok {
			return errors.BadRequest("connection '%s' references unsupported output '%s' for source '%s'", label, ref.Output, topologyNodeRefLabel(source))
		}
	}

	return nil
}

func topologyNodeRefLabel(ref models.TopologyNodeRef) string {
	if ref.Name != "" {
		return fmt.Sprintf("%s:%s", ref.Type, ref.Name)
	}
	if ref.Id != "" {
		return fmt.Sprintf("%s:%s", ref.Type, ref.Id)
	}
	return string(ref.Type)
}

func connectionVolumeMap(existing *models.Stack, spec *models.Stack) map[string]*models.Volume {
	if existing == nil {
		return spec.VolumesMap()
	}

	volumeMap := existing.VolumesMap()
	for name, volume := range spec.VolumesMap() {
		volumeMap[name] = volume
	}
	return volumeMap
}

func validateConnectionTargetResource(resourceMap map[string]*models.StackResource, label string, ref models.TopologyNodeRef) *errors.ServiceError {
	switch ref.Type {
	case models.TopologyNodeTypeStackResource:
		if ref.Name == "" {
			return errors.BadRequest("connection '%s' is missing to.name for stack_resource target", label)
		}
		if _, ok := resourceMap[ref.Name]; !ok {
			return errors.BadRequest("connection '%s' references unknown stack resource '%s'", label, ref.Name)
		}
	case models.TopologyNodeTypeVolume:
		if ref.Name == "" {
			return errors.BadRequest("connection '%s' is missing to.name for volume target", label)
		}
	}
	return nil
}

func validateBuildArtifactSourceConfig(volumeMap map[string]*models.Volume, label string, connection models.StackConnection) *errors.ServiceError {
	if connection.To.Type != models.TopologyNodeTypeVolume {
		return errors.BadRequest("connection '%s' with kind '%s' requires to.type '%s'", label, connection.Kind, models.TopologyNodeTypeVolume)
	}
	if _, ok := volumeMap[connection.To.Name]; !ok {
		return errors.BadRequest("connection '%s' references unknown volume '%s'", label, connection.To.Name)
	}
	if err := validateConfigKeys(connection.Config, map[string]struct{}{
		string(models.ConnectionConfigKeySourcePath):      {},
		string(models.ConnectionConfigKeyDestinationPath): {},
	}, label, "build_artifact_source"); err != nil {
		return err
	}
	sourcePath, _, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeySourcePath), label)
	if err != nil {
		return err
	}
	if sourcePath == "" {
		return errors.BadRequest("connection '%s' requires config.source_path for build artifact sources", label)
	}
	if _, _, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeyDestinationPath), label); err != nil {
		return err
	}
	return nil
}

func (v *stackValidator) validatePostgresConnectionConfig(ctx context.Context, label string, connection models.StackConnection) (*models.PostgresAddon, *errors.ServiceError) {
	if connection.Kind != models.ConnectionKindEnv {
		return nil, errors.BadRequest("connection '%s' with from.type '%s' only supports kind '%s'", label, connection.From.Type, models.ConnectionKindEnv)
	}
	if connection.From.Id == "" {
		return nil, errors.BadRequest("connection '%s' is missing from.id for postgres addon source", label)
	}

	if err := validateConfigKeys(connection.Config, map[string]struct{}{
		string(models.ConnectionConfigKeyDatabase):        {},
		string(models.ConnectionConfigKeyCredentialScope): {},
		string(models.ConnectionConfigKeySuperuser):       {},
	}, label, "postgres"); err != nil {
		return nil, err
	}

	database, _, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeyDatabase), label)
	if err != nil {
		return nil, err
	}

	credentialScope, hasCredentialScope, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeyCredentialScope), label)
	if err != nil {
		return nil, err
	}
	superuser, hasSuperuser, err := getOptionalBoolConfig(connection.Config, string(models.ConnectionConfigKeySuperuser), label)
	if err != nil {
		return nil, err
	}
	if hasCredentialScope && hasSuperuser {
		return nil, errors.BadRequest("connection '%s' cannot set both config.credential_scope and config.superuser", label)
	}

	scope := "owner"
	if hasCredentialScope {
		switch credentialScope {
		case "owner", "superuser":
			scope = credentialScope
		default:
			return nil, errors.BadRequest("connection '%s' has unsupported postgres credential scope '%s'", label, credentialScope)
		}
	} else if hasSuperuser && superuser {
		scope = "superuser"
	}

	addon, serviceErr := v.postgresAddonService.GetPostgresAddon(ctx, connection.From.Id)
	if serviceErr != nil {
		return nil, errors.BadRequest("connection '%s' references non-existent postgres addon '%s'", label, connection.From.Id)
	}

	if scope == "superuser" {
		if !addon.Configuration.EnableSuperuserAccess {
			return nil, errors.BadRequest("connection '%s' requests superuser access but addon '%s' does not have superuser access enabled", label, connection.From.Id)
		}
		return addon, nil
	}

	if database == "" {
		return nil, errors.BadRequest("connection '%s' requires config.database when postgres credential scope is owner", label)
	}
	if !addon.HasDatabase(database) {
		return nil, errors.BadRequest("connection '%s' references non-existent database '%s' in postgres addon '%s'", label, database, connection.From.Id)
	}

	return addon, nil
}

func validateVolumeConnectionConfig(volumeMap map[string]*models.Volume, label string, connection models.StackConnection) *errors.ServiceError {
	if connection.Kind != models.ConnectionKindVolumeMount {
		return errors.BadRequest("connection '%s' with from.type '%s' only supports kind '%s'", label, connection.From.Type, models.ConnectionKindVolumeMount)
	}
	if connection.From.Name == "" {
		return errors.BadRequest("connection '%s' is missing from.name for volume source", label)
	}
	if _, ok := volumeMap[connection.From.Name]; !ok {
		return errors.BadRequest("connection '%s' references unknown volume '%s'", label, connection.From.Name)
	}

	if err := validateConfigKeys(connection.Config, map[string]struct{}{
		string(models.ConnectionConfigKeyMountPath): {},
		string(models.ConnectionConfigKeySubPath):   {},
		string(models.ConnectionConfigKeyReadOnly):  {},
	}, label, "volume"); err != nil {
		return err
	}

	mountPath, _, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeyMountPath), label)
	if err != nil {
		return err
	}
	if mountPath == "" {
		return errors.BadRequest("connection '%s' requires config.mount_path for volume mounts", label)
	}

	if _, _, err := getOptionalStringConfig(connection.Config, string(models.ConnectionConfigKeySubPath), label); err != nil {
		return err
	}
	if _, _, err := getOptionalBoolConfig(connection.Config, string(models.ConnectionConfigKeyReadOnly), label); err != nil {
		return err
	}

	return nil
}

func validateConfigKeys(config map[string]interface{}, allowed map[string]struct{}, label string, configType string) *errors.ServiceError {
	for key := range config {
		if _, ok := allowed[key]; !ok {
			return errors.BadRequest("connection '%s' has unsupported %s config key '%s'", label, configType, key)
		}
	}
	return nil
}

func getOptionalStringConfig(config map[string]interface{}, key string, label string) (string, bool, *errors.ServiceError) {
	value, ok := config[key]
	if !ok {
		return "", false, nil
	}
	asString, ok := value.(string)
	if !ok {
		return "", false, errors.BadRequest("connection '%s' config.%s must be a string", label, key)
	}
	return asString, true, nil
}

func getOptionalBoolConfig(config map[string]interface{}, key string, label string) (bool, bool, *errors.ServiceError) {
	value, ok := config[key]
	if !ok {
		return false, false, nil
	}
	asBool, ok := value.(bool)
	if !ok {
		return false, false, errors.BadRequest("connection '%s' config.%s must be a boolean", label, key)
	}
	return asBool, true, nil
}

func connectionLabel(connection models.StackConnection, index int) string {
	if connection.ID != "" {
		return connection.ID
	}
	return fmt.Sprintf("#%d", index)
}

func (v *stackValidator) validateDomainExistence(ctx context.Context, spec *models.Stack) *errors.ServiceError {
	if !spec.HasExposedPorts() {
		return nil
	}

	orgDomains, err := v.domainService.ListByOrganisationID(ctx, spec.OrganisationID)
	if err != nil {
		return errors.GeneralError("failed to list domains for organisation '%s': %s", spec.OrganisationID, err.Error())
	}

	if len(orgDomains) == 0 {
		return errors.BadRequest("stack '%s' has publicly exposed ports but no domains defined for organisation '%s'", spec.Name, spec.OrganisationID)
	}
	return nil
}

func (v *stackValidator) validateStackPorts(spec *models.Stack) *errors.ServiceError {
	subdomainPrefixes := make(map[string]string)
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		if len(currentResource.Ports) == 0 {
			continue
		}
		currentPorts := currentResource.Ports
		portNames := make(map[string]struct{}, len(currentPorts))
		portNumbers := make(map[int]struct{}, len(currentPorts))

		for _, port := range currentPorts {
			if port.Number <= 0 {
				return errors.BadRequest("stack resource '%s' has invalid port number", currentResource.Name)
			}
			if port.Name == "" {
				return errors.BadRequest("stack resource '%s' has port %d missing name", currentResource.Name, port.Number)
			}
			if _, exists := portNames[port.Name]; exists {
				return errors.BadRequest("stack resource '%s' has duplicate port name '%s'", currentResource.Name, port.Name)
			}
			portNames[port.Name] = struct{}{}
			if _, exists := portNumbers[port.Number]; exists {
				return errors.BadRequest("stack resource '%s' has duplicate port number %d", currentResource.Name, port.Number)
			}
			portNumbers[port.Number] = struct{}{}
			if err := validatePortName(port.Name); err != nil {
				return errors.BadRequest("stack resource '%s' has invalid port name '%s': %s", currentResource.Name, port.Name, err.Error())
			}
			if port.ExposedToPublic && port.SubdomainPrefix != "" {
				if owner, exists := subdomainPrefixes[port.SubdomainPrefix]; exists {
					return errors.BadRequest(
						"duplicate subdomain prefix '%s': used by both resource '%s' and '%s'",
						port.SubdomainPrefix, owner, currentResource.Name)
				}
				subdomainPrefixes[port.SubdomainPrefix] = currentResource.Name
			}
		}
	}
	return nil
}

func validatePortName(name string) error {
	for i, r := range name {
		valid := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if !valid {
			return fmt.Errorf("must contain only lowercase letters, numbers, and hyphens")
		}
		if i == 0 && r == '-' {
			return fmt.Errorf("must start with a lowercase letter or number")
		}
	}
	if name[len(name)-1] == '-' {
		return fmt.Errorf("must end with a lowercase letter or number")
	}
	return nil
}
