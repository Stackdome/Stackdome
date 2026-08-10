package pgstore_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type postgresTestSessionFactory struct {
	database *gorm.DB
}

func (*postgresTestSessionFactory) Init(*config.DatabaseConfig) {}

func (f *postgresTestSessionFactory) DirectDB() *sql.DB {
	database, _ := f.database.DB()
	return database
}

func (f *postgresTestSessionFactory) New(ctx context.Context) *gorm.DB {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	return f.database.Session(&gorm.Session{})
}

func (*postgresTestSessionFactory) CheckConnection() error { return nil }

func (f *postgresTestSessionFactory) Close() error {
	return f.DirectDB().Close()
}

var _ = Describe("ComputeUsageStore PostgreSQL locking", func() {
	It("serializes concurrent stack admissions for one organisation", func() {
		dsn := os.Getenv("STACKDOME_TEST_POSTGRES_DSN")
		if dsn == "" {
			Skip("STACKDOME_TEST_POSTGRES_DSN is not set")
		}

		admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		adminSQL, err := admin.DB()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(adminSQL.Close)
		schemaName := "stack_limit_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		Expect(admin.Exec("CREATE SCHEMA " + schemaName).Error).NotTo(HaveOccurred())
		DeferCleanup(func() {
			Expect(admin.Exec("DROP SCHEMA " + schemaName + " CASCADE").Error).NotTo(HaveOccurred())
		})

		schemaDB, err := gorm.Open(postgres.Open(dsn+" search_path="+schemaName), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())
		Expect(schemaDB.Exec(`
			CREATE TABLE organisations (id TEXT PRIMARY KEY);
			CREATE TABLE stacks (id TEXT PRIMARY KEY, organisation_id TEXT NOT NULL);
			CREATE TABLE stack_resources (id TEXT PRIMARY KEY, stack_id TEXT NOT NULL);
			CREATE TABLE volumes (id TEXT PRIMARY KEY, organisation_id TEXT NOT NULL);
			CREATE TABLE postgres_addons (id TEXT PRIMARY KEY, organisation_id TEXT NOT NULL);
			INSERT INTO organisations (id) VALUES ('org-1');
		`).Error).NotTo(HaveOccurred())

		sessionFactory := &postgresTestSessionFactory{database: schemaDB}
		DeferCleanup(sessionFactory.Close)
		executor := pgstore.NewAtomicExecutor(sessionFactory)
		store := pgstore.NewComputeUsageStore()
		ctx := context.Background()
		start := make(chan struct{})
		results := make(chan *errors.ServiceError, 2)
		var workers sync.WaitGroup

		for index := 0; index < 2; index++ {
			workers.Add(1)
			go func(stackID string) {
				defer workers.Done()
				<-start
				results <- executor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
					usage, serr := store.LockOrganisationAndGetUsageWithTx(txCtx, "org-1", "")
					if serr != nil {
						return serr
					}
					if usage.StackCount >= 1 {
						return errors.BadRequest("stack capacity reached")
					}
					if err := db.TxFromContext(txCtx).Exec(
						"INSERT INTO stacks (id, organisation_id) VALUES (?, ?)", stackID, "org-1",
					).Error; err != nil {
						return errors.GeneralError("insert stack: %s", err.Error())
					}
					return nil
				})
			}(fmt.Sprintf("stack-%d", index))
		}
		close(start)
		workers.Wait()
		close(results)

		var admitted, rejected int
		for serr := range results {
			if serr == nil {
				admitted++
				continue
			}
			if serr.Reason == "stack capacity reached" {
				rejected++
				continue
			}
			Fail(fmt.Sprintf("unexpected stack limit error: %v", serr))
		}
		Expect(admitted).To(Equal(1))
		Expect(rejected).To(Equal(1))

		var persisted int64
		Expect(schemaDB.Table("stacks").Count(&persisted).Error).NotTo(HaveOccurred())
		Expect(persisted).To(Equal(int64(1)))
	})
})
