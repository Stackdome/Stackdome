package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	stderrors "errors"
	"fmt"
	"strings"
	"time"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/auth"
	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/stackfile"
	"github.com/Stackdome/stackdome/pkg/stores"
	"k8s.io/utils/ptr"
)

type PreviewStackService interface {
	BackgroundJobEnqueuerInjectable
	Create(ctx context.Context, preview *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError)
	Sync(ctx context.Context, previewStackID string, opts PreviewSyncOpts) *errors.ServiceError
	Delete(ctx context.Context, previewStackID string) *errors.ServiceError
	Get(ctx context.Context, previewStackID string) (*models.PreviewStack, *errors.ServiceError)
	List(ctx context.Context, projectID string, params stores.ListParams) (*stores.PaginatedResult[*models.PreviewStack], *errors.ServiceError)
	InternalFetchStackfile(ctx context.Context, config *models.StackPreviewConfig, commitSHA string) (content []byte, hash string, opErr *errors.OperationError)
	InternalBuildStackFromContent(ctx context.Context, config *models.StackPreviewConfig, preview *models.PreviewStack, content []byte) (*models.Stack, *errors.OperationError)
}

type PreviewStackServiceSpec struct {
	Store              stores.PreviewStackStore
	ConfigStore        stores.StackPreviewConfigStore
	StackService       StackService
	ReleaseService     StackReleaseService
	SecretService      SecretService
	CredentialResolver CredentialResolver
	Permissions        auth.PermissionService
}

type previewStackService struct {
	store              stores.PreviewStackStore
	configStore        stores.StackPreviewConfigStore
	stackService       StackService
	releaseService     StackReleaseService
	secretService      SecretService
	credentialResolver CredentialResolver
	permissions        auth.PermissionService
	BackgroundJobEnqueuerDep
}

type PreviewSyncOpts struct {
	Commit           string
	StackfileContent *string
	ForceSync        bool
	ImageOverrides   map[string]string
}

func NewPreviewStackService(spec PreviewStackServiceSpec) PreviewStackService {
	return &previewStackService{
		store:              spec.Store,
		configStore:        spec.ConfigStore,
		stackService:       spec.StackService,
		releaseService:     spec.ReleaseService,
		secretService:      spec.SecretService,
		credentialResolver: spec.CredentialResolver,
		permissions:        spec.Permissions,
	}
}

func (s *previewStackService) Create(ctx context.Context, preview *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
	config, sErr := s.configStore.GetByID(ctx, preview.StackPreviewConfigID)
	if sErr != nil {
		return nil, sErr
	}

	if permErr := s.permissions.Check(ctx, config.ProjectID, auth.ResourcePreviewStacks, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}

	if preview.ProjectID != "" && preview.ProjectID != config.ProjectID {
		return nil, errors.BadRequest("preview stack project does not match config project")
	}

	if preview.PRNumber == "" {
		return nil, errors.BadRequest("pr_number is required")
	}
	if preview.Branch == "" {
		return nil, errors.BadRequest("branch is required")
	}

	existing, sErr := s.store.GetByConfigAndPR(ctx, preview.StackPreviewConfigID, preview.PRNumber)
	if sErr != nil && sErr.Code != errors.ErrorNotFound {
		return nil, sErr
	}

	if existing != nil {
		return nil, errors.Conflict("preview stack for PR #%s already exists", preview.PRNumber)
	}

	activeCount, sErr := s.store.CountActiveByConfigID(ctx, preview.StackPreviewConfigID)
	if sErr != nil {
		return nil, sErr
	}
	if activeCount >= int64(config.MaxActivePreviews) {
		return nil, errors.BadRequest("maximum active preview stacks (%d) reached for this config", config.MaxActivePreviews)
	}

	if preview.CommitSHA == "" {
		gitClient, gErr := s.gitClientForConfig(ctx, config)
		if gErr != nil {
			return nil, errors.GeneralError("failed to create git client: %v", gErr)
		}
		result, branchErr := gitClient.GetBranchHeadSHA(ctx, config.GitRepository.RepoURL, preview.Branch)
		if branchErr != nil {
			if stderrors.Is(branchErr, gitclient.ErrNotFound) {
				return nil, errors.BadRequest("branch '%s' not found in repository", preview.Branch)
			}
			return nil, errors.GeneralError("branch validation failed: %v", branchErr)
		}
		preview.CommitSHA = result.HeadSHA
	}

	preview.OrganisationID = config.OrganisationID
	preview.ProjectID = config.ProjectID
	preview.Name = sanitizePreviewName(fmt.Sprintf("pr-%s-%s", preview.PRNumber, config.Name))
	preview.Status = models.PreviewStackStatus{Phase: models.PreviewStackPhaseProvisioning, Reason: "Created"}

	created, sErr := s.store.Create(ctx, preview)
	if sErr != nil {
		return nil, sErr
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.PreviewStack{ID: created.ID}); err != nil {
		return nil, errors.GeneralError("failed to enqueue preview stack processing: %v", err)
	}

	return created, nil
}

