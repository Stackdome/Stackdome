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
	volReadyTestStackID   = "stack-vol-1"
	volReadyTestClusterID = "cluster-vol-1"
)

func volReadyTestRelease(resources []*models.StackResource, connections models.StackConnections) *models.StackRelease {
	return &models.StackRelease{
		ID:      "release-vol-1",
		StackID: volReadyTestStackID,
		State:   models.ReleaseStateInProgress,
		Snapshot: models.StackSnapshot{
			Stack: models.StackShellSnapshot{
				ID: volReadyTestStackID,
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

func notReadyVolume(id string) *models.Volume {
	return &models.Volume{
		ID:     id,
		Name:   id,
		Status: &models.VolumeStatus{Phase: models.VolumePhasePending},
	}
}

func readyVolume(id string) *models.Volume {
	return &models.Volume{
		ID:     id,
		Name:   id,
		Status: &models.VolumeStatus{Phase: models.VolumePhaseReady},
	}
}

// (a) A volume that exists on the stack but isn't referenced by anything in
// this release's snapshot must never gate the release, even if it's not
// Ready. This is the core bug: a broken/unused volume must not block every
// release forever.
func TestApplyReconciler_VolumesReady_UnreferencedVolumeNotReady_DoesNotGate(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := volReadyTestRelease([]*models.StackResource{
		volumeMountResource("web", "v-referenced"),
	}, nil)

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volReadyTestStackID).
		Return([]*models.Volume{
			readyVolume("v-referenced"),
			notReadyVolume("v-unreferenced"),
		}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	ready, err := r.volumesReady(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ready {
		t.Fatalf("expected ready=true (unreferenced volume must not gate), got false")
	}
}

// (b) A volume referenced via a resource's VolumeMounts that isn't Ready must
// gate the release (requeue).
func TestApplyReconciler_VolumesReady_ReferencedByMountNotReady_Gates(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := volReadyTestRelease([]*models.StackResource{
		volumeMountResource("web", "v1"),
	}, nil)

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volReadyTestStackID).
		Return([]*models.Volume{notReadyVolume("v1")}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	ready, err := r.volumesReady(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ready {
		t.Fatalf("expected ready=false, got true")
	}
}

// (c) A volume referenced via BuildConfig.SourceContext.Volume that isn't
// Ready must gate the release.
func TestApplyReconciler_VolumesReady_ReferencedByBuildContextNotReady_Gates(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := volReadyTestRelease([]*models.StackResource{
		buildContextVolumeResource("builder", "v2"),
	}, nil)

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volReadyTestStackID).
		Return([]*models.Volume{notReadyVolume("v2")}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	ready, err := r.volumesReady(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ready {
		t.Fatalf("expected ready=false, got true")
	}
}

// (d) A referenced volume that no longer exists in the DB must fail the
// release outright (not requeue forever) via a missingVolumesError, since
// applying CRs that reference a nonexistent volume would stall the cluster
// agent trying to mount a PVC that will never appear.
func TestApplyReconciler_VolumesReady_ReferencedVolumeMissing_ReturnsMissingVolumesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := volReadyTestRelease([]*models.StackResource{
		volumeMountResource("web", "v-gone"),
	}, nil)

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volReadyTestStackID).
		Return([]*models.Volume{}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	ready, err := r.volumesReady(context.Background(), release)
	if ready {
		t.Fatalf("expected ready=false, got true")
	}
	var missing *missingVolumesError
	if !stderrors.As(err, &missing) {
		t.Fatalf("expected *missingVolumesError, got %v (%T)", err, err)
	}
	if len(missing.refs) != 1 || missing.refs[0] != "v-gone" {
		t.Fatalf("expected missing refs [v-gone], got %+v", missing.refs)
	}
}

// (e) No volumes referenced at all -> ready=true without calling the volume
// service (asserted via zero mock expectations - an unexpected call fails the
// gomock controller).
func TestApplyReconciler_VolumesReady_NoReferences_SkipsListing(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)
	// No EXPECT() set: any call to ListVolumesUsedByStack fails the test.

	release := volReadyTestRelease([]*models.StackResource{
		{Name: "web"},
	}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	ready, err := r.volumesReady(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ready {
		t.Fatalf("expected ready=true, got false")
	}
}

// A volume referenced only via a volume_mount connection (not yet
// materialized into resource.VolumeMounts, since the stackdeploy Resolver
// only does that at render time on a throwaway Stack rebuilt from the
// snapshot) must still gate the release when not Ready. This covers the
// connections gap: release.Snapshot.Resources[].VolumeMounts alone would miss
// this reference.
func TestApplyReconciler_VolumesReady_ReferencedByVolumeMountConnection_Gates(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := volReadyTestRelease(
		[]*models.StackResource{{Name: "web"}},
		models.StackConnections{volumeMountConnection("v3", "data", "web")},
	)

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volReadyTestStackID).
		Return([]*models.Volume{notReadyVolume("v3")}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	ready, err := r.volumesReady(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ready {
		t.Fatalf("expected ready=false, got true")
	}
}

// A volume_mount connection where From.Id is empty and only From.Name
// is set (name-fallback branch) must gate the release when not Ready.
// This covers the production scenario where volumes are referenced by
// name rather than ID.
func TestApplyReconciler_VolumesReady_ReferencedByVolumeMountConnectionNameOnly_Gates(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := volReadyTestRelease(
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

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volReadyTestStackID).
		Return([]*models.Volume{notReadyVolume("data")}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	ready, err := r.volumesReady(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ready {
		t.Fatalf("expected ready=false, got true")
	}
}

// A build_artifact_source connection (resource -> volume, the volume being a
// build-output destination) must NOT gate the release: that volume is never
// referenced by the Stack/StackResource CRs applied here — it's managed by
// the separate volume worker/controller.
func TestApplyReconciler_VolumesReady_BuildArtifactSourceConnection_DoesNotGate(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)
	// No EXPECT() set: the build_artifact_source connection contributes no
	// referenced volume, so with no other references the service must not be
	// called at all.

	release := volReadyTestRelease(
		[]*models.StackResource{{Name: "builder"}},
		models.StackConnections{buildArtifactSourceConnection("builder", "v4", "artifacts")},
	)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	ready, err := r.volumesReady(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ready {
		t.Fatalf("expected ready=true, got false")
	}
}

// recreatedVolumeRelease returns a release whose snapshot references a volume
// by BOTH a (now stale) ID and its name — the shape a persisted VolumeMount
// carries. Used by the recreated-volume tests below.
func recreatedVolumeRelease(staleVolumeID, volumeName string) *models.StackRelease {
	return volReadyTestRelease([]*models.StackResource{
		{
			Name: "web",
			VolumeMounts: []*models.VolumeMount{
				{SourceVolumeID: staleVolumeID, SourceVolumeName: volumeName, TargetPath: "/data"},
			},
		},
	}, nil)
}

// A volume deleted and recreated under the same name while the release is in
// flight gets a new ID: the snapshot ref carries the stale ID plus the name.
// The live volume matches the ref by name, so the release must NOT be failed
// as missing — and with the recreated volume Ready, it proceeds.
func TestApplyReconciler_VolumesReady_RecreatedVolumeMatchedByName_Ready(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := recreatedVolumeRelease("v-stale", "data")

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volReadyTestStackID).
		Return([]*models.Volume{
			{ID: "v-new", Name: "data", Status: &models.VolumeStatus{Phase: models.VolumePhaseReady}},
		}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	ready, err := r.volumesReady(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ready {
		t.Fatalf("expected ready=true (recreated volume matches ref by name), got false")
	}
}

// Same recreated-volume scenario, but the recreated volume isn't Ready yet:
// the ref is satisfied (not missing), so no missingVolumesError — the release
// just gates (requeues) until the volume becomes Ready.
func TestApplyReconciler_VolumesReady_RecreatedVolumeMatchedByName_NotReady_Gates(t *testing.T) {
	ctrl := gomock.NewController(t)
	volSvc := NewMockvolumeService(ctrl)

	release := recreatedVolumeRelease("v-stale", "data")

	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volReadyTestStackID).
		Return([]*models.Volume{
			{ID: "v-new", Name: "data", Status: &models.VolumeStatus{Phase: models.VolumePhasePending}},
		}, nil)

	r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
	ready, err := r.volumesReady(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error (ref satisfied by name, must not be missing): %v", err)
	}
	if ready {
		t.Fatalf("expected ready=false (recreated volume not Ready), got true")
	}
}

// Reconcile-level: when volumesReady reports a *missingVolumesError, Reconcile
// must route it through failRelease — MarkFailed is invoked with a message
// naming the missing ref — and return resultStop with a nil error, matching
// the other terminal failure paths. The snapshot resource carries only a
// volume mount (no image/build config, no secret or postgres connections), so
// the secret-sync steps before volumesReady are all no-ops.
func TestApplyReconciler_Reconcile_MissingVolume_FailsRelease(t *testing.T) {
	ctrl := gomock.NewController(t)

	release := volReadyTestRelease([]*models.StackResource{
		volumeMountResource("web", "v-gone"),
	}, nil)
	release.Manifest = &models.ReleaseManifest{}

	stackSvc := NewMockstackService(ctrl)
	stackSvc.EXPECT().InternalGetStack(gomock.Any(), volReadyTestStackID).
		Return(&models.Stack{ID: volReadyTestStackID, ClusterID: volReadyTestClusterID}, nil)

	clusterMgr := mocks.NewMockClusterManager(ctrl)
	clusterMgr.EXPECT().GetClient(volReadyTestClusterID).
		Return(applySecretsTestClient(t), nil)

	volSvc := NewMockvolumeService(ctrl)
	volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volReadyTestStackID).
		Return([]*models.Volume{}, nil)

	relSvc := NewMockreleaseService(ctrl)
	relSvc.EXPECT().MarkFailed(
		gomock.Any(),
		release.ID,
		gomock.Cond(func(msg string) bool { return strings.Contains(msg, "v-gone") }),
		gomock.Nil(),
	).Return(true, nil)

	rec := NewMockeventRecorder(ctrl)
	rec.EXPECT().RecordReleaseTerminal(gomock.Any(), release, models.ReleaseStateFailed, gomock.Any()).
		Return(nil)

	r := &applyReconciler{
		releaseService: relSvc,
		stackService:   stackSvc,
		clusterManager: clusterMgr,
		volumeService:  volSvc,
		eventRecorder:  rec,
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
