package presenters_test

import (
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster presenter", func() {
	DescribeTable("keeps the deprecated platform alias consistent with shared_compute",
		func(sharedCompute bool) {
			presented := presenters.PresentCluster(&models.Cluster{SharedCompute: sharedCompute})

			Expect(presented.SharedCompute).NotTo(BeNil())
			Expect(presented.Platform).NotTo(BeNil())
			Expect(*presented.SharedCompute).To(Equal(sharedCompute))
			Expect(*presented.Platform).To(Equal(sharedCompute))
		},
		Entry("shared compute", true),
		Entry("tenant owned", false),
	)
})
