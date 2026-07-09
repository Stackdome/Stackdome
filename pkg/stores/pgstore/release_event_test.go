package pgstore_test

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseEventStore", func() {
	var (
		store stores.ReleaseEventStore
		sf    *sqliteSessionFactory
		ctx   context.Context
	)

	const (
		releaseID = "release-1"
		stackID   = "stack-1"
	)

	newEvent := func(dedupeKey string) *models.ReleaseEvent {
		return &models.ReleaseEvent{
			ReleaseID: releaseID,
			StackID:   stackID,
			Source:    models.ReleaseEventSourceHub,
			Scope:     models.ReleaseEventScopeRelease,
			Type:      models.ReleaseEventTypeReleaseStarted,
			Level:     models.ReleaseEventLevelInfo,
			Message:   "release started",
			DedupeKey: dedupeKey,
		}
	}

	BeforeEach(func() {
		sf = newSQLiteSessionFactory(
			`PRAGMA foreign_keys = ON`,
			`CREATE TABLE IF NOT EXISTS stack_releases (
				id TEXT PRIMARY KEY
			)`,
			`CREATE TABLE IF NOT EXISTS release_events (
				id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
				release_id TEXT NOT NULL REFERENCES stack_releases(id) ON DELETE CASCADE,
				stack_id TEXT NOT NULL,
				sequence INT NOT NULL,
				occurred_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				source TEXT NOT NULL,
				scope TEXT NOT NULL,
				resource_name TEXT,
				type TEXT NOT NULL,
				level TEXT NOT NULL,
				message TEXT NOT NULL,
				dedupe_key TEXT NOT NULL,
				links TEXT,
				metadata TEXT,
				created_at DATETIME,
				UNIQUE (release_id, sequence),
				UNIQUE (release_id, dedupe_key)
			)`,
		)
		store = pgstore.NewReleaseEventStore(pgstore.ReleaseEventStoreSpec{SessionFactory: sf})
		ctx = context.Background()

		Expect(sf.New(ctx).Exec(
			`INSERT INTO stack_releases (id) VALUES (?)`, releaseID,
		).Error).NotTo(HaveOccurred())
	})

	Describe("Insert", func() {
		It("assigns sequences 1, 2, 3 across three inserts on one release", func() {
			e1, err := store.Insert(ctx, newEvent("dedupe-1"))
			Expect(err).To(BeNil())
			Expect(e1.Sequence).To(Equal(1))

			e2, err := store.Insert(ctx, newEvent("dedupe-2"))
			Expect(err).To(BeNil())
			Expect(e2.Sequence).To(Equal(2))

			e3, err := store.Insert(ctx, newEvent("dedupe-3"))
			Expect(err).To(BeNil())
			Expect(e3.Sequence).To(Equal(3))
		})

		It("treats a duplicate dedupe key as a no-op, returning (nil, nil)", func() {
			created, err := store.Insert(ctx, newEvent("dedupe-dup"))
			Expect(err).To(BeNil())
			Expect(created).NotTo(BeNil())

			dup, err := store.Insert(ctx, newEvent("dedupe-dup"))
			Expect(err).To(BeNil())
			Expect(dup).To(BeNil())

			var count int64
			Expect(sf.New(ctx).Table("release_events").
				Where("release_id = ? AND dedupe_key = ?", releaseID, "dedupe-dup").
				Count(&count).Error).NotTo(HaveOccurred())
			Expect(count).To(Equal(int64(1)))
		})
	})

	Describe("ListByReleaseID", func() {
		It("returns events with sequence > afterSequence, ascending, capped at limit", func() {
			_, err := store.Insert(ctx, newEvent("dedupe-1"))
			Expect(err).To(BeNil())
			_, err = store.Insert(ctx, newEvent("dedupe-2"))
			Expect(err).To(BeNil())
			_, err = store.Insert(ctx, newEvent("dedupe-3"))
			Expect(err).To(BeNil())

			events, err := store.ListByReleaseID(ctx, releaseID, 1, 100)
			Expect(err).To(BeNil())
			Expect(events).To(HaveLen(2))
			Expect(events[0].Sequence).To(Equal(2))
			Expect(events[1].Sequence).To(Equal(3))

			limited, err := store.ListByReleaseID(ctx, releaseID, 0, 1)
			Expect(err).To(BeNil())
			Expect(limited).To(HaveLen(1))
			Expect(limited[0].Sequence).To(Equal(1))
		})
	})

	Describe("cascade delete", func() {
		It("removes release events when the parent stack_releases row is deleted", func() {
			_, err := store.Insert(ctx, newEvent("dedupe-1"))
			Expect(err).To(BeNil())

			Expect(sf.New(ctx).Exec(`DELETE FROM stack_releases WHERE id = ?`, releaseID).Error).NotTo(HaveOccurred())

			var count int64
			Expect(sf.New(ctx).Table("release_events").
				Where("release_id = ?", releaseID).
				Count(&count).Error).NotTo(HaveOccurred())
			Expect(count).To(Equal(int64(0)))
		})
	})

	Describe("InsertWithTx", func() {
		It("errors when called outside a transaction", func() {
			created, err := store.InsertWithTx(ctx, newEvent("dedupe-1"))
			Expect(created).To(BeNil())
			Expect(err).NotTo(BeNil())
			Expect(err.Reason).To(ContainSubstring("transaction not found in context"))
		})

		It("joins the caller's ambient transaction", func() {
			tx := sf.New(ctx).Begin()
			txCtx := db.CtxWithTransaction(ctx, tx)

			created, err := store.InsertWithTx(txCtx, newEvent("dedupe-1"))
			Expect(err).To(BeNil())
			Expect(created.Sequence).To(Equal(1))

			Expect(tx.Commit().Error).NotTo(HaveOccurred())

			events, err := store.ListByReleaseID(ctx, releaseID, 0, 10)
			Expect(err).To(BeNil())
			Expect(events).To(HaveLen(1))
		})
	})
})
