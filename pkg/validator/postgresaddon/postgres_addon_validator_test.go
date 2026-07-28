package postgresaddon

import (
	"context"
	"strings"

	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func validPostgresAddonSpec(name string) *models.PostgresAddon {
	return &models.PostgresAddon{
		Name:            name,
		OrganisationID:  "org-1",
		UserID:          "user-1",
		ProjectID:       "project-1",
		PostgresVersion: models.PostgresVersion{Major: 16},
		Instances:       models.PostgresInstances{Count: 1},
		Storage:         models.PostgresStorage{Size: "10Gi", StorageClass: "standard"},
	}
}

var _ = Describe("ValidateForCreate name length", func() {
	It("accepts a name of MaxAddonNameLength characters", func() {
		spec := validPostgresAddonSpec(strings.Repeat("a", models.MaxAddonNameLength))

		err := NewPostgresAddonValidator().ValidateForCreate(context.Background(), spec)

		Expect(err).To(BeNil())
	})

	It("rejects a name one character over MaxAddonNameLength", func() {
		spec := validPostgresAddonSpec(strings.Repeat("a", models.MaxAddonNameLength+1))

		err := NewPostgresAddonValidator().ValidateForCreate(context.Background(), spec)

		Expect(err).NotTo(BeNil())
	})
})

var _ = Describe("ValidateForCreate storage class", func() {
	It("accepts an empty storage class (use cluster default)", func() {
		spec := validPostgresAddonSpec("test-pg")
		spec.Storage.StorageClass = ""

		Expect(NewPostgresAddonValidator().ValidateForCreate(context.Background(), spec)).To(BeNil())
	})

	It("accepts an explicit storage class", func() {
		spec := validPostgresAddonSpec("test-pg")
		spec.Storage.StorageClass = "local-path"

		Expect(NewPostgresAddonValidator().ValidateForCreate(context.Background(), spec)).To(BeNil())
	})

	It("still rejects an empty storage size", func() {
		spec := validPostgresAddonSpec("test-pg")
		spec.Storage.Size = ""

		Expect(NewPostgresAddonValidator().ValidateForCreate(context.Background(), spec)).ToNot(BeNil())
	})
})
