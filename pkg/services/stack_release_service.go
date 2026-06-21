package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/clients"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stackrelease"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
)

type StackReleaseService interface {
	CreateRelease(ctx context.Context, stackID string, cause models.ReleaseCause) (*models.StackRelease, *errors.ServiceError)
	RollbackRelease(ctx context.Context, stackID, fromReleaseID string) (*models.StackRelease, *errors.ServiceError)
	GetRelease(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	ListReleases(ctx context.Context, stackID string, params stores.ListParams) (*stores.PaginatedResult[*models.StackRelease], *errors.ServiceError)
	CancelRelease(ctx context.Context, releaseID string) *errors.ServiceError

	// Internal methods are called by workers and controllers; no permission checks.
	InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError)
	InternalListActive(ctx context.Context) ([]*models.StackRelease, *errors.ServiceError)
	MarkInProgress(ctx context.Context, id string) (bool, *errors.ServiceError)
	SaveManifest(ctx context.Context, id string, m *models.ReleaseManifest, rev string, pins models.ReleasePins, rendererVersion string) (bool, *errors.ServiceError)
	MarkReleased(ctx context.Context, id string, outcome models.ReleaseOutcome) (bool, *errors.ServiceError)
	MarkCancelled(ctx context.Context, id string, reasons string) (bool, *errors.ServiceError)
	MarkSuperseded(ctx context.Context, id string, reason string) (bool, *errors.ServiceError)
	MarkFailed(ctx context.Context, id string, message string, outcome *models.ReleaseOutcome) (bool, *errors.ServiceError)
	AppendImageDigests(ctx context.Context, id string, digests map[string]string) *errors.ServiceError

	BackgroundJobEnqueuerInjectable
}

type StackReleaseServiceSpec struct {
	Store            stores.StackReleaseStore
	StackService     StackService
	SecretService    SecretService
	Permissions      auth.PermissionService
	ReferenceService ReferenceService
}

type stackReleaseService struct {
	store            stores.StackReleaseStore
	stackQuery       StackService
	secretService    SecretService
	permissions      auth.PermissionService
	referenceService ReferenceService
	BackgroundJobEnqueuerDep
}

func NewStackReleaseService(spec StackReleaseServiceSpec) StackReleaseService {
	return &stackReleaseService{
		store:            spec.Store,
		stackQuery:       spec.StackService,
		secretService:    spec.SecretService,
		permissions:      spec.Permissions,
		referenceService: spec.ReferenceService,
	}
}

func (s *stackReleaseService) CreateRelease(ctx context.Context, stackID string, cause models.ReleaseCause) (*models.StackRelease, *errors.ServiceError) {
	stack, sErr := s.stackQuery.GetStack(ctx, stackID)
	if sErr != nil {
		return nil, sErr
	}

	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	if stack.DeletionTimestamp != nil {
		return nil, errors.BadRequest("cannot create release for a stack that is being deleted")
	}

	snapshot, err := models.NewStackSnapshot(stack)
	if err != nil {
		return nil, errors.GeneralError("failed to create stack snapshot: %s", err.Error())
	}

	pins, pinErr := s.resolvePins(ctx, stack)
	if pinErr != nil {
		return nil, pinErr
	}
	applyPinsToSnapshot(&snapshot, pins)

	snapshotRev, err := stackrelease.HashJSON(snapshot)
	if err != nil {
		return nil, errors.GeneralError("failed to compute snapshot revision: %s", err.Error())
	}

	identity := auth.GetIdentityFromCtx(ctx)
	createdBy := identity.UserID

	release := &models.StackRelease{
		StackID:          stackID,
		State:            models.ReleaseStatePending,
		Cause:            cause,
		Snapshot:         snapshot,
		SnapshotRevision: snapshotRev,
		Pins:             pins,
		CreatedBy:        createdBy,
	}

	var created *models.StackRelease
	if txErr := s.store.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		var e *errors.ServiceError
		created, e = s.store.Create(txCtx, release)
		if e != nil {
			return e
		}
		return s.referenceService.ProjectRelease(txCtx, created)
	}); txErr != nil {
		return nil, txErr
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.StackRelease{ID: created.ID}); err != nil {
		return nil, errors.GeneralError("failed to enqueue release: %s", err.Error())
	}
	return created, nil
}

