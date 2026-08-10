package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

func TestExtractReferences(t *testing.T) {
	stack := &models.Stack{
		Volumes: []*models.Volume{{ID: "vol-1", Name: "uploads"}},
		Connections: models.StackConnections{
			{Kind: models.ConnectionKindEnv, From: models.TopologyNodeRef{Type: models.TopologyNodeTypeSecret, Id: "sec-1"}, To: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"}},
			{Kind: models.ConnectionKindEnv, From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"}, To: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"}},
			{Kind: models.ConnectionKindVolumeMount, From: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "uploads"}, To: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"}},
		},
		StackResources: []*models.StackResource{
			{Name: "web", ImageConfig: &models.ImageConfigSpec{Image: "x", RegistryCredentialID: "rc-web"}},
			{Name: "web", ImageConfig: &models.ImageConfigSpec{Image: "x", RegistryCredentialID: "rc-web"}}, // dup => deduped
			{Name: "api", ImageConfig: &models.ImageConfigSpec{Image: "x", RegistryCredentialID: "rc-pull"}},
			{Name: "builder", BuildConfig: &models.BuildConfigSpec{
				PushRegistryCredentialID: "rc-push",
				SourceContext:            models.BuildContextSource{Git: &models.GitBuildSource{RepoURL: "https://example.com/a.git", IntegrationID: "gi-1"}},
			}},
		},
	}

	refs := extractReferences(stack)

	want := map[string]bool{
		"secret|sec-1|env":                       true,
		"postgres_addon|pg-1|env":                true,
		"volume|vol-1|volume_declaration":        true,
		"volume|vol-1|volume_mount":              true,
		"registry_credential|rc-web|image_pull":  true,
		"registry_credential|rc-pull|image_pull": true,
		"registry_credential|rc-push|image_push": true,
		"git_integration|gi-1|git_credential":    true,
	}
	if len(refs) != len(want) {
		t.Fatalf("got %d refs, want %d: %+v", len(refs), len(want), refs)
	}
	for _, r := range refs {
		key := string(r.ReferentType) + "|" + r.ReferentID + "|" + string(r.RelationKind)
		if !want[key] {
			t.Fatalf("unexpected reference %q", key)
		}
	}
}

func TestReferenceService_IsReferentInUse(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockResourceReferenceStore(ctrl)
	store.EXPECT().ListByReferent(gomock.Any(), models.ReferentSecret, "sec-1").
		Return([]models.ResourceReference{{StackID: "s1", ReferentType: models.ReferentSecret, ReferentID: "sec-1", RelationKind: models.RelationEnv}}, nil)

	svc := &referenceService{store: store}
	inUse, refs, err := svc.IsReferentInUse(context.Background(), models.ReferentSecret, "sec-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inUse || len(refs) != 1 {
		t.Fatalf("expected in-use with 1 ref, got inUse=%v refs=%+v", inUse, refs)
	}
}

func TestReferenceService_ReprojectSpecLoadsStackAndReplaces(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockResourceReferenceStore(ctrl)
	stackStore := mocks.NewMockStackStore(ctrl)

	stackStore.EXPECT().GetByID(gomock.Any(), "stack-1").Return(&models.Stack{
		Connections: models.StackConnections{
			{Kind: models.ConnectionKindEnv, From: models.TopologyNodeRef{Type: models.TopologyNodeTypeSecret, Id: "sec-1"}, To: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"}},
		},
	}, nil)
	store.EXPECT().ReplaceSpecWithTx(gomock.Any(), "stack-1", gomock.Len(1)).Return(nil)

	svc := &referenceService{store: store, stackStore: stackStore}
	if err := svc.ReprojectSpec(context.Background(), "stack-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
