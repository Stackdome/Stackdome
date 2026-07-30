package builders

import (
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("postgresClusterBuilder.BuildPostgresClusterCR", func() {
	newAddon := func() *models.PostgresAddon {
		return &models.PostgresAddon{
			Name:            "my-db",
			Namespace:       "ns-1",
			PostgresVersion: models.PostgresVersion{Major: 16},
			Instances:       models.PostgresInstances{Count: 1},
			Storage:         models.PostgresStorage{Size: "10Gi", StorageClass: "local-path"},
		}
	}

	It("rejects an empty storage class", func() {
		builder := NewPostgresClusterBuilder()
		addon := newAddon()
		addon.Storage.StorageClass = ""

		cr, err := builder.BuildPostgresClusterCR(addon, PostgresClusterBuildContext{})
		Expect(err).To(HaveOccurred())
		Expect(cr).To(BeNil())
	})

	It("builds the CR when a storage class is set", func() {
		builder := NewPostgresClusterBuilder()
		addon := newAddon()

		cr, err := builder.BuildPostgresClusterCR(addon, PostgresClusterBuildContext{})
		Expect(err).ToNot(HaveOccurred())
		Expect(cr.Spec.StorageSpec.StorageClassName).To(Equal("local-path"))
	})
})
