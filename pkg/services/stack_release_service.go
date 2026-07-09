package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/auth"
	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stackrelease"
	"github.com/Stackdome/stackdome/pkg/stores"
)

//go:generate mockgen -source=stack_release_service.go -destination=stack_release_service_mock_test.go -package=services -exclude_interfaces StackReleaseService

// sourceGitClientProvider builds git clients so resolvePins's git resolution
// can be faked in tests.
type sourceGitClientProvider interface {
	ClientFor(repoURL string, creds gitclient.GitCredentials) (gitclient.GitClient, error)
}

type defaultSourceGitClientProvider struct{}

func (defaultSourceGitClientProvider) ClientFor(repoURL string, creds gitclient.GitCredentials) (gitclient.GitClient, error) {
	return gitclient.NewGitClientForRepo(repoURL, creds)
}

type StackReleaseService interface {
	CreateRelease(ctx context.Context, stackID string, cause models.ReleaseCause) (*models.StackRelease, *errors.ServiceError)
	RollbackRelease(ctx context.Context, stackID, fromReleaseID string) (*models.StackRelease, *errors.ServiceError)
	GetRelease(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	ListReleases(ctx context.Context, stackID string, params stores.ListParams) (*stores.PaginatedResult[*models.StackRelease], *errors.ServiceError)
	ListReleaseEvents(ctx context.Context, stackID, releaseID string, afterSequence, limit int) (*ReleaseEventPage, *errors.ServiceError)
	CancelRelease(ctx context.Context, releaseID string) *errors.ServiceError

	// Internal methods are called by workers and controllers; no permission checks.
	InternalCreateRelease(ctx context.Context, stackID string, cause models.ReleaseCause) (*models.StackRelease, *errors.ServiceError)
	InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError)
	InternalListActive(ctx context.Context) ([]*models.StackRelease, *errors.ServiceError)
	MarkInProgress(ctx context.Context, id string) (bool, *errors.ServiceError)
	SaveManifest(ctx context.Context, id string, m *models.ReleaseManifest, rev string, pins models.ReleasePins, rendererVersion string) (bool, *errors.ServiceError)
	MarkReleased(ctx context.Context, id string, outcome models.ReleaseOutcome) (bool, *errors.ServiceError)
	MarkCancelled(ctx context.Context, id string, reasons string) (bool, *errors.ServiceError)
	MarkSuperseded(ctx context.Context, id string, reason string) (bool, *errors.ServiceError)
	MarkFailed(ctx context.Context, id string, message string, outcome *models.ReleaseOutcome) (bool, *errors.ServiceError)
	MarkFailedWithValidationErrors(ctx context.Context, id, message string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError)
	AppendImageDigests(ctx context.Context, id string, digests map[string]string) *errors.ServiceError

	BackgroundJobEnqueuerInjectable
}

type StackReleaseServiceSpec struct {
	Store              stores.StackReleaseStore
	StackService       StackService
	CredentialResolver CredentialResolver
	Permissions        auth.PermissionService
	ReferenceService   ReferenceService
	EventStore         stores.ReleaseEventStore
	EventRecorder      ReleaseEventRecorder
	// GitClients is optional; it defaults to real git clients.
	GitClients sourceGitClientProvider
}

type stackReleaseService struct {
	store              stores.StackReleaseStore
	stackQuery         StackService
	credentialResolver CredentialResolver
	permissions        auth.PermissionService
	referenceService   ReferenceService
	eventStore         stores.ReleaseEventStore
	eventRecorder      ReleaseEventRecorder
	gitClients         sourceGitClientProvider
	logger             logger.Logger
	BackgroundJobEnqueuerDep
}

