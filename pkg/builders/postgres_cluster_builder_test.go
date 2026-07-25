package builders_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/models"
)

func TestPostgresClusterBuilder(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Postgres Cluster Builder Suite")
}

func addonForBuild() *models.PostgresAddon {
	return &models.PostgresAddon{
		Name:            "test-pg",
		Namespace:       "ns-1",
		ID:              "addon-1",
		PostgresVersion: models.PostgresVersion{Major: 15},
		Instances:       models.PostgresInstances{Count: 1},
		Storage:         models.PostgresStorage{Size: "10Gi"},
	}
}

var _ = Describe("BuildPostgresClusterCR storage class", func() {
	var b = builders.NewPostgresClusterBuilder()

	It("leaves StorageClassName empty when the addon has no storage class (use cluster default)", func() {
		addon := addonForBuild()
		addon.Storage.StorageClass = ""

		cr, err := b.BuildPostgresClusterCR(addon, builders.PostgresClusterBuildContext{})

		Expect(err).ToNot(HaveOccurred())
		Expect(cr.Spec.StorageSpec.StorageClassName).To(BeEmpty())
	})

	It("honors an explicit storage class", func() {
		addon := addonForBuild()
		addon.Storage.StorageClass = "local-path"

		cr, err := b.BuildPostgresClusterCR(addon, builders.PostgresClusterBuildContext{})

		Expect(err).ToNot(HaveOccurred())
		Expect(cr.Spec.StorageSpec.StorageClassName).To(Equal("local-path"))
	})
})
