package release

import (
	"context"
	stderrors "errors"
	"strings"
	"testing"

	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

const (
	volExistTestStackID   = "stack-vol-1"
	volExistTestClusterID = "cluster-vol-1"
)

func volExistTestRelease(resources []*models.StackResource, connections models.StackConnections) *models.StackRelease {
	return &models.StackRelease{
		ID:      "release-vol-1",
		StackID: volExistTestStackID,
		State:   models.ReleaseStateInProgress,
		Snapshot: models.StackSnapshot{
			Stack: models.StackShellSnapshot{
				ID: volExistTestStackID,
			},
			Resources:   resources,
			Connections: connections,
		},
	}
}

func volumeMountResource(name, sourceVolumeID string) *models.StackResource {
	return &models.StackResource{
		Name: name,
		VolumeMounts: []*models.VolumeMount{
			{SourceVolumeID: sourceVolumeID, TargetPath: "/data"},
		},
	}
}

func buildContextVolumeResource(name, sourceVolumeID string) *models.StackResource {
	return &models.StackResource{
		Name: name,
		BuildConfig: &models.BuildConfigSpec{
			SourceContext: models.BuildContextSource{
				Volume: &models.VolumeBuildSource{SourceVolumeID: sourceVolumeID, SourceVolumeName: "src"},
			},
		},
	}
}

func volumeMountConnection(volumeID, volumeName, resourceName string) models.StackConnection {
	return models.StackConnection{
		ID:   "conn-" + volumeID,
		Kind: models.ConnectionKindVolumeMount,
		From: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Id: volumeID, Name: volumeName},
		To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: resourceName},
	}
}

func buildArtifactSourceConnection(resourceName, volumeID, volumeName string) models.StackConnection {
	return models.StackConnection{
		ID:   "conn-bas-" + volumeID,
		Kind: models.ConnectionKindBuildArtifactSource,
		From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: resourceName},
		To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Id: volumeID, Name: volumeName},
	}
}

func liveVolume(id string) *models.Volume {
	return &models.Volume{ID: id, Name: id}
}

// Every snapshot-referenced volume (via resource VolumeMounts and via
// BuildConfig.SourceContext.Volume) exists in the DB -> no error. Extra live
// volumes not referenced by the snapshot are irrelevant. Readiness is
// deliberately not part of the contract: the cluster agent waits for PVCs
// itself, so the hub only checks existence.
func TestApplyReconciler_VerifyReferencedVolumesExist_AllExist_NoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := volExistTestRelease([]*models.StackResource{
		volumeMountResource("web", "v1"),
		buildContextVolumeResource("builder", "v2"),
	}, nil)

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volExistTestStackID).
		Return([]*models.Volume{
			liveVolume("v1"),
			liveVolume("v2"),
			liveVolume("v-unreferenced"),
		}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	if err := r.verifyReferencedVolumesExist(context.Background(), release); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// A referenced volume that no longer exists in the DB must fail the release
// outright via a missingVolumesError, since applying CRs that reference a
// nonexistent volume would stall the cluster agent trying to mount a PVC
// that will never appear.
func TestApplyReconciler_VerifyReferencedVolumesExist_ReferencedVolumeMissing_ReturnsMissingVolumesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := volExistTestRelease([]*models.StackResource{
		volumeMountResource("web", "v-gone"),
	}, nil)

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volExistTestStackID).
		Return([]*models.Volume{}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	err := r.verifyReferencedVolumesExist(context.Background(), release)
	var missing *missingVolumesError
	if !stderrors.As(err, &missing) {
		t.Fatalf("expected *missingVolumesError, got %v (%T)", err, err)
	}
	if len(missing.refs) != 1 || missing.refs[0] != "v-gone" {
		t.Fatalf("expected missing refs [v-gone], got %+v", missing.refs)
	}
}

// No volumes referenced at all -> no error without calling the volume service
// (asserted via zero mock expectations - an unexpected call fails the gomock
// controller).
func TestApplyReconciler_VerifyReferencedVolumesExist_NoReferences_SkipsListing(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)
	// No EXPECT() set: any call to ListVolumesUsedByStack fails the test.

	release := volExistTestRelease([]*models.StackResource{
		{Name: "web"},
	}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	if err := r.verifyReferencedVolumesExist(context.Background(), release); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// A volume referenced only via a volume_mount connection (not yet
// materialized into resource.VolumeMounts, since the stackdeploy Resolver
// only does that at render time on a throwaway Stack rebuilt from the
// snapshot) must still be existence-checked. This covers the connections gap:
// release.Snapshot.Resources[].VolumeMounts alone would miss this reference.
func TestApplyReconciler_VerifyReferencedVolumesExist_VolumeMountConnectionMissing_ReturnsMissingVolumesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := volExistTestRelease(
		[]*models.StackResource{{Name: "web"}},
		models.StackConnections{volumeMountConnection("v3", "data", "web")},
	)

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volExistTestStackID).
		Return([]*models.Volume{}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	err := r.verifyReferencedVolumesExist(context.Background(), release)
	var missing *missingVolumesError
	if !stderrors.As(err, &missing) {
		t.Fatalf("expected *missingVolumesError, got %v (%T)", err, err)
	}
	if len(missing.refs) != 1 || missing.refs[0] != "data" {
		t.Fatalf("expected missing refs [data] (name preferred as label), got %+v", missing.refs)
	}
}

