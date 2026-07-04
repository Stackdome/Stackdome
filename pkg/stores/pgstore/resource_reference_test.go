package pgstore

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type refTestSessionFactory struct{ db *gorm.DB }

func newRefTestSessionFactory(t *testing.T) *refTestSessionFactory {
	t.Helper()
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := gdb.Exec(`
		CREATE TABLE resource_references (
			id text PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			stack_id text NOT NULL,
			release_id text,
			referent_type text NOT NULL,
			referent_id text NOT NULL,
			relation_kind text NOT NULL,
			created_at datetime
		)
	`).Error; err != nil {
		t.Fatalf("create resource_references: %v", err)
	}
	if err := gdb.Exec(`
		CREATE TABLE stack_releases (
			id text PRIMARY KEY,
			state text NOT NULL
		)
	`).Error; err != nil {
		t.Fatalf("create stack_releases: %v", err)
	}
	return &refTestSessionFactory{db: gdb}
}

func (f *refTestSessionFactory) Init(*config.DatabaseConfig)  {}
func (f *refTestSessionFactory) DirectDB() *sql.DB            { d, _ := f.db.DB(); return d }
func (f *refTestSessionFactory) New(context.Context) *gorm.DB { return f.db.Session(&gorm.Session{}) }
func (f *refTestSessionFactory) CheckConnection() error       { return nil }
func (f *refTestSessionFactory) Close() error                 { d, _ := f.db.DB(); return d.Close() }

func (f *refTestSessionFactory) withTx(t *testing.T, fn func(ctx context.Context)) {
	t.Helper()
	tx := f.db.Begin()
	ctx := db.CtxWithTransaction(context.Background(), tx)
	fn(ctx)
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func (f *refTestSessionFactory) seedRelease(t *testing.T, id string, state models.StackReleaseState) {
	t.Helper()
	if err := f.db.Exec("INSERT INTO stack_releases (id, state) VALUES (?, ?)", id, state).Error; err != nil {
		t.Fatalf("seed release: %v", err)
	}
}

func TestResourceReferenceStore_ReplaceSpecAndList(t *testing.T) {
	f := newRefTestSessionFactory(t)
	store := NewResourceReferenceStore(ResourceReferenceStoreSpec{SessionFactory: f})

	f.withTx(t, func(ctx context.Context) {
		if err := store.ReplaceSpecWithTx(ctx, "stack-1", []models.ResourceReference{
			{ReferentType: models.ReferentSecret, ReferentID: "sec-1", RelationKind: models.RelationEnv},
			{ReferentType: models.ReferentVolume, ReferentID: "vol-1", RelationKind: models.RelationVolumeMount},
		}); err != nil {
			t.Fatalf("ReplaceSpecWithTx: %v", err)
		}
	})

	got, err := store.ListByReferent(context.Background(), models.ReferentSecret, "sec-1")
	if err != nil {
		t.Fatalf("ListByReferent: %v", err)
	}
	if len(got) != 1 || got[0].StackID != "stack-1" || got[0].ReleaseID != nil {
		t.Fatalf("unexpected spec rows: %+v", got)
	}

	f.withTx(t, func(ctx context.Context) {
		if err := store.ReplaceSpecWithTx(ctx, "stack-1", nil); err != nil {
			t.Fatalf("ReplaceSpecWithTx empty: %v", err)
		}
	})
	got, _ = store.ListByReferent(context.Background(), models.ReferentSecret, "sec-1")
	if len(got) != 0 {
		t.Fatalf("expected spec rows cleared, got %+v", got)
	}
}

func TestResourceReferenceStore_InsertReleaseIsScoped(t *testing.T) {
	f := newRefTestSessionFactory(t)
	store := NewResourceReferenceStore(ResourceReferenceStoreSpec{SessionFactory: f})
	f.seedRelease(t, "rel-1", models.ReleaseStateReleased)

	f.withTx(t, func(ctx context.Context) {
		_ = store.ReplaceSpecWithTx(ctx, "stack-1", []models.ResourceReference{
			{ReferentType: models.ReferentSecret, ReferentID: "sec-1", RelationKind: models.RelationEnv},
		})
		if err := store.InsertReleaseWithTx(ctx, "rel-1", "stack-1", []models.ResourceReference{
			{ReferentType: models.ReferentSecret, ReferentID: "sec-1", RelationKind: models.RelationImagePull},
		}); err != nil {
			t.Fatalf("InsertReleaseWithTx: %v", err)
		}
	})

	got, _ := store.ListByReferent(context.Background(), models.ReferentSecret, "sec-1")
	if len(got) != 2 {
		t.Fatalf("expected spec + release rows, got %d: %+v", len(got), got)
	}
	var release int
	for _, r := range got {
		if r.ReleaseID != nil && *r.ReleaseID == "rel-1" {
			release++
		}
	}
	if release != 1 {
		t.Fatalf("expected one release-scoped row, got %d", release)
	}
}

func TestResourceReferenceStore_FailedReleaseDoesNotGrip(t *testing.T) {
	f := newRefTestSessionFactory(t)
	store := NewResourceReferenceStore(ResourceReferenceStoreSpec{SessionFactory: f})

	f.seedRelease(t, "rel-released", models.ReleaseStateReleased)
	f.seedRelease(t, "rel-failed", models.ReleaseStateFailed)
	f.seedRelease(t, "rel-superseded", models.ReleaseStateSuperseded)
	f.seedRelease(t, "rel-cancelled", models.ReleaseStateCancelled)
	f.seedRelease(t, "rel-inprogress", models.ReleaseStateInProgress)

	f.withTx(t, func(ctx context.Context) {
		for _, rel := range []string{"rel-released", "rel-failed", "rel-superseded", "rel-cancelled", "rel-inprogress"} {
			if err := store.InsertReleaseWithTx(ctx, rel, "stack-1", []models.ResourceReference{
				{ReferentType: models.ReferentSecret, ReferentID: "sec-1", RelationKind: models.RelationEnv},
			}); err != nil {
				t.Fatalf("InsertReleaseWithTx(%s): %v", rel, err)
			}
		}
	})

	got, err := store.ListByReferent(context.Background(), models.ReferentSecret, "sec-1")
	if err != nil {
		t.Fatalf("ListByReferent: %v", err)
	}

	// Only Released and InProgress should grip. Failed/Superseded/Cancelled should not.
	gripping := map[string]bool{}
	for _, r := range got {
		if r.ReleaseID != nil {
			gripping[*r.ReleaseID] = true
		}
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 gripping rows (Released + InProgress), got %d: %+v", len(got), gripping)
	}
	if !gripping["rel-released"] {
		t.Fatal("Released release should grip")
	}
	if !gripping["rel-inprogress"] {
		t.Fatal("InProgress release should grip")
	}
	if gripping["rel-failed"] {
		t.Fatal("Failed release should NOT grip")
	}
}
