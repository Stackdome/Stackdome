package computequota

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/computeaccess"
	stackerrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

type fakeComputeAccessService struct {
	organisationID string
	err            *stackerrors.ServiceError
}

func (f *fakeComputeAccessService) Activate(_ context.Context, organisationID string) (*computeaccess.ComputeAccess, *stackerrors.ServiceError) {
	f.organisationID = organisationID
	return &computeaccess.ComputeAccess{}, f.err
}

type fakeComputeUsageStore struct {
	usage          ComputeUsage
	err            *stackerrors.ServiceError
	organisationID string
	excludeStackID string
}

func (f *fakeComputeUsageStore) LockOrganisationAndGetUsage(_ context.Context, organisationID, excludeStackID string) (ComputeUsage, *stackerrors.ServiceError) {
	f.organisationID = organisationID
	f.excludeStackID = excludeStackID
	return f.usage, f.err
}

var _ = ginkgo.Describe("Compute policy", func() {
	ginkgo.Describe("self-hosted", func() {
		ginkgo.It("does not impose managed-compute defaults or limits", func() {
			policy := NewSelfHostedPolicy()
			replicas := int32(3)
			resource := &models.StackResource{Replicas: &replicas}
			volume := &models.Volume{Size: "5Gi", StorageClass: "fast"}
			addon := &models.PostgresAddon{
				Instances: models.PostgresInstances{Count: 2},
				Storage:   models.PostgresStorage{Size: "10Gi"},
			}

			gomega.Expect(policy.EnsureAccess(context.Background(), "org-1")).NotTo(gomega.HaveOccurred())
			gomega.Expect(policy.ValidateStackLimits(context.Background(), StackLimitChange{Operation: "unknown"})).NotTo(gomega.HaveOccurred())
			gomega.Expect(policy.ValidateVolumeLimits(context.Background(), "org-1", "not-a-quantity")).NotTo(gomega.HaveOccurred())
			gomega.Expect(policy.ValidatePostgresAddonLimits(context.Background(), "org-1", "", addon)).NotTo(gomega.HaveOccurred())

			policy.ApplyStackResourceDefaults(resource)
			policy.ApplyVolumeDefaults(volume)
			policy.ApplyPostgresAddonDefaults(addon)

			gomega.Expect(*resource.Replicas).To(gomega.Equal(int32(3)))
			gomega.Expect(volume).To(gomega.Equal(&models.Volume{Size: "5Gi", StorageClass: "fast"}))
			gomega.Expect(addon.Instances.Count).To(gomega.Equal(2))
			gomega.Expect(addon.Storage.Size).To(gomega.Equal("10Gi"))
		})
	})

	ginkgo.Describe("Stackdome Cloud", func() {
		var (
			access *fakeComputeAccessService
			usage  *fakeComputeUsageStore
			policy Policy
		)

		ginkgo.BeforeEach(func() {
			access = &fakeComputeAccessService{}
			usage = &fakeComputeUsageStore{}
			policy = NewStackdomeCloudPolicy(StackdomeCloudPolicySpec{
				ComputeAccess: access,
				ComputeUsage:  usage,
				Limits: ComputeLimits{
					MaxStacksPerOrganization:         2,
					MaxStackResourcesPerOrganization: 6,
					ReplicasPerStackResource:         1,
					MaxVolumesPerOrganization:        2,
					MaxVolumeSize:                    "2Gi",
					VolumeStorageClass:               "longhorn",
					MaxPostgresAddonsPerOrganization: 1,
					PostgresInstances:                1,
					MaxPostgresStorageSize:           "2Gi",
				},
			})
		})

		ginkgo.It("propagates compute access denials", func() {
			access.err = stackerrors.ComputeAccessInactive()

			err := policy.EnsureAccess(context.Background(), "org-1")

			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Code).To(gomega.Equal(stackerrors.ErrorComputeAccessInactive))
			gomega.Expect(access.organisationID).To(gomega.Equal("org-1"))
		})

		ginkgo.It("applies managed-compute defaults", func() {
			resource := &models.StackResource{}
			volume := &models.Volume{}
			addon := &models.PostgresAddon{}

			policy.ApplyStackResourceDefaults(resource)
			policy.ApplyVolumeDefaults(volume)
			policy.ApplyPostgresAddonDefaults(addon)

			gomega.Expect(resource.Replicas).NotTo(gomega.BeNil())
			gomega.Expect(*resource.Replicas).To(gomega.Equal(int32(1)))
			gomega.Expect(volume.Size).To(gomega.Equal("2Gi"))
			gomega.Expect(volume.StorageClass).To(gomega.Equal("longhorn"))
			gomega.Expect(addon.Instances.Count).To(gomega.Equal(1))
			gomega.Expect(addon.Storage.Size).To(gomega.Equal("2Gi"))
		})

		ginkgo.It("replaces a stack against usage that excludes its old resources", func() {
			usage.usage = ComputeUsage{StackCount: 2, StackResourceCount: 2}

			err := policy.ValidateStackLimits(context.Background(), StackLimitChange{
				Operation:            StackLimitReplaceStack,
				OrganisationID:       "org-1",
				StackID:              "stack-1",
				DesiredResourceCount: 4,
			})

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(usage.organisationID).To(gomega.Equal("org-1"))
			gomega.Expect(usage.excludeStackID).To(gomega.Equal("stack-1"))
		})

		ginkgo.DescribeTable("rejects stack usage beyond a configured limit",
			func(current ComputeUsage, change StackLimitChange) {
				usage.usage = current

				err := policy.ValidateStackLimits(context.Background(), change)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Code).To(gomega.Equal(stackerrors.ErrorComputeQuotaExceeded))
			},
			ginkgo.Entry("stack count",
				ComputeUsage{StackCount: 2},
				StackLimitChange{Operation: StackLimitCreateStack, OrganisationID: "org-1"},
			),
			ginkgo.Entry("resource count",
				ComputeUsage{StackCount: 1, StackResourceCount: 6},
				StackLimitChange{Operation: StackLimitAddResource, OrganisationID: "org-1"},
			),
		)

		ginkgo.DescribeTable("validates volume requests",
			func(currentVolumeCount int64, size string, expectedCode stackerrors.ServiceErrorCode) {
				usage.usage = ComputeUsage{VolumeCount: currentVolumeCount}

				err := policy.ValidateVolumeLimits(context.Background(), "org-1", size)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Code).To(gomega.Equal(expectedCode))
			},
			ginkgo.Entry("volume count exceeds the limit", int64(2), "1Gi", stackerrors.ErrorComputeQuotaExceeded),
			ginkgo.Entry("volume size exceeds the limit", int64(0), "3Gi", stackerrors.ErrorComputeQuotaExceeded),
			ginkgo.Entry("volume size is malformed", int64(0), "invalid", stackerrors.ErrorBadRequest),
		)

		ginkgo.DescribeTable("validates PostgreSQL addon requests",
			func(currentAddonCount int64, existingAddonID string, addon *models.PostgresAddon, expectedCode stackerrors.ServiceErrorCode) {
				usage.usage = ComputeUsage{PostgresAddonCount: currentAddonCount}

				err := policy.ValidatePostgresAddonLimits(context.Background(), "org-1", existingAddonID, addon)

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Code).To(gomega.Equal(expectedCode))
			},
			ginkgo.Entry("addon count exceeds the limit", int64(1), "", postgresAddon(1, "1Gi"), stackerrors.ErrorComputeQuotaExceeded),
			ginkgo.Entry("instance count differs from the policy", int64(0), "", postgresAddon(2, "1Gi"), stackerrors.ErrorComputeQuotaExceeded),
			ginkgo.Entry("storage exceeds the limit", int64(0), "", postgresAddon(1, "3Gi"), stackerrors.ErrorComputeQuotaExceeded),
			ginkgo.Entry("storage size is malformed", int64(0), "", postgresAddon(1, "invalid"), stackerrors.ErrorBadRequest),
		)
	})
})

func postgresAddon(instances int, storageSize string) *models.PostgresAddon {
	return &models.PostgresAddon{
		Instances: models.PostgresInstances{Count: instances},
		Storage:   models.PostgresStorage{Size: storageSize},
	}
}