func (s *previewStackService) Sync(ctx context.Context, previewStackID string, opts PreviewSyncOpts) *errors.ServiceError {
	preview, sErr := s.store.GetByID(ctx, previewStackID)
	if sErr != nil {
		return sErr
	}

	if permErr := s.permissions.Check(ctx, preview.ProjectID, auth.ResourcePreviewStacks, previewStackID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	if preview.StackID == nil {
		return errors.BadRequest("preview stack has not been provisioned yet")
	}
	if preview.DeletionTimestamp != nil {
		return errors.BadRequest("preview stack is being deleted")
	}

	config, sErr := s.configStore.GetByID(ctx, preview.StackPreviewConfigID)
	if sErr != nil {
		return sErr
	}

	targetCommit, sErr := s.resolveTargetCommit(ctx, config, preview.Branch, opts.Commit)
	if sErr != nil {
		return sErr
	}

	if s.isSyncNoop(preview, targetCommit, opts) {
		return nil
	}

	preview.CommitSHA = targetCommit
	if opts.ForceSync {
		preview.ForceSyncRequestedAt = ptr.To(time.Now().UTC())
	}
	if opts.StackfileContent != nil {
		preview.StackfileContent = opts.StackfileContent
	}
	if len(opts.ImageOverrides) > 0 {
		preview.ImageOverrides = opts.ImageOverrides
	}
	preview.Status = models.PreviewStackStatus{Phase: models.PreviewStackPhaseProvisioning, Reason: "SyncTriggered"}

	if _, sErr := s.store.Update(ctx, preview); sErr != nil {
		return sErr
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.PreviewStack{ID: preview.ID}); err != nil {
		return errors.GeneralError("failed to enqueue preview stack sync: %v", err)
	}

	return nil
}

func (s *previewStackService) resolveTargetCommit(ctx context.Context, config *models.StackPreviewConfig, branch, explicitCommit string) (string, *errors.ServiceError) {
	if explicitCommit != "" {
		return explicitCommit, nil
	}
	gitClient, gErr := s.gitClientForConfig(ctx, config)
	if gErr != nil {
		return "", errors.GeneralError("failed to create git client: %v", gErr)
	}
	result, branchErr := gitClient.GetBranchHeadSHA(ctx, config.GitRepository.RepoURL, branch)
	if branchErr != nil {
		return "", errors.GeneralError("failed to resolve branch head: %v", branchErr)
	}
	return result.HeadSHA, nil
}

func (s *previewStackService) isSyncNoop(preview *models.PreviewStack, targetCommit string, opts PreviewSyncOpts) bool {
	if opts.ForceSync {
		return false
	}
	if preview.Status.Phase == models.PreviewStackPhaseFailed {
		return false
	}
	if opts.StackfileContent != nil {
		if ContentHash([]byte(*opts.StackfileContent)) != preview.ReconcilerStatus.LastAppliedStackfileHash {
			return false
		}
	}
	return targetCommit == preview.CommitSHA && overridesEqual(opts.ImageOverrides, preview.ImageOverrides)
}

func overridesEqual(incoming map[string]string, stored models.ImageOverrides) bool {
	if len(incoming) == 0 && len(stored) == 0 {
		return true
	}
	if len(incoming) != len(stored) {
		return false
	}
	for k, v := range incoming {
		if stored[k] != v {
			return false
		}
	}
	return true
}

func (s *previewStackService) Delete(ctx context.Context, previewStackID string) *errors.ServiceError {
	preview, sErr := s.store.GetByID(ctx, previewStackID)
	if sErr != nil {
		return sErr
	}

	if permErr := s.permissions.Check(ctx, preview.ProjectID, auth.ResourcePreviewStacks, previewStackID, auth.ActionDelete); permErr != nil {
		return permErr
	}

	preview.Status = models.PreviewStackStatus{Phase: models.PreviewStackPhaseDeleting, Reason: "DeleteRequested"}
	preview.DeletionTimestamp = ptr.To(time.Now().UTC())
	if _, sErr := s.store.Update(ctx, preview); sErr != nil {
		return sErr
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.PreviewStack{ID: previewStackID}); err != nil {
		return errors.GeneralError("failed to enqueue preview stack deletion: %v", err)
	}

	return nil
}

func (s *previewStackService) Get(ctx context.Context, previewStackID string) (*models.PreviewStack, *errors.ServiceError) {
	preview, sErr := s.store.GetByID(ctx, previewStackID)
	if sErr != nil {
		return nil, sErr
	}

	if permErr := s.permissions.Check(ctx, preview.ProjectID, auth.ResourcePreviewStacks, previewStackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}

	return preview, nil
}

func (s *previewStackService) List(ctx context.Context, projectID string, params stores.ListParams) (*stores.PaginatedResult[*models.PreviewStack], *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, projectID, auth.ResourcePreviewStacks, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}

	return s.store.ListByProjectID(ctx, projectID, params)
}

