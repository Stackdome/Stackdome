package pgstore_test

import (
	"context"
	"net/http"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TrialAllocationStore", func() {
	var (
		ctx      context.Context
		sf       *sqliteSessionFactory
		store    stores.TrialAllocationStore
		executor stores.AtomicExecutor
		now      time.Time
	)

	BeforeEach(func() {
		ctx = context.Background()
		now = time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
		sf = newSQLiteSessionFactory(`
			CREATE TABLE trial_allocations (
				id TEXT PRIMARY KEY,
				organisation_id TEXT NOT NULL UNIQUE,
				state TEXT NOT NULL,
				activated_at DATETIME NOT NULL,
				expires_at DATETIME NOT NULL,
				cleanup_started_at DATETIME,
				cleaned_up_at DATETIME,
				error_at DATETIME,
				error_message TEXT,
				created_at DATETIME,
				updated_at DATETIME
			)
		`)
		store = pgstore.NewTrialAllocationStore(pgstore.TrialAllocationStoreSpec{SessionFactory: sf})
		executor = pgstore.NewAtomicExecutor(sf)
	})

	acquire := func(orgID string, capacity int) (*models.TrialAllocation, *errors.ServiceError) {
		var allocation *models.TrialAllocation
		serr := executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
			var acquireErr *errors.ServiceError
			allocation, acquireErr = store.AcquireWithTx(txCtx, orgID, now, now.Add(6*time.Hour), capacity)
			return acquireErr
		})
		return allocation, serr
	}

	It("requires the caller's outer transaction", func() {
		allocation, serr := store.AcquireWithTx(ctx, "org-1", now, now.Add(6*time.Hour), 1)
		Expect(allocation).To(BeNil())
		Expect(serr).ToNot(BeNil())
		Expect(serr.Error()).To(ContainSubstring("transaction"))
	})

	It("returns the same active allocation on retry", func() {
		first, serr := acquire("org-1", 1)
		Expect(serr).To(BeNil())
		second, serr := acquire("org-1", 1)
		Expect(serr).To(BeNil())
		Expect(second.ID).To(Equal(first.ID))
		Expect(second.ExpiresAt).To(BeTemporally("==", first.ExpiresAt))
	})

	It("returns a stable 503 without inserting past capacity", func() {
		_, serr := acquire("org-1", 1)
		Expect(serr).To(BeNil())

		allocation, serr := acquire("org-2", 1)
		Expect(allocation).To(BeNil())
		Expect(serr.HttpCode).To(Equal(http.StatusServiceUnavailable))
		Expect(serr.Reason).To(Equal(errors.ErrorCodeCapacityReached))
		Expect(serr.Details).To(Equal(errors.CodeErrorDetails{Code: errors.ErrorCodeCapacityReached}))
		_, getErr := store.GetByOrganisationID(ctx, "org-2")
		Expect(getErr.Code).To(Equal(errors.ErrorNotFound))
	})

	It("rolls the allocation back when later stack persistence fails", func() {
		serr := executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
			allocation, acquireErr := store.AcquireWithTx(txCtx, "org-1", now, now.Add(6*time.Hour), 1)
			Expect(acquireErr).To(BeNil())
			Expect(allocation).ToNot(BeNil())
			return errors.GeneralError("stack insert failed")
		})
		Expect(serr).ToNot(BeNil())
		_, getErr := store.GetByOrganisationID(ctx, "org-1")
		Expect(getErr.Code).To(Equal(errors.ErrorNotFound))
	})

	It("does not reactivate an expired allocation", func() {
		existing := &models.TrialAllocation{
			ID:             "allocation-1",
			OrganisationID: "org-1",
			State:          models.TrialAllocationStateActive,
			ActivatedAt:    now.Add(-7 * time.Hour),
			ExpiresAt:      now,
		}
		Expect(sf.New(ctx).Create(existing).Error).NotTo(HaveOccurred())

		allocation, serr := acquire("org-1", 1)
		Expect(allocation).To(BeNil())
		Expect(serr.Reason).To(Equal(errors.ErrorCodeTrialInactive))
		persisted, getErr := store.GetByOrganisationID(ctx, "org-1")
		Expect(getErr).To(BeNil())
		Expect(persisted.ExpiresAt).To(BeTemporally("==", now))
	})

	It("revalidates an active allocation without changing its expiry", func() {
		existing := &models.TrialAllocation{
			ID: "allocation-1", OrganisationID: "org-1", State: models.TrialAllocationStateActive,
			ActivatedAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour),
		}
		Expect(sf.New(ctx).Create(existing).Error).NotTo(HaveOccurred())

		var got *models.TrialAllocation
		serr := executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
			var revalidateErr *errors.ServiceError
			got, revalidateErr = store.RevalidateWithTx(txCtx, "org-1", now)
			return revalidateErr
		})
		Expect(serr).To(BeNil())
		Expect(got.ID).To(Equal(existing.ID))
		Expect(got.ExpiresAt).To(BeTemporally("==", existing.ExpiresAt))
	})

	It("allows draft mutations before an allocation exists", func() {
		var got *models.TrialAllocation
		serr := executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
			var revalidateErr *errors.ServiceError
			got, revalidateErr = store.RevalidateIfExistsWithTx(txCtx, "org-1", now)
			return revalidateErr
		})
		Expect(serr).To(BeNil())
		Expect(got).To(BeNil())
	})

	It("rejects mutations after an allocation becomes inactive", func() {
		existing := &models.TrialAllocation{
			ID: "allocation-1", OrganisationID: "org-1", State: models.TrialAllocationStateCleanupPending,
			ActivatedAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour),
		}
		Expect(sf.New(ctx).Create(existing).Error).NotTo(HaveOccurred())

		serr := executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
			_, revalidateErr := store.RevalidateIfExistsWithTx(txCtx, "org-1", now)
			return revalidateErr
		})
		Expect(serr.Reason).To(Equal(errors.ErrorCodeTrialInactive))
	})

	DescribeTable("rejects rollback revalidation without an active unexpired allocation",
		func(existing *models.TrialAllocation) {
			if existing != nil {
				Expect(sf.New(ctx).Create(existing).Error).NotTo(HaveOccurred())
			}
			serr := executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
				_, revalidateErr := store.RevalidateWithTx(txCtx, "org-1", now)
				return revalidateErr
			})
			Expect(serr.Reason).To(Equal(errors.ErrorCodeTrialInactive))
		},
		Entry("missing", nil),
		Entry("expired", &models.TrialAllocation{ID: "allocation-1", OrganisationID: "org-1", State: models.TrialAllocationStateActive, ActivatedAt: now.Add(-2 * time.Hour), ExpiresAt: now}),
		Entry("cleanup pending", &models.TrialAllocation{ID: "allocation-1", OrganisationID: "org-1", State: models.TrialAllocationStateCleanupPending, ActivatedAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour)}),
	)

	It("reads only active unexpired allocations for worker admission", func() {
		existing := &models.TrialAllocation{
			ID: "allocation-1", OrganisationID: "org-1", State: models.TrialAllocationStateActive,
			ActivatedAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour),
		}
		Expect(sf.New(ctx).Create(existing).Error).NotTo(HaveOccurred())
		got, serr := store.GetActiveByOrganisationID(ctx, "org-1", now)
		Expect(serr).To(BeNil())
		Expect(got.ID).To(Equal(existing.ID))

		_, serr = store.GetActiveByOrganisationID(ctx, "org-1", existing.ExpiresAt)
		Expect(serr.Reason).To(Equal(errors.ErrorCodeTrialInactive))
	})

	DescribeTable("counts every non-cleaned lifecycle state as capacity",
		func(state models.TrialAllocationState, wantCapacityError bool) {
			existing := &models.TrialAllocation{
				ID:             "allocation-1",
				OrganisationID: "org-1",
				State:          state,
				ActivatedAt:    now.Add(-time.Hour),
				ExpiresAt:      now.Add(time.Hour),
			}
			Expect(sf.New(ctx).Create(existing).Error).NotTo(HaveOccurred())

			allocation, serr := acquire("org-2", 1)
			if wantCapacityError {
				Expect(allocation).To(BeNil())
				Expect(serr.Reason).To(Equal(errors.ErrorCodeCapacityReached))
				return
			}
			Expect(serr).To(BeNil())
			Expect(allocation).ToNot(BeNil())
		},
		Entry("active", models.TrialAllocationStateActive, true),
		Entry("cleanup pending", models.TrialAllocationStateCleanupPending, true),
		Entry("cleaning", models.TrialAllocationStateCleaning, true),
		Entry("error", models.TrialAllocationStateError, true),
		Entry("unknown future state", models.TrialAllocationState("future_state"), true),
		Entry("cleaned", models.TrialAllocationStateCleaned, false),
	)
})
