package pgstore

import (
	"context"
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/glebarez/sqlite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

const volumesTableDDL = `
	CREATE TABLE volumes (
		id text PRIMARY KEY,
		organisation_id text NOT NULL,
		project_id text,
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

func TestVolumeStoreListByProjectIDOrdersByCreatedAtDesc(t *testing.T) {
	ctx := context.Background()
	sf := newVolumesTestSessionFactory(t)
	store := NewVolumeStore(VolumeStoreSpec{SessionFactory: sf})

	older := time.Now().Add(-time.Hour)
	newer := time.Now()

	if err := sf.db.Create(&models.Volume{
		ID:             "vol-old",
		OrganisationID: "org-1",
		ProjectID:         "project-1",
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
		ProjectID:         "project-1",
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

	res, lerr := store.ListByProjectID(ctx, "project-1")
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

func TestVolumeStoreListByProjectIDFiltersByProject(t *testing.T) {
	ctx := context.Background()
	sf := newVolumesTestSessionFactory(t)
	store := NewVolumeStore(VolumeStoreSpec{SessionFactory: sf})

	now := time.Now()
	if err := sf.db.Create(&models.Volume{
		ID:             "vol-a",
		OrganisationID: "org-1",
		ProjectID:         "project-1",
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
		ProjectID:         "project-2",
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

	res, lerr := store.ListByProjectID(ctx, "project-1")
	if lerr != nil {
		t.Fatalf("unexpected error: %v", lerr)
	}
	if len(res) != 1 || res[0].ID != "vol-a" {
		t.Fatalf("expected only vol-a, got %+v", res)
	}
}

// Suite bootstrapped by TestPgstore in pgstore_suite_test.go.

var _ = Describe("VolumeStore InternalListNotReady", func() {
	var (
		sf    *testSessionFactory
		store stores.VolumeStore
		ctx   context.Context
	)

	seedVolume := func(id string, status *models.VolumeStatus) {
		Expect(sf.db.Create(&models.Volume{
			ID:             id,
			OrganisationID: "org-1",
			ProjectID:         "project-1",
			UserID:         "u-1",
			Name:           id,
			NamespaceID:    "ns-1",
			Namespace:      "ns-1",
			Size:           "1Gi",
			AccessMode:     models.READ_WRITE_ONCE,
			Status:         status,
		}).Error).To(Succeed())
	}

	BeforeEach(func() {
		gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		Expect(err).ToNot(HaveOccurred())
		Expect(gdb.Exec(volumesTableDDL).Error).To(Succeed())
		sf = &testSessionFactory{db: gdb}
		store = NewVolumeStore(VolumeStoreSpec{SessionFactory: sf})
		ctx = context.Background()
	})

	It("returns volumes with no status or a non-Ready phase and skips Ready ones", func() {
		seedVolume("vol-no-status", nil)
		seedVolume("vol-pending", &models.VolumeStatus{Phase: models.VolumePhasePending})
		seedVolume("vol-ready", &models.VolumeStatus{Phase: models.VolumePhaseReady})

		res, serr := store.InternalListNotReady(ctx)
		Expect(serr).To(BeNil())

		ids := make([]string, 0, len(res))
		for _, v := range res {
			ids = append(ids, v.ID)
		}
		Expect(ids).To(ConsistOf("vol-no-status", "vol-pending"))
	})
})