func NewStackReleaseService(spec StackReleaseServiceSpec) StackReleaseService {
	if spec.EventStore == nil {
		panic("StackReleaseService requires a ReleaseEventStore")
	}
	if spec.EventRecorder == nil {
		panic("StackReleaseService requires a ReleaseEventRecorder")
	}
	gitClients := spec.GitClients
	if gitClients == nil {
		gitClients = defaultSourceGitClientProvider{}
	}
	return &stackReleaseService{
		store:              spec.Store,
		stackQuery:         spec.StackService,
		credentialResolver: spec.CredentialResolver,
		permissions:        spec.Permissions,
		referenceService:   spec.ReferenceService,
		eventStore:         spec.EventStore,
		eventRecorder:      spec.EventRecorder,
		gitClients:         gitClients,
		logger:             logger.NewLoggerWithPrefix(context.Background(), "stack-release-service"),
	}
}

func (s *stackReleaseService) CreateRelease(ctx context.Context, stackID string, cause models.ReleaseCause) (*models.StackRelease, *errors.ServiceError) {
	stack, sErr := s.stackQuery.GetStack(ctx, stackID)
	if sErr != nil {
		return nil, sErr
	}

	if labels := stack.Labels.ToMap(); labels[models.PreviewStackLabel] == "true" {
		return nil, errors.BadRequest("cannot create releases on preview-managed stacks; use the preview sync API")
	}

	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	if stack.DeletionTimestamp != nil {
		return nil, errors.BadRequest("cannot create release for a stack that is being deleted")
	}

	identity := auth.GetIdentityFromCtx(ctx)
	return s.createReleaseForStack(ctx, stack, cause, identity.UserID)
}

func (s *stackReleaseService) InternalCreateRelease(ctx context.Context, stackID string, cause models.ReleaseCause) (*models.StackRelease, *errors.ServiceError) {
	stack, sErr := s.stackQuery.InternalGetStack(ctx, stackID)
	if sErr != nil {
		return nil, sErr
	}

	if stack.DeletionTimestamp != nil {
		return nil, errors.BadRequest("cannot create release for a stack that is being deleted")
	}

	return s.createReleaseForStack(ctx, stack, cause, models.ReleaseCreatedByPreviewSync)
}

func (s *stackReleaseService) createReleaseForStack(ctx context.Context, stack *models.Stack, cause models.ReleaseCause, createdBy string) (*models.StackRelease, *errors.ServiceError) {
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

	release := &models.StackRelease{
		StackID:          stack.ID,
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
		if e := s.referenceService.ProjectRelease(txCtx, created); e != nil {
			return e
		}
		return s.eventRecorder.RecordReleaseCreated(txCtx, created)
	}); txErr != nil {
		return nil, txErr
	}

	if err := s.BackgroundJobEnqueuer.EnqueueAfterCommit(ctx, &models.StackRelease{ID: created.ID}); err != nil {
		return nil, errors.GeneralError("failed to enqueue release: %s", err.Error())
	}
	return created, nil
}

