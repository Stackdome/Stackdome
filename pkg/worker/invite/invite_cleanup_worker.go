package invite

import (
	"context"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker"
	"github.com/openshift-online/ocm-sdk-go/leadership"
)

const InviteCleanupWorkerName = "invite-cleanup-worker"

type InviteCleanupBatch struct {
	Items []*models.OrgInvite
}

type inviteCleanupWorker struct {
	inviteService  services.OrgInviteService
	leadershipFlag *leadership.Flag
	worker.BaseWorker
}

type InviteCleanupWorkerSpec struct {
	InviteService  services.OrgInviteService
	LeadershipFlag *leadership.Flag
	Env            string
}

func NewInviteCleanupWorker(spec InviteCleanupWorkerSpec) worker.Worker {
	return &inviteCleanupWorker{
		inviteService:  spec.InviteService,
		leadershipFlag: spec.LeadershipFlag,
		BaseWorker:     worker.NewBaseWorker(InviteCleanupWorkerName, spec.Env),
	}
}

func (w *inviteCleanupWorker) Interval() time.Duration {
	return 30 * time.Minute
}

func (w *inviteCleanupWorker) GetInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	if w.leadershipFlag != nil && !w.leadershipFlag.Raised() {
		w.Logger().Infof("Not the leader, invite cleanup worker will not run")
		return nil, nil
	}

	params := stores.ListParams{Page: 1, PageSize: stores.MaxPageSize}
	result, serr := w.inviteService.InternalListExpiredOrPastDue(ctx, time.Now().UTC(), params)
	if serr != nil {
		return nil, w.WorkerError.NewError("failed to list expired invites: %v", serr)
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	return []worker.Operand{&InviteCleanupBatch{Items: result.Items}}, nil
}

func (w *inviteCleanupWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	batch, ok := operand.(*InviteCleanupBatch)
	if !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected *InviteCleanupBatch")
	}
	invites := batch.Items

	w.Logger().Infof("Cleaning up %d expired invites", len(invites))

	if serr := w.inviteService.InternalMarkExpiredAndDelete(ctx, invites); serr != nil {
		w.Logger().Errorf("failed to clean up expired invites: %s", serr.Error())
		return worker.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	w.Logger().Infof("Cleaned up %d expired invites", len(invites))
	return worker.Result{}, nil
}