func (s *previewStackService) gitClientForConfig(ctx context.Context, config *models.StackPreviewConfig) (gitclient.GitClient, error) {
	selector := credentials.GitAuthSelector{
		IntegrationID: config.GitRepository.IntegrationID,
	}

	resolved, sErr := s.credentialResolver.GitCredentials(ctx, config.OrganisationID, config.GitRepository.RepoURL, selector)
	if sErr != nil {
		return nil, sErr
	}
	return gitclient.NewGitClientForRepo(config.GitRepository.RepoURL, resolved.Credentials)
}

// InternalFetchStackfile retrieves the stackfile from git and returns its content along
// with a sha256 hex-encoded hash of the content.
func (s *previewStackService) InternalFetchStackfile(ctx context.Context, config *models.StackPreviewConfig, commitSHA string) ([]byte, string, *errors.OperationError) {
	gitClient, gErr := s.gitClientForConfig(ctx, config)
	if gErr != nil {
		return nil, "", errors.Transient("GitClientFailed", fmt.Sprintf("failed to create git client: %v", gErr))
	}

	content, fetchErr := gitClient.FetchFile(ctx, config.GitRepository.RepoURL, commitSHA, config.StackfilePath)
	if fetchErr != nil {
		if stderrors.Is(fetchErr, gitclient.ErrNotFound) {
			return nil, "", errors.Permanent("StackfileNotFound",
				fmt.Sprintf("stackfile '%s' not found at commit %s", config.StackfilePath, commitSHA))
		}
		if stderrors.Is(fetchErr, gitclient.ErrAuthFailed) {
			return nil, "", errors.Permanent("GitAuthFailed",
				fmt.Sprintf("git authentication failed: %v", fetchErr))
		}
		if stderrors.Is(fetchErr, gitclient.ErrRateLimited) {
			return nil, "", errors.Transient("GitRateLimited",
				fmt.Sprintf("git API rate limited: %v", fetchErr))
		}
		return nil, "", errors.Transient("StackfileFetchFailed",
			fmt.Sprintf("failed to fetch stackfile: %v", fetchErr))
	}

	return content, ContentHash(content), nil
}

func ContentHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// InternalBuildStackFromContent parses stackfile content, resolves references, applies
// preview transforms, and returns the resulting stack model.
func (s *previewStackService) InternalBuildStackFromContent(ctx context.Context, config *models.StackPreviewConfig, preview *models.PreviewStack, content []byte) (*models.Stack, *errors.OperationError) {
	sf, err := stackfile.Load(content)
	if err != nil {
		return nil, errors.Permanent("InvalidStackfile", fmt.Sprintf("failed to parse stackfile: %v", err))
	}

	stack, err := sf.ToStack()
	if err != nil {
		return nil, errors.Permanent("InvalidStackfile", fmt.Sprintf("failed to convert stackfile: %v", err))
	}

	resolver := &previewSecretResolver{
		secretService: s.secretService,
		orgID:         config.OrganisationID,
	}
	if err := stackfile.ResolveStack(ctx, &stack, resolver); err != nil {
		return nil, errors.Permanent("StackfileResolutionFailed",
			fmt.Sprintf("failed to resolve stackfile references: %v", err))
	}

	s.applyBranchSwap(&stack, config.GitRepository.RepoURL, preview.Branch, preview.CommitSHA)
	s.applyImageOverrides(&stack, preview.ImageOverrides)
	s.applyDomainPrefix(&stack, preview.PRNumber)

	model := presenters.ConvertStack(&stack)
	s.applyIntegrationFromConfig(model, config)
	model.Name = preview.Name
	model.OrganisationID = config.OrganisationID
	model.ProjectID = config.ProjectID
	model.UserID = preview.UserID
	model.Labels = append(model.Labels,
		models.Label{Key: models.PreviewStackLabel, Value: models.LabelValueTrue},
		models.Label{Key: models.PreviewConfigIDLabel, Value: preview.StackPreviewConfigID},
		models.Label{Key: models.PreviewPRNumberLabel, Value: preview.PRNumber},
		models.Label{Key: models.PreviewStackIDLabel, Value: preview.ID},
	)

	return model, nil
}