func (s *stackReleaseService) RollbackRelease(ctx context.Context, stackID, fromReleaseID string) (
	*models.StackRelease, *errors.ServiceError) {
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

	if labels := stack.Labels.ToMap(); labels[models.PreviewStackLabel] == "true" {
		return nil, errors.BadRequest("cannot roll back a stack that is managed by preview sync")
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
		if e := s.referenceService.ProjectRelease(txCtx, created); e != nil {
			return e
		}
		return s.eventRecorder.RecordReleaseCreated(txCtx, created)
	}); txErr != nil {
		return nil, txErr
	}

	if err := s.BackgroundJobEnqueuer.EnqueueAfterCommit(ctx, &models.StackRelease{ID: created.ID}); err != nil {
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

const (
	releaseEventsDefaultLimit = 100
	releaseEventsMaxLimit     = 500
)

// releaseCancelledMessage is the user-facing message recorded on the
// release_cancelled terminal event.
const releaseCancelledMessage = "Release cancelled"

// ReleaseEventPage is a sequence-cursor page of release events.
type ReleaseEventPage struct {
	Events            []*models.ReleaseEvent
	NextAfterSequence int
}

func (s *stackReleaseService) ListReleaseEvents(ctx context.Context, stackID, releaseID string, afterSequence, limit int) (*ReleaseEventPage, *errors.ServiceError) {
	release, sErr := s.store.GetByID(ctx, releaseID)
	if sErr != nil {
		return nil, sErr
	}

	if release.StackID != stackID {
		return nil, errors.NotFound("release '%s' does not belong to stack '%s'", releaseID, stackID)
	}

	stack, sErr := s.stackQuery.GetStack(ctx, stackID)
	if sErr != nil {
		return nil, sErr
	}

	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}

	if limit <= 0 {
		limit = releaseEventsDefaultLimit
	}
	if limit > releaseEventsMaxLimit {
		limit = releaseEventsMaxLimit
	}

	events, sErr := s.eventStore.ListByReleaseID(ctx, releaseID, afterSequence, limit)
	if sErr != nil {
		return nil, sErr
	}

	next := afterSequence
	if len(events) > 0 {
		next = events[len(events)-1].Sequence
	}

	return &ReleaseEventPage{Events: events, NextAfterSequence: next}, nil
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
	// The CAS win already persisted the cancellation; recording is best-effort.
	rel.State = models.ReleaseStateCancelled
	if recErr := s.eventRecorder.RecordReleaseTerminal(ctx, rel, models.ReleaseStateCancelled, releaseCancelledMessage); recErr != nil {
		s.logger.Errorf("release %s: failed to record release_cancelled event: %v", rel.ID, recErr)
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

func (s *stackReleaseService) MarkFailedWithValidationErrors(ctx context.Context, id, message string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError) {
	return s.store.MarkFailedWithValidationErrors(ctx, id, message, verrs)
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

		rp, err := s.pinResource(ctx, stack.OrganisationID, res)
		if err != nil {
			return pins, err
		}
		if rp != nil {
			pins.Resources[res.Name] = *rp
		}
	}
	return pins, nil
}

func (s *stackReleaseService) pinResource(ctx context.Context, orgID string, res *models.StackResource) (*models.ResourcePins, *errors.ServiceError) {
	var rp models.ResourcePins

	if rev := res.BuildConfig.SourceRevision.Git; rev != nil {
		sha, branch, err := s.resolveGitSHA(ctx, orgID, res, rev)
		if err != nil {
			return nil, err
		}
		rp.GitSHA = sha
		rp.Branch = branch
	}

	if res.BuildConfig.SourceRevision.Volume != nil {
		if res.BuildConfig.SourceRevision.Volume.CurrentVolumeHash == "" {
			return nil, errors.ValidationFailed([]errors.FieldError{{
				Field:   fmt.Sprintf("resources[%s].source.volume.current_volume_hash", res.Name),
				Code:    errors.VErrVolumeHashMissing,
				Message: fmt.Sprintf("resource '%s': current volume revision is empty", res.Name),
			}})
		}
		rp.VolumeHash = res.BuildConfig.SourceRevision.Volume.CurrentVolumeHash
	}

	return &rp, nil
}

// resolveGitSHA resolves a git revision to a commit SHA. When the revision
// specifies neither branch nor tag, it also resolves and returns the
// repository's default branch so the pin (and the snapshot) can record it —
// the CRD requires branch-or-tag on git revisions.
func (s *stackReleaseService) resolveGitSHA(ctx context.Context, orgID string, res *models.StackResource, rev *models.GitRevision) (string, string, *errors.ServiceError) {
	field := fmt.Sprintf("resources[%s].source.git", res.Name)

	if rev.Commit != "" {
		if rev.Branch == "" && rev.Tag == "" {
			return "", "", errors.ValidationFailed([]errors.FieldError{{
				Field:   field,
				Code:    errors.VErrGitCommitRequiresRef,
				Message: fmt.Sprintf("resource '%s': a commit pin requires a branch or tag (the cluster needs a fetchable ref)", res.Name),
			}})
		}
		return rev.Commit, "", nil
	}

	gitSource := res.BuildConfig.SourceContext.Git
	repoURL := gitSource.RepoURL

	resolved, serr := s.credentialResolver.GitCredentials(ctx, orgID, repoURL, credentials.GitAuthSelector{
		// This can be empty if not explicitly set in the stack resource. In that case, the git integration will be used based on the host.
		IntegrationID: gitSource.IntegrationID,
	})
	if serr != nil {
		if serr.Is404() {
			return "", "", errors.ValidationFailed([]errors.FieldError{{
				Field:   field + ".integration_id",
				Code:    errors.VErrGitIntegrationNotFound,
				Message: fmt.Sprintf("resource '%s': failed to resolve git credentials: %v", res.Name, serr),
			}})
		}
		return "", "", serr
	}
	gitClient, err := s.gitClients.ClientFor(repoURL, resolved.Credentials)
	if err != nil {
		return "", "", errors.ValidationFailed([]errors.FieldError{{
			Field:   field + ".repo_url",
			Code:    errors.VErrGitRepoUnreachable,
			Message: fmt.Sprintf("resource '%s': %v", res.Name, err),
		}})
	}

	if rev.Tag != "" {
		sha, err := gitClient.GetTagSHA(ctx, repoURL, rev.Tag)
		if err != nil {
			return "", "", gitFieldError(res.Name, field+".tag", errors.VErrGitTagNotFound,
				fmt.Sprintf("cannot resolve tag '%s'", rev.Tag), err)
		}
		return sha, "", nil
	}

	branch := rev.Branch
	resolvedDefault := ""
	if branch == "" {
		// No branch or tag: pin the repository's default branch at release time.
		defaultBranch, err := gitClient.GetDefaultBranch(ctx, repoURL)
		if err != nil {
			return "", "", gitFieldError(res.Name, field+".repo_url", errors.VErrGitRepoUnreachable,
				"cannot resolve default branch", err)
		}
		branch = defaultBranch
		resolvedDefault = defaultBranch
	}

	result, err := gitClient.GetBranchHeadSHA(ctx, repoURL, branch)
	if err != nil {
		return "", "", gitFieldError(res.Name, field+".branch", errors.VErrGitBranchNotFound,
			fmt.Sprintf("cannot resolve branch '%s'", branch), err)
	}
	return result.HeadSHA, resolvedDefault, nil
}

// gitFieldError maps a git client error onto a structured release validation
// failure, distinguishing auth and rate-limit failures from missing refs.
func gitFieldError(resourceName, field, notFoundCode string, action string, err error) *errors.ServiceError {
	code := notFoundCode
	switch {
	case stderrors.Is(err, gitclient.ErrAuthFailed):
		code = errors.VErrGitAuthFailed
	case stderrors.Is(err, gitclient.ErrRateLimited):
		code = errors.VErrGitRateLimited
	}
	return errors.ValidationFailed([]errors.FieldError{{
		Field:   field,
		Code:    code,
		Message: fmt.Sprintf("resource '%s': %s: %v", resourceName, action, err),
	}})
}

func applyPinsToSnapshot(snapshot *models.StackSnapshot, pins models.ReleasePins) {
	for _, resource := range snapshot.Resources {
		rp, ok := pins.Resources[resource.Name]
		if !ok {
			continue
		}
		if rp.GitSHA != "" &&
			resource.BuildConfig != nil &&
			resource.BuildConfig.SourceRevision.Git != nil {
			resource.BuildConfig.SourceRevision.Git.Commit = rp.GitSHA
			if rp.Branch != "" && resource.BuildConfig.SourceRevision.Git.Branch == "" {
				resource.BuildConfig.SourceRevision.Git.Branch = rp.Branch
			}
		}
	}
}
