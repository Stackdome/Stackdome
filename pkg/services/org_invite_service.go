package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/mail"
	"strings"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

const (
	inviteTokenByteLen = 16
	minExpiryDays      = 1
	maxExpiryDays      = 30
)

type OrgInviteService interface {
	BackgroundJobEnqueuerInjectable
	Create(ctx context.Context, email, teamName string, role models.TeamRole, expiresInDays int) (*models.OrgInvite, string, *errors.ServiceError)
	List(ctx context.Context, orgID string, params stores.ListParams) (*stores.PaginatedResult[*models.OrgInvite], *errors.ServiceError)
	GetByID(ctx context.Context, orgID, id string) (*models.OrgInvite, *errors.ServiceError)
	Revoke(ctx context.Context, orgID, id string) *errors.ServiceError
	Resend(ctx context.Context, orgID, id string) *errors.ServiceError
	PublicGetInviteInfo(ctx context.Context, rawToken string) (*models.OrgInvite, *errors.ServiceError)
	ValidateAndConsume(ctx context.Context, rawToken, email string) (*models.OrgInvite, *errors.ServiceError)
	InternalGetByID(ctx context.Context, id string) (*models.OrgInvite, *errors.ServiceError)
	InternalGetPendingUnsent(ctx context.Context, params stores.ListParams) (*stores.PaginatedResult[*models.OrgInvite], *errors.ServiceError)
	InternalMarkEmailSent(ctx context.Context, id string) *errors.ServiceError
	InternalMarkEmailError(ctx context.Context, id string, errMsg string) *errors.ServiceError
	InternalDecryptToken(ctx context.Context, encryptedToken string) (string, *errors.ServiceError)
	InternalListExpiredOrPastDue(ctx context.Context, now time.Time, params stores.ListParams) (*stores.PaginatedResult[*models.OrgInvite], *errors.ServiceError)
	InternalMarkExpiredAndDelete(ctx context.Context, invites []*models.OrgInvite) *errors.ServiceError
}

type OrgInviteServiceSpec struct {
	SessionFactory    db.SessionFactory
	TeamService       TeamService
	UserService       UserService
	EncryptionService EncryptionService
	Permissions       auth.PermissionService
	Logger            logger.Logger
}

type orgInviteService struct {
	inviteStore       stores.OrgInviteStore
	teamService       TeamService
	userService       UserService
	encryptionService EncryptionService
	permissions       auth.PermissionService
	logger            logger.Logger
	BackgroundJobEnqueuerDep
}

