package pgstore

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/models"
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
			discriminator text NOT NULL DEFAULT '',
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

	createConnectionRecord(t, sf, models.StackConnection{
		ID:      "conn-1",
		StackID: "stack-1",
		Kind:    models.ConnectionKindEnv,
		From:    models.TopologyNodeRef{Type: models.TopologyNodeTypeSecret, Id: "sec-1"},
		To:      models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
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

	createConnectionRecord(t, sf, models.StackConnection{
		ID:      "conn-1",
		StackID: "stack-1",
		Kind:    models.ConnectionKindEnv,
		From:    models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "api"},
		To:      models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"},
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

	createConnectionRecord(t, sf, models.StackConnection{
		ID:      "conn-1",
		StackID: "stack-1",
		Kind:    models.ConnectionKindVolumeMount,
		From:    models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "uploads"},
		To:      models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
	})
	createConnectionRecord(t, sf, models.StackConnection{
		ID:      "conn-2",
		StackID: "stack-2",
		Kind:    models.ConnectionKindVolumeMount,
		From:    models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "uploads"},
		To:      models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
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

func TestStackConnectionStoreAllowsMultipleConnectionsToSameAddonWithDifferentDatabases(t *testing.T) {
	sf := newTestSessionFactory(t)

	pgFrom := models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"}
	webTo := models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"}

	conn1 := models.StackConnection{
		ID:      "conn-1",
		StackID: "stack-1",
		Kind:    models.ConnectionKindEnv,
		From:    pgFrom,
		To:      webTo,
		Config:  models.ConnectionConfig{"database": "app_db"},
	}
	conn1.Discriminator = conn1.ComputeDiscriminator()

	conn2 := models.StackConnection{
		ID:      "conn-2",
		StackID: "stack-1",
		Kind:    models.ConnectionKindEnv,
		From:    pgFrom,
		To:      webTo,
		Config:  models.ConnectionConfig{"database": "tooljet_db"},
	}
	conn2.Discriminator = conn2.ComputeDiscriminator()

	createConnectionRecord(t, sf, conn1)
	createConnectionRecord(t, sf, conn2)

	var count int64
	if err := sf.db.Model(&models.StackConnection{}).Where("stack_id = ?", "stack-1").Count(&count).Error; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 connections, got %d", count)
	}
}

func TestComputeDiscriminatorReturnsDatabase(t *testing.T) {
	conn := models.StackConnection{
		Kind:   models.ConnectionKindEnv,
		From:   models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"},
		Config: models.ConnectionConfig{"database": "app_db"},
	}
	if got := conn.ComputeDiscriminator(); got != "app_db" {
		t.Fatalf("expected discriminator 'app_db', got '%s'", got)
	}
}

func TestComputeDiscriminatorReturnsEmptyForNonPostgres(t *testing.T) {
	conn := models.StackConnection{
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{Type: models.TopologyNodeTypeSecret, Id: "sec-1"},
	}
	if got := conn.ComputeDiscriminator(); got != "" {
		t.Fatalf("expected empty discriminator for secret connection, got '%s'", got)
	}
}

func TestComputeDiscriminatorReturnsEmptyForPostgresWithNilConfig(t *testing.T) {
	conn := models.StackConnection{
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "pg-1"},
	}
	if got := conn.ComputeDiscriminator(); got != "" {
		t.Fatalf("expected empty discriminator for postgres addon with nil config, got '%s'", got)
	}
}

func TestComputeDiscriminatorReturnsEmptyForVolumeMount(t *testing.T) {
	conn := models.StackConnection{
		Kind: models.ConnectionKindVolumeMount,
		From: models.TopologyNodeRef{Type: models.TopologyNodeTypeVolume, Name: "data"},
	}
	if got := conn.ComputeDiscriminator(); got != "" {
		t.Fatalf("expected empty discriminator for volume mount, got '%s'", got)
	}
}

func createConnectionRecord(t *testing.T, sf *testSessionFactory, record models.StackConnection) {
	t.Helper()
	if err := sf.db.Create(&record).Error; err != nil {
		t.Fatalf("failed to create stack connection record: %v", err)
	}
}
