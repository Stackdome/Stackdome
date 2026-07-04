package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

//go:generate mockgen -destination=../mocks/mock_reference_service.go -package=mocks github.com/ashishmax31/stackdome-api-server/pkg/services ReferenceService
type ReferenceService interface {
	ReprojectSpec(ctx context.Context, stackID string) *errors.ServiceError
	ProjectRelease(ctx context.Context, release *models.StackRelease) *errors.ServiceError
	IsReferentInUse(ctx context.Context, referentType models.ReferentType, referentID string) (
		bool,
		[]models.ResourceReference,
		*errors.ServiceError,
	)
}

type ReferenceServiceSpec struct {
	SessionFactory db.SessionFactory
	StackStore     stores.StackStore
}

type referenceService struct {
	store      stores.ResourceReferenceStore
	stackStore stores.StackStore
}

func NewReferenceService(spec ReferenceServiceSpec) ReferenceService {
	return &referenceService{
		store: pgstore.NewResourceReferenceStore(
			pgstore.ResourceReferenceStoreSpec{
				SessionFactory: spec.SessionFactory,
			},
		),
		stackStore: spec.StackStore,
	}
}

// ReprojectSpec rebuilds the live-spec reference rows for a stack. It MUST run
// inside a DB transaction — the store returns an error if none is present — and
// every spec mutation already calls it within its transaction.
func (s *referenceService) ReprojectSpec(ctx context.Context, stackID string) *errors.ServiceError {
	stack, err := s.stackStore.GetByID(ctx, stackID)
	if err != nil {
		return err
	}
	return s.store.ReplaceSpecWithTx(ctx, stackID, extractReferences(stack))
}

// ProjectRelease writes the immutable reference rows for a release. It MUST run
// inside the same DB transaction that inserts the release.
func (s *referenceService) ProjectRelease(ctx context.Context, release *models.StackRelease) *errors.ServiceError {
	stack := release.Snapshot.ToStack()
	return s.store.InsertReleaseWithTx(ctx, release.ID, release.StackID, extractReferences(stack))
}

func (s *referenceService) IsReferentInUse(ctx context.Context, referentType models.ReferentType, referentID string) (bool, []models.ResourceReference, *errors.ServiceError) {
	refs, err := s.store.ListByReferent(ctx, referentType, referentID)
	if err != nil {
		return false, nil, err
	}
	return len(refs) > 0, refs, nil
}

// extractReferences derives the deletable referents a stack consumes, from its
// connections and embedded (implicit) secret refs. Scope fields (StackID/ReleaseID)
// are set by the store, not here. Results are deduped by (referent, kind).
func extractReferences(stack *models.Stack) []models.ResourceReference {
	seen := make(map[string]struct{})
	var refs []models.ResourceReference

	add := func(rt models.ReferentType, id string, kind models.RelationKind) {
		if id == "" {
			return
		}
		key := string(rt) + "|" + id + "|" + string(kind)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		refs = append(refs, models.ResourceReference{ReferentType: rt, ReferentID: id, RelationKind: kind})
	}

	volumesByName := stack.VolumesMap()
	for _, c := range stack.Connections {
		for _, ep := range []models.TopologyNodeRef{c.From, c.To} {
			switch ep.Type {
			case models.TopologyNodeTypeSecret:
				add(models.ReferentSecret, ep.Id, models.RelationEnv)
			case models.TopologyNodeTypePostgresAddon:
				add(models.ReferentPostgresAddon, ep.Id, models.RelationEnv)
			case models.TopologyNodeTypeVolume:
				id := ep.Id
				if id == "" {
					if v, ok := volumesByName[ep.Name]; ok {
						id = v.ID
					}
				}
				add(models.ReferentVolume, id, models.RelationVolumeMount)
			}
		}
	}

	for _, r := range stack.StackResources {
		if r.HasGitCredentials() {
			add(models.ReferentSecret, r.BuildConfig.SourceContext.Git.GitSecretRef.SecretID, models.RelationGitCredential)
		}
		if r.HasImagePullSecrets() {
			add(models.ReferentSecret, r.ImageConfig.PullSecretRef.SecretID, models.RelationImagePull)
		}
		if r.HasImagePushSecrets() {
			add(models.ReferentSecret, r.BuildConfig.RegistrySecretRef.SecretID, models.RelationImagePush)
		}
		add(models.ReferentRegistryCredential, r.RegistryPullCredentialID(), models.RelationImagePull)
		add(models.ReferentRegistryCredential, r.RegistryPushCredentialID(), models.RelationImagePush)
		add(models.ReferentGitIntegration, r.GitIntegrationID(), models.RelationGitCredential)
	}
	return refs
}

// describeReferences renders blocking references for an error message.
func describeReferences(refs []models.ResourceReference) string {
	stacks := map[string]struct{}{}
	releases := 0
	for _, r := range refs {
		if r.ReleaseID == nil {
			stacks[r.StackID] = struct{}{}
		} else {
			releases++
		}
	}
	parts := []string{}
	if len(stacks) > 0 {
		parts = append(parts, fmt.Sprintf("%d stack spec(s)", len(stacks)))
	}
	if releases > 0 {
		parts = append(parts, fmt.Sprintf("%d release(s)", releases))
	}
	if len(parts) == 0 {
		return "active resources"
	}
	return strings.Join(parts, " and ")
}