func NewOrgInviteService(spec OrgInviteServiceSpec) OrgInviteService {
	return &orgInviteService{
		inviteStore: pgstore.NewOrgInviteStore(pgstore.OrgInviteStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		teamService:       spec.TeamService,
		userService:       spec.UserService,
		encryptionService: spec.EncryptionService,
		permissions:       spec.Permissions,
		logger:            spec.Logger,
	}
}

func (s *orgInviteService) Create(ctx context.Context, email, teamName string, role models.TeamRole, expiresInDays int) (*models.OrgInvite, string, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return nil, "", errors.Unauthorized("not authenticated")
	}

	if permErr := s.permissions.Check(ctx, identity.OrgID, auth.ResourceInvites, "", auth.ActionCreate); permErr != nil {
		return nil, "", permErr
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, "", errors.BadRequest("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, "", errors.BadRequest("invalid email address")
	}

	if role != models.DeveloperRole && role != models.ViewerRole {
		return nil, "", errors.BadRequest("role must be Developer or Viewer")
	}

	if expiresInDays < minExpiryDays || expiresInDays > maxExpiryDays {
		return nil, "", errors.BadRequest("expires_in_days must be between %d and %d", minExpiryDays, maxExpiryDays)
	}

	existingUser, _ := s.userService.InternalGetByEmail(ctx, email)
	if existingUser != nil {
		return nil, "", errors.Conflict("user with this email already exists")
	}

	existingInvite, checkErr := s.inviteStore.GetPendingByOrgAndEmail(ctx, identity.OrgID, email)
	if checkErr != nil {
		return nil, "", checkErr
	}
	if existingInvite != nil {
		return nil, "", errors.Conflict("pending invite already exists for this email")
	}

	team, serr := s.teamService.GetTeamByOrgAndName(ctx, identity.OrgID, teamName)
	if serr != nil {
		return nil, "", errors.BadRequest("team '%s' not found in organization", teamName)
	}

	rawBytes := make([]byte, inviteTokenByteLen)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, "", errors.GeneralError("failed to generate invite token")
	}
	rawToken := hex.EncodeToString(rawBytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	// We encrypt the raw token and store it in the DB because background workers might need to
	// access the raw token to send emails.
	encryptedToken, encErr := s.encryptionService.EncryptData([]byte(rawToken))
	if encErr != nil {
		return nil, "", errors.GeneralError("failed to encrypt invite token")
	}

	invite := &models.OrgInvite{
		Email:          email,
		OrganisationID: identity.OrgID,
		TeamID:         team.ID,
		TeamRole:       role,
		TokenHash:      tokenHash,
		EncryptedToken: encryptedToken,
		Status:         models.InviteStatusPending,
		ExpiresAt:      time.Now().UTC().Add(time.Duration(expiresInDays) * 24 * time.Hour),
		InvitedByID:    identity.UserID,
	}

	created, serr := s.inviteStore.Create(ctx, invite)
	if serr != nil {
		return nil, "", serr
	}

	if s.BackgroundJobEnqueuer != nil {
		if err := s.BackgroundJobEnqueuer.Enqueue(&models.OrgInvite{ID: created.ID}); err != nil {
			s.logger.Errorf("failed to enqueue invite email job: %s", err.Error())
		}
	}

	return created, rawToken, nil
}

func (s *orgInviteService) List(ctx context.Context, orgID string, params stores.ListParams) (*stores.PaginatedResult[*models.OrgInvite], *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceInvites, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.inviteStore.ListByOrgID(ctx, orgID, params)
}

func (s *orgInviteService) GetByID(ctx context.Context, orgID, id string) (*models.OrgInvite, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceInvites, id, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	invite, serr := s.inviteStore.GetByID(ctx, id)
	if serr != nil {
		return nil, serr
	}
	if invite.OrganisationID != orgID {
		return nil, errors.NotFound("invite not found")
	}
	return invite, nil
}

func (s *orgInviteService) Revoke(ctx context.Context, orgID, id string) *errors.ServiceError {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceInvites, id, auth.ActionDelete); permErr != nil {
		return permErr
	}
	invite, serr := s.inviteStore.GetByID(ctx, id)
	if serr != nil {
		return serr
	}
	if invite.OrganisationID != orgID {
		return errors.NotFound("invite not found")
	}
	if invite.Status != models.InviteStatusPending {
		return errors.BadRequest("can only revoke pending invites")
	}
	return s.inviteStore.UpdateStatus(ctx, id, models.InviteStatusRevoked)
}

func (s *orgInviteService) Resend(ctx context.Context, orgID, id string) *errors.ServiceError {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceInvites, id, auth.ActionWrite); permErr != nil {
		return permErr
	}
	invite, serr := s.inviteStore.GetByID(ctx, id)
	if serr != nil {
		return serr
	}
	if invite.OrganisationID != orgID {
		return errors.NotFound("invite not found")
	}
	if invite.Status != models.InviteStatusPending {
		return errors.BadRequest("can only resend pending invites")
	}
	if invite.ExpiresAt.Before(time.Now().UTC()) {
		return errors.BadRequest("invite has expired")
	}
	if serr := s.inviteStore.ResetEmailStatus(ctx, id); serr != nil {
		return serr
	}

	if s.BackgroundJobEnqueuer != nil {
		if err := s.BackgroundJobEnqueuer.Enqueue(&models.OrgInvite{ID: id}); err != nil {
			s.logger.Errorf("failed to enqueue invite email job for resend: %s", err.Error())
		}
	}
	return nil
}