// applyBranchSwap sets the branch to the preview branch for resources whose git
// repo URL matches the config's repo URL, pinning to the resolved commit SHA.
func (s *previewStackService) applyBranchSwap(stack *openapi.Stack, repoURL, branch, commitSHA string) {
	for i := range stack.Spec.StackResources {
		res := &stack.Spec.StackResources[i]
		if res.Source == nil || res.Source.Git == nil {
			continue
		}
		if normalizeRepoURL(res.Source.Git.RepoUrl) != normalizeRepoURL(repoURL) {
			continue
		}
		res.Source.Git.SetBranch(branch)
		res.Source.Git.SetCommit(commitSHA)
		res.Source.Git.Tag = nil
	}
}

// applyImageOverrides replaces build sources with image sources for matching resources.
func (s *previewStackService) applyImageOverrides(stack *openapi.Stack, overrides models.ImageOverrides) {
	if len(overrides) == 0 {
		return
	}
	for i := range stack.Spec.StackResources {
		res := &stack.Spec.StackResources[i]
		image, ok := overrides[res.Name]
		if !ok {
			continue
		}
		res.Source = &openapi.SourceSpec{Image: openapi.NewImageSource(image)}
	}
}

// applyDomainPrefix prepends a PR-specific prefix to subdomain prefixes on exposed ports.
func (s *previewStackService) applyDomainPrefix(stack *openapi.Stack, prNumber string) {
	prefix := fmt.Sprintf("pr-%s-", prNumber)
	for i := range stack.Spec.StackResources {
		res := &stack.Spec.StackResources[i]
		for j := range res.Ports {
			port := &res.Ports[j]
			if port.ExposedToPublic {
				existing := port.GetSubdomainPrefix()
				prefixed := prefix + existing
				port.SubdomainPrefix = &prefixed
			}
		}
	}
}

// applyIntegrationFromConfig pins the preview config's git integration onto
// build configs that match the config's repo URL and don't already have clone
// auth. This runs on the converted model so preview builds follow the same
// credential ladder as the config: integration id, then host auto-match, then
// anonymous.
func (s *previewStackService) applyIntegrationFromConfig(stack *models.Stack, config *models.StackPreviewConfig) {
	if config.GitRepository.IntegrationID == "" {
		return
	}
	for _, res := range stack.StackResources {
		if res.BuildConfig == nil || res.BuildConfig.SourceContext.Git == nil {
			continue
		}
		git := res.BuildConfig.SourceContext.Git
		if git.IntegrationID != "" {
			continue
		}
		if normalizeRepoURL(git.RepoURL) != normalizeRepoURL(config.GitRepository.RepoURL) {
			continue
		}
		git.IntegrationID = config.GitRepository.IntegrationID
	}
}

// sanitizePreviewName cleans a preview stack name to be valid.
func sanitizePreviewName(name string) string {
	name = strings.ToLower(name)
	// Replace invalid characters with hyphens.
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	result := b.String()
	// Trim leading/trailing hyphens and collapse runs.
	parts := strings.Split(result, "-")
	var clean []string
	for _, p := range parts {
		if p != "" {
			clean = append(clean, p)
		}
	}
	result = strings.Join(clean, "-")
	// Truncate to a reasonable length.
	if len(result) > 63 {
		result = result[:63]
	}
	return result
}

// normalizeRepoURL strips trailing .git and slashes for comparison.
func normalizeRepoURL(url string) string {
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimSuffix(url, "/")
	return strings.ToLower(url)
}

// previewSecretResolver implements stackfile.Resolver for preview environments.
type previewSecretResolver struct {
	secretService SecretService
	orgID         string
}

func (r *previewSecretResolver) ResolveSecretByName(ctx context.Context, name string) (string, error) {
	secret, sErr := r.secretService.InternalGetByName(ctx, r.orgID, name)
	if sErr != nil {
		return "", fmt.Errorf("secret '%s' not found: %w", name, sErr)
	}
	return secret.ID, nil
}

// TODO: Implement addon resolution for preview stacks.?
func (r *previewSecretResolver) ResolveAddonByName(ctx context.Context, addonType, name string) (string, error) {
	// Addon resolution is not supported for preview stacks currently.
	return "", fmt.Errorf("addon '%s' of type '%s' cannot be resolved in preview context", name, addonType)
}
