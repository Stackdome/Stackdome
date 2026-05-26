package pgstore

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ashishmax31/stackdome-api-server/config"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type testSessionFactory struct {
	db *gorm.DB
}

func newTestSessionFactory(t *testing.T) *testSessionFactory {
	t.Helper()

	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := gdb.Exec(`
		CREATE TABLE stack_connections (
			id text PRIMARY KEY,
			stack_id text NOT NULL,
			kind text NOT NULL,
			from_ref jsonb NOT NULL,
			to_ref jsonb NOT NULL,
			mappings jsonb,
			config jsonb,
			created_at datetime,
			updated_at datetime
		)
	`).Error; err != nil {
		t.Fatalf("failed to create stack_connections table: %v", err)
	}
	return &testSessionFactory{db: gdb}
}

func (f *testSessionFactory) Init(*config.DatabaseConfig) {}

func (f *testSessionFactory) DirectDB() *sql.DB {
	db, _ := f.db.DB()
	return db
}

func (f *testSessionFactory) New(context.Context) *gorm.DB {
	return f.db.Session(&gorm.Session{})
}

func (f *testSessionFactory) CheckConnection() error { return nil }

func (f *testSessionFactory) Close() error {
	db, _ := f.db.DB()
	return db.Close()
}

func TestStackConnectionStoreReferencesSourceByTypeAndID(t *testing.T) {
	ctx := context.Background()
	sf := newTestSessionFactory(t)
	store := NewStackConnectionStore(StackConnectionStoreSpec{SessionFactory: sf})

	createConnectionRecord(t, sf, models.StackConnectionRecord{
		ID:      "conn-1",
		StackID: "stack-1",
		Kind:    models.ConnectionKindEnv,
		FromRef: models.TopologyNodeRef{Type: models.TopologyNodeTypeSecret, Id: "sec-1"},
		ToRef:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
	})

	referenced, err := store.IsNodeReferencedAsSource(ctx, "", models.TopologyNodeRef{
		Type: models.TopologyNodeTypeSecret,
		Id:   "sec-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !referenced {
		t.Fatalf("expected secret source to be referenced")
	}
}

func TestStackConnectionStoreReferencesTargetByTypeAndID(t *testing.T) {
	ctx := context.Background()
	sf := newTestSessionFactory(t)
	store := NewStackConnectionStore(StackConnectionStoreSpec{SessionFactory: sf})

	createConnectionRecord(t, sf, models.StackConnectionRecord{
		ID:      "conn-1",
		StackID: "stack-1",
		Kind:    models.ConnectionKindEnv,
		FromRef: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
		ToRef:   models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"},
	})

	referenced, err := store.IsNodeReferencedAsTarget(ctx, "", models.TopologyNodeRef{
		Type: models.TopologyNodeTypePostgresAddon,
		Id:   "pg-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !referenced {
		t.Fatalf("expected postgres target to be referenced")
	}
}

func TestStackConnectionStoreReferencesSourceByStackScopedName(t *testing.T) {
	ctx := context.Background()
	sf := newTestSessionFactory(t)
	store := NewStackConnectionStore(StackConnectionStoreSpec{SessionFactory: sf})

	createConnectionRecord(t, sf, models.StackConnectionRecord{
		ID:      "conn-1",
		StackID: "stack-1",
		Kind:    models.ConnectionKindVolumeMount,
		FromRef: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "uploads"},
		ToRef:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
	})
	createConnectionRecord(t, sf, models.StackConnectionRecord{
		ID:      "conn-2",
		StackID: "stack-2",
		Kind:    models.ConnectionKindVolumeMount,
		FromRef: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "uploads"},
		ToRef:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
	})

	referenced, err := store.IsNodeReferencedAsSource(ctx, "stack-1", models.TopologyNodeRef{
		Type: models.TopologyNodeTypeVolume,
		Name: "uploads",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !referenced {
		t.Fatalf("expected stack-local volume source to be referenced")
	}

	referenced, err = store.IsNodeReferencedAsSource(ctx, "stack-1", models.TopologyNodeRef{
		Type: models.TopologyNodeTypeVolume,
		Name: "missing",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if referenced {
		t.Fatalf("expected missing stack-local volume source to be unreferenced")
	}
}

func TestStackConnectionStoreReferencesTargetByStackScopedName(t *testing.T) {
	ctx := context.Background()
	sf := newTestSessionFactory(t)
	store := NewStackConnectionStore(StackConnectionStoreSpec{SessionFactory: sf})

	createConnectionRecord(t, sf, models.StackConnectionRecord{
		ID:      "conn-1",
		StackID: "stack-1",
		Kind:    models.ConnectionKindBuildArtifactSource,
		FromRef: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "builder"},
		ToRef:   models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "assets"},
	})

	referenced, err := store.IsNodeReferencedAsTarget(ctx, "stack-1", models.TopologyNodeRef{
		Type: models.TopologyNodeTypeVolume,
		Name: "assets",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !referenced {
		t.Fatalf("expected stack-local volume target to be referenced")
	}
}

func createConnectionRecord(t *testing.T, sf *testSessionFactory, record models.StackConnectionRecord) {
	t.Helper()
	if err := sf.db.Create(&record).Error; err != nil {
		t.Fatalf("failed to create stack connection record: %v", err)
	}
}
