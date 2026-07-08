package pgstore_test

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResourceValidationRecordStore", func() {
	var (
		store stores.ResourceValidationRecordStore
		ctx   context.Context
	)

	const (
		stackID = "stack-1"
	)

	BeforeEach(func() {
		sf := newSQLiteSessionFactory(`
			CREATE TABLE IF NOT EXISTS resource_validation_records (
				stack_id TEXT NOT NULL,
				resource_name TEXT NOT NULL,
				check_kind TEXT NOT NULL,
				fingerprint TEXT NOT NULL,
				validated_at DATETIME NOT NULL,
				PRIMARY KEY (stack_id, resource_name, check_kind)
			)
		`)
		store = pgstore.NewResourceValidationRecordStore(pgstore.ResourceValidationRecordStoreSpec{SessionFactory: sf})
		ctx = context.Background()
	})

	It("returns not found when no record exists", func() {
		_, err := store.Get(ctx, stackID, "web", models.ValidationCheckImagePull)
		Expect(err).NotTo(BeNil())
		Expect(err.Is404()).To(BeTrue())
	})

	It("upserts and gets a record", func() {
		validatedAt := time.Now().UTC().Truncate(time.Second)
		record := &models.ResourceValidationRecord{
			StackID:      stackID,
			ResourceName: "web",
			CheckKind:    models.ValidationCheckImagePull,
			Fingerprint:  "sha256:abc",
			ValidatedAt:  validatedAt,
		}
		Expect(store.Upsert(ctx, record)).To(BeNil())

		got, err := store.Get(ctx, stackID, "web", models.ValidationCheckImagePull)
		Expect(err).To(BeNil())
		Expect(got.Fingerprint).To(Equal("sha256:abc"))
		Expect(got.ValidatedAt.Equal(validatedAt)).To(BeTrue())
	})

	It("overwrites the fingerprint on repeated upserts for the same key", func() {
		first := &models.ResourceValidationRecord{
			StackID:      stackID,
			ResourceName: "web",
			CheckKind:    models.ValidationCheckImagePull,
			Fingerprint:  "sha256:old",
			ValidatedAt:  time.Now().UTC().Truncate(time.Second),
		}
		Expect(store.Upsert(ctx, first)).To(BeNil())

		second := &models.ResourceValidationRecord{
			StackID:      stackID,
			ResourceName: "web",
			CheckKind:    models.ValidationCheckImagePull,
			Fingerprint:  "sha256:new",
			ValidatedAt:  time.Now().UTC().Truncate(time.Second),
		}
		Expect(store.Upsert(ctx, second)).To(BeNil())

		got, err := store.Get(ctx, stackID, "web", models.ValidationCheckImagePull)
		Expect(err).To(BeNil())
		Expect(got.Fingerprint).To(Equal("sha256:new"))
	})

	It("distinguishes records by check kind", func() {
		imagePull := &models.ResourceValidationRecord{
			StackID:      stackID,
			ResourceName: "web",
			CheckKind:    models.ValidationCheckImagePull,
			Fingerprint:  "sha256:pull",
			ValidatedAt:  time.Now().UTC().Truncate(time.Second),
		}
		pushAccess := &models.ResourceValidationRecord{
			StackID:      stackID,
			ResourceName: "web",
			CheckKind:    models.ValidationCheckPushAccess,
			Fingerprint:  "sha256:push",
			ValidatedAt:  time.Now().UTC().Truncate(time.Second),
		}
		Expect(store.Upsert(ctx, imagePull)).To(BeNil())
		Expect(store.Upsert(ctx, pushAccess)).To(BeNil())

		got, err := store.Get(ctx, stackID, "web", models.ValidationCheckPushAccess)
		Expect(err).To(BeNil())
		Expect(got.Fingerprint).To(Equal("sha256:push"))
	})
})
