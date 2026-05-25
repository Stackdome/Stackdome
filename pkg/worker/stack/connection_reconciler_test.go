package stack

import (
	"context"
	"testing"

	serrors "github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func TestConnectionReconcilerResolvesVolumeMountConnections(t *testing.T) {
	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		VolumeService: fakeVolumeService{
			volumes: []*models.Volume{
				{ID: "vol-1", Name: "uploads", VolumeSource: nil},
			},
		},
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{ID: "res-1", Name: "web"},
		},
		Connections: models.StackConnections{
			{
				Id:   "vol-web",
				Kind: models.ConnectionKindVolumeMount,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "uploads"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
				Config: map[string]interface{}{
					"mount_path": "/uploads",
					"sub_path":   "data",
				},
			},
		},
	}

	result, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	mounts := stack.StackResources[0].VolumeMounts
	if len(mounts) != 1 {
		t.Fatalf("expected 1 volume mount, got %d", len(mounts))
	}
	if mounts[0].SourceVolumeName != "uploads" {
		t.Fatalf("expected source volume name 'uploads', got '%s'", mounts[0].SourceVolumeName)
	}
	if mounts[0].SourceVolumeID != "vol-1" {
		t.Fatalf("expected source volume ID 'vol-1', got '%s'", mounts[0].SourceVolumeID)
	}
	if mounts[0].TargetPath != "/uploads" {
		t.Fatalf("expected target path '/uploads', got '%s'", mounts[0].TargetPath)
	}
	if mounts[0].SourceSubPath != "data" {
		t.Fatalf("expected sub path 'data', got '%s'", mounts[0].SourceSubPath)
	}
}

func TestConnectionReconcilerResolvesBuildArtifactSourceConnections(t *testing.T) {
	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		VolumeService: fakeVolumeService{
			volumes: []*models.Volume{
				{ID: "vol-1", Name: "assets"},
			},
		},
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{ID: "res-1", Name: "builder", BuildConfig: &models.BuildConfigSpec{}},
		},
		Connections: models.StackConnections{
			{
				Id:   "build-assets",
				Kind: models.ConnectionKindBuildArtifactSource,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "builder"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "assets"},
				Config: map[string]interface{}{
					"source_path":      "/app/public",
					"destination_path": "/",
				},
			},
		},
	}

	result, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected nil result, got %#v", result)
	}

	volume := stack.Volumes[0]
	if volume.VolumeSource == nil {
		t.Fatalf("expected volume source to be set")
	}
	if len(volume.VolumeSource.BuildSource) != 1 {
		t.Fatalf("expected 1 build source, got %d", len(volume.VolumeSource.BuildSource))
	}
	bs := volume.VolumeSource.BuildSource[0]
	if bs.ResourceName != "builder" {
		t.Fatalf("expected resource name 'builder', got '%s'", bs.ResourceName)
	}
	if bs.SourcePath != "/app/public" {
		t.Fatalf("expected source path '/app/public', got '%s'", bs.SourcePath)
	}
	if bs.DestinationPath != "/" {
		t.Fatalf("expected destination path '/', got '%s'", bs.DestinationPath)
	}
}

func TestConnectionReconcilerSkipsWhenNoMountOrBuildConnections(t *testing.T) {
	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		VolumeService: fakeVolumeService{},
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{ID: "res-1", Name: "web"},
		},
		Connections: models.StackConnections{
			{
				Id:   "env-conn",
				Kind: models.ConnectionKindEnv,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
			},
		},
	}

	result, err := reconciler.Reconcile(context.Background(), stack)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected nil result")
	}
	if len(stack.Volumes) != 0 {
		t.Fatalf("expected no volumes to be loaded, got %d", len(stack.Volumes))
	}
}

func TestConnectionReconcilerErrorsOnUnknownVolume(t *testing.T) {
	reconciler := NewConnectionReconciler(ConnectionReconcilerSpec{
		VolumeService: fakeVolumeService{volumes: []*models.Volume{}},
	})

	stack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{ID: "res-1", Name: "web"},
		},
		Connections: models.StackConnections{
			{
				Id:   "bad-vol",
				Kind: models.ConnectionKindVolumeMount,
				From: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "nonexistent"},
				To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
				Config: map[string]interface{}{
					"mount_path": "/data",
				},
			},
		},
	}

	_, err := reconciler.Reconcile(context.Background(), stack)
	if err == nil {
		t.Fatal("expected error for unknown volume")
	}
}

type fakeVolumeService struct {
	volumes []*models.Volume
}

func (f fakeVolumeService) ListVolumesUsedByStack(_ context.Context, _ string) ([]*models.Volume, *serrors.ServiceError) {
	return f.volumes, nil
}

func (f fakeVolumeService) InternalDeleteVolumesUsedByStackFromDB(_ context.Context, _ string) *serrors.ServiceError {
	return nil
}
