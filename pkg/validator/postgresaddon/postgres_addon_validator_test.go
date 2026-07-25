package postgresaddon_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/validator/postgresaddon"
)

func TestPostgresAddonValidator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Postgres Addon Validator Suite")
}

func validAddon() *models.PostgresAddon {
	return &models.PostgresAddon{
		Name:            "test-pg",
		OrganisationID:  "org-1",
		UserID:          "user-1",
		ProjectID:       "project-1",
		PostgresVersion: models.PostgresVersion{Major: 15},
		Instances:       models.PostgresInstances{Count: 1},
		Storage:         models.PostgresStorage{Size: "10Gi"},
	}
}

var _ = Describe("PostgresAddonValidator storage class", func() {
	var v = postgresaddon.NewPostgresAddonValidator()

	It("accepts an empty storage class (use cluster default)", func() {
		addon := validAddon()
		addon.Storage.StorageClass = ""

		Expect(v.ValidateForCreate(context.Background(), addon)).To(BeNil())
	})

	It("accepts an explicit storage class", func() {
		addon := validAddon()
		addon.Storage.StorageClass = "local-path"

		Expect(v.ValidateForCreate(context.Background(), addon)).To(BeNil())
	})

	It("still rejects an empty storage size", func() {
		addon := validAddon()
		addon.Storage.Size = ""

		Expect(v.ValidateForCreate(context.Background(), addon)).ToNot(BeNil())
	})
})