func (s *stackReleaseService) RollbackRelease(ctx context.Context, stackID, fromReleaseID string) (*models.StackRelease, *errors.ServiceError) {
	src, sErr := s.store.GetByID(ctx, fromReleaseID)
	if sErr != nil {
		return nil, sErr
	}

	if src.StackID != stackID {
		return nil, errors.NotFound("release '%s' does not belong to stack '%s'", fromReleaseID, stackID)
	}

	if src.State != models.ReleaseStateReleased {
		return nil, errors.BadRequest("can only roll back to a successfully released deployment (#%d is %s)", src.Sequence, src.State)
	}

	// Permission check via the stack's team.
	stack, sErr := s.stackQuery.GetStack(ctx, stackID)
	if sErr != nil {
		return nil, sErr
	}
	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	if stack.DeletionTimestamp != nil {
		return nil, errors.BadRequest("cannot roll back a stack that is being deleted")
	}

	identity := auth.GetIdentityFromCtx(ctx)
	createdBy := ""
	if identity != nil {
		createdBy = identity.UserID
	}

	now := time.Now().UTC()
	release := &models.StackRelease{
		StackID:          stackID,
		State:            models.ReleaseStatePending,
		Cause:            models.ReleaseCause{Kind: models.ReleaseCauseRollback, Detail: fmt.Sprintf("rollback to release #%d", src.Sequence)},
		Message:          fmt.Sprintf("rollback to release #%d", src.Sequence),
		Snapshot:         src.Snapshot,
		SnapshotRevision: src.SnapshotRevision,
		Manifest:         src.Manifest,
		ManifestRevision: src.ManifestRevision,
		Pins:             src.Pins,
		RendererVersion:  src.RendererVersion,
		RenderedAt:       &now,
		CreatedBy:        createdBy,
	}

	var created *models.StackRelease
	if txErr := s.store.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		var e *errors.ServiceError
		created, e = s.store.Create(txCtx, release)
		if e != nil {
			return e
		}
		return s.referenceService.ProjectRelease(txCtx, created)
	}); txErr != nil {
		return nil, txErr
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.StackRelease{ID: created.ID}); err != nil {
		return nil, errors.GeneralError("failed to enqueue release: %s", err.Error())
	}
	return created, nil
}

func (s *stackReleaseService) GetRelease(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError) {
	rel, sErr := s.store.GetByID(ctx, releaseID)
	if sErr != nil {
		return nil, sErr
	}

	stack, sErr := s.stackQuery.GetStack(ctx, rel.StackID)
	if sErr != nil {
		return nil, sErr
	}

	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, rel.StackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}

	return rel, nil
}

func (s *stackReleaseService) ListReleases(ctx context.Context, stackID string, params stores.ListParams) (*stores.PaginatedResult[*models.StackRelease], *errors.ServiceError) {
	stack, sErr := s.stackQuery.GetStack(ctx, stackID)
	if sErr != nil {
		return nil, sErr
	}

	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}

	return s.store.ListByStackID(ctx, stackID, params)
}

func (s *stackReleaseService) CancelRelease(ctx context.Context, releaseID string) *errors.ServiceError {
	rel, sErr := s.store.GetByID(ctx, releaseID)
	if sErr != nil {
		return sErr
	}

	stack, sErr := s.stackQuery.GetStack(ctx, rel.StackID)
	if sErr != nil {
		return sErr
	}

	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, rel.StackID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	if rel.State == models.ReleaseStateInProgress {
		return errors.BadRequest("cannot cancel release #%d: it is already in progress", rel.Sequence)
	}
	if rel.State.Terminal() {
		return errors.BadRequest("cannot cancel release #%d: it is already %s", rel.Sequence, rel.State)
	}

	won, sErr := s.store.Cancel(ctx, releaseID)
	if sErr != nil {
		return sErr
	}
	if !won {
		return errors.Conflict("release #%d is no longer pending (it may have already started processing)", rel.Sequence)
	}
	return nil
}

// --- Internal methods (no permission checks, called by workers/controllers) ---

func (s *stackReleaseService) InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError) {
	return s.store.GetByID(ctx, releaseID)
}

func (s *stackReleaseService) InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError) {
	return s.store.GetActiveByStackID(ctx, stackID)
}

func (s *stackReleaseService) InternalListActive(ctx context.Context) ([]*models.StackRelease, *errors.ServiceError) {
	return s.store.ListActive(ctx)
}

func (s *stackReleaseService) MarkInProgress(ctx context.Context, id string) (bool, *errors.ServiceError) {
	return s.store.MarkInProgress(ctx, id)
}

func (s *stackReleaseService) SaveManifest(ctx context.Context, id string, m *models.ReleaseManifest, rev string, pins models.ReleasePins, rendererVersion string) (bool, *errors.ServiceError) {
	return s.store.SaveManifest(ctx, id, m, rev, pins, rendererVersion)
}

func (s *stackReleaseService) MarkCancelled(ctx context.Context, id string, reason string) (bool, *errors.ServiceError) {
	return s.store.MarkCancelled(ctx, id, reason)
}

func (s *stackReleaseService) MarkSuperseded(ctx context.Context, id string, reason string) (bool, *errors.ServiceError) {
	return s.store.MarkSuperseded(ctx, id, reason)
}