// A volume_mount connection where From.Id is empty and only From.Name is set
// (name-fallback branch) must be satisfied by a live volume matching that
// name. This covers the production scenario where volumes are referenced by
// name rather than ID.
func TestApplyReconciler_VerifyReferencedVolumesExist_NameOnlyConnectionReference_Found(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := volExistTestRelease(
		[]*models.StackResource{{Name: "web"}},
		models.StackConnections{
			models.StackConnection{
				ID:   "conn-data",
				Kind: models.ConnectionKindVolumeMount,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Id: "", Name: "data"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
			},
		},
	)

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volExistTestStackID).
		Return([]*models.Volume{liveVolume("data")}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	if err := r.verifyReferencedVolumesExist(context.Background(), release); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// A build_artifact_source connection (resource -> volume, the volume being a
// build-output destination) contributes no referenced volume: that volume is
// never referenced by the Stack/StackResource CRs applied here — it's managed
// by the separate volume worker/controller. With no other references the
// service must not be called at all.
func TestApplyReconciler_VerifyReferencedVolumesExist_BuildArtifactSourceConnection_NotReferenced(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)
	// No EXPECT() set: any call to ListVolumesUsedByStack fails the test.

	release := volExistTestRelease(
		[]*models.StackResource{{Name: "builder"}},
		models.StackConnections{buildArtifactSourceConnection("builder", "v4", "artifacts")},
	)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	if err := r.verifyReferencedVolumesExist(context.Background(), release); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// A volume deleted and recreated under the same name while the release is in
// flight gets a new ID: the snapshot ref carries the stale ID plus the name
// (the shape a persisted VolumeMount carries). The live volume matches the
// ref by name, so the release must NOT be failed as missing.
func TestApplyReconciler_VerifyReferencedVolumesExist_RecreatedVolumeMatchedByName_Found(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := volExistTestRelease([]*models.StackResource{
		{
			Name: "web",
			VolumeMounts: []*models.VolumeMount{
				{SourceVolumeID: "v-stale", SourceVolumeName: "data", TargetPath: "/data"},
			},
		},
	}, nil)

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volExistTestStackID).
		Return([]*models.Volume{
			{ID: "v-new", Name: "data"},
		}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	if err := r.verifyReferencedVolumesExist(context.Background(), release); err != nil {
		t.Fatalf("unexpected error (recreated volume matches ref by name): %v", err)
	}
}

// Reconcile-level: when verifyReferencedVolumesExist reports a
// *missingVolumesError, Reconcile must route it through failRelease —
// MarkFailed is invoked with a message naming the missing ref — and return
// resultStop with a nil error, matching the other terminal failure paths. The
// snapshot resource carries only a volume mount (no image/build config, no
// secret or postgres connections), so the secret-sync steps before the
// existence check are all no-ops.
func TestApplyReconciler_Reconcile_MissingVolume_FailsRelease(t *testing.T) {
	ctrl := gomock.NewController(t)

	release := volExistTestRelease([]*models.StackResource{
		volumeMountResource("web", "v-gone"),
	}, nil)
	release.Manifest = &models.ReleaseManifest{}

	stackSvc := NewMockstackService(ctrl)
	stackSvc.EXPECT().InternalGetStack(gomock.Any(), volExistTestStackID).
		Return(&models.Stack{ID: volExistTestStackID, ClusterID: volExistTestClusterID}, nil)

	clusterMgr := mocks.NewMockClusterManager(ctrl)
	clusterMgr.EXPECT().GetClient(volExistTestClusterID).
		Return(applySecretsTestClient(t), nil)

	volSvc := NewMockvolumeService(ctrl)
	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volExistTestStackID).
		Return([]*models.Volume{}, nil)

	relSvc := NewMockreleaseService(ctrl)
	relSvc.EXPECT().MarkFailed(
		gomock.Any(),
		release.ID,
		gomock.Cond(func(msg string) bool { return strings.Contains(msg, "v-gone") }),
		gomock.Nil(),
	).Return(true, nil)

	r := &applyReconciler{
		releaseService: relSvc,
		stackService:   stackSvc,
		clusterManager: clusterMgr,
		volumeService:  volSvc,
		logger:         testLogger(),
	}

	result, err := r.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("expected nil error (terminal failure is routed via failRelease), got: %v", err)
	}
	if !result.resultStop {
		t.Fatal("expected resultStop when a referenced volume is missing")
	}
}
