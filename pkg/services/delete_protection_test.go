package services

import (
	"context"
	"testing"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services/clusterresource"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
)

func TestSecretDeleteIsBlockedByConnectionUsage(t *testing.T) {
	svc := &secretService{
		secretStore: &fakeSecretStore{
			secret: &models.Secret{ID: "sec-1", TeamID: "team-1"},
		},
		connectionUsageChecker: &fakeConnectionUsageChecker{sourceReferenced: true},
		resourceUsageService:   &fakeResourceUsageService{},
		permissions:            allowPermissions{},
	}

	err := svc.Delete(context.Background(), "sec-1")
	if err == nil {
		t.Fatalf("expected delete to be blocked")
	}
	if got, want := err.Error(), "error: secret with ID 'sec-1' is in use by stack connections"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestSecretDeleteIsAllowedAfterConnectionRemoval(t *testing.T) {
	store := &fakeSecretStore{
		secret: &models.Secret{ID: "sec-1", TeamID: "team-1"},
	}
	svc := &secretService{
		secretStore:            store,
		connectionUsageChecker: &fakeConnectionUsageChecker{},
		resourceUsageService:   &fakeResourceUsageService{},
		permissions:            allowPermissions{},
	}

	if err := svc.Delete(context.Background(), "sec-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !store.deleted {
		t.Fatalf("expected secret to be deleted")
	}
}

func TestSecretDeleteIsBlockedByDirectConfigUsage(t *testing.T) {
	svc := &secretService{
		secretStore: &fakeSecretStore{
			secret: &models.Secret{ID: "sec-1", TeamID: "team-1"},
		},
		connectionUsageChecker: &fakeConnectionUsageChecker{},
		resourceUsageService:   &fakeResourceUsageService{inUse: true},
		permissions:            allowPermissions{},
	}

	err := svc.Delete(context.Background(), "sec-1")
	if err == nil {
		t.Fatalf("expected delete to be blocked")
	}
	if got, want := err.Error(), "error: secret with ID 'sec-1' is in use by stacks"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestPostgresAddonDeleteIsBlockedByConnectionUsage(t *testing.T) {
	svc := &postgresAddonService{
		postgresAddonStore: &fakePostgresAddonStore{
			addon: &models.PostgresAddon{ID: "pg-1", TeamID: "team-1"},
		},
		connectionUsageChecker: &fakeConnectionUsageChecker{sourceReferenced: true},
		permissions:            allowPermissions{},
	}

	_, err := svc.DeletePostgresAddon(context.Background(), "pg-1")
	if err == nil {
		t.Fatalf("expected delete to be blocked")
	}
	if got, want := err.Error(), "error: addon is in use by one or more stacks and cannot be deleted"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestVolumeDeleteIsBlockedByVolumeMountConnectionSource(t *testing.T) {
	svc := newVolumeServiceForDeleteProtectionTest(&fakeConnectionUsageChecker{sourceReferenced: true})

	err := svc.Delete(context.Background(), "vol-1")
	if err == nil {
		t.Fatalf("expected delete to be blocked")
	}
	if got, want := err.Error(), "error: volume is in use by one or more stack connections"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestVolumeDeleteIsBlockedByBuildArtifactConnectionTarget(t *testing.T) {
	svc := newVolumeServiceForDeleteProtectionTest(&fakeConnectionUsageChecker{targetReferenced: true})

	err := svc.Delete(context.Background(), "vol-1")
	if err == nil {
		t.Fatalf("expected delete to be blocked")
	}
	if got, want := err.Error(), "error: volume is in use by one or more stack connections"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func newVolumeServiceForDeleteProtectionTest(checker connectionUsageChecker) *volumeService {
	return &volumeService{
		volumeStore: &fakeVolumeStore{
			volume: &models.Volume{
				ID:     "vol-1",
				Name:   "uploads",
				TeamID: "team-1",
			},
		},
		stackVolumeStore:       &fakeStackVolumeStore{stackID: "stack-1"},
		connectionUsageChecker: checker,
		clusterResourceService: fakeVolumeClusterResourceService{},
		permissions:            allowPermissions{},
	}
}

type allowPermissions struct{}

func (allowPermissions) Check(context.Context, string, string, string, string) *errors.ServiceError {
	return nil
}

type fakeConnectionUsageChecker struct {
	sourceReferenced bool
	targetReferenced bool
}

func (f *fakeConnectionUsageChecker) IsNodeReferencedAsSource(_ context.Context, _ string, _ models.TopologyNodeRef) (bool, error) {
	return f.sourceReferenced, nil
}

func (f *fakeConnectionUsageChecker) IsNodeReferencedAsTarget(_ context.Context, _ string, _ models.TopologyNodeRef) (bool, error) {
	return f.targetReferenced, nil
}

func (f *fakeConnectionUsageChecker) IsNodeReferenced(ctx context.Context, stackID string, ref models.TopologyNodeRef) (bool, error) {
	source, err := f.IsNodeReferencedAsSource(ctx, stackID, ref)
	if err != nil || source {
		return source, err
	}
	return f.IsNodeReferencedAsTarget(ctx, stackID, ref)
}

type fakeResourceUsageService struct {
	inUse bool
}

func (f *fakeResourceUsageService) Create(context.Context, *models.ResourceUsage) error { return nil }
func (f *fakeResourceUsageService) IsResourceInUse(context.Context, string, string) (bool, error) {
	return f.inUse, nil
}
func (f *fakeResourceUsageService) GetByStackID(context.Context, string) ([]*models.ResourceUsage, error) {
	return nil, nil
}
func (f *fakeResourceUsageService) DeleteByStackID(context.Context, string) error { return nil }
func (f *fakeResourceUsageService) Delete(context.Context, string, string, string) error {
	return nil
}

type fakeSecretStore struct {
	secret  *models.Secret
	deleted bool
}

func (f *fakeSecretStore) Create(context.Context, *models.Secret) (*models.Secret, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeSecretStore) GetByID(context.Context, string) (*models.Secret, *errors.ServiceError) {
	return f.secret, nil
}
func (f *fakeSecretStore) GetByName(context.Context, string, string) (*models.Secret, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeSecretStore) Update(context.Context, *models.Secret) (*models.Secret, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeSecretStore) UpdateEncryptedData(context.Context, string, string, []string, string) *errors.ServiceError {
	panic("not used")
}
func (f *fakeSecretStore) Delete(context.Context, string) *errors.ServiceError {
	f.deleted = true
	return nil
}
func (f *fakeSecretStore) ListByOrganisation(context.Context, string) ([]*models.Secret, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeSecretStore) ListByTeamID(context.Context, string) ([]*models.Secret, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeSecretStore) ListByTeamIDs(context.Context, []string) ([]*models.Secret, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeSecretStore) ListByUser(context.Context, string, string) ([]*models.Secret, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeSecretStore) ListByType(context.Context, models.SecretType, models.SecretType) ([]*models.Secret, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeSecretStore) ValidateSecretExists(context.Context, string) (bool, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeSecretStore) GetSecretKeys(context.Context, string) ([]string, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeSecretStore) ValidateSecretHasKeys(context.Context, string, []string) (bool, []string, *errors.ServiceError) {
	panic("not used")
}

type fakePostgresAddonStore struct {
	addon *models.PostgresAddon
}

func (f *fakePostgresAddonStore) Create(context.Context, *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) CreateWithTx(context.Context, *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) GetByID(context.Context, string) (*models.PostgresAddon, *errors.ServiceError) {
	return f.addon, nil
}
func (f *fakePostgresAddonStore) GetByName(context.Context, string, string) (*models.PostgresAddon, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) Update(context.Context, *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) UpdateWithTx(context.Context, *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) Delete(context.Context, string) *errors.ServiceError {
	panic("not used")
}
func (f *fakePostgresAddonStore) ListByOrganisation(context.Context, string) ([]*models.PostgresAddon, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) ListByTeamID(context.Context, string) ([]*models.PostgresAddon, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) ListByTeamIDs(context.Context, []string) ([]*models.PostgresAddon, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) ListByCluster(context.Context, string) ([]*models.PostgresAddon, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) ValidateAddonExists(context.Context, string) (bool, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) ValidateAddonNameUnique(context.Context, string, string, string) (bool, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) UpdateStatus(context.Context, string, *models.PostgresAddonStatus) *errors.ServiceError {
	panic("not used")
}
func (f *fakePostgresAddonStore) UpdateBackupRequestedAt(context.Context, string, *time.Time) *errors.ServiceError {
	panic("not used")
}
func (f *fakePostgresAddonStore) InternalList(context.Context, string, ...any) ([]*models.PostgresAddon, *errors.ServiceError) {
	panic("not used")
}
func (f *fakePostgresAddonStore) WithTransaction(context.Context, func(context.Context) *errors.ServiceError) *errors.ServiceError {
	panic("not used")
}

type fakeVolumeStore struct {
	volume  *models.Volume
	deleted bool
}

func (f *fakeVolumeStore) Create(context.Context, *models.Volume) (*models.Volume, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeVolumeStore) CreateWithTx(context.Context, *models.Volume) (*models.Volume, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeVolumeStore) InternalList(context.Context, []string) ([]*models.Volume, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeVolumeStore) GetByID(context.Context, string) (*models.Volume, *errors.ServiceError) {
	return f.volume, nil
}
func (f *fakeVolumeStore) UpdateGitRepoSourceRevision(context.Context, string, models.GitRepoRevision) (*models.Volume, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeVolumeStore) UpdateGitRepoSourceRevisionWithTx(context.Context, string, models.GitRepoRevision) (*models.Volume, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeVolumeStore) UpdateRemoteDirSourceHash(context.Context, string, string) (*models.Volume, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeVolumeStore) UpdateRemoteDirSourceHashWithTx(context.Context, string, string) (*models.Volume, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeVolumeStore) UpdateStatus(context.Context, string, *models.VolumeStatus) *errors.ServiceError {
	panic("not used")
}
func (f *fakeVolumeStore) Delete(context.Context, string) *errors.ServiceError {
	f.deleted = true
	return nil
}
func (f *fakeVolumeStore) DeleteWithTx(context.Context, string) *errors.ServiceError {
	f.deleted = true
	return nil
}
func (f *fakeVolumeStore) GetByUserID(context.Context, string) ([]*models.Volume, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeVolumeStore) GetByVolumeNameAndNamespace(context.Context, string, string) (*models.Volume, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeVolumeStore) ListByTeamID(context.Context, string) ([]*models.Volume, *errors.ServiceError) {
	panic("not used")
}
func (f *fakeVolumeStore) WithTransaction(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
	return fn(ctx)
}

type fakeStackVolumeStore struct {
	stackID string
}

func (fakeStackVolumeStore) Create(context.Context, *models.StackVolume) *errors.ServiceError {
	panic("not used")
}
func (fakeStackVolumeStore) Delete(context.Context, string, string) *errors.ServiceError {
	panic("not used")
}
func (f fakeStackVolumeStore) GetByVolumeID(context.Context, string) (*models.StackVolume, *errors.ServiceError) {
	return &models.StackVolume{StackID: f.stackID}, nil
}
func (fakeStackVolumeStore) ListVolumesByStackID(context.Context, string) ([]*models.Volume, *errors.ServiceError) {
	panic("not used")
}
func (fakeStackVolumeStore) CreateWithTx(context.Context, *models.StackVolume) *errors.ServiceError {
	panic("not used")
}
func (fakeStackVolumeStore) DeleteWithTx(context.Context, string, string) *errors.ServiceError {
	panic("not used")
}

type fakeVolumeClusterResourceService struct{}

func (fakeVolumeClusterResourceService) UpdateVolumeRemoteDirRevisionInCluster(context.Context, *models.Volume) *clusterresource.ClusterResourceError {
	panic("not used")
}
func (fakeVolumeClusterResourceService) UpdateVolumeGitRevisionInCluster(context.Context, *models.Volume) *clusterresource.ClusterResourceError {
	panic("not used")
}
func (fakeVolumeClusterResourceService) CreateVolumeInCluster(context.Context, *models.Volume) *clusterresource.ClusterResourceError {
	panic("not used")
}
func (fakeVolumeClusterResourceService) DeleteVolumeInCluster(context.Context, *models.Volume) *clusterresource.ClusterResourceError {
	return nil
}

var _ auth.PermissionService = allowPermissions{}
var _ stores.SecretStore = (*fakeSecretStore)(nil)
var _ stores.PostgresAddonStore = (*fakePostgresAddonStore)(nil)
var _ stores.VolumeStore = (*fakeVolumeStore)(nil)
var _ stores.StackVolumeStore = fakeStackVolumeStore{}
