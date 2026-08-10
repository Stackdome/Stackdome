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

var _ = Describe("ComputeAccessStore", func() {
	var (
		ctx      context.Context
		sf       *sqliteSessionFactory
		store    stores.ComputeAccessStore
		executor stores.AtomicExecutor
		now      time.Time
	)

	BeforeEach(func() {
		ctx = context.Background()
		now = time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
		sf = newSQLiteSessionFactory(`
			CREATE TABLE compute_entitlements (
				id TEXT PRIMARY KEY,
				organisation_id TEXT NOT NULL,
				source TEXT NOT NULL,
				status TEXT NOT NULL,
				starts_at DATETIME NOT NULL,
				expires_at DATETIME,
				created_at DATETIME,
				updated_at DATETIME,
				UNIQUE (organisation_id, source)
			);
			CREATE TABLE shared_compute_leases (
				id TEXT PRIMARY KEY,
				organisation_id TEXT NOT NULL,
				entitlement_id TEXT NOT NULL UNIQUE,
				state TEXT NOT NULL,
				activated_at DATETIME NOT NULL,
				cleanup_started_at DATETIME,
				cleaned_up_at DATETIME,
				error_at DATETIME,
				error_message TEXT,
				created_at DATETIME,
				updated_at DATETIME
			)
		`)
		store = pgstore.NewComputeAccessStore(pgstore.ComputeAccessStoreSpec{
			SessionFactory: sf, MaxActiveSharedComputeLeases: 1,
		})
		executor = pgstore.NewAtomicExecutor(sf)
	})

	activate := func(orgID string) (*models.ComputeAccess, *errors.ServiceError) {
		var access *models.ComputeAccess
		expiresAt := now.Add(6 * time.Hour)
		serr := executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
			var activationErr *errors.ServiceError
			access, activationErr = store.ActivateWithTx(txCtx, stores.ComputeAccessActivation{
				OrganisationID: orgID, EntitlementSource: models.ComputeEntitlementSourceTrial,
				StartsAt: now, ExpiresAt: &expiresAt,
			})
			return activationErr
		})
		return access, serr
	}

	persistAccess := func(orgID string, entitlementStatus models.ComputeEntitlementStatus, leaseState models.SharedComputeLeaseState, expiresAt *time.Time) *models.ComputeAccess {
		entitlement := &models.ComputeEntitlement{
			ID: "entitlement-" + orgID, OrganisationID: orgID,
			Source: models.ComputeEntitlementSourceTrial, Status: entitlementStatus,
			StartsAt: now.Add(-time.Hour), ExpiresAt: expiresAt,
		}
		lease := &models.SharedComputeLease{
			ID: "lease-" + orgID, OrganisationID: orgID, EntitlementID: entitlement.ID,
			State: leaseState, ActivatedAt: now.Add(-time.Hour),
		}
		Expect(sf.New(ctx).Create(entitlement).Error).NotTo(HaveOccurred())
		Expect(sf.New(ctx).Create(lease).Error).NotTo(HaveOccurred())
		return &models.ComputeAccess{Entitlement: entitlement, Lease: lease}
	}

	It("requires the caller's outer transaction", func() {
		expiresAt := now.Add(time.Hour)
		access, serr := store.ActivateWithTx(ctx, stores.ComputeAccessActivation{
			OrganisationID: "org-1", EntitlementSource: models.ComputeEntitlementSourceTrial,
			StartsAt: now, ExpiresAt: &expiresAt,
		})
		Expect(access).To(BeNil())
		Expect(serr.Error()).To(ContainSubstring("transaction"))
	})

	It("returns the same entitlement and lease on retry", func() {
		first, serr := activate("org-1")
		Expect(serr).To(BeNil())
		second, serr := activate("org-1")
		Expect(serr).To(BeNil())
		Expect(second.Entitlement.ID).To(Equal(first.Entitlement.ID))
		Expect(second.Lease.ID).To(Equal(first.Lease.ID))
		Expect(second.Lease.EntitlementID).To(Equal(second.Entitlement.ID))
	})

	It("returns a stable 503 without inserting past shared capacity", func() {
		_, serr := activate("org-1")
		Expect(serr).To(BeNil())

		access, serr := activate("org-2")
		Expect(access).To(BeNil())
		Expect(serr.HttpCode).To(Equal(http.StatusServiceUnavailable))
		Expect(serr.Reason).To(Equal(errors.ErrorCodeCapacityReached))
		Expect(serr.Details).To(Equal(errors.CodeErrorDetails{Code: errors.ErrorCodeCapacityReached}))

		var entitlements, leases int64
		Expect(sf.New(ctx).Model(&models.ComputeEntitlement{}).Where("organisation_id = ?", "org-2").Count(&entitlements).Error).NotTo(HaveOccurred())
		Expect(sf.New(ctx).Model(&models.SharedComputeLease{}).Where("organisation_id = ?", "org-2").Count(&leases).Error).NotTo(HaveOccurred())
		Expect(entitlements).To(BeZero())
		Expect(leases).To(BeZero())
	})

	It("rolls both records back when later release persistence fails", func() {
		expiresAt := now.Add(6 * time.Hour)
		serr := executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
			access, activationErr := store.ActivateWithTx(txCtx, stores.ComputeAccessActivation{
				OrganisationID: "org-1", EntitlementSource: models.ComputeEntitlementSourceTrial,
				StartsAt: now, ExpiresAt: &expiresAt,
			})
			Expect(activationErr).To(BeNil())
			Expect(access).ToNot(BeNil())
			return errors.GeneralError("release insert failed")
		})
		Expect(serr).ToNot(BeNil())

		var entitlements, leases int64
		Expect(sf.New(ctx).Model(&models.ComputeEntitlement{}).Count(&entitlements).Error).NotTo(HaveOccurred())
		Expect(sf.New(ctx).Model(&models.SharedComputeLease{}).Count(&leases).Error).NotTo(HaveOccurred())
		Expect(entitlements).To(BeZero())
		Expect(leases).To(BeZero())
	})

	It("does not reactivate expired compute access", func() {
		expiresAt := now
		persisted := persistAccess("org-1", models.ComputeEntitlementStatusActive, models.SharedComputeLeaseStateActive, &expiresAt)

		access, serr := activate("org-1")
		Expect(access).To(BeNil())
		Expect(serr.Reason).To(Equal(errors.ErrorCodeComputeAccessInactive))
		var entitlement models.ComputeEntitlement
		Expect(sf.New(ctx).First(&entitlement, "id = ?", persisted.Entitlement.ID).Error).NotTo(HaveOccurred())
		Expect(entitlement.ExpiresAt).NotTo(BeNil())
		Expect(*entitlement.ExpiresAt).To(BeTemporally("==", now))
	})

	It("admits database-only drafts before a lease exists", func() {
		var access *models.ComputeAccess
		serr := executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
			var admissionErr *errors.ServiceError
			access, admissionErr = store.AdmitComputeMutationWithTx(txCtx, "org-1", now)
			return admissionErr
		})
		Expect(serr).To(BeNil())
		Expect(access).To(BeNil())
	})

	DescribeTable("requires an active entitlement and lease",
		func(expiresAt *time.Time, leaseState models.SharedComputeLeaseState) {
			if expiresAt != nil || leaseState != "" {
				persistAccess("org-1", models.ComputeEntitlementStatusActive, leaseState, expiresAt)
			}
			serr := executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
				_, accessErr := store.RequireWithTx(txCtx, "org-1", now)
				return accessErr
			})
			Expect(serr.Reason).To(Equal(errors.ErrorCodeComputeAccessInactive))
		},
		Entry("missing", nil, models.SharedComputeLeaseState("")),
		Entry("expired", func() *time.Time { value := now; return &value }(), models.SharedComputeLeaseStateActive),
		Entry("cleanup pending", func() *time.Time { value := now.Add(time.Hour); return &value }(), models.SharedComputeLeaseStateCleanupPending),
	)

	DescribeTable("counts every non-cleaned lease state as shared capacity",
		func(state models.SharedComputeLeaseState, wantCapacityError bool) {
			expiresAt := now.Add(time.Hour)
			persistAccess("org-1", models.ComputeEntitlementStatusActive, state, &expiresAt)

			access, serr := activate("org-2")
			if wantCapacityError {
				Expect(access).To(BeNil())
				Expect(serr.Reason).To(Equal(errors.ErrorCodeCapacityReached))
				return
			}
			Expect(serr).To(BeNil())
			Expect(access).ToNot(BeNil())
		},
		Entry("active", models.SharedComputeLeaseStateActive, true),
		Entry("cleanup pending", models.SharedComputeLeaseStateCleanupPending, true),
		Entry("cleaning", models.SharedComputeLeaseStateCleaning, true),
		Entry("error", models.SharedComputeLeaseStateError, true),
		Entry("cleaned", models.SharedComputeLeaseStateCleaned, false),
	)
})
