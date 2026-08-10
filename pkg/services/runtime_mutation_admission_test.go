package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("cloud runtime mutation admission", func() {
	It("rejects a stack shell update before persistence", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockStackStore(ctrl)
		svc := &stackService{stackStore: store, runtimePolicy: policy}
		policy.EXPECT().AdmitMutationWithTx(gomock.Any(), "org-1").Return(errors.TrialInactive())

		updated, serr := svc.InternalUpdateShellWithTx(context.Background(), &models.Stack{}, &models.Stack{ID: "stack-1", OrganisationID: "org-1"})

		Expect(updated).To(BeNil())
		Expect(serr.Reason).To(Equal(errors.ErrorCodeTrialInactive))
	})

	It("rejects connection creation inside its write transaction", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockStackStore(ctrl)
		svc := &stackService{stackStore: store, runtimePolicy: policy}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		policy.EXPECT().AdmitMutationWithTx(gomock.Any(), "org-1").Return(errors.TrialInactive())

		created, serr := svc.createStackConnection(context.Background(), &models.Stack{ID: "stack-1", OrganisationID: "org-1"}, &models.StackConnection{})

		Expect(created).To(BeNil())
		Expect(serr.Reason).To(Equal(errors.ErrorCodeTrialInactive))
	})

	It("rejects resource restart before its direct Kubernetes path", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		stacks := mocks.NewMockStackStore(ctrl)
		resources := mocks.NewMockStackResourceStore(ctrl)
		permissions := mocks.NewMockPermissionService(ctrl)
		svc := &stackResourceService{stackStore: stacks, stackResourceStore: resources, permissions: permissions, runtimePolicy: policy}
		stack := &models.Stack{ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1"}
		stacks.EXPECT().GetByID(gomock.Any(), "stack-1").Return(stack, nil)
		permissions.EXPECT().Check(gomock.Any(), "project-1", auth.ResourceStacks, "stack-1", auth.ActionWrite).Return(nil)
		resources.EXPECT().GetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(&models.StackResource{ID: "resource-1"}, nil)
		stacks.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		policy.EXPECT().AdmitMutationWithTx(gomock.Any(), "org-1").Return(errors.TrialInactive())

		updated, serr := svc.Restart(context.Background(), "stack-1", "web")

		Expect(updated).To(BeNil())
		Expect(serr.Reason).To(Equal(errors.ErrorCodeTrialInactive))
	})

	It("rejects a volume source revision before database or cluster mutation", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockVolumeStore(ctrl)
		svc := &volumeService{volumeStore: store, runtimePolicy: policy}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		store.EXPECT().GetByID(gomock.Any(), "volume-1").Return(&models.Volume{ID: "volume-1", OrganisationID: "org-1"}, nil)
		policy.EXPECT().AdmitMutationWithTx(gomock.Any(), "org-1").Return(errors.TrialInactive())

		updated, serr := svc.UpdateGitRepoSourceRevision(context.Background(), "volume-1", models.GitRepoRevision{Branch: "main"})

		Expect(updated).To(BeNil())
		Expect(serr.Reason).To(Equal(errors.ErrorCodeTrialInactive))
	})

	DescribeTable("rejects PostgreSQL lifecycle mutation before persistence and enqueue",
		func(invoke func(*postgresAddonService) *errors.ServiceError) {
			ctrl := gomock.NewController(GinkgoT())
			policy := NewMockRuntimePolicy(ctrl)
			store := mocks.NewMockPostgresAddonStore(ctrl)
			permissions := mocks.NewMockPermissionService(ctrl)
			svc := &postgresAddonService{postgresAddonStore: store, permissions: permissions, runtimePolicy: policy}
			store.EXPECT().GetByID(gomock.Any(), "addon-1").Return(&models.PostgresAddon{ID: "addon-1", OrganisationID: "org-1", ProjectID: "project-1"}, nil)
			permissions.EXPECT().Check(gomock.Any(), "project-1", auth.ResourceAddonsPostgres, "addon-1", auth.ActionWrite).Return(nil)
			store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
					return fn(ctx)
				},
			)
			policy.EXPECT().AdmitMutationWithTx(gomock.Any(), "org-1").Return(errors.TrialInactive())

			serr := invoke(svc)

			Expect(serr.Reason).To(Equal(errors.ErrorCodeTrialInactive))
		},
		Entry("backup", func(svc *postgresAddonService) *errors.ServiceError {
			return svc.TriggerBackup(context.Background(), "addon-1")
		}),
		Entry("hibernate", func(svc *postgresAddonService) *errors.ServiceError {
			return svc.TriggerHibernate(context.Background(), "addon-1", true)
		}),
		Entry("fence", func(svc *postgresAddonService) *errors.ServiceError {
			return svc.TriggerFence(context.Background(), "addon-1", true)
		}),
	)
})