func (s *stackReleaseService) MarkReleased(ctx context.Context, id string, outcome models.ReleaseOutcome) (bool, *errors.ServiceError) {
	return s.store.MarkReleased(ctx, id, outcome)
}

func (s *stackReleaseService) MarkFailed(ctx context.Context, id string, message string, outcome *models.ReleaseOutcome) (bool, *errors.ServiceError) {
	return s.store.MarkFailed(ctx, id, message, outcome)
}

func (s *stackReleaseService) AppendImageDigests(ctx context.Context, id string, digests map[string]string) *errors.ServiceError {
	return s.store.AppendImageDigests(ctx, id, digests)
}

func (s *stackReleaseService) resolvePins(ctx context.Context, stack *models.Stack) (models.ReleasePins, *errors.ServiceError) {
	pins := models.ReleasePins{
		Resources: make(map[string]models.ResourcePins),
	}
	for _, res := range stack.StackResources {
		if res.BuildConfig == nil {
			continue
		}

		rp, err := s.pinResource(ctx, res)
		if err != nil {
			return pins, err
		}
		if rp != nil {
			pins.Resources[res.Name] = *rp
		}
	}
	return pins, nil
}

func (s *stackReleaseService) pinResource(ctx context.Context, res *models.StackResource) (*models.ResourcePins, *errors.ServiceError) {
	var rp models.ResourcePins

	if rev := res.BuildConfig.SourceRevision.Git; rev != nil {
		sha, err := s.resolveGitSHA(ctx, res, rev)
		if err != nil {
			return nil, err
		}
		rp.GitSHA = sha
	}

	if res.BuildConfig.SourceRevision.Volume != nil {
		rp.VolumeHash = res.BuildConfig.SourceRevision.Volume.CurrentVolumeHash
	}

	return &rp, nil
}

func (s *stackReleaseService) resolveGitSHA(ctx context.Context, res *models.StackResource, rev *models.GitRevision) (string, *errors.ServiceError) {
	if rev.Commit != "" {
		return rev.Commit, nil
	}

	repoURL := res.BuildConfig.SourceContext.Git.RepoURL
	gitClient, err := s.gitClientForResource(ctx, res)
	if err != nil {
		return "", errors.GeneralError("resource '%s': %v", res.Name, err)
	}

	if rev.Tag != "" {
		sha, err := gitClient.GetTagSHA(ctx, repoURL, rev.Tag)
		if err != nil {
			return "", errors.GeneralError("resource '%s': cannot resolve tag '%s': %v", res.Name, rev.Tag, err)
		}
		return sha, nil
	}

	if rev.Branch != nil {
		if rev.Branch.HeadSha != "" && rev.Branch.HeadSha != "HEAD" {
			return rev.Branch.HeadSha, nil
		}
		result, err := gitClient.GetBranchHeadSHA(ctx, repoURL, rev.Branch.Name)
		if err != nil {
			return "", errors.GeneralError("resource '%s': cannot resolve branch '%s': %v", res.Name, rev.Branch.Name, err)
		}
		return result.HeadSHA, nil
	}

	return "", errors.GeneralError("resource '%s': git revision has no commit, tag, or branch", res.Name)
}

func (s *stackReleaseService) gitClientForResource(ctx context.Context, res *models.StackResource) (clients.GitClient, error) {
	gitSource := res.BuildConfig.SourceContext.Git
	if gitSource.GitSecretRef == nil {
		return clients.NewGitClientAnonymous()
	}

	secret, serr := s.secretService.InternalGetByID(ctx, gitSource.GitSecretRef.SecretID)
	if serr != nil {
		return nil, fmt.Errorf("git secret '%s': %w", gitSource.GitSecretRef.SecretID, serr)
	}

	if token, ok := secret.Data[models.TokenSecretKey]; ok {
		return clients.NewGitClientWithToken(token)
	}
	return clients.NewGitClient(secret.Data[models.UsernameSecretKey], secret.Data[models.PasswordSecretKey])
}

func applyPinsToSnapshot(snapshot *models.StackSnapshot, pins models.ReleasePins) {
	for _, resource := range snapshot.Resources {
		rp, ok := pins.Resources[resource.Name]
		if !ok {
			continue
		}
		if rp.GitSHA != "" &&
			resource.BuildConfig != nil &&
			resource.BuildConfig.SourceRevision.Git != nil &&
			resource.BuildConfig.SourceRevision.Git.Branch != nil {
			resource.BuildConfig.SourceRevision.Git.Branch.HeadSha = rp.GitSHA
		}
		if rp.GitSHA != "" && resource.BuildConfig.SourceRevision.Git != nil {
			resource.BuildConfig.SourceRevision.Git.Commit = rp.GitSHA
		}
	}
}
