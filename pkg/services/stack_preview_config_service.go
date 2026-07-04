package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/auth"
	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/validator"
	"github.com/Stackdome/stackdome/pkg/validator/secret"
	"github.com/google/uuid"
)

type StackPreviewConfigService interface {
	Create(ctx context.Context, config *models.StackPreviewConfig) (*models.StackPreviewConfig, *errors.ServiceError)
	Get(ctx context.Context, id string) (*models.StackPreviewConfig, *errors.ServiceError)
	Update(ctx context.Context, id string, config *models.StackPreviewConfig) (*models.StackPreviewConfig, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	List(ctx context.Context, teamID string, params stores.ListParams) (*stores.PaginatedResult[*models.StackPreviewConfig], *errors.ServiceError)
}

type StackPreviewConfigServiceSpec struct {
	Store                   stores.StackPreviewConfigStore
	PreviewStackStore       stores.PreviewStackStore
	SecretService           SecretService
	CredentialResolver      CredentialResolver
	SourceCredentialService SourceCredentialService
	Permissions             auth.PermissionService
}

type stackPreviewConfigService struct {
	store              stores.StackPreviewConfigStore
	previewStackStore  stores.PreviewStackStore
	secretService      SecretService
	credentialResolver CredentialResolver
	sourceCredentials  SourceCredentialService
	secretValidator    validator.SecretValidator
	permissions        auth.PermissionService
}

func NewStackPreviewConfigService(spec StackPreviewConfigServiceSpec) StackPreviewConfigService {
	return &stackPreviewConfigService{
		store:              spec.Store,
		previewStackStore:  spec.PreviewStackStore,
		secretService:      spec.SecretService,
		credentialResolver: spec.CredentialResolver,
		sourceCredentials:  spec.SourceCredentialService,
		secretValidator:    secret.NewSecretValidator(),
		permissions:        spec.Permissions,
	}
}

func (s *stackPreviewConfigService) Create(ctx context.Context, config *models.StackPreviewConfig) (*models.StackPreviewConfig, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, config.TeamID, auth.ResourcePreviewConfigs, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}
	if config.MaxActivePreviews <= 0 {
		config.MaxActivePreviews = models.DefaultMaxActivePreviews
	}
	if config.GitRepository.BaseBranch == "" {
		config.GitRepository.BaseBranch = models.DefaultBaseBranch
	}

	// Pre-generate the config ID so managed secrets can reference their owner
	// before the row exists.
	if config.ID == "" {
		config.ID = uuid.NewString()
	}
	// PreparePreviewConfig materializes managed secrets before validation, so
	// keep a rollback handle and undo the credential rows if validation or the
	// store write fails — otherwise they orphan under a config that never
	// existed.
	var sourceRollback SourceRollback
	if s.sourceCredentials != nil {
		rollback, err := s.sourceCredentials.PreparePreviewConfig(ctx, config)
		if err != nil {
			return nil, err
		}
		sourceRollback = rollback
	}

	if err := s.validate(ctx, config); err != nil {
		rollbackSources(ctx, sourceRollback)
		return nil, err
	}

	if err := s.validateGitRepo(ctx, config); err != nil {
		rollbackSources(ctx, sourceRollback)
		return nil, err
	}

	created, err := s.store.Create(ctx, config)
	if err != nil {
		rollbackSources(ctx, sourceRollback)
		return nil, err
	}
	return created, nil
}

func (s *stackPreviewConfigService) Get(ctx context.Context, id string) (*models.StackPreviewConfig, *errors.ServiceError) {
	config, sErr := s.store.GetByID(ctx, id)
	if sErr != nil {
		return nil, sErr
	}

	if permErr := s.permissions.Check(ctx, config.TeamID, auth.ResourcePreviewConfigs, id, auth.ActionRead); permErr != nil {
		return nil, permErr
	}

	return config, nil
}

