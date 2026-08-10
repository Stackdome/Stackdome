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
	"github.com/Stackdome/stackdome/pkg/interfaces"
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

func selectAuthoritativeWorkloadRelease(latest, converged *models.StackRelease) *models.StackRelease {
	if latest != nil {
		switch latest.State {
		case models.ReleaseStatePending, models.ReleaseStateInProgress:
			return latest
		}
	}
	return converged
}

func resolveAuthoritativeWorkloadRelease(ctx context.Context, store stores.StackReleaseStore, stack *models.Stack) (*models.StackRelease, *errors.ServiceError) {
	latest, serr := store.GetLatestByStackID(ctx, stack.ID)
	if serr != nil {
		return nil, serr
	}

	var converged *models.StackRelease
	convergedID := stack.GetConvergedReleaseID()
	if latest == nil || !latest.State.Active() {
		if convergedID != "" {
			converged, serr = store.GetByID(ctx, convergedID)
			if serr != nil {
				return nil, serr
			}
		}
	}

	selected := selectAuthoritativeWorkloadRelease(latest, converged)
	if selected == nil || !selected.IsAuthoritativeWorkloadRelease(stack, latest) {
		return nil, nil
	}
	return selected, nil
}

type StackReleaseService interface {
	CreateRelease(ctx context.Context, stackID string, cause models.ReleaseCause) (*models.StackRelease, *errors.ServiceError)
	RollbackRelease(ctx context.Context, stackID, fromReleaseID string) (*models.StackRelease, *errors.ServiceError)
	GetRelease(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	GetReleaseDetail(ctx context.Context, releaseID string) (*models.StackRelease, *models.ReleaseLiveStatus, *errors.ServiceError)
	ListReleases(ctx context.Context, stackID string, params stores.ListParams) (*stores.PaginatedResult[*models.StackRelease], *errors.ServiceError)
	ListReleaseEvents(ctx context.Context, stackID, releaseID string, afterSequence, limit int) (*ReleaseEventPage, *errors.ServiceError)
	StreamReleaseEvents(ctx context.Context, stackID, releaseID string, afterSequence int) (interfaces.ServerSideStreamable, *errors.ServiceError)
	CancelRelease(ctx context.Context, releaseID string) *errors.ServiceError

	// Internal methods are called by workers and controllers; no permission checks.
	InternalCreateRelease(ctx context.Context, stackID string, cause models.ReleaseCause) (*models.StackRelease, *errors.ServiceError)
	InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	InternalResolveAuthoritativeWorkloadRelease(ctx context.Context, stack *models.Stack) (*models.StackRelease, *errors.ServiceError)
	InternalListAuthoritativeWorkload(ctx context.Context) (*models.WorkloadAuthorityScan, *errors.ServiceError)
	InternalGetReleaseRefs(ctx context.Context, stacks []*models.Stack) (map[string]models.StackReleaseRefs, *errors.ServiceError)
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError)
	InternalGetLatestByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError)
	InternalListActive(ctx context.Context) ([]*models.StackRelease, *errors.ServiceError)
	MarkInProgress(ctx context.Context, id string) (bool, *errors.ServiceError)
	SaveManifest(ctx context.Context, id string, m *models.ReleaseManifest, rev string, pins models.ReleasePins, rendererVersion string) (bool, *errors.ServiceError)
	MarkReleased(ctx context.Context, id string, outcome models.ReleaseOutcome) (bool, *errors.ServiceError)
	MarkCancelled(ctx context.Context, id string, reasons string) (bool, *errors.ServiceError)
	MarkSuperseded(ctx context.Context, id string, reason string) (bool, *errors.ServiceError)
	MarkFailed(ctx context.Context, id string, message string, outcome *models.ReleaseOutcome) (bool, *errors.ServiceError)
	SetConvergeClockStartedAt(ctx context.Context, id string, startedAt *time.Time) *errors.ServiceError
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
	GitClients    sourceGitClientProvider
	RuntimePolicy RuntimePolicy
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
	runtimePolicy      RuntimePolicy
	logger             logger.Logger
	BackgroundJobEnqueuerDep
}

