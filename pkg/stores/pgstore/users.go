package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type dbUserStore struct {
	sessionFactory db.SessionFactory
}

type UserStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewUserStore(spec UserStoreSpec) stores.UserStore {
	return &dbUserStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (d dbUserStore) Create(ctx context.Context, user *models.User) (*models.User, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&user).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create user: %s", err.Error())
	}
	return d.GetByEmail(ctx, user.Email)
}

func (d dbUserStore) GetByID(ctx context.Context, id string) (*models.User, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var user models.User
	err := grm.Model(&models.User{}).
		Preload(clause.Associations).Where("id = ?", id).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("user with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch user: %s", err.Error())
	}
	return &user, nil
}

func (d dbUserStore) GetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var user models.User
	err := grm.Model(&models.User{}).
		Preload(clause.Associations).Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("user with email '%s' not found", email)
		}
		return nil, errors.GeneralError("failed to fetch user: %s", err.Error())
	}
	return &user, nil
}

func (d dbUserStore) Update(ctx context.Context, id string, user *models.User) (*models.User, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	if err := grm.Model(&models.User{}).Where("id = ?", id).Updates(user).Error; err != nil {
		return nil, errors.GeneralError("failed to update user: %s", err.Error())
	}
	return d.GetByID(ctx, id)
}

func (d dbUserStore) ListByOrgID(ctx context.Context, orgID string, params stores.PaginationParams) (*stores.PaginatedResult[*models.User], *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)

	var total int64
	if err := grm.Model(&models.User{}).Where("organisation_id = ?", orgID).Count(&total).Error; err != nil {
		return nil, errors.GeneralError("failed to count users: %s", err.Error())
	}

	var users []*models.User
	if err := grm.Preload(clause.Associations).
		Where("organisation_id = ?", orgID).
		Order("created_at ASC").
		Limit(params.Limit()).
		Offset(params.Offset()).
		Find(&users).Error; err != nil {
		return nil, errors.GeneralError("failed to list users by org: %s", err.Error())
	}

	totalPages := int(total) / params.Limit()
	if int(total)%params.Limit() > 0 {
		totalPages++
	}

	return &stores.PaginatedResult[*models.User]{
		Items:      users,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.Limit(),
		TotalPages: totalPages,
	}, nil
}

func (d dbUserStore) ListByOrgAndRole(ctx context.Context, orgID string, role models.UserRole) ([]*models.User, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var users []*models.User
	if err := grm.Where("organisation_id = ? AND role = ?", orgID, string(role)).Find(&users).Error; err != nil {
		return nil, errors.GeneralError("failed to list users by org and role: %s", err.Error())
	}
	return users, nil
}
