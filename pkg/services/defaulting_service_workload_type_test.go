package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/models"
)

var _ = Describe("workload type defaulting", func() {
	// Stackfile-created stacks (PR previews, compose import) omit workload_type,
	// which the OpenAPI enum has no value for. Persisting "" holds the resource
	// out of the editor's draft, so the stack reads as "resource removed".
	It("defaults an empty workload type to Service", func() {
		stack := &models.Stack{
			StackResources: []*models.StackResource{
				{Name: "app"},
				{Name: "mysql"},
			},
		}

		result, err := NewStackDefaultingService().PopulateDefaultValues(stack)

		Expect(err).ToNot(HaveOccurred())
		Expect(result.StackResources[0].WorkloadType).To(Equal(models.WorkloadTypeService))
		Expect(result.StackResources[1].WorkloadType).To(Equal(models.WorkloadTypeService))
	})

	It("leaves an explicit workload type untouched", func() {
		stack := &models.Stack{
			StackResources: []*models.StackResource{
				{Name: "cron", WorkloadType: models.WorkloadTypeCronJob},
				{Name: "queue", WorkloadType: models.WorkloadTypeWorker},
			},
		}

		result, err := NewStackDefaultingService().PopulateDefaultValues(stack)

		Expect(err).ToNot(HaveOccurred())
		Expect(result.StackResources[0].WorkloadType).To(Equal(models.WorkloadTypeCronJob))
		Expect(result.StackResources[1].WorkloadType).To(Equal(models.WorkloadTypeWorker))
	})
})
