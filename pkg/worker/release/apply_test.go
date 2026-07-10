package release

import (
	"context"
	stderrors "errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
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

var _ = Describe("ApplyReconciler volume existence", func() {
	var (
		ctrl   *gomock.Controller
		volSvc *MockvolumeService
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		volSvc = NewMockvolumeService(ctrl)
	})

	// Every snapshot-referenced volume (via resource VolumeMounts and via
	// BuildConfig.SourceContext.Volume) exists in the DB -> no error. Extra live
	// volumes not referenced by the snapshot are irrelevant. Readiness is
	// deliberately not part of the contract: the cluster agent waits for PVCs
	// itself, so the hub only checks existence.
	It("returns no error when every referenced volume exists", func() {
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
		Expect(r.verifyReferencedVolumesExist(context.Background(), release)).To(Succeed())
	})

	// A referenced volume that no longer exists in the DB must fail the release
	// outright via a missingVolumesError, since applying CRs that reference a
	// nonexistent volume would stall the cluster agent trying to mount a PVC
	// that will never appear.
	It("returns a missingVolumesError when a referenced volume no longer exists", func() {
		release := volExistTestRelease([]*models.StackResource{
			volumeMountResource("web", "v-gone"),
		}, nil)

		volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volExistTestStackID).
			Return([]*models.Volume{}, nil)

		r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
		err := r.verifyReferencedVolumesExist(context.Background(), release)
		var missing *missingVolumesError
		Expect(stderrors.As(err, &missing)).To(BeTrue(), "expected *missingVolumesError, got %v (%T)", err, err)
		Expect(missing.refs).To(Equal([]string{"v-gone"}))
	})

	// No volumes referenced at all -> no error without calling the volume service
	// (asserted via zero mock expectations - an unexpected call fails the gomock
	// controller).
	It("skips listing when no volumes are referenced", func() {
		// No EXPECT() set: any call to ListVolumesUsedByStack fails the test.
		release := volExistTestRelease([]*models.StackResource{
			{Name: "web"},
		}, nil)

		r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
		Expect(r.verifyReferencedVolumesExist(context.Background(), release)).To(Succeed())
	})

	// A volume referenced only via a volume_mount connection (not yet
	// materialized into resource.VolumeMounts, since the stackdeploy Resolver
	// only does that at render time on a throwaway Stack rebuilt from the
	// snapshot) must still be existence-checked. This covers the connections gap:
	// release.Snapshot.Resources[].VolumeMounts alone would miss this reference.
	It("returns a missingVolumesError for a missing volume referenced only via a volume_mount connection", func() {
		release := volExistTestRelease(
			[]*models.StackResource{{Name: "web"}},
			models.StackConnections{volumeMountConnection("v3", "data", "web")},
		)

		volSvc.EXPECT().ListVolumesUsedByStack(gomock.Any(), volExistTestStackID).
			Return([]*models.Volume{}, nil)

		r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
		err := r.verifyReferencedVolumesExist(context.Background(), release)
		var missing *missingVolumesError
		Expect(stderrors.As(err, &missing)).To(BeTrue(), "expected *missingVolumesError, got %v (%T)", err, err)
		Expect(missing.refs).To(Equal([]string{"data"}), "expected missing refs [data] (name preferred as label)")
	})

	// A volume_mount connection where From.Id is empty and only From.Name is set
	// (name-fallback branch) must be satisfied by a live volume matching that
	// name. This covers the production scenario where volumes are referenced by
	// name rather than ID.
	It("matches a name-only connection reference against a live volume", func() {
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
		Expect(r.verifyReferencedVolumesExist(context.Background(), release)).To(Succeed())
	})

	// A build_artifact_source connection (resource -> volume, the volume being a
	// build-output destination) contributes no referenced volume: that volume is
	// never referenced by the Stack/StackResource CRs applied here — it's managed
	// by the separate volume worker/controller. With no other references the
	// service must not be called at all.
	It("does not treat a build_artifact_source connection as a volume reference", func() {
		// No EXPECT() set: any call to ListVolumesUsedByStack fails the test.
		release := volExistTestRelease(
			[]*models.StackResource{{Name: "builder"}},
			models.StackConnections{buildArtifactSourceConnection("builder", "v4", "artifacts")},
		)

		r := &applyReconciler{volumeService: volSvc, logger: testLogger()}
		Expect(r.verifyReferencedVolumesExist(context.Background(), release)).To(Succeed())
	})

	// A volume deleted and recreated under the same name while the release is in
	// flight gets a new ID: the snapshot ref carries the stale ID plus the name
	// (the shape a persisted VolumeMount carries). The live volume matches the
	// ref by name, so the release must NOT be failed as missing.
	It("matches a recreated volume by name despite a stale ID", func() {
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
		Expect(r.verifyReferencedVolumesExist(context.Background(), release)).To(Succeed(),
			"recreated volume matches ref by name")
	})

	// Reconcile-level: when verifyReferencedVolumesExist reports a
	// *missingVolumesError, Reconcile must route it through failRelease —
	// MarkFailed is invoked with a message naming the missing ref — and return
	// resultStop with a nil error, matching the other terminal failure paths. The
	// snapshot resource carries only a volume mount (no image/build config, no
	// secret or postgres connections), so the secret-sync steps before the
	// existence check are all no-ops.
	It("fails the release via failRelease when Reconcile finds a missing volume", func() {
		release := volExistTestRelease([]*models.StackResource{
			volumeMountResource("web", "v-gone"),
		}, nil)
		release.Manifest = &models.ReleaseManifest{}

		stackSvc := NewMockstackService(ctrl)
		stackSvc.EXPECT().InternalGetStack(gomock.Any(), volExistTestStackID).
			Return(&models.Stack{ID: volExistTestStackID, ClusterID: volExistTestClusterID}, nil)

		clusterMgr := mocks.NewMockClusterManager(ctrl)
		clusterMgr.EXPECT().GetClient(volExistTestClusterID).
			Return(applySecretsTestClient(GinkgoT()), nil)

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
		Expect(err).ToNot(HaveOccurred(), "terminal failure is routed via failRelease")
		Expect(result.resultStop).To(BeTrue(), "expected resultStop when a referenced volume is missing")
	})
})
