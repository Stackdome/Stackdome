package pgstore_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ComputeAccessStore PostgreSQL locking", func() {
	It("serializes concurrent outer stack transactions at active capacity", func() {
		if os.Getenv("STACKDOME_PG_COMPUTE_ACCESS_TEST") != "1" {
			Skip("set STACKDOME_PG_COMPUTE_ACCESS_TEST=1 with DB_* variables to run PostgreSQL locking coverage")
		}

		port, err := strconv.Atoi(os.Getenv("DB_PORT"))
		Expect(err).NotTo(HaveOccurred())
		sessionFactory := db.NewSessionFactory(&config.DatabaseConfig{
			Dialect:            "postgres",
			SSLMode:            config.DBSSLModeDisable,
			MaxOpenConnections: 30,
			DBConnectionConfig: config.DBConnectionConfig{
				Host: os.Getenv("DB_HOST"), Port: port, Name: os.Getenv("DB_NAME"),
				Username: os.Getenv("DB_USERNAME"), Password: os.Getenv("DB_PASSWORD"),
			},
		})
		DeferCleanup(sessionFactory.Close)

		ctx := context.Background()
		database := sessionFactory.New(ctx)
		Expect(database.Exec("DELETE FROM shared_compute_leases WHERE organisation_id LIKE 'pr2-lock-org-%'").Error).NotTo(HaveOccurred())
		Expect(database.Exec("DELETE FROM compute_entitlements WHERE organisation_id LIKE 'pr2-lock-org-%'").Error).NotTo(HaveOccurred())
		Expect(database.Exec("DELETE FROM organisations WHERE id LIKE 'pr2-lock-org-%'").Error).NotTo(HaveOccurred())
		DeferCleanup(func() {
			Expect(database.Exec("DELETE FROM shared_compute_leases WHERE organisation_id LIKE 'pr2-lock-org-%'").Error).NotTo(HaveOccurred())
			Expect(database.Exec("DELETE FROM compute_entitlements WHERE organisation_id LIKE 'pr2-lock-org-%'").Error).NotTo(HaveOccurred())
			Expect(database.Exec("DELETE FROM organisations WHERE id LIKE 'pr2-lock-org-%'").Error).NotTo(HaveOccurred())
		})

		const (
			capacity      = 10
			organisations = 24
		)
		for i := 0; i < organisations; i++ {
			id := fmt.Sprintf("pr2-lock-org-%02d", i)
			Expect(database.Exec("INSERT INTO organisations (id, name, platform) VALUES (?, ?, false)", id, id).Error).NotTo(HaveOccurred())
		}

		store := pgstore.NewComputeAccessStore(pgstore.ComputeAccessStoreSpec{
			SessionFactory: sessionFactory, MaxActiveSharedComputeLeases: capacity,
		})
		executor := pgstore.NewAtomicExecutor(sessionFactory)
		now := time.Now().UTC()
		start := make(chan struct{})
		results := make(chan *errors.ServiceError, organisations)
		var workers sync.WaitGroup
		for i := 0; i < organisations; i++ {
			workers.Add(1)
			go func(index int) {
				defer workers.Done()
				<-start
				results <- executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
					expiresAt := now.Add(6 * time.Hour)
					_, serr := store.ActivateWithTx(txCtx, stores.ComputeAccessActivation{
						OrganisationID:    fmt.Sprintf("pr2-lock-org-%02d", index),
						EntitlementSource: models.ComputeEntitlementSourceTrial,
						StartsAt:          now, ExpiresAt: &expiresAt,
					})
					return serr
				})
			}(i)
		}
		close(start)
		workers.Wait()
		close(results)

		var admitted, rejected int
		for serr := range results {
			switch {
			case serr == nil:
				admitted++
			case serr.Reason == errors.ErrorCodeCapacityReached:
				rejected++
			default:
				Fail(fmt.Sprintf("unexpected acquisition error: %v", serr))
			}
		}
		Expect(admitted).To(Equal(capacity))
		Expect(rejected).To(Equal(organisations - capacity))

		var activeOrganisationID string
		Expect(database.Raw(`SELECT organisation_id FROM shared_compute_leases WHERE state = 'active' LIMIT 1`).
			Scan(&activeOrganisationID).Error).NotTo(HaveOccurred())
		Expect(activeOrganisationID).NotTo(BeEmpty())
		Expect(database.Exec("DELETE FROM organisations WHERE id = ?", activeOrganisationID).Error).To(HaveOccurred(),
			"compute access ownership must block organisation deletion until cleanup removes its records")

		capacityLock := database.Begin()
		Expect(capacityLock.Error).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(capacityLock.Rollback().Error).NotTo(HaveOccurred()) })
		// Hold the production capacity lock to prove active retries use only their allocation row lock.
		Expect(capacityLock.Exec("SELECT pg_advisory_xact_lock(?)", int64(0x535441434b444f4d)).Error).NotTo(HaveOccurred())

		retryCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		DeferCleanup(cancel)
		serr := executor.WithTransaction(retryCtx, func(txCtx context.Context) *errors.ServiceError {
			expiresAt := now.Add(6 * time.Hour)
			_, acquireErr := store.ActivateWithTx(txCtx, stores.ComputeAccessActivation{
				OrganisationID: activeOrganisationID, EntitlementSource: models.ComputeEntitlementSourceTrial,
				StartsAt: now, ExpiresAt: &expiresAt,
			})
			return acquireErr
		})
		Expect(serr).To(BeNil(), "an admitted organisation must not wait for the global capacity lock")
	})
})
