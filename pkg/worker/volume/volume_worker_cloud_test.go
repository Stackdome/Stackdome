package volume

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("VolumeWorker cloud admission", func() {
	It("does not reconcile a draft volume without an active allocation", func() {
		ctrl := gomock.NewController(GinkgoT())
		volumes := NewMockvolumeService(ctrl)
		stackVolumes := NewMockstackVolumeStore(ctrl)
		stacks := NewMockstackService(ctrl)
		volumes.EXPECT().InternalGet(gomock.Any(), "volume-1").Return(&models.Volume{ID: "volume-1"}, nil)
		stackVolumes.EXPECT().GetByVolumeID(gomock.Any(), "volume-1").Return(&models.StackVolume{StackID: "stack-1"}, nil)
		stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(&models.Stack{ID: "stack-1", OrganisationID: "org-1"}, nil)
		w := &volumeWorker{volumeService: volumes, stackVolumeStore: stackVolumes, stackService: stacks, runtimePolicy: &inactiveVolumeRuntimePolicy{}, BaseWorker: worker.NewBaseWorker(VolumeWorkerName, "test")}

		result, serr := w.Execute(context.Background(), models.VolumeOperand{ID: "volume-1"})
		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})
})

type inactiveVolumeRuntimePolicy struct{}

func (*inactiveVolumeRuntimePolicy) OrganisationProvisioningMode() services.ProvisioningMode {
	return services.ProvisioningModeDatabaseOnly
}
func (*inactiveVolumeRuntimePolicy) DraftProvisioningMode() services.ProvisioningMode {
	return services.ProvisioningModeDatabaseOnly
}
func (*inactiveVolumeRuntimePolicy) AdmitFirstReleaseWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveVolumeRuntimePolicy) AdmitRollbackWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveVolumeRuntimePolicy) RequireActiveAllocation(context.Context, string) *errors.ServiceError {
	return errors.TrialInactive()
}
