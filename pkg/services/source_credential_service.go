package services

import (
	"context"
	stderrors "errors"

	gitclient "github.com/ashishmax31/stackdome-api-server/pkg/clients/git"
	"github.com/ashishmax31/stackdome-api-server/pkg/credentials"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

//go:generate mockgen -destination=../mocks/mock_source_credential_service.go -package=mocks github.com/ashishmax31/stackdome-api-server/pkg/services SourceCredentialService

// SourceRollback reverts the managed-secret mutations performed by a reversible
// prepare. The type is declared in pkg/credentials to keep generated mocks free
// of an import cycle; see credentials.SourceRollback for semantics.
type SourceRollback = credentials.SourceRollback

// SourceCredentialService materializes inline source credentials into managed
// secrets, wires the internal refs, resolves default branches, and garbage
// collects managed secrets when their slot disappears.
type SourceCredentialService interface {
	// PrepareStackSources runs PrepareResourceSource for every resource. The
	// stack ID must already be set (pre-generated for creates). It returns a
	// rollback handle so the caller can undo materialized/overwritten managed
	// secrets if a later step (e.g. validation) fails before the changes are
	// persisted. If preparation itself fails, any partial work is rolled back
	// before returning and the handle is nil.
	PrepareStackSources(ctx context.Context, stack *models.Stack) (credentials.SourceRollback, *errors.ServiceError)
	PrepareResourceSource(ctx context.Context, stack *models.Stack, resource *models.StackResource) *errors.ServiceError
	// CleanupResourceSources removes all managed secrets owned by a stack
	// resource. Called when the resource is deleted.
	CleanupResourceSources(ctx context.Context, stackID, resourceName string) *errors.ServiceError
	// CleanupStackSources removes managed secrets for every resource in a
	// stack. Called when the stack is hard-deleted.
	CleanupStackSources(ctx context.Context, stack *models.Stack) *errors.ServiceError
	// PreparePreviewConfig materializes the preview config's inline credentials
	// and returns a rollback handle with the same semantics as
	// PrepareStackSources.
	PreparePreviewConfig(ctx context.Context, config *models.StackPreviewConfig) (credentials.SourceRollback, *errors.ServiceError)
	CleanupPreviewConfig(ctx context.Context, configID string) *errors.ServiceError
}

//go:generate mockgen -source=source_credential_service.go -destination=source_credential_service_mock_test.go -package=services -exclude_interfaces SourceCredentialService
type managedSecretService interface {
	CreateManaged(ctx context.Context, spec ManagedSecretSpec) (*models.Secret, *errors.ServiceError)
	GetManagedFor(ctx context.Context, ownerKind, ownerID string, slot models.ManagedSecretSlot) (*models.Secret, *errors.ServiceError)
	DeleteManagedFor(ctx context.Context, ownerKind, ownerID string, slots ...models.ManagedSecretSlot) *errors.ServiceError
	// RestoreManaged puts a previously captured managed secret snapshot back
	// exactly as it was (same ID and encrypted data), re-creating it if it was
	// deleted. Used only by rollback.
	RestoreManaged(ctx context.Context, snapshot *models.Secret) *errors.ServiceError
}

type orgRegistryCredentialCreator interface {
	Create(ctx context.Context, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError)
}

type orgGitIntegrationCreator interface {
	Create(ctx context.Context, integration *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError)
}

// sourceGitClientProvider builds git clients so default-branch resolution can
// be faked in tests.
type sourceGitClientProvider interface {
	ClientFor(repoURL string, creds gitclient.GitCredentials) (gitclient.GitClient, error)
}

type defaultSourceGitClientProvider struct{}

func (defaultSourceGitClientProvider) ClientFor(repoURL string, creds gitclient.GitCredentials) (gitclient.GitClient, error) {
	return gitclient.NewGitClientForRepo(repoURL, creds)
}

type SourceCredentialServiceSpec struct {
	SecretService             managedSecretService
	RegistryCredentialService orgRegistryCredentialCreator
	GitIntegrationService     orgGitIntegrationCreator
	CredentialResolver        CredentialResolver
	Logger                    logger.Logger
	// GitClients is optional; it defaults to real git clients.
	GitClients sourceGitClientProvider
}

type sourceCredentialService struct {
	secrets             managedSecretService
	registryCredentials orgRegistryCredentialCreator
	gitIntegrations     orgGitIntegrationCreator
	resolver            CredentialResolver
	gitClients          sourceGitClientProvider
	logger              logger.Logger
}

func NewSourceCredentialService(spec SourceCredentialServiceSpec) SourceCredentialService {
	gitClients := spec.GitClients
	if gitClients == nil {
		gitClients = defaultSourceGitClientProvider{}
	}
	return &sourceCredentialService{
		secrets:             spec.SecretService,
		registryCredentials: spec.RegistryCredentialService,
		gitIntegrations:     spec.GitIntegrationService,
		resolver:            spec.CredentialResolver,
		gitClients:          gitClients,
		logger:              spec.Logger,
	}
}

func (s *sourceCredentialService) PrepareStackSources(ctx context.Context, stack *models.Stack) (SourceRollback, *errors.ServiceError) {
	rec := newManagedSecretRecorder(s.secrets, s.logger)
	clone := *s
	clone.secrets = rec
	for _, resource := range stack.StackResources {
		if err := clone.PrepareResourceSource(ctx, stack, resource); err != nil {
			// Undo any managed secrets already materialized/overwritten before
			// the failing resource so a partial prepare leaves nothing behind.
			rec.Rollback(ctx)
			return nil, err
		}
	}
	return rec, nil
}

func (s *sourceCredentialService) PrepareResourceSource(ctx context.Context, stack *models.Stack, resource *models.StackResource) *errors.ServiceError {
	ownerID := models.StackResourceManagedOwnerID(stack.ID, resource.Name)

	if err := s.prepareGitSlot(ctx, stack, resource, ownerID); err != nil {
		return err
	}
	if err := s.preparePullSlot(ctx, stack, resource, ownerID); err != nil {
		return err
	}
	if err := s.preparePushSlot(ctx, stack, resource, ownerID); err != nil {
		return err
	}
	return s.resolveDefaultBranch(ctx, stack, resource)
}

func (s *sourceCredentialService) prepareGitSlot(ctx context.Context, stack *models.Stack, resource *models.StackResource, ownerID string) *errors.ServiceError {
	if resource.BuildConfig == nil || resource.BuildConfig.SourceContext.Git == nil {
		return s.dropSlot(ctx, ownerID, models.ManagedSecretSlotGit)
	}

	git := resource.BuildConfig.SourceContext.Git
	if git.InlineCredentials != nil {
		inline := git.InlineCredentials
		if inline.SaveToOrg {
			if err := s.promoteToOrgGitIntegration(ctx, stack.OrganisationID, git.RepoURL, inline); err != nil {
				return err.WithPrefix("resource '%s'", resource.Name)
			}
		}
		secret, serr := s.secrets.CreateManaged(ctx, ManagedSecretSpec{
			OrganisationID: stack.OrganisationID,
			TeamID:         stack.TeamID,
			UserID:         stack.UserID,
			OwnerKind:      models.ManagedByKindStackResource,
			OwnerID:        ownerID,
			Slot:           models.ManagedSecretSlotGit,
			Name:           models.ManagedSecretName(stack.Name, resource.Name, models.ManagedSecretSlotGit),
			Type:           models.SecretTypeGitCredentials,
			Data: map[string]string{
				models.UsernameSecretKey: inline.Username,
				models.PasswordSecretKey: inline.Password,
			},
		})
		if serr != nil {
			return serr.WithPrefix("resource '%s': failed to store git credentials", resource.Name)
		}
		git.GitSecretRef = &models.SecretReference{SecretID: secret.ID}
		git.InlineCredentials = nil
		return nil
	}

	if git.IntegrationID != "" {
		// Slot satisfied by an integration; any leftover managed secret is stale.
		return s.dropSlot(ctx, ownerID, models.ManagedSecretSlotGit)
	}

	if git.GitSecretRef == nil {
		// Keep previously materialized credentials until the slot goes away or
		// new credentials replace them — writeOnly fields cannot be re-sent.
		existing, serr := s.secrets.GetManagedFor(ctx, models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit)
		if serr != nil {
			if serr.Is404() {
				return nil
			}
			return serr
		}
		git.GitSecretRef = &models.SecretReference{SecretID: existing.ID}
	}
	return nil
}

func (s *sourceCredentialService) preparePullSlot(ctx context.Context, stack *models.Stack, resource *models.StackResource, ownerID string) *errors.ServiceError {
	if resource.ImageConfig == nil {
		return s.dropSlot(ctx, ownerID, models.ManagedSecretSlotPull)
	}

	img := resource.ImageConfig
	if img.InlineCredentials != nil {
		inline := img.InlineCredentials
		if inline.SaveToOrg {
			if err := s.promoteToOrgRegistryCredential(ctx, stack, img.Image, models.RegistryCredentialPurposePull, inline); err != nil {
				return err.WithPrefix("resource '%s'", resource.Name)
			}
		}
		secret, serr := s.secrets.CreateManaged(ctx, ManagedSecretSpec{
			OrganisationID: stack.OrganisationID,
			TeamID:         stack.TeamID,
			UserID:         stack.UserID,
			OwnerKind:      models.ManagedByKindStackResource,
			OwnerID:        ownerID,
			Slot:           models.ManagedSecretSlotPull,
			Name:           models.ManagedSecretName(stack.Name, resource.Name, models.ManagedSecretSlotPull),
			Type:           models.SecretTypeDockerRegistry,
			Data: map[string]string{
				models.UsernameSecretKey: inline.Username,
				models.PasswordSecretKey: inline.Password,
			},
		})
		if serr != nil {
			return serr.WithPrefix("resource '%s': failed to store pull credentials", resource.Name)
		}
		img.PullSecretRef = &models.SecretReference{SecretID: secret.ID}
		img.InlineCredentials = nil
		return nil
	}

	if img.RegistryCredentialID != "" {
		return s.dropSlot(ctx, ownerID, models.ManagedSecretSlotPull)
	}

	if img.PullSecretRef == nil {
		existing, serr := s.secrets.GetManagedFor(ctx, models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPull)
		if serr != nil {
			if serr.Is404() {
				return nil
			}
			return serr
		}
		img.PullSecretRef = &models.SecretReference{SecretID: existing.ID}
	}
	return nil
}

func (s *sourceCredentialService) preparePushSlot(ctx context.Context, stack *models.Stack, resource *models.StackResource, ownerID string) *errors.ServiceError {
	hasPushTarget := resource.BuildConfig != nil &&
		!resource.BuildConfig.BuildImageRepository.UseInClusterRegistry &&
		resource.BuildConfig.BuildImageRepository.ExternalImageRef != ""
	if !hasPushTarget {
		return s.dropSlot(ctx, ownerID, models.ManagedSecretSlotPush)
	}

	bc := resource.BuildConfig
	if bc.PushInlineCredentials != nil {
		inline := bc.PushInlineCredentials
		if inline.SaveToOrg {
			if err := s.promoteToOrgRegistryCredential(ctx, stack, bc.BuildImageRepository.ExternalImageRef, models.RegistryCredentialPurposePush, inline); err != nil {
				return err.WithPrefix("resource '%s'", resource.Name)
			}
		}
		secret, serr := s.secrets.CreateManaged(ctx, ManagedSecretSpec{
			OrganisationID: stack.OrganisationID,
			TeamID:         stack.TeamID,
			UserID:         stack.UserID,
			OwnerKind:      models.ManagedByKindStackResource,
			OwnerID:        ownerID,
			Slot:           models.ManagedSecretSlotPush,
			Name:           models.ManagedSecretName(stack.Name, resource.Name, models.ManagedSecretSlotPush),
			Type:           models.SecretTypeDockerRegistry,
			Data: map[string]string{
				models.UsernameSecretKey: inline.Username,
				models.PasswordSecretKey: inline.Password,
			},
		})
		if serr != nil {
			return serr.WithPrefix("resource '%s': failed to store push credentials", resource.Name)
		}
		bc.RegistrySecretRef = &models.SecretReference{SecretID: secret.ID}
		bc.PushInlineCredentials = nil
		return nil
	}

	if bc.PushRegistryCredentialID != "" {
		return s.dropSlot(ctx, ownerID, models.ManagedSecretSlotPush)
	}

	if bc.RegistrySecretRef == nil {
		existing, serr := s.secrets.GetManagedFor(ctx, models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPush)
		if serr != nil {
			if serr.Is404() {
				return nil
			}
			return serr
		}
		bc.RegistrySecretRef = &models.SecretReference{SecretID: existing.ID}
	}
	return nil
}

// promoteToOrgRegistryCredential handles save_to_org for registry slots.
func (s *sourceCredentialService) promoteToOrgRegistryCredential(ctx context.Context, stack *models.Stack, ref string, purpose models.RegistryCredentialPurpose, inline *models.InlineCredentials) *errors.ServiceError {
	if s.registryCredentials == nil {
		return errors.GeneralError("registry credential service is not configured")
	}
	_, serr := s.registryCredentials.Create(ctx, &models.RegistryCredential{
		OrganisationID: stack.OrganisationID,
		Host:           ref,
		Purpose:        purpose,
		Username:       inline.Username,
		Password:       inline.Password,
	})
	if serr != nil {
		return serr.WithPrefix("failed to save credentials to the organisation")
	}
	return nil
}

// resolveDefaultBranch fills in the repository's default branch when a git
// source specifies neither branch nor tag. The result is stored on the
// resource so releases and pins stay reproducible.
func (s *sourceCredentialService) resolveDefaultBranch(ctx context.Context, stack *models.Stack, resource *models.StackResource) *errors.ServiceError {
	if resource.BuildConfig == nil || resource.BuildConfig.SourceContext.Git == nil {
		return nil
	}
	rev := resource.BuildConfig.SourceRevision.Git
	if rev == nil || rev.Branch != "" || rev.Tag != "" {
		return nil
	}

	git := resource.BuildConfig.SourceContext.Git
	resolved, serr := s.resolver.GitCredentials(ctx, stack.OrganisationID, git.RepoURL, credentials.GitAuthSelector{
		SecretRef:     git.GitSecretRef,
		IntegrationID: git.IntegrationID,
	})
	if serr != nil {
		return serr.WithPrefix("resource '%s': failed to resolve git credentials", resource.Name)
	}

	client, err := s.gitClients.ClientFor(git.RepoURL, resolved.Credentials)
	if err != nil {
		return errors.GeneralError("resource '%s': failed to create git client: %s", resource.Name, err.Error())
	}

	branch, err := client.GetDefaultBranch(ctx, git.RepoURL)
	if err != nil {
		target := errors.CredentialErrorTarget{
			Kind: errors.CredentialTargetKindGitClone,
			Host: gitRepoHost(git.RepoURL),
			Ref:  git.RepoURL,
		}
		if stderrors.Is(err, gitclient.ErrAuthFailed) {
			if resolved.Source == credentials.SourceAnonymous {
				return errors.CredentialsRequired(target, "resource '%s': repository '%s' requires credentials to resolve its default branch", resource.Name, git.RepoURL)
			}
			return errors.CredentialsInvalid(target, "resource '%s': configured git credentials were rejected by '%s'", resource.Name, git.RepoURL)
		}
		return errors.BadRequest("resource '%s': failed to resolve default branch for '%s': %s", resource.Name, git.RepoURL, err.Error())
	}

	rev.Branch = branch
	return nil
}

func (s *sourceCredentialService) dropSlot(ctx context.Context, ownerID string, slot models.ManagedSecretSlot) *errors.ServiceError {
	return s.secrets.DeleteManagedFor(ctx, models.ManagedByKindStackResource, ownerID, slot)
}

// promoteToOrgGitIntegration handles save_to_org for git slots.
func (s *sourceCredentialService) promoteToOrgGitIntegration(ctx context.Context, orgID, repoURL string, inline *models.InlineCredentials) *errors.ServiceError {
	if s.gitIntegrations == nil {
		return errors.GeneralError("git integration service is not configured")
	}
	_, serr := s.gitIntegrations.Create(ctx, &models.GitIntegration{
		OrganisationID: orgID,
		Type:           models.GitIntegrationTypeGitCredentials,
		Host:           gitclient.RepoHost(repoURL),
		Auth: &models.GitIntegrationAuth{
			Basic: &models.GitIntegrationBasicAuth{
				Username: inline.Username,
				Password: inline.Password,
			},
		},
	})
	if serr != nil {
		return serr.WithPrefix("failed to save credentials to the organisation")
	}
	return nil
}

func (s *sourceCredentialService) CleanupResourceSources(ctx context.Context, stackID, resourceName string) *errors.ServiceError {
	return s.secrets.DeleteManagedFor(ctx, models.ManagedByKindStackResource, models.StackResourceManagedOwnerID(stackID, resourceName))
}

// CleanupStackSources removes managed secrets for every resource in a stack.
// Called when the stack is hard-deleted.
func (s *sourceCredentialService) CleanupStackSources(ctx context.Context, stack *models.Stack) *errors.ServiceError {
	for _, resource := range stack.StackResources {
		if err := s.CleanupResourceSources(ctx, stack.ID, resource.Name); err != nil {
			return err
		}
	}
	return nil
}

func (s *sourceCredentialService) PreparePreviewConfig(ctx context.Context, config *models.StackPreviewConfig) (SourceRollback, *errors.ServiceError) {
	rec := newManagedSecretRecorder(s.secrets, s.logger)
	clone := *s
	clone.secrets = rec
	if err := clone.preparePreviewConfig(ctx, config); err != nil {
		rec.Rollback(ctx)
		return nil, err
	}
	return rec, nil
}

func (s *sourceCredentialService) preparePreviewConfig(ctx context.Context, config *models.StackPreviewConfig) *errors.ServiceError {
	repo := &config.GitRepository
	if repo.InlineCredentials != nil {
		inline := repo.InlineCredentials
		if inline.SaveToOrg {
			if err := s.promoteToOrgGitIntegration(ctx, config.OrganisationID, repo.RepoURL, inline); err != nil {
				return err
			}
		}
		secret, serr := s.secrets.CreateManaged(ctx, ManagedSecretSpec{
			OrganisationID: config.OrganisationID,
			TeamID:         config.TeamID,
			UserID:         config.UserID,
			OwnerKind:      models.ManagedByKindPreviewConfig,
			OwnerID:        config.ID,
			Slot:           models.ManagedSecretSlotGit,
			Name:           models.ManagedSecretName(config.Name, models.ManagedByKindPreviewConfig, models.ManagedSecretSlotGit),
			Type:           models.SecretTypeGitCredentials,
			Data: map[string]string{
				models.UsernameSecretKey: inline.Username,
				models.PasswordSecretKey: inline.Password,
			},
		})
		if serr != nil {
			return serr.WithPrefix("failed to store preview git credentials")
		}
		repo.GitSecretID = &secret.ID
		repo.InlineCredentials = nil
		return nil
	}

	if repo.IntegrationID != "" {
		return s.secrets.DeleteManagedFor(ctx, models.ManagedByKindPreviewConfig, config.ID, models.ManagedSecretSlotGit)
	}

	if repo.GitSecretID == nil || *repo.GitSecretID == "" {
		existing, serr := s.secrets.GetManagedFor(ctx, models.ManagedByKindPreviewConfig, config.ID, models.ManagedSecretSlotGit)
		if serr != nil {
			if serr.Is404() {
				return nil
			}
			return serr
		}
		repo.GitSecretID = &existing.ID
	}
	return nil
}

func (s *sourceCredentialService) CleanupPreviewConfig(ctx context.Context, configID string) *errors.ServiceError {
	return s.secrets.DeleteManagedFor(ctx, models.ManagedByKindPreviewConfig, configID)
}

func gitRepoHost(repoURL string) string {
	return gitclient.RepoHost(repoURL)
}

// managedSecretUndo records how to revert a single managed-secret mutation.
// When snapshot is set, the prior secret is restored; otherwise the newly
// created secret at owner/slot is deleted.
type managedSecretUndo struct {
	snapshot  *models.Secret
	ownerKind string
	ownerID   string
	slot      models.ManagedSecretSlot
}

// managedSecretRecorder wraps a managedSecretService, capturing enough state
// before each mutation to undo it. It implements managedSecretService so a
// prepare flow can run unchanged against the recorder, then Rollback reverses
// every mutation in LIFO order.
type managedSecretRecorder struct {
	inner  managedSecretService
	logger logger.Logger
	undos  []managedSecretUndo
}

func newManagedSecretRecorder(inner managedSecretService, log logger.Logger) *managedSecretRecorder {
	return &managedSecretRecorder{inner: inner, logger: log}
}

func (r *managedSecretRecorder) CreateManaged(ctx context.Context, spec ManagedSecretSpec) (*models.Secret, *errors.ServiceError) {
	// Snapshot the prior secret (if any) before the upsert overwrites it.
	prior, gerr := r.inner.GetManagedFor(ctx, spec.OwnerKind, spec.OwnerID, spec.Slot)
	if gerr != nil && !gerr.Is404() {
		return nil, gerr
	}
	created, cerr := r.inner.CreateManaged(ctx, spec)
	if cerr != nil {
		return nil, cerr
	}
	if gerr == nil && prior != nil {
		r.undos = append(r.undos, managedSecretUndo{snapshot: prior})
	} else {
		r.undos = append(r.undos, managedSecretUndo{
			ownerKind: spec.OwnerKind,
			ownerID:   spec.OwnerID,
			slot:      spec.Slot,
		})
	}
	return created, nil
}

func (r *managedSecretRecorder) DeleteManagedFor(ctx context.Context, ownerKind, ownerID string, slots ...models.ManagedSecretSlot) *errors.ServiceError {
	// Snapshot each targeted slot so a rollback can put it back. Only the
	// slot-scoped form is used during prepare (dropSlot); the delete-all form
	// is left un-snapshotted since it never runs inside a reversible prepare.
	for _, slot := range slots {
		existing, gerr := r.inner.GetManagedFor(ctx, ownerKind, ownerID, slot)
		if gerr != nil {
			if gerr.Is404() {
				continue
			}
			return gerr
		}
		r.undos = append(r.undos, managedSecretUndo{snapshot: existing})
	}
	return r.inner.DeleteManagedFor(ctx, ownerKind, ownerID, slots...)
}

func (r *managedSecretRecorder) GetManagedFor(ctx context.Context, ownerKind, ownerID string, slot models.ManagedSecretSlot) (*models.Secret, *errors.ServiceError) {
	return r.inner.GetManagedFor(ctx, ownerKind, ownerID, slot)
}

func (r *managedSecretRecorder) RestoreManaged(ctx context.Context, snapshot *models.Secret) *errors.ServiceError {
	return r.inner.RestoreManaged(ctx, snapshot)
}

func (r *managedSecretRecorder) Rollback(ctx context.Context) {
	for i := len(r.undos) - 1; i >= 0; i-- {
		undo := r.undos[i]
		if undo.snapshot != nil {
			if err := r.inner.RestoreManaged(ctx, undo.snapshot); err != nil {
				r.logErrorf("failed to restore managed secret id=%s owner=%s/%s slot=%s: %v",
					undo.snapshot.ID, undo.snapshot.ManagedByKind, undo.snapshot.ManagedByID, undo.snapshot.ManagedSlot, err)
			}
			continue
		}
		if err := r.inner.DeleteManagedFor(ctx, undo.ownerKind, undo.ownerID, undo.slot); err != nil {
			r.logErrorf("failed to delete orphaned managed secret owner=%s/%s slot=%s: %v",
				undo.ownerKind, undo.ownerID, undo.slot, err)
		}
	}
	r.undos = nil
}

func (r *managedSecretRecorder) logErrorf(format string, args ...any) {
	if r.logger != nil {
		r.logger.Errorf(format, args...)
	}
}