func NewStackReleaseService(spec StackReleaseServiceSpec) StackReleaseService {
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
		runtimePolicy:      spec.RuntimePolicy,
		logger:             logger.NewLoggerWithPrefix(context.Background(), "stack-release-service"),
	}
}

func (s *stackReleaseService) CreateRelease(ctx context.Context, stackID string, cause models.ReleaseCause) (*models.StackRelease, *errors.ServiceError) {
	stack, sErr := s.stackQuery.GetStack(ctx, stackID)
	if sErr != nil {
		return nil, sErr
	}

	if labels := stack.Labels.ToMap(); labels[models.PreviewStackLabel] == models.LabelValueTrue {
		return nil, errors.BadRequest("cannot create releases on preview-managed stacks; use the preview sync API")
	}

	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
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
	for _, resource := range stack.StackResources {
		if resource != nil {
			s.runtimePolicy.ApplyStackResourceDefaults(resource)
		}
	}
	snapshot, err := models.NewStackSnapshot(stack)
	if err != nil {
		return nil, errors.GeneralError("failed to create stack snapshot: %s", err.Error())
	}
	if freezeErr := s.freezeVolumeGitRevisions(ctx, &snapshot); freezeErr != nil {
		return nil, freezeErr
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
		if admissionErr := s.runtimePolicy.AdmitFirstReleaseWithTx(txCtx, stack.OrganisationID); admissionErr != nil {
			return admissionErr
		}
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

	if err := s.BackgroundJobEnqueuer.EnqueueAfterCommit(ctx, models.StackReleaseOperand{ID: created.ID}); err != nil {
		return nil, errors.GeneralError("failed to enqueue release: %s", err.Error())
	}
	return created, nil
}

// freezeVolumeGitRevisions resolves mutable branch and tag references before
// the snapshot is hashed and persisted. Retries and rollbacks can then consume
// the same content without consulting a repository again.
func (s *stackReleaseService) freezeVolumeGitRevisions(ctx context.Context, snapshot *models.StackSnapshot) *errors.ServiceError {
	for _, volume := range snapshot.Volumes {
		if volume == nil || volume.VolumeSource == nil || volume.VolumeSource.GitRepoSource == nil {
			continue
		}
		source := volume.VolumeSource.GitRepoSource
		if source.Revision.Commit != "" {
			continue
		}
		gitClient, err := s.gitClients.ClientFor(source.RepoUrl, gitclient.GitCredentials{})
		if err != nil {
			return errors.BadRequest("volume '%s': failed to create git client: %s", volume.Name, err.Error())
		}
		resolved, err := gitclient.ResolveGitRepoRevision(ctx, gitClient, source.RepoUrl, source.Revision)
		if err != nil {
			return errors.BadRequest("volume '%s': failed to resolve git revision: %s", volume.Name, err.Error())
		}
		if resolved.Commit == "" {
			return errors.BadRequest("volume '%s': resolved git revision has no commit SHA", volume.Name)
		}
		source.Revision = resolved
	}
	return nil
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
	if compatibilityErr := models.ValidatePinnedVolumeGitRevisions(src.Snapshot); compatibilityErr != nil {
		return nil, errors.BadRequest("cannot roll back release #%d: %s", src.Sequence, compatibilityErr.Error())
	}

	// Permission check via the stack's project.
	stack, sErr := s.stackQuery.GetStack(ctx, stackID)
	if sErr != nil {
		return nil, sErr
	}

	if labels := stack.Labels.ToMap(); labels[models.PreviewStackLabel] == models.LabelValueTrue {
		return nil, errors.BadRequest("cannot roll back a stack that is managed by preview sync")
	}

	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
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
		if admissionErr := s.runtimePolicy.AdmitRollbackWithTx(txCtx, stack.OrganisationID); admissionErr != nil {
			return admissionErr
		}
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

	if err := s.BackgroundJobEnqueuer.EnqueueAfterCommit(ctx, models.StackReleaseOperand{ID: created.ID}); err != nil {
		return nil, errors.GeneralError("failed to enqueue release: %s", err.Error())
	}
	return created, nil
}

// getReleaseWithStack fetches a release and its owning stack. GetStack
// already perm-checks (same project, same stack, auth.ActionRead), so callers
// don't need a separate permissions.Check.
func (s *stackReleaseService) getReleaseWithStack(ctx context.Context, releaseID string) (*models.StackRelease, *models.Stack, *errors.ServiceError) {
	rel, sErr := s.store.GetByID(ctx, releaseID)
	if sErr != nil {
		return nil, nil, sErr
	}

	stack, sErr := s.stackQuery.GetStack(ctx, rel.StackID)
	if sErr != nil {
		return nil, nil, sErr
	}

	return rel, stack, nil
}

func (s *stackReleaseService) GetRelease(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError) {
	rel, _, sErr := s.getReleaseWithStack(ctx, releaseID)
	if sErr != nil {
		return nil, sErr
	}
	return rel, nil
}

// GetReleaseDetail returns the release plus its live-status overlay, computed
// from the release's stack. The overlay is nil unless the release is active
// or is the stack's currently converged (live) release.
func (s *stackReleaseService) GetReleaseDetail(ctx context.Context, releaseID string) (*models.StackRelease, *models.ReleaseLiveStatus, *errors.ServiceError) {
	release, stack, sErr := s.getReleaseWithStack(ctx, releaseID)
	if sErr != nil {
		return nil, nil, sErr
	}

	return release, models.BuildReleaseLiveStatus(release, stack), nil
}

func (s *stackReleaseService) ListReleases(ctx context.Context, stackID string, params stores.ListParams) (*stores.PaginatedResult[*models.StackRelease], *errors.ServiceError) {
	stack, sErr := s.stackQuery.GetStack(ctx, stackID)
	if sErr != nil {
		return nil, sErr
	}

	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}

	return s.store.ListByStackID(ctx, stackID, params)
}

