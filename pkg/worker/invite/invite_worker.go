package invite

import (
	"context"
	"time"

	emailpkg "github.com/Stackdome/stackdome/pkg/email"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/worker"
	"github.com/openshift-online/ocm-sdk-go/leadership"
)

const InviteWorkerName = "invite-email-worker"

type inviteWorker struct {
	inviteService  services.OrgInviteService
	emailService   emailpkg.EmailService
	leadershipFlag *leadership.Flag
	worker.BaseWorker
}

type InviteWorkerSpec struct {
	InviteService  services.OrgInviteService
	EmailService   emailpkg.EmailService
	LeadershipFlag *leadership.Flag
	Env            string
}

func NewInviteWorker(spec InviteWorkerSpec) worker.Worker {
	return &inviteWorker{
		inviteService:  spec.InviteService,
		emailService:   spec.EmailService,
		leadershipFlag: spec.LeadershipFlag,
		BaseWorker:     worker.NewBaseWorker(InviteWorkerName, spec.Env),
	}
}

func (w *inviteWorker) Interval() time.Duration {
	return 30 * time.Second
}

func (w *inviteWorker) GetInput(ctx context.Context) ([]worker.Operand, *errors.ServiceError) {
	if w.leadershipFlag != nil && !w.leadershipFlag.Raised() {
		w.Logger().Infof("Not the leader, invite email worker will not send emails")
		return nil, nil
	}

	params := stores.ListParams{Page: 1, PageSize: stores.MaxPageSize}
	result, serr := w.inviteService.InternalGetPendingUnsent(ctx, params)
	if serr != nil {
		return nil, w.WorkerError.NewError("failed to list pending unsent invites: %v", serr)
	}

	operands := make([]worker.Operand, len(result.Items))
	for i, invite := range result.Items {
		operands[i] = invite
	}
	return operands, nil
}

func (w *inviteWorker) Execute(ctx context.Context, operand worker.Operand) (worker.Result, *errors.ServiceError) {
	id, ok := operand.(*models.OrgInvite)
	if !ok {
		return worker.Result{}, w.WorkerError.NewError("invalid operand type, expected *models.OrgInvite")
	}

	invite, serr := w.inviteService.InternalGetByID(ctx, id.ID)
	if serr != nil {
		if serr.Is404() {
			w.Logger().Infof("Invite %s not found, skipping", id.ID)
			return worker.Result{}, nil
		}
		return worker.Result{}, serr
	}

	if invite.EmailSent || invite.Status != models.InviteStatusPending {
		return worker.Result{}, nil
	}

	rawToken, decErr := w.inviteService.InternalDecryptToken(ctx, invite.EncryptedToken)
	if decErr != nil {
		w.Logger().Errorf("failed to decrypt token for invite %s: %s", invite.ID, decErr.Error())
		if markErr := w.inviteService.InternalMarkEmailError(ctx, invite.ID, "failed to decrypt token"); markErr != nil {
			w.Logger().Errorf("failed to mark email error for invite %s: %s", invite.ID, markErr.Error())
		}
		return worker.Result{}, nil
	}

	orgName := ""
	if invite.Organisation != nil {
		orgName = invite.Organisation.Name
	}
	projectName := ""
	if invite.Project != nil {
		projectName = invite.Project.Name
	}
	inviterName := ""
	if invite.InvitedBy != nil {
		inviterName = invite.InvitedBy.Name
	}

	w.Logger().Infof("Sending invite email to %s for org %s", invite.Email, orgName)

	txErr := w.inviteService.InternalWithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		if markErr := w.inviteService.InternalMarkEmailSent(txCtx, invite.ID); markErr != nil {
			return markErr
		}
		if emailErr := w.emailService.SendInviteEmail(txCtx, emailpkg.InviteEmailParams{
			ToEmail:     invite.Email,
			OrgName:     orgName,
			ProjectName:    projectName,
			InviterName: inviterName,
			InviteToken: rawToken,
			ExpiresAt:   invite.ExpiresAt,
		}); emailErr != nil {
			return errors.GeneralError("failed to send invite email: %s", emailErr.Error())
		}
		return nil
	})
	if txErr != nil {
		w.Logger().Errorf("failed to send invite email to %s: %s", invite.Email, txErr.Error())
		if markErr := w.inviteService.InternalMarkEmailError(ctx, invite.ID, txErr.Error()); markErr != nil {
			w.Logger().Errorf("failed to mark email error for invite %s: %s", invite.ID, markErr.Error())
		}
		return worker.Result{RequeueAfter: 5 * time.Minute}, nil
	}

	w.Logger().Infof("Invite email sent to %s", invite.Email)
	return worker.Result{}, nil
}