func (s *orgInviteService) PublicGetInviteInfo(ctx context.Context, rawToken string) (*models.OrgInvite, *errors.ServiceError) {
	invite, serr := s.findByToken(ctx, rawToken)
	if serr != nil {
		return nil, serr
	}
	if invite.Status != models.InviteStatusPending {
		return nil, errors.Gone("invite is no longer valid")
	}
	if invite.ExpiresAt.Before(time.Now().UTC()) {
		return nil, errors.Gone("invite has expired")
	}
	return invite, nil
}

func (s *orgInviteService) ValidateAndConsume(ctx context.Context, rawToken, email string) (*models.OrgInvite, *errors.ServiceError) {
	invite, serr := s.findByToken(ctx, rawToken)
	if serr != nil {
		return nil, serr
	}

	if invite.Status == models.InviteStatusAccepted {
		return nil, errors.Gone("invite has already been accepted")
	}
	if invite.Status == models.InviteStatusRevoked {
		return nil, errors.Gone("invite has been revoked")
	}
	if invite.Status != models.InviteStatusPending {
		return nil, errors.Gone("invite is no longer valid")
	}
	if invite.ExpiresAt.Before(time.Now().UTC()) {
		return nil, errors.Gone("invite has expired")
	}

	if !strings.EqualFold(invite.Email, email) {
		return nil, errors.BadRequest("email does not match invite")
	}

	if serr := s.inviteStore.MarkAccepted(ctx, invite.ID); serr != nil {
		return nil, serr
	}

	return invite, nil
}

func (s *orgInviteService) InternalGetByID(ctx context.Context, id string) (*models.OrgInvite, *errors.ServiceError) {
	return s.inviteStore.GetByID(ctx, id)
}

func (s *orgInviteService) InternalGetPendingUnsent(ctx context.Context, params stores.ListParams) (*stores.PaginatedResult[*models.OrgInvite], *errors.ServiceError) {
	return s.inviteStore.ListPendingUnsent(ctx, params)
}

func (s *orgInviteService) InternalMarkEmailSent(ctx context.Context, id string) *errors.ServiceError {
	return s.inviteStore.MarkEmailSent(ctx, id)
}

func (s *orgInviteService) InternalMarkEmailError(ctx context.Context, id string, errMsg string) *errors.ServiceError {
	return s.inviteStore.MarkEmailError(ctx, id, errMsg)
}

func (s *orgInviteService) InternalDecryptToken(ctx context.Context, encryptedToken string) (string, *errors.ServiceError) {
	rawBytes, err := s.encryptionService.DecryptData(encryptedToken)
	if err != nil {
		return "", errors.GeneralError("failed to decrypt invite token")
	}
	return string(rawBytes), nil
}

func (s *orgInviteService) InternalListExpiredOrPastDue(ctx context.Context, now time.Time, params stores.ListParams) (*stores.PaginatedResult[*models.OrgInvite], *errors.ServiceError) {
	return s.inviteStore.ListExpiredOrPastDue(ctx, now, params)
}

func (s *orgInviteService) InternalMarkExpiredAndDelete(ctx context.Context, invites []*models.OrgInvite) *errors.ServiceError {
	ids := make([]string, len(invites))
	for i, invite := range invites {
		ids[i] = invite.ID
	}
	return s.inviteStore.DeleteIDs(ctx, ids)
}

func (s *orgInviteService) findByToken(ctx context.Context, rawToken string) (*models.OrgInvite, *errors.ServiceError) {
	if rawToken == "" {
		return nil, errors.NotFound("invite not found")
	}

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	return s.inviteStore.GetByTokenHash(ctx, tokenHash)
}