const (
	releaseEventsDefaultLimit = 100
	releaseEventsMaxLimit     = 500
	// releaseCancelledMessage is the user-facing message recorded on the
	// release_cancelled terminal event.
	releaseCancelledMessage = "Release cancelled"

	// releaseLiveMessage is the user-facing message recorded on the
	// release_released terminal event.
	releaseLiveMessage = "Release is live"
)

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

	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionRead); permErr != nil {
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

// StreamReleaseEvents returns a live SSE stream of release events after
// afterSequence. It enforces the exact same ownership and read-permission
// checks as ListReleaseEvents before handing back the streamer.
func (s *stackReleaseService) StreamReleaseEvents(ctx context.Context, stackID, releaseID string, afterSequence int) (interfaces.ServerSideStreamable, *errors.ServiceError) {
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

	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}

	return &releaseEventStreamer{
		events:        s.eventStore,
		releases:      s.store,
		releaseID:     releaseID,
		afterSeq:      afterSequence,
		pollInterval:  releaseEventStreamPollInterval,
		graceInterval: releaseEventStreamTerminalGraceInterval,
		presentEvent:  defaultPresentReleaseEvent,
	}, nil
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

	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, rel.StackID, auth.ActionWrite); permErr != nil {
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

// InternalGetReleaseRefs batch-resolves the latest and currently converged
// release for each given stack. No permission checks: callers have already
// authorized the stacks.
func (s *stackReleaseService) InternalGetReleaseRefs(ctx context.Context, stacks []*models.Stack) (map[string]models.StackReleaseRefs, *errors.ServiceError) {
	stackIDs := make([]string, len(stacks))
	for i, stack := range stacks {
		stackIDs[i] = stack.ID
	}

	latestReleasesByStackID, sErr := s.store.GetLatestByStackIDs(ctx, stackIDs)
	if sErr != nil {
		return nil, sErr
	}

	covergedReleaseIDs := make([]string, 0)
	for _, stack := range stacks {
		latestRelease := latestReleasesByStackID[stack.ID]
		// We dont want to fetch the converged release if:
		// - There is no latest release (the stack is not deployed, so no question of convergence)
		// - The stack has no converged release (the stack is not converged)
		// - The converged release is the same as the latest release (the stack is converged to the latest release)
		if latestRelease == nil || !stack.HasConvergedRelease() || stack.GetConvergedReleaseID() == latestRelease.ID {
			continue
		}
		covergedReleaseIDs = append(covergedReleaseIDs, stack.GetConvergedReleaseID())
	}

	convergedReleases, sErr := s.store.GetByIDs(ctx, covergedReleaseIDs)
	if sErr != nil {
		return nil, sErr
	}

	convergedReleasesByStackID := make(map[string]*models.StackRelease)
	for _, release := range convergedReleases {
		convergedReleasesByStackID[release.StackID] = release
	}

	refs := make(map[string]models.StackReleaseRefs, len(stacks))
	for _, stack := range stacks {
		latestRelease := latestReleasesByStackID[stack.ID]
		entry := models.StackReleaseRefs{
			Latest: latestRelease,
		}
		// If the stack has a converged release and it is the same as the latest release,
		// set the converged release to the latest release.
		if stack.HasConvergedRelease() && latestRelease != nil && stack.GetConvergedReleaseID() == latestRelease.ID {
			entry.Converged = latestRelease
		} else {
			entry.Converged = convergedReleasesByStackID[stack.ID]
		}
		refs[stack.ID] = entry
	}

	return refs, nil
}

func (s *stackReleaseService) InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError) {
	return s.store.GetByID(ctx, releaseID)
}

