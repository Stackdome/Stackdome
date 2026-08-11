package computequota

import (
	stackerrors "github.com/Stackdome/stackdome/pkg/errors"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Stack limit arithmetic", func() {
	ginkgo.Describe("stackResourceExclusionFor", func() {
		ginkgo.It("excludes the replaced stack from persisted resource usage", func() {
			excludedStackID, err := stackResourceExclusionFor(StackLimitChange{
				Operation: StackLimitReplaceStack,
				StackID:   "stack-1",
			})

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(excludedStackID).To(gomega.Equal("stack-1"))
		})

		ginkgo.DescribeTable("counts every persisted resource for non-replacement operations",
			func(operation StackLimitOperation) {
				excludedStackID, err := stackResourceExclusionFor(StackLimitChange{Operation: operation})

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(excludedStackID).To(gomega.BeEmpty())
			},
			ginkgo.Entry("stack creation", StackLimitCreateStack),
			ginkgo.Entry("resource creation", StackLimitAddResource),
			ginkgo.Entry("resource update", StackLimitUpdateResource),
		)

		ginkgo.It("rejects a replacement without a stack ID", func() {
			_, err := stackResourceExclusionFor(StackLimitChange{Operation: StackLimitReplaceStack})

			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Code).To(gomega.Equal(stackerrors.ErrorGeneral))
		})

		ginkgo.It("rejects an unsupported operation", func() {
			_, err := stackResourceExclusionFor(StackLimitChange{Operation: "unknown"})

			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Code).To(gomega.Equal(stackerrors.ErrorGeneral))
		})
	})

	ginkgo.Describe("stackUsageAfterChange", func() {
		current := ComputeUsage{
			StackCount:         2,
			StackResourceCount: 5,
			VolumeCount:        7,
			PostgresAddonCount: 11,
		}

		ginkgo.DescribeTable("projects the committed stack usage",
			func(change StackLimitChange, expected ComputeUsage) {
				proposed, err := stackUsageAfterChange(current, change)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(proposed).To(gomega.Equal(expected))
			},
			ginkgo.Entry("creates a stack and all of its resources",
				StackLimitChange{Operation: StackLimitCreateStack, DesiredResourceCount: 3},
				ComputeUsage{StackCount: 3, StackResourceCount: 8, VolumeCount: 7, PostgresAddonCount: 11},
			),
			ginkgo.Entry("replaces the excluded stack resources without changing the stack count",
				StackLimitChange{Operation: StackLimitReplaceStack, StackID: "stack-1", DesiredResourceCount: 4},
				ComputeUsage{StackCount: 2, StackResourceCount: 9, VolumeCount: 7, PostgresAddonCount: 11},
			),
			ginkgo.Entry("adds one resource",
				StackLimitChange{Operation: StackLimitAddResource},
				ComputeUsage{StackCount: 2, StackResourceCount: 6, VolumeCount: 7, PostgresAddonCount: 11},
			),
			ginkgo.Entry("updates a resource in place",
				StackLimitChange{Operation: StackLimitUpdateResource},
				ComputeUsage{StackCount: 2, StackResourceCount: 5, VolumeCount: 7, PostgresAddonCount: 11},
			),
		)

		ginkgo.DescribeTable("rejects a negative whole-stack resource count",
			func(operation StackLimitOperation) {
				_, err := stackUsageAfterChange(current, StackLimitChange{
					Operation:            operation,
					DesiredResourceCount: -1,
				})

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Code).To(gomega.Equal(stackerrors.ErrorGeneral))
			},
			ginkgo.Entry("stack creation", StackLimitCreateStack),
			ginkgo.Entry("stack replacement", StackLimitReplaceStack),
		)

		ginkgo.It("rejects an unsupported operation", func() {
			_, err := stackUsageAfterChange(current, StackLimitChange{Operation: "unknown"})

			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Code).To(gomega.Equal(stackerrors.ErrorGeneral))
		})
	})
})
