package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
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
	if user.DefaultUser && user.Role != models.PlatformAdminRole {
		return nil, errors.GeneralError("default user must have role %s", models.PlatformAdminRole)
	}
	err := grm.Create(&user).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create user: %s", err.Error())
	}
	return d.GetByEmail(ctx, user.Email)
}

func (d dbUserStore) GetByID(ctx context.Context, id string) (*models.User, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var user models.User
	err := grm.Model(&models.User{}).Where("id = ?", id).First(&user).Error
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
	err := grm.Model(&models.User{}).Where("email = ?", email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("user with email '%s' not found", email)
		}
		return nil, errors.GeneralError("failed to fetch user: %s", err.Error())
	}
	return &user, nil
}

func (d dbUserStore) GetDefaultUser(ctx context.Context) (*models.User, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var user models.User
	if err := grm.Model(&models.User{}).Where("default_user = ?", true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("default user not found")
		}
		return nil, errors.GeneralError("failed to fetch default user: %s", err.Error())
	}
	return &user, nil
}