func (s *stackReleaseService) InternalResolveAuthoritativeWorkloadRelease(ctx context.Context, stack *models.Stack) (*models.StackRelease, *errors.ServiceError) {
	return resolveAuthoritativeWorkloadRelease(ctx, s.store, stack)
}

// InternalListAuthoritativeWorkload returns only IDs needed to enqueue cloud
// reconciliation. Full stacks and release snapshots are loaded by Execute.
func (s *stackReleaseService) InternalListAuthoritativeWorkload(ctx context.Context) (*models.WorkloadAuthorityScan, *errors.ServiceError) {
	stacks, serr := s.stackQuery.InternalListWorkloadAuthorityCandidates(ctx)
	if serr != nil {
		return nil, serr
	}
	scan := &models.WorkloadAuthorityScan{
		Releases:         make([]models.WorkloadReleaseRef, 0, len(stacks)),
		DeletingStackIDs: make([]string, 0),
	}
	activeStacks := make([]*models.Stack, 0, len(stacks))
	for _, stack := range stacks {
		if stack.DeletionTimestamp != nil {
			scan.DeletingStackIDs = append(scan.DeletingStackIDs, stack.ID)
			continue
		}
		activeStacks = append(activeStacks, stack)
	}
	if len(activeStacks) == 0 {
		return scan, nil
	}

	stackIDs := make([]string, len(activeStacks))
	for i, stack := range activeStacks {
		stackIDs[i] = stack.ID
	}
	latestByStackID, serr := s.store.GetLatestSummariesByStackIDs(ctx, stackIDs)
	if serr != nil {
		return nil, serr
	}
	convergedIDs := make([]string, 0, len(activeStacks))
	for _, stack := range activeStacks {
		latest := latestByStackID[stack.ID]
		if latest != nil && latest.State.Active() {
			continue
		}
		if convergedID := stack.GetConvergedReleaseID(); convergedID != "" {
			convergedIDs = append(convergedIDs, convergedID)
		}
	}
	convergedByID, serr := s.store.GetSummariesByIDs(ctx, convergedIDs)
	if serr != nil {
		return nil, serr
	}
	for _, stack := range activeStacks {
		latest := latestByStackID[stack.ID]
		if latest != nil && latest.State.Active() {
			scan.Releases = append(scan.Releases, models.WorkloadReleaseRef{StackID: stack.ID, ReleaseID: latest.ID})
			continue
		}
		converged := convergedByID[stack.GetConvergedReleaseID()]
		if converged == nil || converged.StackID != stack.ID {
			continue
		}
		switch converged.State {
		case models.ReleaseStateInProgress, models.ReleaseStateReleased, models.ReleaseStateSuperseded:
			scan.Releases = append(scan.Releases, models.WorkloadReleaseRef{StackID: stack.ID, ReleaseID: converged.ID})
		}
	}
	return scan, nil
}

