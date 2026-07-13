package pgstore

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
)

type orgInviteStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

type OrgInviteStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewOrgInviteStore(spec OrgInviteStoreSpec) stores.OrgInviteStore {
	return &orgInviteStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (s *orgInviteStore) Create(ctx context.Context, invite *models.OrgInvite) (*models.OrgInvite, *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	if err := session.Create(invite).Error; err != nil {
		return nil, errors.GeneralError("failed to create invite: %s", err.Error())
	}
	return s.GetByID(ctx, invite.ID)
}

func (s *orgInviteStore) GetByID(ctx context.Context, id string) (*models.OrgInvite, *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	var invite models.OrgInvite
	if err := session.Preload("Organisation").Preload("Project").Preload("InvitedBy").Where("id = ?", id).First(&invite).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("invite not found")
		}
		return nil, errors.GeneralError("failed to get invite: %s", err.Error())
	}
	return &invite, nil
}

func (s *orgInviteStore) GetByTokenHash(ctx context.Context, tokenHash string) (*models.OrgInvite, *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	var invite models.OrgInvite
	if err := session.Preload("Organisation").Preload("Project").Preload("InvitedBy").Where("token_hash = ?", tokenHash).First(&invite).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("invite not found")
		}
		return nil, errors.GeneralError("failed to get invite by token: %s", err.Error())
	}
	return &invite, nil
}

func (s *orgInviteStore) ListByOrgID(ctx context.Context, orgID string, params stores.ListParams) (*stores.PaginatedResult[*models.OrgInvite], *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	params = params.WithDefaultOrder("created_at DESC")

	var total int64
	baseQuery := session.Model(&models.OrgInvite{}).Where("organisation_id = ?", orgID)
	if err := params.ApplyFiltersOnly(baseQuery).Count(&total).Error; err != nil {
		return nil, errors.GeneralError("failed to count invites: %s", err.Error())
	}

	var invites []*models.OrgInvite
	query := session.Preload("Organisation").Preload("Project").Preload("InvitedBy").Where("organisation_id = ?", orgID)
	if err := params.Apply(query).Find(&invites).Error; err != nil {
		return nil, errors.GeneralError("failed to list invites: %s", err.Error())
	}

	return &stores.PaginatedResult[*models.OrgInvite]{
		Items:      invites,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.Limit(),
		TotalPages: params.TotalPages(total),
	}, nil
}

func (s *orgInviteStore) UpdateStatus(ctx context.Context, id string, status models.InviteStatus) *errors.ServiceError {
	session := s.sessionFactory.New(ctx)
	result := session.Model(&models.OrgInvite{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return errors.GeneralError("failed to update invite status: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("invite not found")
	}
	return nil
}

func (s *orgInviteStore) MarkAccepted(ctx context.Context, id string) *errors.ServiceError {
	session := s.sessionFactory.New(ctx)
	now := time.Now().UTC()
	result := session.Model(&models.OrgInvite{}).Where("id = ? AND status = ?", id, models.InviteStatusPending).Updates(map[string]any{
		"status":      models.InviteStatusAccepted,
		"accepted_at": now,
	})
	if result.Error != nil {
		return errors.GeneralError("failed to mark invite accepted: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.Conflict("invite has already been accepted or is no longer pending")
	}
	return nil
}

func (s *orgInviteStore) MarkEmailSent(ctx context.Context, id string) *errors.ServiceError {
	session := s.sessionFactory.New(ctx)
	now := time.Now().UTC()
	result := session.Model(&models.OrgInvite{}).Where("id = ?", id).Updates(map[string]any{
		"email_sent":    true,
		"email_sent_at": now,
		"email_error":   nil,
	})
	if result.Error != nil {
		return errors.GeneralError("failed to mark email sent: %s", result.Error.Error())
	}
	return nil
}

func (s *orgInviteStore) MarkEmailError(ctx context.Context, id string, errMsg string) *errors.ServiceError {
	session := s.sessionFactory.New(ctx)
	result := session.Model(&models.OrgInvite{}).Where("id = ?", id).Update("email_error", errMsg)
	if result.Error != nil {
		return errors.GeneralError("failed to mark email error: %s", result.Error.Error())
	}
	return nil
}

func (s *orgInviteStore) ResetEmailStatus(ctx context.Context, id string) *errors.ServiceError {
	session := s.sessionFactory.New(ctx)
	result := session.Model(&models.OrgInvite{}).Where("id = ?", id).Updates(map[string]any{
		"email_sent":  false,
		"email_error": nil,
	})
	if result.Error != nil {
		return errors.GeneralError("failed to reset email status: %s", result.Error.Error())
	}
	return nil
}

func (s *orgInviteStore) ListPendingUnsent(ctx context.Context, params stores.ListParams) (*stores.PaginatedResult[*models.OrgInvite], *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	now := time.Now().UTC()
	params = params.WithDefaultOrder("created_at ASC")

	baseCondition := "status = ? AND email_sent = false AND expires_at > ?"

	var total int64
	baseQuery := session.Model(&models.OrgInvite{}).Where(baseCondition, models.InviteStatusPending, now)
	if err := params.ApplyFiltersOnly(baseQuery).Count(&total).Error; err != nil {
		return nil, errors.GeneralError("failed to count pending unsent invites: %s", err.Error())
	}

	var invites []*models.OrgInvite
	query := session.Preload("Organisation").Preload("Project").Preload("InvitedBy").
		Where(baseCondition, models.InviteStatusPending, now)
	if err := params.Apply(query).Find(&invites).Error; err != nil {
		return nil, errors.GeneralError("failed to list pending unsent invites: %s", err.Error())
	}

	return &stores.PaginatedResult[*models.OrgInvite]{
		Items:      invites,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.Limit(),
		TotalPages: params.TotalPages(total),
	}, nil
}

func (s *orgInviteStore) GetPendingByOrgAndEmail(ctx context.Context, orgID, email string) (*models.OrgInvite, *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	var invite models.OrgInvite
	if err := session.Preload("Project").Preload("InvitedBy").
		Where("organisation_id = ? AND email = ? AND status = ?", orgID, email, models.InviteStatusPending).
		First(&invite).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.GeneralError("failed to check pending invite: %s", err.Error())
	}
	return &invite, nil
}

func (s *orgInviteStore) DeleteIDs(ctx context.Context, ids []string) *errors.ServiceError {
	if len(ids) == 0 {
		return nil
	}
	session := s.sessionFactory.New(ctx)
	if err := session.Where("id IN ?", ids).Delete(&models.OrgInvite{}).Error; err != nil {
		return errors.GeneralError("failed to delete invites: %s", err.Error())
	}
	return nil
}

func (s *orgInviteStore) ListExpiredOrPastDue(ctx context.Context, now time.Time, params stores.ListParams) (*stores.PaginatedResult[*models.OrgInvite], *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	params = params.WithDefaultOrder("expires_at ASC")

	condition := "(status = ? OR (status = ? AND expires_at < ?))"
	args := []any{models.InviteStatusExpired, models.InviteStatusPending, now}

	var total int64
	if err := session.Model(&models.OrgInvite{}).Where(condition, args...).Count(&total).Error; err != nil {
		return nil, errors.GeneralError("failed to count expired invites: %s", err.Error())
	}

	var invites []*models.OrgInvite
	query := session.Where(condition, args...)
	if err := params.Apply(query).Find(&invites).Error; err != nil {
		return nil, errors.GeneralError("failed to list expired invites: %s", err.Error())
	}

	return &stores.PaginatedResult[*models.OrgInvite]{
		Items:      invites,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.Limit(),
		TotalPages: params.TotalPages(total),
	}, nil
}
