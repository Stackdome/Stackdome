package pgstore

import (
	"context"
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const volumesTableDDL = `
	CREATE TABLE volumes (
		id text PRIMARY KEY,
		organisation_id text NOT NULL,
		team_id text,
		user_id text NOT NULL,
		name text NOT NULL,
		namespace_id text NOT NULL,
		namespace text NOT NULL,
		labels jsonb,
		annotations jsonb,
		size text,
		storage_class text,
		access_mode text NOT NULL,
		volume_source jsonb,
		sync_before_use boolean,
		status jsonb,
		created_at datetime,
		updated_at datetime
	)
`

func newVolumesTestSessionFactory(t *testing.T) *testSessionFactory {
	t.Helper()

	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := gdb.Exec(volumesTableDDL).Error; err != nil {
		t.Fatalf("failed to create volumes table: %v", err)
	}
	return &testSessionFactory{db: gdb}
}

func TestVolumeStoreListByTeamIDOrdersByCreatedAtDesc(t *testing.T) {
	ctx := context.Background()
	sf := newVolumesTestSessionFactory(t)
	store := NewVolumeStore(VolumeStoreSpec{SessionFactory: sf})

	older := time.Now().Add(-time.Hour)
	newer := time.Now()

	if err := sf.db.Create(&models.Volume{
		ID:             "vol-old",
		OrganisationID: "org-1",
		TeamID:         "team-1",
		UserID:         "u-1",
		Name:           "older",
		NamespaceID:    "ns-1",
		Namespace:      "ns-1",
		Size:           "1Gi",
		AccessMode:     models.READ_WRITE_ONCE,
		CreatedAt:      older,
		UpdatedAt:      older,
	}).Error; err != nil {
		t.Fatalf("seed older: %v", err)
	}
	if err := sf.db.Create(&models.Volume{
		ID:             "vol-new",
		OrganisationID: "org-1",
		TeamID:         "team-1",
		UserID:         "u-1",
		Name:           "newer",
		NamespaceID:    "ns-1",
		Namespace:      "ns-1",
		Size:           "1Gi",
		AccessMode:     models.READ_WRITE_ONCE,
		CreatedAt:      newer,
		UpdatedAt:      newer,
	}).Error; err != nil {
		t.Fatalf("seed newer: %v", err)
	}

	res, lerr := store.ListByTeamID(ctx, "team-1")
	if lerr != nil {
		t.Fatalf("unexpected error: %v", lerr)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(res))
	}
	if res[0].ID != "vol-new" {
		t.Fatalf("expected vol-new first (created_at DESC), got %s", res[0].ID)
	}
	if res[1].ID != "vol-old" {
		t.Fatalf("expected vol-old second, got %s", res[1].ID)
	}
}

func TestVolumeStoreListByTeamIDFiltersByTeam(t *testing.T) {
	ctx := context.Background()
	sf := newVolumesTestSessionFactory(t)
	store := NewVolumeStore(VolumeStoreSpec{SessionFactory: sf})

	now := time.Now()
	if err := sf.db.Create(&models.Volume{
		ID:             "vol-a",
		OrganisationID: "org-1",
		TeamID:         "team-1",
		UserID:         "u-1",
		Name:           "a",
		NamespaceID:    "ns-1",
		Namespace:      "ns-1",
		Size:           "1Gi",
		AccessMode:     models.READ_WRITE_ONCE,
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error; err != nil {
		t.Fatalf("seed a: %v", err)
	}
	if err := sf.db.Create(&models.Volume{
		ID:             "vol-b",
		OrganisationID: "org-1",
		TeamID:         "team-2",
		UserID:         "u-1",
		Name:           "b",
		NamespaceID:    "ns-1",
		Namespace:      "ns-1",
		Size:           "1Gi",
		AccessMode:     models.READ_WRITE_ONCE,
		CreatedAt:      now,
		UpdatedAt:      now,
	}).Error; err != nil {
		t.Fatalf("seed b: %v", err)
	}

	res, lerr := store.ListByTeamID(ctx, "team-1")
	if lerr != nil {
		t.Fatalf("unexpected error: %v", lerr)
	}
	if len(res) != 1 || res[0].ID != "vol-a" {
		t.Fatalf("expected only vol-a, got %+v", res)
	}
}
