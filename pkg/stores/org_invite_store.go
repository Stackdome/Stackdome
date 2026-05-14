package stores

import (
	"context"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

//go:generate mockgen -source=org_invite_store.go -destination=../mocks/mock_org_invite_store.go -package=mocks
type OrgInviteStore interface {
	AtomicExecutor
	Create(ctx context.Context, invite *models.OrgInvite) (*models.OrgInvite, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.OrgInvite, *errors.ServiceError)
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.OrgInvite, *errors.ServiceError)
	ListByOrgID(ctx context.Context, orgID string, params ListParams) (*PaginatedResult[*models.OrgInvite], *errors.ServiceError)
	UpdateStatus(ctx context.Context, id string, status models.InviteStatus) *errors.ServiceError
	MarkAccepted(ctx context.Context, id string) *errors.ServiceError
	MarkEmailSent(ctx context.Context, id string) *errors.ServiceError
	MarkEmailError(ctx context.Context, id string, errMsg string) *errors.ServiceError
	ResetEmailStatus(ctx context.Context, id string) *errors.ServiceError
	ListPendingUnsent(ctx context.Context, params ListParams) (*PaginatedResult[*models.OrgInvite], *errors.ServiceError)
	GetPendingByOrgAndEmail(ctx context.Context, orgID, email string) (*models.OrgInvite, *errors.ServiceError)
	DeleteIDs(ctx context.Context, ids []string) *errors.ServiceError
	ListExpiredOrPastDue(ctx context.Context, now time.Time, params ListParams) (*PaginatedResult[*models.OrgInvite], *errors.ServiceError)
}
