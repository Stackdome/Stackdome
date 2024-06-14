package workspaceprovisionrequests

import (
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	workerlib "github.com/ashishmax31/stackdome-api-server/pkg/worker"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type subReconcilerResult struct {
	resultNil          bool
	resultStop         bool
	resultRequeue      bool
	resultRequeueAfter *time.Duration
}

var (
	resultNil          = subReconcilerResult{resultNil: true}
	resultStop         = subReconcilerResult{resultStop: true}
	resultRequeue      = subReconcilerResult{resultRequeue: true}
	resultRequeueAfter = func(t time.Duration) subReconcilerResult {
		return subReconcilerResult{resultRequeueAfter: &t}
	}
)

const (
	WorkspaceProvisionRequestReconcileWorker = "WorkspaceProvisionRequestReconcileWorker"
)

var _ workerlib.Worker = (*workspaceProvisionRequestReconcileWorker)(nil)

type WorkspaceProvisionRequestReconcileWorkerSpec struct {
	ClusterClient       client.Client
	Env                 string
	WPRService          services.WorkspaceProvisionRequestService
	UserService         services.UserService
	ClusterService      services.ClusterService
	OrganisationService services.OrganisationService
}

type workspaceProvisionRequestReconcileWorker struct {
	clusterClient       client.Client
	userService         services.UserService
	wprService          services.WorkspaceProvisionRequestService
	clusterService      services.ClusterService
	organisationService services.OrganisationService
	workerlib.BaseWorker
}

func NewWorkspaceProvisionRequestWorker(spec WorkspaceProvisionRequestReconcileWorkerSpec) *workspaceProvisionRequestReconcileWorker {
	a := &workspaceProvisionRequestReconcileWorker{
		BaseWorker:          workerlib.NewBaseWorker(WorkspaceProvisionRequestReconcileWorker, spec.Env),
		clusterClient:       spec.ClusterClient,
		userService:         spec.UserService,
		wprService:          spec.WPRService,
		clusterService:      spec.ClusterService,
		organisationService: spec.OrganisationService,
	}
	return a
}