func (s *stackPreviewConfigService) Update(ctx context.Context, id string, updated *models.StackPreviewConfig) (*models.StackPreviewConfig, *errors.ServiceError) {
	existing, sErr := s.store.GetByID(ctx, id)
	if sErr != nil {
		return nil, sErr
	}

	if permErr := s.permissions.Check(ctx, existing.TeamID, auth.ResourcePreviewConfigs, id, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	updated.ID = existing.ID
	updated.OrganisationID = existing.OrganisationID
	updated.TeamID = existing.TeamID
	updated.UserID = existing.UserID
	updated.Name = existing.Name

	var sourceRollback SourceRollback
	if s.sourceCredentials != nil {
		rollback, err := s.sourceCredentials.PreparePreviewConfig(ctx, updated)
		if err != nil {
			return nil, err
		}
		sourceRollback = rollback
	}

	if err := s.validate(ctx, updated); err != nil {
		rollbackSources(ctx, sourceRollback)
		return nil, err
	}

	if err := s.validateGitRepo(ctx, updated); err != nil {
		rollbackSources(ctx, sourceRollback)
		return nil, err
	}

	result, err := s.store.Update(ctx, updated)
	if err != nil {
		rollbackSources(ctx, sourceRollback)
		return nil, err
	}
	return result, nil
}

func (s *stackPreviewConfigService) Delete(ctx context.Context, id string) *errors.ServiceError {
	config, sErr := s.store.GetByID(ctx, id)
	if sErr != nil {
		return sErr
	}

	if permErr := s.permissions.Check(ctx, config.TeamID, auth.ResourcePreviewConfigs, id, auth.ActionDelete); permErr != nil {
		return permErr
	}

	activeCount, sErr := s.previewStackStore.CountActiveByConfigID(ctx, id)
	if sErr != nil {
		return sErr
	}
	if activeCount > 0 {
		return errors.Conflict("cannot delete preview config with %d active preview stack(s)", activeCount)
	}

	// Clean up managed secrets before deleting the row. Cleanup is idempotent
	// (a no-op once the secrets are gone), so a retry after a partial failure
	// re-runs cleanup harmlessly and then deletes the row. Doing it in this
	// order means a cleanup failure leaves the row intact for the retry rather
	// than orphaning the secrets forever.
	if s.sourceCredentials != nil {
		if cleanupErr := s.sourceCredentials.CleanupPreviewConfig(ctx, id); cleanupErr != nil {
			return cleanupErr
		}
	}
	if err := s.store.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

func (s *stackPreviewConfigService) List(ctx context.Context, teamID string, params stores.ListParams) (*stores.PaginatedResult[*models.StackPreviewConfig], *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, teamID, auth.ResourcePreviewConfigs, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}

	return s.store.ListByTeamID(ctx, teamID, params)
}

func (s *stackPreviewConfigService) validate(ctx context.Context, config *models.StackPreviewConfig) *errors.ServiceError {
	if config.Name == "" {
		return errors.Validation("name is required")
	}
	if config.GitRepository.RepoURL == "" {
		return errors.Validation("git_repository.repo_url is required")
	}

	if config.GitRepository.BaseBranch == "" {
		return errors.Validation("git_repository.base_branch is required")
	}

	if config.MaxActivePreviews != 0 && (config.MaxActivePreviews < 1 || config.MaxActivePreviews > models.MaxMaxActivePreviews) {
		return errors.Validation("max_active_previews must be between 1 and %d", models.MaxMaxActivePreviews)
	}

	if config.StackfilePath == "" {
		config.StackfilePath = models.DefaultStackfilePath
	}

	if config.UsesGitSecret() {
		secret, gErr := s.secretService.InternalGetByID(ctx, *config.GitSecretID())
		if gErr != nil {
			if gErr.Code == errors.ErrorNotFound {
				return errors.Validation("referenced git secret with id '%s' not found", *config.GitSecretID())
			}
			return errors.InternalServerError("failed to get git secret: %v", gErr)
		}
		if err := s.secretValidator.ValidateSecretType(models.SecretTypeGitCredentials, secret); err != nil {
			return err
		}
	}

	return nil
}

// validateGitRepo checks that the git repo is accessible and the stackfile exists.
// For GitHub repos, it uses the Contents API (fast, no clone).
// For non-GitHub repos, it validates credentials and branch existence.
// If stackfile content is provided inline, the file existence check is skipped,
// but git credentials are still validated when a secret ref is present (needed for builds).
func (s *stackPreviewConfigService) validateGitRepo(ctx context.Context, config *models.StackPreviewConfig) *errors.ServiceError {
	// Ensure repo is accessible
	gitClient, err := s.gitClientForConfig(ctx, config)
	if err != nil {
		return errors.BadRequest("invalid git credentials: %v", err)
	}

	if _, err := gitClient.CheckAccess(ctx, config.GitRepository.RepoURL); err != nil {
		return errors.BadRequest("git access check failed: %v", err)
	}

	return nil
}

func (s *stackPreviewConfigService) gitClientForConfig(ctx context.Context, config *models.StackPreviewConfig) (gitclient.GitClient, error) {
	selector := credentials.GitAuthSelector{
		IntegrationID: config.GitRepository.IntegrationID,
	}
	if config.UsesGitSecret() {
		selector.SecretRef = &models.SecretReference{SecretID: *config.GitSecretID()}
	}

	resolved, sErr := s.credentialResolver.GitCredentials(ctx, config.OrganisationID, config.GitRepository.RepoURL, selector)
	if sErr != nil {
		return nil, sErr
	}
	return gitclient.NewGitClientForRepo(config.GitRepository.RepoURL, resolved.Credentials)
}
