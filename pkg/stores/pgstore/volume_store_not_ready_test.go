package pgstore

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/glebarez/sqlite"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

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
			TeamID:         "team-1",
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