func (s *stackReleaseService) InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError) {
	return s.store.GetActiveByStackID(ctx, stackID)
}

func (s *stackReleaseService) InternalGetLatestByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError) {
	return s.store.GetLatestByStackID(ctx, stackID)
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
	won, serr := s.store.MarkCancelled(ctx, id, reason)
	if serr != nil || !won {
		return won, serr
	}
	s.recordTerminalEvent(ctx, id, models.ReleaseStateCancelled, releaseCancelledMessage)
	return true, nil
}

func (s *stackReleaseService) MarkSuperseded(ctx context.Context, id string, reason string) (bool, *errors.ServiceError) {
	won, serr := s.store.MarkSuperseded(ctx, id, reason)
	if serr != nil || !won {
		return won, serr
	}
	s.recordTerminalEvent(ctx, id, models.ReleaseStateSuperseded, reason)
	return true, nil
}

func (s *stackReleaseService) MarkReleased(ctx context.Context, id string, outcome models.ReleaseOutcome) (bool, *errors.ServiceError) {
	won, serr := s.store.MarkReleased(ctx, id, outcome)
	if serr != nil || !won {
		return won, serr
	}
	s.recordTerminalEvent(ctx, id, models.ReleaseStateReleased, releaseLiveMessage)
	return true, nil
}

func (s *stackReleaseService) MarkFailed(ctx context.Context, id string, message string, outcome *models.ReleaseOutcome) (bool, *errors.ServiceError) {
	won, serr := s.store.MarkFailed(ctx, id, message, outcome)
	if serr != nil || !won {
		return won, serr
	}
	s.recordTerminalEvent(ctx, id, models.ReleaseStateFailed, message)
	return true, nil
}

func (s *stackReleaseService) SetConvergeClockStartedAt(ctx context.Context, id string, startedAt *time.Time) *errors.ServiceError {
	return s.store.SetConvergeClockStartedAt(ctx, id, startedAt)
}

func (s *stackReleaseService) MarkFailedWithValidationErrors(ctx context.Context, id, message string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError) {
	won, serr := s.store.MarkFailedWithValidationErrors(ctx, id, message, verrs)
	if serr != nil || !won {
		return won, serr
	}
	rel, getErr := s.store.GetByID(ctx, id)
	if getErr != nil {
		s.logger.Errorf("release %s: failed to load release for failure events: %v", id, getErr)
		return true, nil
	}
	for _, verr := range verrs {
		if recErr := s.eventRecorder.RecordReleaseCheckFailed(ctx, rel, verr.ResourceName, validationCheckKey(verr), verr.Message); recErr != nil {
			s.logger.Errorf("release %s: failed to record release_check_failed event: %v", id, recErr)
		}
	}
	if recErr := s.eventRecorder.RecordReleaseTerminal(ctx, rel, models.ReleaseStateFailed, message); recErr != nil {
		s.logger.Errorf("release %s: failed to record release_failed event: %v", id, recErr)
	}
	return true, nil
}

// recordTerminalEvent runs after a won terminal CAS: the state change is
// already persisted, so recording is best-effort and only the winner reaches
// here — a requeue that lost the race cannot double-emit.
func (s *stackReleaseService) recordTerminalEvent(ctx context.Context, id string, state models.StackReleaseState, message string) {
	rel, serr := s.store.GetByID(ctx, id)
	if serr != nil {
		s.logger.Errorf("release %s: failed to load release for %s event: %v", id, state, serr)
		return
	}
	if recErr := s.eventRecorder.RecordReleaseTerminal(ctx, rel, state, message); recErr != nil {
		s.logger.Errorf("release %s: failed to record %s event: %v", id, state, recErr)
	}
}

// validationCheckKey identifies a check failure for event dedupe. The error
// Code alone can collide across the pull and push checks of one resource,
// so Field — which encodes the check surface — is prefixed to keep the
// dedupe key unique per (resource, check).
func validationCheckKey(verr models.ReleaseValidationError) string {
	return fmt.Sprintf("%s:%s", verr.Field, verr.Code)
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
