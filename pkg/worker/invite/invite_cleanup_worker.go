package invite

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/worker"
	"github.com/openshift-online/ocm-sdk-go/leadership"
)

const InviteCleanupWorkerName = "invite-cleanup-worker"

// InviteCleanupRequest is a comparable singleton operand: successive ticks
// dedup to one queue entry, and Execute queries the expired set itself.
type InviteCleanupRequest struct{}

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
		w.Logger().Info(ctx, "Not the leader, invite cleanup worker will not run")
		return nil, nil
	}

	return []worker.Operand{InviteCleanupRequest{}}, nil
}

func (w *inviteCleanupWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	if _, ok := operand.(InviteCleanupRequest); !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected InviteCleanupRequest")
	}

	params := stores.ListParams{Page: 1, PageSize: stores.MaxPageSize}
	result, serr := w.inviteService.InternalListExpiredOrPastDue(ctx, time.Now().UTC(), params)
	if serr != nil {
		return worker.Result{}, w.WorkerError.NewError("failed to list expired invites: %v", serr)
	}
	invites := result.Items
	if len(invites) == 0 {
		return worker.Result{}, nil
	}

	w.Logger().Info(ctx, "Cleaning up %d expired invites", len(invites))

	if serr := w.inviteService.InternalMarkExpiredAndDelete(ctx, invites); serr != nil {
		w.Logger().Error(ctx, "failed to clean up expired invites: %s", serr.Error())
		return worker.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	w.Logger().Info(ctx, "Cleaned up %d expired invites", len(invites))
	return worker.Result{}, nil
}
